package instance

import (
	"context"
	"sync"
	"testing"
	"time"

	domaininstance "github.com/AeolianCloud/pveCloud/server/internal/domain/instance"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
	"gorm.io/gorm"
)

func TestClaimTasksDoesNotDoubleClaimConcurrentWorkers(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, asyncTasksSchema)

	repo := NewRepository(db)
	now := time.Now().Add(-time.Minute).Truncate(time.Millisecond)
	if err := repo.CreateTask(context.Background(), nil, &Task{
		TaskNo:      "TASK-claim-once",
		TaskType:    domaininstance.TaskTypeOperationSync,
		Status:      domaininstance.TaskStatusPending,
		MaxAttempts: 3,
		ScheduledAt: now,
	}); err != nil {
		t.Fatalf("create task: %v", err)
	}

	start := make(chan struct{})
	results := make(chan int, 2)
	var wg sync.WaitGroup
	for _, workerID := range []string{"worker-a", "worker-b"} {
		wg.Add(1)
		go func(workerID string) {
			defer wg.Done()
			<-start
			err := db.Transaction(func(tx *gorm.DB) error {
				rows, err := repo.ClaimTasks(context.Background(), tx, workerID, 1, time.Now().Add(time.Minute))
				if err != nil {
					return err
				}
				// Hold the transaction briefly so the second worker must contend
				// with the row lock instead of only observing already-committed state.
				time.Sleep(150 * time.Millisecond)
				results <- len(rows)
				return nil
			})
			if err != nil {
				t.Errorf("claim task for %s: %v", workerID, err)
				results <- -1
			}
		}(workerID)
	}
	close(start)
	wg.Wait()
	close(results)

	claimed := 0
	for count := range results {
		if count < 0 {
			t.Fatal("claim failed")
		}
		claimed += count
	}
	if claimed != 1 {
		t.Fatalf("task should be claimed exactly once, got %d claims", claimed)
	}

	task, err := repo.TaskByNo(context.Background(), "TASK-claim-once")
	if err != nil {
		t.Fatalf("load task: %v", err)
	}
	if task.Status != domaininstance.TaskStatusRunning || task.Attempts != 1 || task.LockedBy == nil {
		t.Fatalf("claimed task state mismatch: status=%s attempts=%d locked_by=%v", task.Status, task.Attempts, task.LockedBy)
	}
}

const asyncTasksSchema = `
CREATE TABLE async_tasks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  task_no VARCHAR(64) NOT NULL,
  task_type VARCHAR(64) NOT NULL,
  idempotency_key VARCHAR(191) NULL,
  status VARCHAR(32) NOT NULL,
  object_type VARCHAR(64) NULL,
  object_no VARCHAR(64) NULL,
  payload TEXT NULL,
  result TEXT NULL,
  attempts INT NOT NULL DEFAULT 0,
  max_attempts INT NOT NULL DEFAULT 3,
  scheduled_at DATETIME(3) NOT NULL,
  locked_by VARCHAR(128) NULL,
  locked_until DATETIME(3) NULL,
  last_error_code VARCHAR(64) NULL,
  last_error_message VARCHAR(500) NULL,
  completed_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  active_idempotency_key VARCHAR(191) GENERATED ALWAYS AS (IF(status <> 'cancelled', idempotency_key, NULL)) STORED,
  UNIQUE KEY uk_async_tasks_task_no (task_no),
  UNIQUE KEY uk_async_tasks_active_idempotency (task_type, active_idempotency_key),
  KEY idx_async_tasks_status_schedule (status, scheduled_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
