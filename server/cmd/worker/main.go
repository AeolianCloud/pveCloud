package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to yaml config file")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := bootstrap.NewApp(ctx, *configPath)
	if err != nil {
		log.Fatalf("bootstrap worker app: %v", err)
	}

	worker := bootstrap.NewWorker(app)
	if err := worker.Run(ctx); err != nil {
		log.Fatalf("run worker: %v", err)
	}
}
