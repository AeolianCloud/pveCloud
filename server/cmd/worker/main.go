package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/AeolianCloud/pveCloud/server/internal/platform/bootstrap"
)

func main() {
	configPath := flag.String("config", "config.yaml", "YAML 配置文件路径")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := bootstrap.NewApp(ctx, *configPath)
	if err != nil {
		log.Fatalf("初始化 Worker 应用失败：%v", err)
	}

	worker := bootstrap.NewWorker(app)
	if err := worker.Run(ctx); err != nil {
		log.Fatalf("运行 Worker 失败：%v", err)
	}
}
