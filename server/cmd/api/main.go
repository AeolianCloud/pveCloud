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
	configPath := flag.String("config", "config.yaml", "path to yaml config file")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := bootstrap.NewApp(ctx, *configPath)
	if err != nil {
		log.Fatalf("bootstrap api app: %v", err)
	}

	server := &http.Server{
		Addr:              app.Config.App.Addr,
		Handler:           routes.NewRouter(app),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		app.Logger.Info("api server listening", "addr", app.Config.App.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.Logger.Error("api server stopped unexpectedly", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.Config.App.ShutdownTimeout())
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		app.Logger.Error("api server graceful shutdown failed", "error", err)
	}
}
