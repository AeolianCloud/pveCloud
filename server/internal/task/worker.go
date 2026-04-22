package task

import (
	"context"
	"time"
)

type Worker struct {
	repo       WorkerRepository
	workerName string
	now        func() time.Time
}

func NewWorker(repo WorkerRepository, workerName string) *Worker {
	return &Worker{
		repo:       repo,
		workerName: workerName,
		now:        time.Now,
	}
}

func (w *Worker) ClaimNext(ctx context.Context) (*Task, error) {
	return w.repo.ClaimPendingTask(ctx, w.now(), w.workerName)
}
