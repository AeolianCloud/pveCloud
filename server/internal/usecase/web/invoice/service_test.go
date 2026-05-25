package invoice

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	domaininvoice "github.com/AeolianCloud/pveCloud/server/internal/domain/invoice"
	domainorder "github.com/AeolianCloud/pveCloud/server/internal/domain/order"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
)

func TestCreateInvoiceIsIdempotentAndCancelReleasesOrder(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, webInvoiceUsersSchema, webInvoiceOrdersSchema, webInvoiceRefundsSchema, webInvoiceApplicationsSchema, webInvoiceApplicationOrdersSchema)
	seedWebInvoiceUser(t, db, 10)
	seedWebInvoiceOrder(t, db, 10, 1001, "ORD-invoice-1", domainorder.StatusFulfilled, domainorder.PaymentStatusPaid, 1200)
	seedWebInvoiceOrder(t, db, 10, 1002, "ORD-invoice-2", domainorder.StatusFulfilled, domainorder.PaymentStatusPaid, 1800)

	service := NewService(db, config.StorageConfig{LocalPath: t.TempDir()})
	created, err := service.Create(context.Background(), 10, webdto.InvoiceCreateRequest{
		OrderNos:    []string{"ORD-invoice-1", "ORD-invoice-2"},
		TitleType:   domaininvoice.TitleTypeCompany,
		Title:       "测试企业",
		TaxNo:       stringPtr("913300000000000000"),
		ClientToken: "invoice-token-1",
	})
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}
	if created.AmountCents != 3000 || len(created.Orders) != 2 || created.Status != domaininvoice.StatusPending {
		t.Fatalf("unexpected created invoice: %#v", created)
	}

	repeated, err := service.Create(context.Background(), 10, webdto.InvoiceCreateRequest{
		OrderNos:    []string{"ORD-invoice-1"},
		TitleType:   domaininvoice.TitleTypePersonal,
		Title:       "重复提交",
		ClientToken: "invoice-token-1",
	})
	if err != nil {
		t.Fatalf("repeat invoice: %v", err)
	}
	if repeated.InvoiceNo != created.InvoiceNo {
		t.Fatalf("same client token should return existing invoice, got %s want %s", repeated.InvoiceNo, created.InvoiceNo)
	}

	_, err = service.Create(context.Background(), 10, webdto.InvoiceCreateRequest{
		OrderNos:    []string{"ORD-invoice-1"},
		TitleType:   domaininvoice.TitleTypePersonal,
		Title:       "二次开票",
		ClientToken: "invoice-token-2",
	})
	if err == nil {
		t.Fatalf("active invoice order should block duplicate application")
	}

	cancelled, err := service.Cancel(context.Background(), 10, created.InvoiceNo, webdto.InvoiceCancelRequest{Reason: stringPtr("重新申请")})
	if err != nil {
		t.Fatalf("cancel invoice: %v", err)
	}
	if cancelled.Status != domaininvoice.StatusCancelled {
		t.Fatalf("cancel should update invoice status, got %s", cancelled.Status)
	}
	var activeCount int64
	if err := db.Table("invoice_application_orders").Where("order_no = ? AND status_snapshot IN ?", "ORD-invoice-1", []string{domaininvoice.StatusPending, domaininvoice.StatusProcessing, domaininvoice.StatusIssued}).Count(&activeCount).Error; err != nil {
		t.Fatalf("count active invoice orders: %v", err)
	}
	if activeCount != 0 {
		t.Fatalf("cancelled invoice should release active occupancy, got %d", activeCount)
	}

	createdAgain, err := service.Create(context.Background(), 10, webdto.InvoiceCreateRequest{
		OrderNos:    []string{"ORD-invoice-1"},
		TitleType:   domaininvoice.TitleTypePersonal,
		Title:       "个人抬头",
		ClientToken: "invoice-token-3",
	})
	if err != nil {
		t.Fatalf("create after cancel should succeed: %v", err)
	}
	if createdAgain.InvoiceNo == created.InvoiceNo || createdAgain.AmountCents != 1200 {
		t.Fatalf("unexpected invoice after cancel: %#v", createdAgain)
	}
}

