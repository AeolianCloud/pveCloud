package order

import (
	"context"
	"testing"
	"time"

	domaininstance "github.com/AeolianCloud/pveCloud/server/internal/domain/instance"
	domainorder "github.com/AeolianCloud/pveCloud/server/internal/domain/order"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
)

func TestRenewalExpiresAtExtendsFromCurrentExpiryWhenActive(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 0, 0, 123456789, time.UTC)
	current := time.Date(2026, 6, 1, 8, 30, 0, 987654321, time.UTC)

	got := renewalExpiresAt(now, &current, 3)
	want := time.Date(2026, 9, 1, 8, 30, 0, 987000000, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("active instance should renew from current expiry, got %s want %s", got, want)
	}
}

func TestRenewalExpiresAtExtendsFromNowWhenExpired(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 0, 0, 123456789, time.UTC)
	current := time.Date(2026, 5, 1, 8, 30, 0, 0, time.UTC)

	got := renewalExpiresAt(now, &current, 1)
	want := time.Date(2026, 6, 23, 12, 0, 0, 123000000, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("expired instance should renew from now, got %s want %s", got, want)
	}
}

func TestRenewalInstanceUpdatesRespectAutoReleaseSwitch(t *testing.T) {
	nextExpiresAt := time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC)

	enabled := renewalInstanceUpdates(config.InstanceLifecycleConfig{AutoReleaseEnabled: true, ExpireReleaseAfterSeconds: 3600}, nextExpiresAt)
	wantReleaseAt := nextExpiresAt.Add(time.Hour)
	if got := enabled["expire_release_scheduled_at"]; got != wantReleaseAt {
		t.Fatalf("auto release enabled should schedule release, got %#v want %s", got, wantReleaseAt)
	}
	if enabled["expire_notice_sent_at"] != nil {
		t.Fatalf("renewal must clear old notice marker, got %#v", enabled["expire_notice_sent_at"])
	}

	disabled := renewalInstanceUpdates(config.InstanceLifecycleConfig{AutoReleaseEnabled: false, ExpireReleaseAfterSeconds: 3600}, nextExpiresAt)
	if disabled["expire_release_scheduled_at"] != nil {
		t.Fatalf("auto release disabled must clear release schedule, got %#v", disabled["expire_release_scheduled_at"])
	}
}

