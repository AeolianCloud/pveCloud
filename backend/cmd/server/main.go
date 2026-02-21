package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"pvecloud/backend/internal/config"
	"pvecloud/backend/internal/database"
	"pvecloud/backend/internal/security"
	"pvecloud/backend/internal/session"
	"pvecloud/backend/internal/router"
)

func main() {
	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	log := newLogger(cfg.Log)
	defer log.Sync() //nolint:errcheck

	// 初始化数据库
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal("数据库初始化失败", zap.Error(err))
	}

	os.Setenv("GIN_MODE", cfg.Server.Mode)

	// 初始化 Redis（会话存储）
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer rdb.Close() // 关闭连接池（不影响优雅退出逻辑）
	// 启动时做一次连通性检查，避免运行中才暴露问题
	{
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			log.Fatal("Redis 连接失败", zap.Error(err), zap.String("地址", cfg.Redis.Addr()))
		}
	}
	sessStore := session.NewRedisStore(rdb, cfg.Redis.KeyPrefix)
	loginGuard := security.NewLoginGuard(rdb, cfg.Redis.KeyPrefix, cfg.Security.Login)

	// 构建路由
	r := router.New(db, log, cfg, sessStore, loginGuard)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 异步启动
	go func() {
		log.Info("服务启动", zap.String("地址", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("监听失败", zap.Error(err))
		}
	}()

	// 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("正在关闭服务...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("关闭服务失败", zap.Error(err))
	}
	log.Info("服务已停止")
}

func newLogger(cfg config.LogConfig) *zap.Logger {
	level := zap.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zap.DebugLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "时间"
	encCfg.MessageKey = "消息"
	encCfg.LevelKey = "级别"
	encCfg.CallerKey = "位置"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var enc zapcore.Encoder
	if cfg.Encoding == "console" {
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		enc = zapcore.NewConsoleEncoder(encCfg)
	} else {
		enc = zapcore.NewJSONEncoder(encCfg)
	}

	core := zapcore.NewCore(enc, zapcore.AddSync(os.Stdout), level)
	return zap.New(core, zap.AddCaller())
}
