package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	appconfig "pvecloud/backend/internal/config"
	"pvecloud/backend/internal/handler"
	"pvecloud/backend/internal/job"
	"pvecloud/backend/internal/middleware"
	"pvecloud/backend/internal/model"
	"pvecloud/backend/internal/pveclient"
	"pvecloud/backend/internal/repository"
	"pvecloud/backend/internal/security"
	"pvecloud/backend/internal/service"
)

// Run 启动服务。cfgPath 是配置文件路径，通常为 "config/config.yaml"。
func Run(cfgPath string) error {
	cfg, err := appconfig.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config failed: %w", err)
	}

	db, err := gorm.Open(mysql.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("connect database failed: %w", err)
	}
	if err := db.AutoMigrate(model.AllModels()...); err != nil {
		return fmt.Errorf("auto migrate failed: %w", err)
	}

	var redisClient *redis.Client
	redisClient = redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB})
	if pingErr := redisClient.Ping(context.Background()).Err(); pingErr != nil {
		log.Printf("redis unavailable, token store disabled: %v", pingErr)
		redisClient = nil
	}
	tokenStore := security.NewTokenStore(redisClient)

	jwtManager := middleware.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessTokenExpireH, cfg.JWT.RefreshTokenExpireDH)
	rateLimiter := middleware.NewLoginRateLimiter(5, 30*time.Minute)

	var pve pveclient.PVEClient
	if cfg.PVEClientMode == "http" {
		pve = pveclient.NewHTTPClient(pveclient.HTTPClientConfig{
			BaseURL:                 cfg.PVE.BaseURL,
			APIKey:                  cfg.PVE.APIKey,
			APISecret:               cfg.PVE.APISecret,
			Timeout:                 time.Duration(cfg.PVE.TimeoutSeconds) * time.Second,
			MaxRetries:              cfg.PVE.MaxRetries,
			RetryBackoff:            time.Duration(cfg.PVE.RetryBackoffMS) * time.Millisecond,
			CircuitFailureThreshold: cfg.PVE.CircuitFailureThreshold,
			CircuitOpenDuration:     time.Duration(cfg.PVE.CircuitOpenSeconds) * time.Second,
		})
	} else {
		pve = pveclient.NewMockClient()
	}

	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	instanceRepo := repository.NewInstanceRepository(db)
	snapshotRepo := repository.NewSnapshotRepository(db)
	ticketRepo := repository.NewTicketRepository(db)

	userService := service.NewUserService(userRepo, jwtManager, tokenStore, service.ConsoleEmailSender{})
	productService := service.NewProductService(productRepo)
	billingService := service.NewBillingService(walletRepo)
	orderService := service.NewOrderService(db, productRepo, walletRepo, orderRepo, taskRepo, instanceRepo, pve)
	instanceService := service.NewInstanceService(instanceRepo, snapshotRepo, pve)
	ticketService := service.NewTicketService(ticketRepo)

	authHandler := handler.NewAuthHandler(userService, rateLimiter)
	productHandler := handler.NewProductHandler(productService)
	walletHandler := handler.NewWalletHandler(billingService)
	orderHandler := handler.NewOrderHandler(orderService)
	taskHandler := handler.NewTaskHandler(orderService)
	instanceHandler := handler.NewInstanceHandler(instanceService)
	snapshotHandler := handler.NewSnapshotHandler(instanceService)
	ticketHandler := handler.NewTicketHandler(ticketService)
	adminHandler := handler.NewAdminHandler(userRepo, orderRepo, ticketRepo, tokenStore, pve)

	r := gin.Default()

	api := r.Group("/api/v1")
	pub := api.Group("/pub")
	user := api.Group("/user", middleware.AuthMiddleware(jwtManager, tokenStore))
	admin := api.Group("/admin", middleware.AuthMiddleware(jwtManager, tokenStore), middleware.AdminOnlyMiddleware())

	authHandler.RegisterRoutes(pub, user)
	productHandler.RegisterRoutes(pub, admin)
	walletHandler.RegisterRoutes(user)
	orderHandler.RegisterRoutes(user, admin)
	taskHandler.RegisterRoutes(user)
	instanceHandler.RegisterRoutes(user)
	snapshotHandler.RegisterRoutes(user)
	ticketHandler.RegisterRoutes(user, admin)
	adminHandler.RegisterRoutes(admin)

	ctx := context.Background()
	go job.NewTaskSyncer(taskRepo, orderRepo, instanceRepo, walletRepo, pve).Start(ctx)
	go job.NewInstanceStatusSyncer(instanceRepo, instanceService).Start(ctx)
	go job.NewExpireChecker(instanceRepo, instanceService).Start(ctx)
	go job.NewHourlyBilling(instanceRepo, billingService, pve).Start(ctx)

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	if err := r.Run(addr); err != nil {
		return fmt.Errorf("run server failed: %w", err)
	}
	return nil
}
