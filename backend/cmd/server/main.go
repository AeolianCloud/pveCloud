package main

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

// main 负责组装依赖、配置路由并启动 HTTP 服务。
func main() {
	cfg, err := appconfig.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	db, err := gorm.Open(mysql.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect database failed: %v", err)
	}
	if err := db.AutoMigrate(model.AllModels()...); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	// Redis 用于 token 黑名单与 refresh token 会话存储；不可用时退化为无状态模式。
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

	userService := service.NewUserService(userRepo, jwtManager, tokenStore)
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
	adminHandler := handler.NewAdminHandler(userRepo, orderRepo, tokenStore, pve)

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
	go job.NewExpireChecker(instanceRepo, instanceService).Start(ctx)
	go job.NewHourlyBilling(instanceRepo, billingService).Start(ctx)

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("run server failed: %v", err)
	}
}
