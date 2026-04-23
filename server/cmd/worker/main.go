package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/audit"
	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap"
	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap/config"
	"github.com/AeolianCloud/pveCloud/server/internal/instance"
	"github.com/AeolianCloud/pveCloud/server/internal/notification"
	"github.com/AeolianCloud/pveCloud/server/internal/resource"
	"github.com/AeolianCloud/pveCloud/server/internal/task"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	app, db, err := bootstrap.NewWorkerApp(cfg)
	if err != nil {
		log.Fatal(err)
	}

	pollInterval, err := time.ParseDuration(cfg.Worker.PollInterval)
	if err != nil {
		log.Fatal(err)
	}

	repo := task.NewMySQLRepository(db)
	logRepo := task.NewMySQLLogRepository(db)
	vmClient := buildVMClient(cfg)
	instanceSvc := instance.NewService(
		instance.NewMySQLRepository(db),
		vmClient,
		audit.NewService(audit.NewMySQLRepository(db)),
		notification.NewService(),
	)
	executor := task.NewDispatchingExecutor(func(ctx context.Context, orderID uint64) error {
		_, err := instanceSvc.HandleCreateInstanceTask(ctx, orderID)
		if err == nil {
			return nil
		}

		var providerErr *resource.ProviderError
		if errors.As(err, &providerErr) && providerErr.Retryable {
			return task.Retryable(err, providerErr.Delay)
		}

		return err
	})
	worker := task.NewWorker(repo, logRepo, executor, "worker-1")
	worker.SetPollInterval(pollInterval)
	worker.SetBatchSize(cfg.Worker.BatchSize)

	go func() {
		if err := app.Server().ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	if err := worker.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func buildVMClient(cfg config.Config) resource.VMClient {
	switch cfg.Resource.Provider {
	case "mock":
		return resource.NewMockClient()
	default:
		return resource.NewMockClient()
	}
}
