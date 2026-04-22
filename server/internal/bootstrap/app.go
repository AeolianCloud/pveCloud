package bootstrap

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/adminuser"
	adminhandler "github.com/AeolianCloud/pveCloud/server/internal/adminuser/handler"
	"github.com/AeolianCloud/pveCloud/server/internal/auth"
	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap/config"
	"github.com/AeolianCloud/pveCloud/server/internal/common/cache"
	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
	loggerx "github.com/AeolianCloud/pveCloud/server/internal/common/logger"
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

func NewWorkerApp(cfg config.Config) (App, error) {
	return newHTTPApp("worker", cfg.WorkerAddr, cfg)
}

func (a *app) Handler() http.Handler {
	return a.server.Handler
}

func (a *app) Server() *http.Server {
	return a.server
}

func newHTTPApp(serviceName, addr string, cfg config.Config) (App, error) {
	db, err := database.Open(cfg.MySQLDSN)
	if err != nil {
		return nil, err
	}

	logger := loggerx.New(cfg.AppEnv)
	redisClient := cache.NewClient(cfg.RedisAddr)

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
		userSvc := user.NewService(db, auth.NewJWTSigner(cfg.JWTWebSecret))
		authHandler := userhandler.NewAuthHandler(userSvc)
		registerHandler := userhandler.NewRegisterHandler(userSvc)
		mux.HandleFunc("POST /auth/login", authHandler.Login)
		mux.HandleFunc("POST /auth/register", registerHandler.Register)
	case "admin-api":
		adminSvc := adminuser.NewService(db, auth.NewJWTSigner(cfg.JWTAdminSecret))
		authHandler := adminhandler.NewAuthHandler(adminSvc)
		mux.HandleFunc("POST /auth/login", authHandler.Login)
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
