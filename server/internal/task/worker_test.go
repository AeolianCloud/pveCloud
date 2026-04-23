package task

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeWorkerRepo struct {
	tasks      []*Task
	claimedNow time.Time
	claimedBy  string
	successIDs []uint64
	retryIDs   []uint64
	retryAt    []time.Time
	failedIDs  []uint64
}

func (f *fakeWorkerRepo) ClaimPendingTask(ctx context.Context, now time.Time, workerName string) (*Task, error) {
	f.claimedNow = now
	f.claimedBy = workerName
	if len(f.tasks) == 0 {
		return nil, nil
	}
	row := f.tasks[0]
	f.tasks = f.tasks[1:]
	return row, nil
}

func (f *fakeWorkerRepo) MarkTaskSuccess(ctx context.Context, taskID uint64, now time.Time) error {
	f.successIDs = append(f.successIDs, taskID)
	return nil
}

func (f *fakeWorkerRepo) MarkTaskRetry(ctx context.Context, taskID uint64, nextRunAt time.Time, now time.Time) error {
	f.retryIDs = append(f.retryIDs, taskID)
	f.retryAt = append(f.retryAt, nextRunAt)
	return nil
}

func (f *fakeWorkerRepo) MarkTaskFailed(ctx context.Context, taskID uint64, now time.Time) error {
	f.failedIDs = append(f.failedIDs, taskID)
	return nil
}

type fakeLogRepo struct {
	levels   []string
	messages []string
}

func (f *fakeLogRepo) AppendLog(ctx context.Context, taskID uint64, level, message string) error {
	f.levels = append(f.levels, level)
	f.messages = append(f.messages, message)
	return nil
}

type fakeExecutor struct {
	err      error
	executed []Task
}

func (f *fakeExecutor) Execute(ctx context.Context, row Task) error {
	f.executed = append(f.executed, row)
	return f.err
}

func TestWorkerClaimNextUsesWorkerName(t *testing.T) {
	repo := &fakeWorkerRepo{
		tasks: []*Task{{
			ID:       8001,
			TaskNo:   "T8001",
			TaskType: "create_instance",
			Status:   "processing",
		}},
	}
	worker := NewWorker(repo, nil, nil, "worker-1")

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

func TestWorkerRunOnceMarksTaskSuccess(t *testing.T) {
	repo := &fakeWorkerRepo{
		tasks: []*Task{{
			ID:            8001,
			TaskNo:        "T8001",
			TaskType:      "create_instance",
			BusinessType:  "order",
			BusinessID:    5001,
			Status:        "processing",
			MaxRetryCount: 5,
		}},
	}
	logRepo := &fakeLogRepo{}
	executor := &fakeExecutor{}
	worker := NewWorker(repo, logRepo, executor, "worker-1")

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("run once: %v", err)
	}
	if len(executor.executed) != 1 || executor.executed[0].TaskNo != "T8001" {
		t.Fatalf("expected executor to run T8001, got %+v", executor.executed)
	}
	if len(repo.successIDs) != 1 || repo.successIDs[0] != 8001 {
		t.Fatalf("expected task 8001 marked success, got %+v", repo.successIDs)
	}
	if len(logRepo.levels) != 2 || logRepo.levels[0] != "info" || logRepo.levels[1] != "info" {
		t.Fatalf("expected info start/success logs, got %+v", logRepo.levels)
	}
}

func TestWorkerRunOnceSchedulesRetryForRetryableError(t *testing.T) {
	now := time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC)
	repo := &fakeWorkerRepo{
		tasks: []*Task{{
			ID:            8002,
			TaskNo:        "T8002",
			TaskType:      "create_instance",
			BusinessType:  "order",
			BusinessID:    5002,
			Status:        "processing",
			RetryCount:    1,
			MaxRetryCount: 5,
		}},
	}
	logRepo := &fakeLogRepo{}
	executor := &fakeExecutor{err: Retryable(errors.New("temporary failure"), 2*time.Minute)}
	worker := NewWorker(repo, logRepo, executor, "worker-1")
	worker.now = func() time.Time { return now }

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("run once: %v", err)
	}
	if len(repo.retryIDs) != 1 || repo.retryIDs[0] != 8002 {
		t.Fatalf("expected task 8002 marked retry, got %+v", repo.retryIDs)
	}
	if len(repo.retryAt) != 1 || !repo.retryAt[0].Equal(now.Add(2*time.Minute)) {
		t.Fatalf("expected retry at %s, got %+v", now.Add(2*time.Minute), repo.retryAt)
	}
	if got := logRepo.levels[len(logRepo.levels)-1]; got != "warn" {
		t.Fatalf("expected final warn log, got %s", got)
	}
}

func TestWorkerRunOnceMarksTaskFailedForTerminalError(t *testing.T) {
	repo := &fakeWorkerRepo{
		tasks: []*Task{{
			ID:            8003,
			TaskNo:        "T8003",
			TaskType:      "create_instance",
			BusinessType:  "order",
			BusinessID:    5003,
			Status:        "processing",
			RetryCount:    4,
			MaxRetryCount: 5,
		}},
	}
	logRepo := &fakeLogRepo{}
	executor := &fakeExecutor{err: errors.New("permanent failure")}
	worker := NewWorker(repo, logRepo, executor, "worker-1")

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("run once: %v", err)
	}
	if len(repo.failedIDs) != 1 || repo.failedIDs[0] != 8003 {
		t.Fatalf("expected task 8003 marked failed, got %+v", repo.failedIDs)
	}
	if got := logRepo.levels[len(logRepo.levels)-1]; got != "error" {
		t.Fatalf("expected final error log, got %s", got)
	}
}
