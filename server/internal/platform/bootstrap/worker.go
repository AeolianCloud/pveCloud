package bootstrap

import (
	"context"

	"github.com/AeolianCloud/pveCloud/server/internal/job"
	"github.com/AeolianCloud/pveCloud/server/internal/job/handlers"
)

type Worker struct {
	dispatcher *job.Dispatcher
}

/**
 * NewWorker 创建异步任务 Worker。
 *
 * @param app API 和 Worker 共享的运行时依赖容器
 * @return *Worker 异步任务 Worker
 */
func NewWorker(app *App) *Worker {
	cfg := job.Config{
		WorkerID:     app.Config.Worker.ID,
		PollInterval: app.Config.Worker.PollInterval(),
		LockTTL:      app.Config.Worker.LockTTL(),
		BatchSize:    app.Config.Worker.BatchSize,
	}
	store := job.NewGormStore(app.DB, cfg.WorkerID, cfg.LockTTL)
	dispatcher := job.NewDispatcher(cfg, store, handlers.NewRegistry(), app.Logger)
	return &Worker{dispatcher: dispatcher}
}

/**
 * Run 启动 Worker 主循环。
 *
 * @param ctx Worker 生命周期上下文
 * @return error Worker 调度循环错误
 */
func (w *Worker) Run(ctx context.Context) error {
	return w.dispatcher.Run(ctx)
}
