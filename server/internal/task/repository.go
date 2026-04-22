package task

import (
	"context"
	"time"
)

type Repository interface {
	FindByBusinessKey(ctx context.Context, taskType, businessType string, businessID uint64) (Task, bool, error)
	CreateTask(ctx context.Context, in CreateTaskParams) (Task, error)
}

type WorkerRepository interface {
	ClaimPendingTask(ctx context.Context, now time.Time, workerName string) (*Task, error)
}
