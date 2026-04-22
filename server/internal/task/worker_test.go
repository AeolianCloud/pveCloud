package task_test

import (
	"context"
	"testing"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/task"
)

type fakeWorkerRepo struct {
	task       *task.Task
	claimedNow time.Time
	claimedBy  string
}

func (f *fakeWorkerRepo) ClaimPendingTask(ctx context.Context, now time.Time, workerName string) (*task.Task, error) {
	f.claimedNow = now
	f.claimedBy = workerName
	return f.task, nil
}

func TestWorkerClaimNextUsesWorkerName(t *testing.T) {
	repo := &fakeWorkerRepo{
		task: &task.Task{
			ID:       8001,
			TaskNo:   "T8001",
			TaskType: "create_instance",
			Status:   "processing",
		},
	}
	worker := task.NewWorker(repo, "worker-1")

	row, err := worker.ClaimNext(context.Background())
	if err != nil {
		t.Fatalf("claim next: %v", err)
	}
	if row == nil || row.TaskNo != "T8001" {
		t.Fatalf("expected task T8001, got %+v", row)
	}
	if repo.claimedBy != "worker-1" {
		t.Fatalf("expected worker name worker-1, got %s", repo.claimedBy)
	}
	if repo.claimedNow.IsZero() {
		t.Fatalf("expected claim time to be recorded")
	}
}
