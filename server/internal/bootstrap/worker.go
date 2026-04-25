package bootstrap

import (
	"context"
	"time"
)

type Worker struct {
	app *App
}

func NewWorker(app *App) *Worker {
	return &Worker{app: app}
}

func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.app.Config.Worker.PollInterval())
	defer ticker.Stop()

	w.app.Logger.Info("worker started", "worker_id", w.app.Config.Worker.ID)

	for {
		select {
		case <-ctx.Done():
			w.app.Logger.Info("worker stopped", "worker_id", w.app.Config.Worker.ID)
			return nil
		case <-ticker.C:
			w.app.Logger.Debug("worker tick", "worker_id", w.app.Config.Worker.ID)
		}
	}
}
