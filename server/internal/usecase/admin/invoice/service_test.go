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
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
)

func TestAcceptAndIssueSyncSnapshotsFileReferenceAndAudit(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, adminInvoiceUsersSchema, adminInvoiceApplicationsSchema, adminInvoiceApplicationOrdersSchema, adminInvoiceFilesSchema, adminInvoiceFileReferencesSchema, adminInvoiceAuditLogsSchema)
	seedAdminInvoiceUser(t, db, 20)
	seedAdminInvoiceApplication(t, db, 2001, "INV-admin-1", 20, domaininvoice.StatusPending)
	seedAdminInvoiceOrder(t, db, 2001, "INV-admin-1", 3001, "ORD-admin-1", domaininvoice.StatusPending)
	seedAdminInvoicePDF(t, db, 5001)

	service := NewService(db, nil, config.StorageConfig{LocalPath: t.TempDir()})
	accepted, err := service.Accept(context.Background(), 99, "INV-admin-1")
	if err != nil {
		t.Fatalf("accept invoice: %v", err)
	}
	if accepted.Status != domaininvoice.StatusProcessing {
		t.Fatalf("accept should move invoice to processing, got %s", accepted.Status)
	}
	requireAdminInvoiceSnapshot(t, db, "INV-admin-1", domaininvoice.StatusProcessing)

	issuedAt := time.Now().Truncate(time.Millisecond)
	issued, err := service.Issue(context.Background(), 99, "INV-admin-1", admindto.InvoiceIssueRequest{InvoiceNumber: "FP-001", IssuedAt: issuedAt, FileID: 5001})
	if err != nil {
		t.Fatalf("issue invoice: %v", err)
	}
	if issued.Status != domaininvoice.StatusIssued || issued.File == nil || issued.File.ID != 5001 {
		t.Fatalf("issue should bind pdf and move invoice to issued, got %#v", issued)
	}
	requireAdminInvoiceSnapshot(t, db, "INV-admin-1", domaininvoice.StatusIssued)
	var refCount int64
	if err := db.Table("file_attachment_references").Where("file_id = ? AND ref_type = ? AND ref_id = ?", 5001, domaininvoice.FileRefType, "2001").Count(&refCount).Error; err != nil {
		t.Fatalf("count file references: %v", err)
	}
	if refCount != 1 {
		t.Fatalf("issue should create one file reference, got %d", refCount)
	}
	var auditCount int64
	if err := db.Table("admin_audit_logs").Where("admin_id = ? AND action IN ?", 99, []string{"invoice.accept", "invoice.issue"}).Count(&auditCount).Error; err != nil {
		t.Fatalf("count audit logs: %v", err)
	}
	if auditCount != 2 {
		t.Fatalf("accept and issue should write audit logs, got %d", auditCount)
	}
}

