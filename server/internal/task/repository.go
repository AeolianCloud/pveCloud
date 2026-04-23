package task

import (
	"context"
	"time"
)

type Repository interface {
	FindByBusinessKey(ctx context.Context, taskType, businessType string, businessID uint64) (Task, bool, error)
	CreateTask(ctx context.Context, in CreateTaskParams) (Task, error)
	ListTasks(ctx context.Context, limit int) ([]Task, error)
}

type WorkerRepository interface {
	ClaimPendingTask(ctx context.Context, now time.Time, workerName string) (*Task, error)
	MarkTaskSuccess(ctx context.Context, taskID uint64, now time.Time) error
	MarkTaskRetry(ctx context.Context, taskID uint64, nextRunAt time.Time, now time.Time) error
	MarkTaskFailed(ctx context.Context, taskID uint64, now time.Time) error
}