func TestConfirmRenewalPersistsOrderInstanceTasksAndAudit(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, renewalUsersSchema, renewalOrdersSchema, renewalInstancesSchema, renewalAsyncTasksSchema, renewalAdminAuditLogsSchema)

	currentExpiresAt := time.Now().AddDate(0, 1, 0).Truncate(time.Millisecond)
	expectedExpiresAt := currentExpiresAt.AddDate(0, 3, 0).Truncate(time.Millisecond)
	instanceNo := "INS-renewal-1"
	orderNo := "ORD-renewal-1"

	if err := db.Exec(`INSERT INTO users (id, username, email, password_hash, status) VALUES (?, ?, ?, ?, ?)`, 11, "renew-user", "renew@example.com", "hash", "active").Error; err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if err := db.Exec(`
INSERT INTO orders (
  id, order_no, user_id, client_token, status, order_type, related_instance_no,
  product_no, product_type, product_name, plan_no, plan_code, plan_name,
  cpu_cores, memory_mb, system_disk_gb, data_disk_gb, bandwidth_mbps,
  public_ip_count, virtualization, architecture, billing_cycle, price_cents,
  currency, quantity, total_amount_cents, payment_status, region_no,
  region_code, region_name, network_type_no, network_type_code, network_type_name,
  template_no, template_code, template_name, os_family, os_distribution,
  os_version, os_architecture
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		21, orderNo, 11, "renew-token", domainorder.StatusPending, domainorder.TypeRenewal, instanceNo,
		"PROD-1", "server", "Server", "PLAN-1", "basic", "Basic",
		2, 4096, 40, 0, 100, 1, "kvm", "x86_64", "quarterly", 3000,
		"CNY", 1, 3000, domainorder.PaymentStatusUnpaid, "REG-1",
		"cn", "China", "NET-1", "classic", "Classic",
		"TPL-1", "ubuntu", "Ubuntu", "linux", "ubuntu", "22.04", "x86_64",
	).Error; err != nil {
		t.Fatalf("insert renewal order: %v", err)
	}
	if err := db.Exec(`
INSERT INTO instances (
  id, instance_no, user_id, order_id, order_no, status, product_no, product_name,
  plan_no, plan_name, cpu_cores, memory_mb, system_disk_gb, data_disk_gb,
  bandwidth_mbps, region_no, region_name, network_type_no, network_type_name,
  template_no, template_name, os_family, os_distribution, os_version,
  external_node, external_vmid, expires_at, expire_notice_sent_at, expire_release_scheduled_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		31, instanceNo, 11, 20, "ORD-purchase-1", domaininstance.StatusRunning, "PROD-1", "Server",
		"PLAN-1", "Basic", 2, 4096, 40, 0, 100, "REG-1", "China", "NET-1", "Classic",
		"TPL-1", "Ubuntu", "linux", "ubuntu", "22.04", "node-a", 1001, currentExpiresAt, currentExpiresAt.Add(-24*time.Hour), currentExpiresAt.Add(time.Hour),
	).Error; err != nil {
		t.Fatalf("insert instance: %v", err)
	}

	remark := "确认续费到账"
	service := NewService(db, nil, config.InstanceLifecycleConfig{AutoReleaseEnabled: true, ExpireNoticeBeforeSeconds: 86400, ExpireReleaseAfterSeconds: 7200})
	detail, err := service.ConfirmRenewal(context.Background(), 99, orderNo, admindto.OrderRenewalConfirmRequest{Remark: &remark})
	if err != nil {
		t.Fatalf("confirm renewal: %v", err)
	}
	if detail.Status != domainorder.StatusFulfilled || detail.PaymentStatus != domainorder.PaymentStatusManualConfirmed || detail.PaidAt == nil {
		t.Fatalf("renewal order should be fulfilled and manually confirmed: %#v", detail.AdminOrderItem)
	}

	var instance struct {
		ExpiresAt                time.Time  `gorm:"column:expires_at"`
		ExpireNoticeSentAt       *time.Time `gorm:"column:expire_notice_sent_at"`
		ExpireReleaseScheduledAt time.Time  `gorm:"column:expire_release_scheduled_at"`
	}
	if err := db.Table("instances").Select("expires_at, expire_notice_sent_at, expire_release_scheduled_at").Where("instance_no = ?", instanceNo).Take(&instance).Error; err != nil {
		t.Fatalf("load instance after renewal: %v", err)
	}
	if !instance.ExpiresAt.Equal(expectedExpiresAt) {
		t.Fatalf("instance expiry should extend from current expiry, got %s want %s", instance.ExpiresAt, expectedExpiresAt)
	}
	if instance.ExpireNoticeSentAt != nil {
		t.Fatalf("renewal should clear old notice marker, got %s", *instance.ExpireNoticeSentAt)
	}
	if wantRelease := expectedExpiresAt.Add(2 * time.Hour); !instance.ExpireReleaseScheduledAt.Equal(wantRelease) {
		t.Fatalf("renewal should reschedule auto release, got %s want %s", instance.ExpireReleaseScheduledAt, wantRelease)
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
		t.Fatalf("renewal should enqueue notice and release tasks, got %#v", counts)
	}

	var auditCount int64
	if err := db.Table("admin_audit_logs").Where("admin_id = ? AND action = ? AND object_type = ? AND object_id = ?", 99, "order.renewal.confirm", "order", orderNo).Count(&auditCount).Error; err != nil {
		t.Fatalf("count audit logs: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("confirm renewal should write exactly one audit log, got %d", auditCount)
	}
}

const renewalUsersSchema = `
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

const renewalOrdersSchema = `
CREATE TABLE orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  order_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  client_token VARCHAR(128) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  order_type VARCHAR(32) NOT NULL DEFAULT 'purchase',
  related_instance_no VARCHAR(64) NULL,
  product_no VARCHAR(64) NOT NULL,
  product_type VARCHAR(32) NOT NULL,
  product_name VARCHAR(128) NOT NULL,
  product_summary VARCHAR(255) NULL,
  plan_no VARCHAR(64) NOT NULL,
  plan_code VARCHAR(64) NOT NULL,
  plan_name VARCHAR(128) NOT NULL,
  plan_summary VARCHAR(255) NULL,
  cpu_cores INT NOT NULL,
  memory_mb INT NOT NULL,
  system_disk_gb INT NOT NULL,
  data_disk_gb INT NOT NULL DEFAULT 0,
  bandwidth_mbps INT NOT NULL,
  traffic_gb INT NULL,
  public_ip_count INT NOT NULL DEFAULT 1,
  virtualization VARCHAR(32) NOT NULL,
  architecture VARCHAR(32) NOT NULL,
  billing_cycle VARCHAR(32) NOT NULL,
  price_cents BIGINT UNSIGNED NOT NULL,
  original_price_cents BIGINT UNSIGNED NULL,
  currency VARCHAR(16) NOT NULL,
  quantity INT NOT NULL DEFAULT 1,
  total_amount_cents BIGINT UNSIGNED NOT NULL,
  payment_status VARCHAR(32) NOT NULL DEFAULT 'unpaid',
  paid_at DATETIME(3) NULL,
  payment_provider VARCHAR(32) NULL,
  payment_trade_no VARCHAR(128) NULL,
  payment_callback_payload TEXT NULL,
  region_no VARCHAR(64) NOT NULL,
  region_code VARCHAR(64) NOT NULL,
  region_name VARCHAR(128) NOT NULL,
  network_type_no VARCHAR(64) NOT NULL DEFAULT '',
  network_type_code VARCHAR(64) NOT NULL DEFAULT '',
  network_type_name VARCHAR(128) NOT NULL DEFAULT '',
  template_no VARCHAR(64) NOT NULL,
  template_code VARCHAR(96) NOT NULL,
  template_name VARCHAR(128) NOT NULL,
  os_family VARCHAR(32) NOT NULL,
  os_distribution VARCHAR(64) NOT NULL,
  os_version VARCHAR(64) NOT NULL,
  os_architecture VARCHAR(32) NOT NULL,
  user_note VARCHAR(500) NULL,
  admin_note VARCHAR(1000) NULL,
  cancel_reason VARCHAR(500) NULL,
  closed_reason VARCHAR(500) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  cancelled_at DATETIME(3) NULL,
  closed_at DATETIME(3) NULL,
  UNIQUE KEY uk_orders_order_no (order_no),
  UNIQUE KEY uk_orders_user_client_token (user_id, client_token)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const renewalInstancesSchema = `
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

const renewalAsyncTasksSchema = `
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

const renewalAdminAuditLogsSchema = `
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