func seedAdminInvoiceUser(t *testing.T, db *gorm.DB, id uint64) {
	t.Helper()
	if err := db.Exec(`INSERT INTO users (id, username, email, status) VALUES (?, ?, ?, 'active')`, id, "invoice-user", "invoice@example.com").Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func seedAdminInvoiceApplication(t *testing.T, db *gorm.DB, id uint64, invoiceNo string, userID uint64, status string) {
	t.Helper()
	if err := db.Exec(`INSERT INTO invoice_applications (
id, invoice_no, user_id, client_token, invoice_type, title_type, title, tax_no,
amount_cents, currency, status
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'CNY', ?)`,
		id, invoiceNo, userID, "client-"+invoiceNo, domaininvoice.TypeElectronicNormal, domaininvoice.TitleTypeCompany, "测试企业", "913300000000000000", 3000, status,
	).Error; err != nil {
		t.Fatalf("seed invoice application: %v", err)
	}
}

func seedAdminInvoiceOrder(t *testing.T, db *gorm.DB, invoiceID uint64, invoiceNo string, orderID uint64, orderNo string, status string) {
	t.Helper()
	if err := db.Exec(`INSERT INTO invoice_application_orders (
invoice_id, invoice_no, user_id, order_id, order_no, order_type, order_amount_cents,
currency, payment_status, status_snapshot
) VALUES (?, ?, ?, ?, ?, ?, ?, 'CNY', ?, ?)`,
		invoiceID, invoiceNo, 20, orderID, orderNo, domainorder.TypePurchase, 3000, domainorder.PaymentStatusPaid, status,
	).Error; err != nil {
		t.Fatalf("seed invoice order: %v", err)
	}
}

func seedAdminInvoicePDF(t *testing.T, db *gorm.DB, id uint64) {
	t.Helper()
	if err := db.Exec(`INSERT INTO file_attachments (
id, original_name, stored_name, mime_type, extension, size, storage_path, storage_driver, checksum, status
) VALUES (?, 'invoice.pdf', 'invoice.pdf', 'application/pdf', 'pdf', 128, '2026/05/25/invoice.pdf', 'local', 'checksum', 'active')`, id).Error; err != nil {
		t.Fatalf("seed pdf: %v", err)
	}
}

func requireAdminInvoiceSnapshot(t *testing.T, db *gorm.DB, invoiceNo string, status string) {
	t.Helper()
	var row struct {
		StatusSnapshot string `gorm:"column:status_snapshot"`
	}
	if err := db.Table("invoice_application_orders").Select("status_snapshot").Where("invoice_no = ?", invoiceNo).Take(&row).Error; err != nil {
		t.Fatalf("load snapshot: %v", err)
	}
	if row.StatusSnapshot != status {
		t.Fatalf("snapshot status mismatch: got %s want %s", row.StatusSnapshot, status)
	}
}

const adminInvoiceUsersSchema = `
CREATE TABLE users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(64) NOT NULL,
  email VARCHAR(128) NOT NULL,
  display_name VARCHAR(64) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const adminInvoiceApplicationsSchema = `
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

const adminInvoiceApplicationOrdersSchema = `
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
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  KEY idx_invoice_application_orders_invoice (invoice_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const adminInvoiceFilesSchema = `
CREATE TABLE file_attachments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  original_name VARCHAR(255) NOT NULL,
  stored_name VARCHAR(128) NOT NULL,
  mime_type VARCHAR(128) NOT NULL,
  extension VARCHAR(32) NOT NULL,
  size BIGINT UNSIGNED NOT NULL,
  storage_path VARCHAR(500) NOT NULL,
  storage_driver VARCHAR(32) NOT NULL DEFAULT 'local',
  checksum VARCHAR(64) NOT NULL,
  uploader_id BIGINT UNSIGNED NULL,
  uploader_user_id BIGINT UNSIGNED NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const adminInvoiceFileReferencesSchema = `
CREATE TABLE file_attachment_references (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  file_id BIGINT UNSIGNED NOT NULL,
  ref_type VARCHAR(64) NOT NULL,
  ref_id VARCHAR(128) NOT NULL,
  ref_name VARCHAR(255) NULL,
  ref_path VARCHAR(255) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_file_attachment_references_unique (file_id, ref_type, ref_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const adminInvoiceAuditLogsSchema = `
CREATE TABLE admin_audit_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  admin_id BIGINT UNSIGNED NULL,
  admin_username VARCHAR(64) NULL,
  admin_display_name VARCHAR(64) NULL,
  session_id VARCHAR(128) NULL,
  action VARCHAR(128) NOT NULL,
  object_type VARCHAR(64) NOT NULL,
  object_id VARCHAR(128) NULL,
  before_data JSON NULL,
  after_data JSON NULL,
  remark VARCHAR(500) NULL,
  request_id VARCHAR(128) NULL,
  request_method VARCHAR(16) NULL,
  request_path VARCHAR(255) NULL,
  ip VARCHAR(64) NULL,
  user_agent VARCHAR(500) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
