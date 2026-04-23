package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/adminuser"
	adminhandler "github.com/AeolianCloud/pveCloud/server/internal/adminuser/handler"
	"github.com/AeolianCloud/pveCloud/server/internal/audit"
	"github.com/AeolianCloud/pveCloud/server/internal/auth"
	"github.com/AeolianCloud/pveCloud/server/internal/billing"
	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap/config"
	"github.com/AeolianCloud/pveCloud/server/internal/catalog"
	cataloghandler "github.com/AeolianCloud/pveCloud/server/internal/catalog/handler"
	"github.com/AeolianCloud/pveCloud/server/internal/common/cache"
	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
	loggerx "github.com/AeolianCloud/pveCloud/server/internal/common/logger"
	"github.com/AeolianCloud/pveCloud/server/internal/instance"
	instancehandler "github.com/AeolianCloud/pveCloud/server/internal/instance/handler"
	"github.com/AeolianCloud/pveCloud/server/internal/notification"
	notificationhandler "github.com/AeolianCloud/pveCloud/server/internal/notification/handler"
	"github.com/AeolianCloud/pveCloud/server/internal/order"
	orderhandler "github.com/AeolianCloud/pveCloud/server/internal/order/handler"
	"github.com/AeolianCloud/pveCloud/server/internal/payment"
	paymenthandler "github.com/AeolianCloud/pveCloud/server/internal/payment/handler"
	"github.com/AeolianCloud/pveCloud/server/internal/resource"
	"github.com/AeolianCloud/pveCloud/server/internal/task"
	taskhandler "github.com/AeolianCloud/pveCloud/server/internal/task/handler"
	"github.com/AeolianCloud/pveCloud/server/internal/user"
	userhandler "github.com/AeolianCloud/pveCloud/server/internal/user/handler"
	"github.com/go-redis/redis/v8"
)

type App interface {
	Handler() http.Handler
	Server() *http.Server
}

type app struct {
	server *http.Server
	logger *slog.Logger
	db     *sql.DB
	redis  *redis.Client
}

func NewPublicApp(cfg config.Config) (App, error) {
	return newHTTPApp("public-api", cfg.PublicAPIAddr, cfg)
}

func NewAdminApp(cfg config.Config) (App, error) {
	return newHTTPApp("admin-api", cfg.AdminAPIAddr, cfg)
}

func NewWorkerApp(cfg config.Config) (App, *sql.DB, error) {
	base, err := newHTTPApp("worker", cfg.WorkerAddr, cfg)
	if err != nil {
		return nil, nil, err
	}

	concrete, ok := base.(*app)
	if !ok {
		return nil, nil, errors.New("unexpected worker app type")
	}

	return concrete, concrete.db, nil
}

func (a *app) Handler() http.Handler {
	return a.server.Handler
}

func (a *app) Server() *http.Server {
	return a.server
}

