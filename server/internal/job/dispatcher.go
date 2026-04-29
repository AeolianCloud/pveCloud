package job

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"
)

/**
 * Dispatcher 负责领取任务、分发 handler 并推进任务状态。
 */
type Dispatcher struct {
	cfg      Config
	store    Store
	handlers map[string]Handler
	logger   *slog.Logger
	now      func() time.Time
}

/**
 * NewDispatcher 创建异步任务调度器。
 */
func NewDispatcher(cfg Config, store Store, handlers map[string]Handler, logger *slog.Logger) *Dispatcher {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 1
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 5 * time.Second
	}
	if cfg.LockTTL <= 0 {
		cfg.LockTTL = time.Minute
	}
	if logger == nil {
		logger = slog.Default()
	}
	if handlers == nil {
		handlers = map[string]Handler{}
	}
	return &Dispatcher{
		cfg:      cfg,
		store:    store,
		handlers: handlers,
		logger:   logger,
		now:      time.Now,
	}
}

/**
 * Run 启动任务调度循环。
 */
func (d *Dispatcher) Run(ctx context.Context) error {
	d.logger.Info("Worker 已启动", "worker_id", d.cfg.WorkerID)
	defer d.logger.Info("Worker 已停止", "worker_id", d.cfg.WorkerID)

	for {
		if err := d.RunOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
			d.logger.Error("Worker 单轮执行失败", "worker_id", d.cfg.WorkerID, "error", err)
		}

		timer := time.NewTimer(d.cfg.PollInterval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return nil
		case <-timer.C:
		}
	}
}

/**
 * RunOnce 执行一轮任务领取和处理。
 */
func (d *Dispatcher) RunOnce(ctx context.Context) error {
	for i := 0; i < d.cfg.BatchSize; i++ {
		task, err := d.store.Claim(ctx, d.now())
		if errors.Is(err, ErrNoTask) {
			return nil
		}
		if err != nil {
			return err
		}
		d.process(ctx, task)
	}
	return nil
}

func (d *Dispatcher) process(ctx context.Context, task AsyncTask) {
	handler, ok := d.handlers[strings.TrimSpace(task.TaskType)]
	if !ok {
		if err := d.store.MarkFailed(ctx, task, ErrUnregisteredHandler, d.now()); err != nil {
			d.logger.Error("标记未注册任务失败", "task_id", task.ID, "task_type", task.TaskType, "error", err)
		}
		return
	}

	result := d.callHandler(ctx, handler, task)
	now := d.now()
	if result.Err == nil {
		if err := d.store.MarkSucceeded(ctx, task, result.Result, now); err != nil {
			d.logger.Error("标记任务成功失败", "task_id", task.ID, "task_type", task.TaskType, "error", err)
		}
		return
	}

	if result.Retryable && canRetry(task) {
		nextRunAt := now.Add(d.retryDelay(task))
		if err := d.store.MarkRetryableFailure(ctx, task, result.Err, nextRunAt, now); err != nil {
			d.logger.Error("标记任务重试失败", "task_id", task.ID, "task_type", task.TaskType, "error", err)
		}
		return
	}

	if err := d.store.MarkFailed(ctx, task, result.Err, now); err != nil {
		d.logger.Error("标记任务失败失败", "task_id", task.ID, "task_type", task.TaskType, "error", err)
	}
}

func (d *Dispatcher) callHandler(ctx context.Context, handler Handler, task AsyncTask) (result HandlerResult) {
	defer func() {
		if recovered := recover(); recovered != nil {
			result = RetryableFailure(fmt.Errorf("panic: %v", recovered))
		}
	}()
	return handler(ctx, task)
}

func (d *Dispatcher) retryDelay(task AsyncTask) time.Duration {
	exponent := math.Pow(2, float64(task.RetryCount))
	delay := time.Duration(exponent) * d.cfg.PollInterval
	maxDelay := 30 * time.Minute
	if delay > maxDelay {
		return maxDelay
	}
	if delay <= 0 {
		return d.cfg.PollInterval
	}
	return delay
}

func canRetry(task AsyncTask) bool {
	if task.MaxRetries == 0 {
		return false
	}
	return task.RetryCount+1 < task.MaxRetries
}

func errorMessage(err error) *string {
	if err == nil {
		return nil
	}
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return nil
	}
	if len(message) > 2000 {
		message = message[:2000]
	}
	return &message
}
