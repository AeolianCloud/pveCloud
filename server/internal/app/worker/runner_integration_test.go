package worker

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	domaininstance "github.com/AeolianCloud/pveCloud/server/internal/domain/instance"
	domainorder "github.com/AeolianCloud/pveCloud/server/internal/domain/order"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	mysqlinstance "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/instance"
	mysqlorder "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/order"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
)

func TestExpiryReleaseSkipsStaleLifecycleTask(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, instancesSchema)

	oldExpiresAt := time.Now().Add(-2 * time.Hour).Truncate(time.Millisecond)
	currentExpiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Millisecond)
	if err := db.Exec(`
INSERT INTO instances (
  instance_no, user_id, order_id, order_no, status, product_no, product_name,
  plan_no, plan_name, cpu_cores, memory_mb, system_disk_gb, data_disk_gb,
  bandwidth_mbps, region_no, region_name, template_no, template_name,
  os_family, os_distribution, os_version, external_node, external_vmid,
  expires_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"INS-stale-release", 1, 10, "ORD-1", domaininstance.StatusRunning,
		"PROD-1", "Product", "PLAN-1", "Plan", 2, 2048, 40, 0, 10,
		"REG-1", "Region", "TPL-1", "Template", "linux", "debian", "12",
		"node-a", 1001, currentExpiresAt, time.Now(), time.Now(),
	).Error; err != nil {
		t.Fatalf("insert instance: %v", err)
	}

	payload := fmt.Sprintf(`{"instance_no":"INS-stale-release","expires_at":%q}`, oldExpiresAt.Format(time.RFC3339Nano))
	objectType := "instance"
	objectNo := "INS-stale-release"
	runner := &Runner{
		db:           db,
		tasks:        mysqlinstance.NewRepository(db),
		lifecycleCfg: config.InstanceLifecycleConfig{AutoReleaseEnabled: true},
	}
	err := runner.expiryRelease(context.Background(), mysqlinstance.Task{
		TaskNo:     "TASK-stale-release",
		TaskType:   domaininstance.TaskTypeExpiryRelease,
		Status:     domaininstance.TaskStatusRunning,
		ObjectType: &objectType,
		ObjectNo:   &objectNo,
		Payload:    &payload,
	})
	if err != nil {
		t.Fatalf("stale expiry release should be skipped without error: %v", err)
	}

	var status string
	var releasedAt *time.Time
	if err := db.Table("instances").Select("status", "released_at").Where("instance_no = ?", "INS-stale-release").Row().Scan(&status, &releasedAt); err != nil {
		t.Fatalf("load instance state: %v", err)
	}
	if status != domaininstance.StatusRunning || releasedAt != nil {
		t.Fatalf("stale release task must not release instance, status=%s released_at=%v", status, releasedAt)
	}
}

func TestPaymentProvisionTaskMarksOrderErrorAfterMaxAttempts(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, ordersSchema)
	mysqltest.Exec(t, db, instancesSchema)
	mysqltest.Exec(t, db, asyncTasksSchema)

	now := time.Now().Truncate(time.Millisecond)
	if err := db.Exec(`
INSERT INTO orders (
  order_no, user_id, client_token, status, order_type, product_no, product_type,
  product_name, plan_no, plan_code, plan_name, cpu_cores, memory_mb,
  system_disk_gb, data_disk_gb, bandwidth_mbps, public_ip_count,
  virtualization, architecture, billing_cycle, price_cents, currency, quantity,
  total_amount_cents, payment_status, region_no, region_code, region_name,
  network_type_no, network_type_code, network_type_name, template_no, template_code,
  template_name, os_family, os_distribution, os_version, os_architecture,
  created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"ORD-payment-error", 1, "token-payment-error", domainorder.StatusPending, domainorder.TypePurchase,
		"PROD-1", "server", "Product", "PLAN-1", "basic", "Plan", 2, 2048, 40, 0, 10, 1,
		"kvm", "x86_64", "monthly", 1000, "CNY", 1, 1000, domainorder.PaymentStatusPaid,
		"REG-1", "default", "Region", "NET-1", "default", "Default", "TPL-1", "debian-12",
		"Debian 12", "linux", "debian", "12", "x86_64", now, now,
	).Error; err != nil {
		t.Fatalf("insert order: %v", err)
	}
	objectType := "order"
	objectNo := "ORD-payment-error"
	if err := db.Exec(`
INSERT INTO async_tasks (
  task_no, task_type, idempotency_key, status, object_type, object_no,
  attempts, max_attempts, scheduled_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"TASK-payment-error", domaininstance.TaskTypePaymentProvision, "payment_order_provision:PAY-1",
		domaininstance.TaskStatusRunning, objectType, objectNo, 10, 10, now, now, now,
	).Error; err != nil {
		t.Fatalf("insert task: %v", err)
	}

	runner := &Runner{
		db:     db,
		tasks:  mysqlinstance.NewRepository(db),
		orders: mysqlorder.NewRepository(db),
	}
	task, err := runner.tasks.TaskByNo(context.Background(), "TASK-payment-error")
	if err != nil {
		t.Fatalf("load task: %v", err)
	}
	if err := runner.markFailedOrRetry(context.Background(), task, errors.New("mcp unavailable")); err != nil {
		t.Fatalf("mark payment provision failed: %v", err)
	}

	var orderStatus string
	if err := db.Table("orders").Select("status").Where("order_no = ?", "ORD-payment-error").Row().Scan(&orderStatus); err != nil {
		t.Fatalf("load order status: %v", err)
	}
	if orderStatus != domainorder.StatusError {
		t.Fatalf("paid purchase order should move to error after exhausted automatic provision, got %s", orderStatus)
	}

	updatedTask, err := runner.tasks.TaskByNo(context.Background(), "TASK-payment-error")
	if err != nil {
		t.Fatalf("load updated task: %v", err)
	}
	if updatedTask.Status != domaininstance.TaskStatusFailed {
		t.Fatalf("task status = %s, want failed", updatedTask.Status)
	}
	if updatedTask.LastErrorMessage == nil || *updatedTask.LastErrorMessage == "" {
		t.Fatal("task should keep a sanitized failure summary")
	}
}

const instancesSchema = `
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
  data_disk_gb INT NOT NULL,
  bandwidth_mbps INT NOT NULL,
  region_no VARCHAR(64) NOT NULL,
  region_name VARCHAR(128) NOT NULL,
  network_type_no VARCHAR(64) NULL,
  network_type_name VARCHAR(128) NULL,
  template_no VARCHAR(64) NOT NULL,
  template_name VARCHAR(128) NOT NULL,
  os_family VARCHAR(64) NOT NULL,
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
  UNIQUE KEY uk_instances_instance_no (instance_no),
  UNIQUE KEY uk_instances_order_id (order_id),
  UNIQUE KEY uk_instances_external_vm (external_node, external_vmid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const ordersSchema = `
CREATE TABLE orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  order_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  client_token VARCHAR(128) NOT NULL,
  status VARCHAR(32) NOT NULL,
  order_type VARCHAR(32) NOT NULL DEFAULT 'purchase',
  related_instance_no VARCHAR(64) NULL,
  product_no VARCHAR(64) NOT NULL,
  product_type VARCHAR(32) NOT NULL,
  product_name VARCHAR(128) NOT NULL,
  product_summary VARCHAR(500) NULL,
  plan_no VARCHAR(64) NOT NULL,
  plan_code VARCHAR(64) NOT NULL,
  plan_name VARCHAR(128) NOT NULL,
  plan_summary VARCHAR(500) NULL,
  cpu_cores INT NOT NULL,
  memory_mb INT NOT NULL,
  system_disk_gb INT NOT NULL,
  data_disk_gb INT NOT NULL,
  bandwidth_mbps INT NOT NULL,
  traffic_gb INT NULL,
  public_ip_count INT NOT NULL DEFAULT 0,
  virtualization VARCHAR(32) NOT NULL,
  architecture VARCHAR(32) NOT NULL,
  billing_cycle VARCHAR(32) NOT NULL,
  price_cents BIGINT UNSIGNED NOT NULL,
  original_price_cents BIGINT UNSIGNED NULL,
  currency CHAR(3) NOT NULL,
  quantity INT NOT NULL DEFAULT 1,
  total_amount_cents BIGINT UNSIGNED NOT NULL,
  payment_status VARCHAR(32) NOT NULL DEFAULT 'unpaid',
  paid_at DATETIME(3) NULL,
  payment_provider VARCHAR(32) NULL,
  payment_trade_no VARCHAR(128) NULL,
  payment_callback_payload JSON NULL,
  region_no VARCHAR(64) NOT NULL,
  region_code VARCHAR(64) NOT NULL,
  region_name VARCHAR(128) NOT NULL,
  network_type_no VARCHAR(64) NOT NULL,
  network_type_code VARCHAR(64) NOT NULL,
  network_type_name VARCHAR(128) NOT NULL,
  template_no VARCHAR(64) NOT NULL,
  template_code VARCHAR(64) NOT NULL,
  template_name VARCHAR(128) NOT NULL,
  os_family VARCHAR(64) NOT NULL,
  os_distribution VARCHAR(64) NOT NULL,
  os_version VARCHAR(64) NOT NULL,
  os_architecture VARCHAR(64) NOT NULL,
  user_note VARCHAR(500) NULL,
  admin_note VARCHAR(500) NULL,
  cancel_reason VARCHAR(500) NULL,
  closed_reason VARCHAR(500) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  cancelled_at DATETIME(3) NULL,
  closed_at DATETIME(3) NULL,
  UNIQUE KEY uk_orders_order_no (order_no),
  UNIQUE KEY uk_orders_user_client_token (user_id, client_token)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const asyncTasksSchema = `
CREATE TABLE async_tasks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  task_no VARCHAR(64) NOT NULL,
  task_type VARCHAR(64) NOT NULL,
  idempotency_key VARCHAR(191) NULL,
  status VARCHAR(32) NOT NULL,
  object_type VARCHAR(64) NULL,
  object_no VARCHAR(64) NULL,
  payload JSON NULL,
  result JSON NULL,
  attempts INT NOT NULL DEFAULT 0,
  max_attempts INT NOT NULL DEFAULT 10,
  scheduled_at DATETIME(3) NOT NULL,
  locked_by VARCHAR(128) NULL,
  locked_until DATETIME(3) NULL,
  last_error_code VARCHAR(64) NULL,
  last_error_message VARCHAR(500) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  completed_at DATETIME(3) NULL,
  UNIQUE KEY uk_async_tasks_task_no (task_no),
  UNIQUE KEY uk_async_tasks_idempotency_key (idempotency_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
