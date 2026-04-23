package task

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Worker struct {
	repo         WorkerRepository
	logRepo      LogRepository
	executor     Executor
	workerName   string
	pollInterval time.Duration
	batchSize    int
	now          func() time.Time
}

func NewWorker(repo WorkerRepository, logRepo LogRepository, executor Executor, workerName string) *Worker {
	return &Worker{
		repo:         repo,
		logRepo:      logRepo,
		executor:     executor,
		workerName:   workerName,
		pollInterval: time.Second,
		batchSize:    1,
		now:          time.Now,
	}
}

func (w *Worker) ClaimNext(ctx context.Context) (*Task, error) {
	return w.repo.ClaimPendingTask(ctx, w.now(), w.workerName)
}

func (w *Worker) SetPollInterval(d time.Duration) {
	if d > 0 {
		w.pollInterval = d
	}
}

func (w *Worker) SetBatchSize(n int) {
	if n > 0 {
		w.batchSize = n
	}
}

func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		if err := w.RunOnce(ctx); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (w *Worker) RunOnce(ctx context.Context) error {
	for i := 0; i < w.batchSize; i++ {
		row, err := w.ClaimNext(ctx)
		if err != nil {
			return err
		}
		if row == nil {
			return nil
		}
		if err := w.processTask(ctx, *row); err != nil {
			return err
		}
	}
	return nil
}

func (w *Worker) processTask(ctx context.Context, row Task) error {
	if w.executor == nil || w.logRepo == nil {
		return errors.New("worker dependencies are not configured")
	}

	if err := w.logRepo.AppendLog(ctx, row.ID, "info", fmt.Sprintf("start task %s", row.TaskNo)); err != nil {
		return err
	}

	err := w.executor.Execute(ctx, row)
	if err == nil {
		if err := w.repo.MarkTaskSuccess(ctx, row.ID, w.now()); err != nil {
			return err
		}
		return w.logRepo.AppendLog(ctx, row.ID, "info", fmt.Sprintf("task %s executed successfully", row.TaskNo))
	}

	var retryable *RetryableError
	if errors.As(err, &retryable) && row.RetryCount+1 < row.MaxRetryCount {
		nextRunAt := w.now().Add(retryable.Delay)
		if err := w.repo.MarkTaskRetry(ctx, row.ID, nextRunAt, w.now()); err != nil {
			return err
		}
		return w.logRepo.AppendLog(ctx, row.ID, "warn", fmt.Sprintf("task %s retry scheduled: %v", row.TaskNo, retryable.Err))
	}

	if err := w.repo.MarkTaskFailed(ctx, row.ID, w.now()); err != nil {
		return err
	}
	return w.logRepo.AppendLog(ctx, row.ID, "error", fmt.Sprintf("task %s failed: %v", row.TaskNo, err))
}
