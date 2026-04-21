package database_test

import (
	"os"
	"strings"
	"testing"
)

func TestMigrationsContainChineseComments(t *testing.T) {
	core := loadMigrationFile(t, "../../../migrations/0001_core_auth.sql")
	if !strings.Contains(core, "COMMENT='前台用户主表'") {
		t.Fatalf("expected Chinese table comment for users")
	}

	orders := loadMigrationFile(t, "../../../migrations/0003_orders_payments.sql")
	if !strings.Contains(orders, "COMMENT='订单主表'") {
		t.Fatalf("expected Chinese table comment for orders")
	}
	if !strings.Contains(orders, "订单状态：pending_payment-待支付，paid-已支付，provisioning-开通中，active-已生效，failed-失败，closed-已关闭") {
		t.Fatalf("expected explicit order status comment")
	}

	tasks := loadMigrationFile(t, "../../../migrations/0004_instances_tasks.sql")
	if !strings.Contains(tasks, "COMMENT='异步任务主表'") {
		t.Fatalf("expected Chinese table comment for async_tasks")
	}
}

func loadMigrationFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}

	return string(data)
}
