package bootstrap

import (
	"context"
	"time"
)

type Worker struct {
	app *App
}

/**
 * NewWorker 创建异步任务 Worker。
 *
 * @param app API 和 Worker 共享的运行时依赖容器
 * @return *Worker 异步任务 Worker
 */
func NewWorker(app *App) *Worker {
	return &Worker{app: app}
}

/**
 * Run 启动 Worker 主循环。
 *
 * @param ctx Worker 生命周期上下文
 * @return error 当前阶段固定返回 nil；后续任务执行失败会通过任务状态记录
 */
func (w *Worker) Run(ctx context.Context) error {
	w.app.Logger.Info("Worker 已启动", "worker_id", w.app.Config.Worker.ID)

	timer := time.NewTimer(w.app.Config.Worker.PollInterval())
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			w.app.Logger.Info("Worker 已停止", "worker_id", w.app.Config.Worker.ID)
			return nil
		case <-timer.C:
			w.app.Logger.Debug("Worker 心跳", "worker_id", w.app.Config.Worker.ID)
			timer.Reset(w.app.Config.Worker.PollInterval())
		}
	}
}
