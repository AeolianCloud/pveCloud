package bootstrap

import (
	"context"
	"log/slog"

	"gorm.io/gorm"

	"github.com/AeolianCloud/pveCloud/server/internal/pkg/cache"
	"github.com/AeolianCloud/pveCloud/server/internal/pkg/logger"
)

/**
 * App 表示 API 和 Worker 共享的运行时依赖容器。
 */
type App struct {
	Config *Config
	DB     *gorm.DB
	Redis  *cache.Redis
	Logger *slog.Logger
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
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	log := logger.New(cfg.Log.Level)

	db, err := ConnectDatabase(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	redisClient, err := ConnectRedis(ctx, cfg.Redis)
	if err != nil {
		return nil, err
	}

	return &App{
		Config: cfg,
		DB:     db,
		Redis:  redisClient,
		Logger: log,
	}, nil
}
