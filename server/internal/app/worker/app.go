package worker

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/AeolianCloud/pveCloud/server/internal/integration/mail"
	"github.com/AeolianCloud/pveCloud/server/internal/integration/mcppve"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/cache"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/database"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/logger"
)

type App struct {
	Config *config.Config
	DB     *gorm.DB
	Redis  *cache.Redis
	Logger *slog.Logger
	MCPPVE *mcppve.Client
	Runner *Runner
}

func NewApp(ctx context.Context, configPath string) (*App, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置文件 %q 失败: %w", configPath, err)
	}
	log := logger.New(cfg.Log.Level)
	db, err := database.ConnectDatabase(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}
	redisClient, err := cache.ConnectRedis(ctx, cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("初始化缓存失败: %w", err)
	}
	mcpPVEClient, err := mcppve.NewClient(cfg.MCPPVE)
	if err != nil {
		return nil, fmt.Errorf("初始化虚拟化管理接口失败: %w", err)
	}
	app := &App{Config: cfg, DB: db, Redis: redisClient, Logger: log, MCPPVE: mcpPVEClient}
	app.Runner = NewRunner(db, log, mcpPVEClient, mail.NewSender(cfg.Mail), cfg.Worker, cfg.InstanceLifecycle, cfg.Notification)
	return app, nil
}
