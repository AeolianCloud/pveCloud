package bootstrap

import (
	"context"
	"log/slog"

	"gorm.io/gorm"

	"github.com/AeolianCloud/pveCloud/server/internal/pkg/logger"
)

type App struct {
	Config *Config
	DB     *gorm.DB
	Logger *slog.Logger
}

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

	return &App{
		Config: cfg,
		DB:     db,
		Logger: log,
	}, nil
}