func seedWebInvoiceUser(t *testing.T, db *gorm.DB, id uint64) {
	t.Helper()
	if err := db.Exec(`INSERT INTO users (id, username, email, status) VALUES (?, ?, ?, 'active')`, id, "user", "user@example.com").Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func seedWebInvoiceOrder(t *testing.T, db *gorm.DB, userID uint64, id uint64, orderNo string, status string, paymentStatus string, amount uint64) {
	t.Helper()
	paidAt := time.Now().Truncate(time.Millisecond)
	if err := db.Exec(`INSERT INTO orders (
id, order_no, user_id, status, order_type, related_instance_no, total_amount_cents,
currency, payment_status, paid_at, product_name, plan_name, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, NULL, ?, 'CNY', ?, ?, '云服务器', '基础套餐', ?, ?)`,
		id, orderNo, userID, status, domainorder.TypePurchase, amount, paymentStatus, paidAt, paidAt, paidAt,
	).Error; err != nil {
		t.Fatalf("seed order: %v", err)
	}
}

func stringPtr(value string) *string { return &value }

const webInvoiceUsersSchema = `
CREATE TABLE users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(64) NOT NULL,
  email VARCHAR(128) NOT NULL,
  display_name VARCHAR(64) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const webInvoiceOrdersSchema = `
CREATE TABLE orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  order_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  order_type VARCHAR(32) NOT NULL DEFAULT 'purchase',
  related_instance_no VARCHAR(64) NULL,
  total_amount_cents BIGINT UNSIGNED NOT NULL,
  currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
  payment_status VARCHAR(32) NOT NULL DEFAULT 'unpaid',
  paid_at DATETIME(3) NULL,
  product_name VARCHAR(128) NOT NULL,
  plan_name VARCHAR(128) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_orders_order_no (order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const webInvoiceRefundsSchema = `
CREATE TABLE refund_transactions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  order_id BIGINT UNSIGNED NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const webInvoiceApplicationsSchema = `
CREATE TABLE invoice_applications (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  invoice_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  client_token VARCHAR(128) NOT NULL,
  invoice_type VARCHAR(32) NOT NULL DEFAULT 'electronic_normal',
  title_type VARCHAR(32) NOT NULL,
  title VARCHAR(100) NOT NULL,
  tax_no VARCHAR(64) NULL,
  email VARCHAR(128) NULL,
  amount_cents BIGINT UNSIGNED NOT NULL,
  currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  remark VARCHAR(500) NULL,
  admin_note VARCHAR(1000) NULL,
  reject_reason VARCHAR(500) NULL,
  cancel_reason VARCHAR(500) NULL,
  invoice_code VARCHAR(64) NULL,
  invoice_number VARCHAR(128) NULL,
  invoice_file_id BIGINT UNSIGNED NULL,
  accepted_by_admin_id BIGINT UNSIGNED NULL,
  rejected_by_admin_id BIGINT UNSIGNED NULL,
  issued_by_admin_id BIGINT UNSIGNED NULL,
  accepted_at DATETIME(3) NULL,
  rejected_at DATETIME(3) NULL,
  cancelled_at DATETIME(3) NULL,
  issued_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_invoice_applications_invoice_no (invoice_no),
  UNIQUE KEY uk_invoice_applications_idempotency (user_id, client_token)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const webInvoiceApplicationOrdersSchema = `
CREATE TABLE invoice_application_orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  invoice_id BIGINT UNSIGNED NOT NULL,
  invoice_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  order_id BIGINT UNSIGNED NOT NULL,
  order_no VARCHAR(64) NOT NULL,
  order_type VARCHAR(32) NOT NULL,
  order_amount_cents BIGINT UNSIGNED NOT NULL,
  currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
  payment_status VARCHAR(32) NOT NULL,
  paid_at DATETIME(3) NULL,
  product_name VARCHAR(128) NULL,
  plan_name VARCHAR(128) NULL,
  status_snapshot VARCHAR(32) NOT NULL DEFAULT 'pending',
  active_order_id BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE WHEN status_snapshot IN ('pending', 'processing', 'issued') THEN order_id ELSE NULL END
  ) STORED,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_invoice_application_orders_invoice_order (invoice_id, order_id),
  UNIQUE KEY uk_invoice_application_orders_active_order (active_order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
