package task_test

import (
	"context"
	"testing"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/task"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil"
)

func TestClaimPendingTaskLocksRowOnce(t *testing.T) {
	db := testutil.OpenMariaDB(t)
	repo := task.NewMySQLRepository(db)
	now := time.Now().UTC()

	created, err := repo.CreateTask(context.Background(), task.CreateTaskParams{
		TaskType:      "create_instance",
		BusinessType:  "order",
		BusinessID:    5001,
		Status:        "pending",
		NextRunAt:     now,
		MaxRetryCount: 5,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	first, err := repo.ClaimPendingTask(context.Background(), now, "worker-1")
	if err != nil {
		t.Fatalf("claim first task: %v", err)
	}
	if first == nil || first.ID != created.ID {
		t.Fatalf("expected claimed task %d, got %+v", created.ID, first)
	}

	second, err := repo.ClaimPendingTask(context.Background(), now, "worker-2")
	if err != nil {
		t.Fatalf("claim second task: %v", err)
	}
	if second != nil {
		t.Fatalf("expected no second claim, got %+v", second)
	}
}
