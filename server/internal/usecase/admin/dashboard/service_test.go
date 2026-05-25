package dashboard

import (
	"context"
	"testing"

	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
)

func TestBusinessMetricsAggregateCurrentOperationalBacklog(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db,
		dashboardOrdersSchema,
		dashboardInstancesSchema,
		dashboardAsyncTasksSchema,
		dashboardTicketsSchema,
		dashboardInvoiceApplicationsSchema,
		dashboardPaymentTransactionsSchema,
		dashboardRefundTransactionsSchema,
	)

	mysqltest.Exec(t, db,
		`INSERT INTO orders (status) VALUES ('pending'), ('pending'), ('error'), ('fulfilled')`,
		`INSERT INTO instances (status) VALUES ('error'), ('error'), ('running')`,
		`INSERT INTO async_tasks (status) VALUES ('failed'), ('pending')`,
		`INSERT INTO tickets (status) VALUES ('waiting_admin'), ('waiting_admin'), ('waiting_user'), ('closed')`,
		`INSERT INTO invoice_applications (status) VALUES ('pending'), ('processing'), ('issued'), ('rejected')`,
		`INSERT INTO payment_transactions (status) VALUES ('failed'), ('pending'), ('paid')`,
		`INSERT INTO refund_transactions (status) VALUES ('pending'), ('failed'), ('failed'), ('succeeded')`,
	)

	service := NewAdminDashboardService(db)
	metrics, err := service.businessMetrics(context.Background())
	if err != nil {
		t.Fatalf("aggregate business metrics: %v", err)
	}

	values := map[string]int64{}
	targets := map[string]string{}
	for _, metric := range metrics {
		values[metric.Key] = metric.Value
		if metric.TargetPermission != nil {
			targets[metric.Key] = *metric.TargetPermission
		}
	}

	want := map[string]int64{
		"pending_orders":     2,
		"order_errors":       1,
		"instance_errors":    2,
		"failed_async_tasks": 1,
		"pending_tickets":    2,
		"invoice_todo":       2,
		"payment_exceptions": 4,
	}
	for key, expected := range want {
		if values[key] != expected {
			t.Fatalf("metric %s mismatch: got %d want %d; all=%#v", key, values[key], expected, values)
		}
	}
	if targets["pending_orders"] != "page.orders" || targets["payment_exceptions"] != "page.payments" {
		t.Fatalf("business metrics should carry target page permissions, got %#v", targets)
	}
}

const dashboardOrdersSchema = `
CREATE TABLE orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  status VARCHAR(32) NOT NULL,
  KEY idx_dashboard_orders_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const dashboardInstancesSchema = `
CREATE TABLE instances (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  status VARCHAR(32) NOT NULL,
  KEY idx_dashboard_instances_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const dashboardAsyncTasksSchema = `
CREATE TABLE async_tasks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  status VARCHAR(32) NOT NULL,
  KEY idx_dashboard_async_tasks_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const dashboardTicketsSchema = `
CREATE TABLE tickets (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  status VARCHAR(32) NOT NULL,
  KEY idx_dashboard_tickets_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const dashboardInvoiceApplicationsSchema = `
CREATE TABLE invoice_applications (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  status VARCHAR(32) NOT NULL,
  KEY idx_dashboard_invoice_applications_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const dashboardPaymentTransactionsSchema = `
CREATE TABLE payment_transactions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  status VARCHAR(32) NOT NULL,
  KEY idx_dashboard_payment_transactions_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const dashboardRefundTransactionsSchema = `
CREATE TABLE refund_transactions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  status VARCHAR(32) NOT NULL,
  KEY idx_dashboard_refund_transactions_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
