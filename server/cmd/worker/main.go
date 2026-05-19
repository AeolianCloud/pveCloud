package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/AeolianCloud/pveCloud/server/internal/app/worker"
)

func main() {
	configPath := flag.String("config", "config.yaml", "YAML 配置文件路径")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := worker.NewApp(ctx, *configPath)
	if err != nil {
		log.Fatalf("初始化 Worker 失败：%v", err)
	}

	app.Logger.Info("Worker 进程启动", "worker_id", app.Config.Worker.ID)
	if err := app.Runner.Run(ctx); err != nil {
		app.Logger.Error("Worker 进程异常退出", "error", err)
	}
}
