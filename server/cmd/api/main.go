package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap"
	"github.com/AeolianCloud/pveCloud/server/internal/routes"
)

func main() {
	configPath := flag.String("config", "config.yaml", "YAML 配置文件路径")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := bootstrap.NewApp(ctx, *configPath)
	if err != nil {
		log.Fatalf("初始化 API 应用失败：%v", err)
	}

	server := &http.Server{
		Addr:              app.Config.App.Addr,
		Handler:           routes.NewRouter(app),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		app.Logger.Info("API 服务正在监听", "addr", app.Config.App.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.Logger.Error("API 服务异常停止", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.Config.App.ShutdownTimeout())
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		app.Logger.Error("API 服务优雅退出失败", "error", err)
	}
}
