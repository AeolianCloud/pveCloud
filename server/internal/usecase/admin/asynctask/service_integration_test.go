package asynctask

import (
	"context"
	"testing"
	"time"

	domaininstance "github.com/AeolianCloud/pveCloud/server/internal/domain/instance"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
)

func TestRetryFailedTaskResetsLockAndWritesAudit(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, asyncTasksSchema, adminAuditLogsSchema)

	now := time.Now().Truncate(time.Millisecond)
	lockedUntil := now.Add(time.Hour)
	completedAt := now.Add(-time.Minute)
	if err := db.Exec(`
INSERT INTO async_tasks (
  task_no, task_type, status, attempts, max_attempts, scheduled_at, locked_by,
  locked_until, last_error_code, last_error_message, completed_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"TASK-retry-1", domaininstance.TaskTypeExpiryRelease, domaininstance.TaskStatusFailed,
		3, 3, now.Add(-time.Hour), "worker-a", lockedUntil, "task_failed", "boom", completedAt,
	).Error; err != nil {
		t.Fatalf("insert failed task: %v", err)
	}

	remark := "人工确认后重试"
	item, err := NewService(db, nil).Retry(context.Background(), 42, "TASK-retry-1", admindto.AsyncTaskRetryRequest{Remark: &remark})
	if err != nil {
		t.Fatalf("retry failed task: %v", err)
	}
	if item.Status != domaininstance.TaskStatusPending || item.LockedBy != nil || item.LockedUntil != nil {
		t.Fatalf("retry should reset status and lock fields: %#v", item)
	}
	if item.LastErrorCode != nil || item.LastErrorMessage != nil || item.CompletedAt != nil {
		t.Fatalf("retry should clear failure fields: %#v", item)
	}

	var auditCount int64
	if err := db.Table("admin_audit_logs").
		Where("admin_id = ? AND action = ? AND object_type = ? AND object_id = ?", 42, "async_task.retry", "async_task", "TASK-retry-1").
		Count(&auditCount).Error; err != nil {
		t.Fatalf("count audit logs: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("retry should write exactly one audit log, got %d", auditCount)
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
  UNIQUE KEY uk_async_tasks_active_idempotency (task_type, active_idempotency_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const adminAuditLogsSchema = `
CREATE TABLE admin_audit_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  admin_id BIGINT UNSIGNED NULL,
  admin_username VARCHAR(64) NULL,
  admin_display_name VARCHAR(64) NULL,
  session_id VARCHAR(64) NULL,
  request_id VARCHAR(64) NULL,
  request_method VARCHAR(16) NULL,
  request_path VARCHAR(255) NULL,
  action VARCHAR(128) NOT NULL,
  object_type VARCHAR(64) NOT NULL,
  object_id VARCHAR(128) NULL,
  before_data JSON NULL,
  after_data JSON NULL,
  ip VARCHAR(64) NULL,
  user_agent VARCHAR(500) NULL,
  remark VARCHAR(500) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  KEY idx_admin_audit_logs_action_created (action, created_at),
  KEY idx_admin_audit_logs_object (object_type, object_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
