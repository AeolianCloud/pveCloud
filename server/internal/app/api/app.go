package api

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/AeolianCloud/pveCloud/server/internal/platform/cache"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/database"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/logger"
	logsusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/logs"
	weblogging "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/logging"
)

/**
 * App 表示 API 共享的运行时依赖容器。
 */
type App struct {
	Config      *config.Config
	DB          *gorm.DB
	Redis       *cache.Redis
	Logger      *slog.Logger
	Logs        *logsusecase.Service
	LogRecorder *weblogging.Recorder
	Routes      RouteSets
}

/**
 * NewApp 根据 YAML 配置文件初始化应用依赖。
 *
 * @param ctx 启动上下文，用于数据库和 Redis 连接检查
 * @param configPath YAML 配置文件路径
 * @return *App 初始化后的应用依赖容器
 * @return error 初始化失败原因
 */
func NewApp(ctx context.Context, configPath string) (*App, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config %q: %w", configPath, err)
	}

	log := logger.New(cfg.Log.Level)

	db, err := database.ConnectDatabase(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	redisClient, err := cache.ConnectRedis(ctx, cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("initialize redis: %w", err)
	}

	app := &App{
		Config:      cfg,
		DB:          db,
		Redis:       redisClient,
		Logger:      log,
		Logs:        logsusecase.NewService(db),
		LogRecorder: weblogging.NewRecorder(db),
	}
	app.Routes = NewRouteSets(app)
	return app, nil
}
