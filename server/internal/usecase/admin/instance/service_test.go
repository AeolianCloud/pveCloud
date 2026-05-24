package instance

import (
	"context"
	"testing"
	"time"

	domaininstance "github.com/AeolianCloud/pveCloud/server/internal/domain/instance"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	mysqlinstance "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/instance"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
)

func TestReleaseCompletionUpdatesMarksExpiryOnlyForWorkerRelease(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)

	workerUpdates := releaseCompletionUpdates(mysqlinstance.Operation{Action: domaininstance.OperationRelease}, now)
	if workerUpdates["expire_released_at"] != now {
		t.Fatalf("worker release should mark expire_released_at, got %#v", workerUpdates["expire_released_at"])
	}

	adminID := uint64(7)
	adminUpdates := releaseCompletionUpdates(mysqlinstance.Operation{Action: domaininstance.OperationRelease, AdminID: &adminID}, now)
	if _, ok := adminUpdates["expire_released_at"]; ok {
		t.Fatalf("admin release must not mark expire_released_at: %#v", adminUpdates)
	}
}

func TestUpdateExpiresAtReschedulesLifecycleTasksAndWritesAudit(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, instanceUsersSchema, instanceOrdersSchema, instanceInstancesSchema, instanceOperationsSchema, instanceAsyncTasksSchema, instanceAdminAuditLogsSchema)

	instanceNo := "INS-expires-1"
	oldExpiresAt := time.Now().AddDate(0, 1, 0).Truncate(time.Millisecond)
	oldNoticeAt := oldExpiresAt.Add(-24 * time.Hour)
	oldReleaseAt := oldExpiresAt.Add(2 * time.Hour)
	nextExpiresAt := time.Now().AddDate(0, 2, 0).Truncate(time.Millisecond)

	if err := db.Exec(`INSERT INTO users (id, username, email, password_hash, status) VALUES (?, ?, ?, ?, ?)`, 21, "instance-user", "instance@example.com", "hash", "active").Error; err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if err := db.Exec(`INSERT INTO orders (id, order_no, user_id, client_token, status, order_type, related_instance_no) VALUES (?, ?, ?, ?, ?, ?, ?)`, 31, "ORD-renewal-seed", 21, "renewal-seed-token", "fulfilled", "renewal", instanceNo).Error; err != nil {
		t.Fatalf("insert renewal order seed: %v", err)
	}
	if err := db.Exec(`
INSERT INTO instances (
  id, instance_no, user_id, order_id, order_no, status, product_no, product_name,
  plan_no, plan_name, cpu_cores, memory_mb, system_disk_gb, data_disk_gb,
  bandwidth_mbps, region_no, region_name, network_type_no, network_type_name,
  template_no, template_name, os_family, os_distribution, os_version,
  external_node, external_vmid, expires_at, expire_notice_sent_at, expire_release_scheduled_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		41, instanceNo, 21, 31, "ORD-purchase-2", domaininstance.StatusRunning, "PROD-1", "Server",
		"PLAN-1", "Basic", 2, 4096, 40, 0, 100, "REG-1", "China", "NET-1", "Classic",
		"TPL-1", "Ubuntu", "linux", "ubuntu", "22.04", "node-a", 1002, oldExpiresAt, oldNoticeAt, oldReleaseAt,
	).Error; err != nil {
		t.Fatalf("insert instance: %v", err)
	}

	remark := "延长服务期"
	service := NewService(db, nil, nil, config.InstanceLifecycleConfig{AutoReleaseEnabled: true, ExpireNoticeBeforeSeconds: 86400, ExpireReleaseAfterSeconds: 7200})
	detail, err := service.UpdateExpiresAt(context.Background(), 77, instanceNo, admindto.InstanceExpiresAtRequest{ExpiresAt: nextExpiresAt, Remark: &remark})
	if err != nil {
		t.Fatalf("update expires at: %v", err)
	}
	if detail.ExpiresAt == nil || !detail.ExpiresAt.Equal(nextExpiresAt) {
		t.Fatalf("detail should return updated expires_at, got %#v want %s", detail.ExpiresAt, nextExpiresAt)
	}

	var instance struct {
		ExpiresAt                time.Time  `gorm:"column:expires_at"`
		ExpireNoticeSentAt       *time.Time `gorm:"column:expire_notice_sent_at"`
		ExpireReleaseScheduledAt time.Time  `gorm:"column:expire_release_scheduled_at"`
	}
	if err := db.Table("instances").Select("expires_at, expire_notice_sent_at, expire_release_scheduled_at").Where("instance_no = ?", instanceNo).Take(&instance).Error; err != nil {
		t.Fatalf("load updated instance: %v", err)
	}
	if !instance.ExpiresAt.Equal(nextExpiresAt) {
		t.Fatalf("expires_at should update, got %s want %s", instance.ExpiresAt, nextExpiresAt)
	}
	if instance.ExpireNoticeSentAt != nil {
		t.Fatalf("manual expiry update should clear old notice marker, got %s", *instance.ExpireNoticeSentAt)
	}
	if wantRelease := nextExpiresAt.Add(2 * time.Hour); !instance.ExpireReleaseScheduledAt.Equal(wantRelease) {
		t.Fatalf("manual expiry update should reschedule release, got %s want %s", instance.ExpireReleaseScheduledAt, wantRelease)
	}

	var taskCounts []struct {
		TaskType string
		Count    int64
	}
	if err := db.Table("async_tasks").Select("task_type, COUNT(*) AS count").Group("task_type").Scan(&taskCounts).Error; err != nil {
		t.Fatalf("count lifecycle tasks: %v", err)
	}
	counts := map[string]int64{}
	for _, row := range taskCounts {
		counts[row.TaskType] = row.Count
	}
	if counts[domaininstance.TaskTypeExpiryNotice] != 1 || counts[domaininstance.TaskTypeExpiryRelease] != 1 {
		t.Fatalf("manual expiry update should enqueue notice and release tasks, got %#v", counts)
	}

	var auditCount int64
	if err := db.Table("admin_audit_logs").Where("admin_id = ? AND action = ? AND object_type = ? AND object_id = ?", 77, "instance.expires_at.update", "instance", instanceNo).Count(&auditCount).Error; err != nil {
		t.Fatalf("count audit logs: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("manual expiry update should write exactly one audit log, got %d", auditCount)
	}
}

const instanceUsersSchema = `
CREATE TABLE users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(64) NOT NULL,
  email VARCHAR(191) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  display_name VARCHAR(64) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  UNIQUE KEY uk_users_username (username),
  UNIQUE KEY uk_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const instanceOrdersSchema = `
CREATE TABLE orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  order_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  client_token VARCHAR(128) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  order_type VARCHAR(32) NOT NULL DEFAULT 'purchase',
  related_instance_no VARCHAR(64) NULL,
  product_no VARCHAR(64) NOT NULL DEFAULT '',
  product_type VARCHAR(32) NOT NULL DEFAULT 'server',
  product_name VARCHAR(128) NOT NULL DEFAULT '',
  plan_no VARCHAR(64) NOT NULL DEFAULT '',
  plan_code VARCHAR(64) NOT NULL DEFAULT '',
  plan_name VARCHAR(128) NOT NULL DEFAULT '',
  cpu_cores INT NOT NULL DEFAULT 0,
  memory_mb INT NOT NULL DEFAULT 0,
  system_disk_gb INT NOT NULL DEFAULT 0,
  data_disk_gb INT NOT NULL DEFAULT 0,
  bandwidth_mbps INT NOT NULL DEFAULT 0,
  public_ip_count INT NOT NULL DEFAULT 1,
  virtualization VARCHAR(32) NOT NULL DEFAULT '',
  architecture VARCHAR(32) NOT NULL DEFAULT '',
  billing_cycle VARCHAR(32) NOT NULL DEFAULT 'monthly',
  price_cents BIGINT UNSIGNED NOT NULL DEFAULT 0,
  currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
  quantity INT NOT NULL DEFAULT 1,
  total_amount_cents BIGINT UNSIGNED NOT NULL DEFAULT 0,
  payment_status VARCHAR(32) NOT NULL DEFAULT 'unpaid',
  paid_at DATETIME(3) NULL,
  region_no VARCHAR(64) NOT NULL DEFAULT '',
  region_code VARCHAR(64) NOT NULL DEFAULT '',
  region_name VARCHAR(128) NOT NULL DEFAULT '',
  network_type_no VARCHAR(64) NOT NULL DEFAULT '',
  network_type_code VARCHAR(64) NOT NULL DEFAULT '',
  network_type_name VARCHAR(128) NOT NULL DEFAULT '',
  template_no VARCHAR(64) NOT NULL DEFAULT '',
  template_code VARCHAR(96) NOT NULL DEFAULT '',
  template_name VARCHAR(128) NOT NULL DEFAULT '',
  os_family VARCHAR(32) NOT NULL DEFAULT '',
  os_distribution VARCHAR(64) NOT NULL DEFAULT '',
  os_version VARCHAR(64) NOT NULL DEFAULT '',
  os_architecture VARCHAR(32) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_orders_order_no (order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const instanceInstancesSchema = `
CREATE TABLE instances (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  instance_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  order_id BIGINT UNSIGNED NOT NULL,
  order_no VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  product_no VARCHAR(64) NOT NULL,
  product_name VARCHAR(128) NOT NULL,
  plan_no VARCHAR(64) NOT NULL,
  plan_name VARCHAR(128) NOT NULL,
  cpu_cores INT NOT NULL,
  memory_mb INT NOT NULL,
  system_disk_gb INT NOT NULL,
  data_disk_gb INT NOT NULL DEFAULT 0,
  bandwidth_mbps INT NOT NULL,
  region_no VARCHAR(64) NOT NULL,
  region_name VARCHAR(128) NOT NULL,
  network_type_no VARCHAR(64) NULL,
  network_type_name VARCHAR(128) NULL,
  template_no VARCHAR(64) NOT NULL,
  template_name VARCHAR(128) NOT NULL,
  os_family VARCHAR(32) NOT NULL,
  os_distribution VARCHAR(64) NOT NULL,
  os_version VARCHAR(64) NOT NULL,
  external_node VARCHAR(128) NOT NULL,
  external_vmid INT UNSIGNED NOT NULL,
  external_resource_location VARCHAR(255) NULL,
  last_error_code VARCHAR(64) NULL,
  last_error_message VARCHAR(500) NULL,
  service_started_at DATETIME(3) NULL,
  expires_at DATETIME(3) NULL,
  expire_notice_sent_at DATETIME(3) NULL,
  expire_release_scheduled_at DATETIME(3) NULL,
  expire_released_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  released_at DATETIME(3) NULL,
  UNIQUE KEY uk_instances_instance_no (instance_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const instanceOperationsSchema = `
CREATE TABLE instance_operations (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  operation_no VARCHAR(64) NOT NULL,
  instance_id BIGINT UNSIGNED NOT NULL,
  order_id BIGINT UNSIGNED NULL,
  admin_id BIGINT UNSIGNED NULL,
  user_id BIGINT UNSIGNED NULL,
  action VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL,
  external_operation_id VARCHAR(128) NULL,
  operation_location VARCHAR(255) NULL,
  resource_location VARCHAR(255) NULL,
  error_code VARCHAR(64) NULL,
  error_message VARCHAR(500) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  completed_at DATETIME(3) NULL,
  UNIQUE KEY uk_instance_operations_operation_no (operation_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const instanceAsyncTasksSchema = `
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

const instanceAdminAuditLogsSchema = `
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