func newHTTPApp(serviceName, addr string, cfg config.Config) (App, error) {
	if err := cfg.ValidateForService(serviceName); err != nil {
		return nil, err
	}

	db, redisClient, logger, err := newRuntime(cfg)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"service": serviceName,
			"env":     cfg.AppEnv,
		})
	})

	switch serviceName {
	case "public-api":
		webSigner := auth.NewJWTSigner(cfg.JWTWebSecret)
		userSvc := user.NewService(db, webSigner)
		userAuth := auth.Middleware(webSigner, "user")

		catalogRepo := catalog.NewMySQLRepository(db)
		catalogSvc := catalog.NewService(catalogRepo, 15*time.Minute)
		billingSvc := &orderBillingAdapter{svc: billing.NewService(billing.NewMySQLRepository(db))}
		orderRepo := order.NewMySQLRepository(db)
		paymentRepo := payment.NewMySQLRepository(db)
		callbackStore := payment.NewMySQLCallbackStore(db)
		paymentSvc := payment.NewServiceWithCallbackStore(paymentRepo, callbackStore)
		orderSvc := order.NewService(db, orderRepo, billingSvc, paymentSvc, catalogSvc)
		instanceSvc := instance.NewService(
			instance.NewMySQLRepository(db),
			resource.NewMockClient(),
			audit.NewService(audit.NewMySQLRepository(db)),
			notification.NewService(notification.NewMySQLRepository(db)),
		)

		authHandler := userhandler.NewAuthHandler(userSvc)
		registerHandler := userhandler.NewRegisterHandler(userSvc)
		publicProductsHandler := cataloghandler.NewPublicHandler(catalogSvc)
		verifier := payment.NewProviderVerifier(cfg.Payment)
		callbackHandler := paymenthandler.NewCallbackHandler(paymentSvc, verifier)
		publicOrdersHandler := orderhandler.NewPublicHandler(orderSvc)
		publicPaymentHandler := paymenthandler.NewPublicPaymentHandler(paymentSvc)
		publicInstancesHandler := instancehandler.NewPublicHandler(instanceSvc)
		publicInstanceDetailHandler := instancehandler.NewDetailHandler(instanceSvc)
		notificationSvc := notification.NewService(notification.NewMySQLRepository(db))
		noticeHandler := notificationhandler.NewPublicHandler(notificationSvc)

		mux.HandleFunc("POST /auth/login", authHandler.Login)
		mux.HandleFunc("POST /auth/register", registerHandler.Register)
		mux.HandleFunc("GET /products", publicProductsHandler.ListProducts)
		mux.HandleFunc("GET /products/{productID}", publicProductsHandler.GetProduct)
		mux.HandleFunc("POST /payments/callback", callbackHandler.Handle)
		mux.Handle("POST /orders", userAuth(http.HandlerFunc(publicOrdersHandler.CreateOrder)))
		mux.Handle("GET /orders", userAuth(http.HandlerFunc(publicOrdersHandler.ListOrders)))
		mux.Handle("GET /payments/{paymentOrderNo}", userAuth(http.HandlerFunc(publicPaymentHandler.GetPaymentStatus)))
		mux.Handle("GET /instances", userAuth(http.HandlerFunc(publicInstancesHandler.ListMine)))
		mux.Handle("GET /instances/{instanceID}", userAuth(http.HandlerFunc(publicInstanceDetailHandler.GetMineByID)))
		mux.Handle("GET /notices", userAuth(http.HandlerFunc(noticeHandler.ListNotices)))
		mux.Handle("PUT /notices/{id}/read", userAuth(http.HandlerFunc(noticeHandler.MarkRead)))
	case "admin-api":
		adminSigner := auth.NewJWTSigner(cfg.JWTAdminSecret)
		adminSvc := adminuser.NewService(db, adminSigner)
		adminAuth := auth.Middleware(adminSigner, "admin")

		catalogRepo := catalog.NewMySQLRepository(db)
		catalogSvc := catalog.NewService(catalogRepo, 15*time.Minute)
		orderSvc := order.NewService(
			db,
			order.NewMySQLRepository(db),
			&orderBillingAdapter{svc: billing.NewService(billing.NewMySQLRepository(db))},
			payment.NewService(payment.NewMySQLRepository(db)),
			catalogSvc,
		)
		instanceSvc := instance.NewService(
			instance.NewMySQLRepository(db),
			resource.NewMockClient(),
			audit.NewService(audit.NewMySQLRepository(db)),
			notification.NewService(notification.NewMySQLRepository(db)),
		)
		taskSvc := task.NewService(task.NewMySQLRepository(db))

		authHandler := adminhandler.NewAuthHandler(adminSvc)
		adminProductsHandler := cataloghandler.NewAdminHandler(catalogSvc)
		adminOrdersHandler := orderhandler.NewAdminHandler(orderSvc)
		adminInstancesHandler := instancehandler.NewAdminHandler(instanceSvc)
		adminTasksHandler := taskhandler.NewAdminHandler(taskSvc)
		webSigner := auth.NewJWTSigner(cfg.JWTWebSecret)
		userSvc := user.NewService(db, webSigner)
		adminUsersHandler := userhandler.NewAdminUsersHandler(userSvc)
		adminAdminsHandler := adminhandler.NewAdminAdminsHandler(adminSvc)
		dashboardHandler := adminhandler.NewDashboardHandler(db)

		mux.HandleFunc("POST /auth/login", authHandler.Login)
		mux.Handle("GET /products", adminAuth(http.HandlerFunc(adminProductsHandler.ListProducts)))
		mux.Handle("GET /orders", adminAuth(http.HandlerFunc(adminOrdersHandler.ListOrders)))
		mux.Handle("GET /instances", adminAuth(http.HandlerFunc(adminInstancesHandler.ListAll)))
		mux.Handle("GET /tasks", adminAuth(http.HandlerFunc(adminTasksHandler.ListTasks)))
		mux.Handle("GET /users", adminAuth(http.HandlerFunc(adminUsersHandler.ListUsers)))
		mux.Handle("GET /admins", adminAuth(http.HandlerFunc(adminAdminsHandler.ListAdmins)))
		mux.Handle("GET /dashboard", adminAuth(http.HandlerFunc(dashboardHandler.Stats)))
	}

	return &app{
		server: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
		logger: logger,
		db:     db,
		redis:  redisClient,
	}, nil
}

func newRuntime(cfg config.Config) (*sql.DB, *redis.Client, *slog.Logger, error) {
	db, err := database.Open(cfg.MySQLDSN)
	if err != nil {
		return nil, nil, nil, err
	}

	logger := loggerx.New(cfg.AppEnv)
	redisClient := cache.NewClient(cfg.RedisAddr)

	return db, redisClient, logger, nil
}

type orderBillingAdapter struct {
	svc *billing.Service
}

func (a *orderBillingAdapter) Quote(ctx context.Context, skuID uint64, cycle string) (order.BillingQuote, error) {
	row, err := a.svc.Quote(ctx, skuID, cycle)
	if err != nil {
		return order.BillingQuote{}, err
	}
	return order.BillingQuote{
		Cycle:          row.Cycle,
		OriginalAmount: row.OriginalAmount,
		DiscountAmount: row.DiscountAmount,
		PayableAmount:  row.PayableAmount,
	}, nil
}

func (a *orderBillingAdapter) CreateRecord(ctx context.Context, q database.Querier, in billing.CreateRecordInput) (billing.Record, error) {
	return a.svc.CreateRecord(ctx, q, in)
}
