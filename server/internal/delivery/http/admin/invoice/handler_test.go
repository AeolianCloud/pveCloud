package invoice

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	adminmiddleware "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/middleware"
	domaininvoice "github.com/AeolianCloud/pveCloud/server/internal/domain/invoice"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
	invoiceusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/invoice"
)

func TestAdminInvoiceRoutesEnforcePermissionsAndValidateRequests(t *testing.T) {
	lowPermissionRouter := newAdminInvoicePermissionRouter(nil, []string{"page.invoices"})
	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "accept", method: http.MethodPost, path: "/invoices/INV-ADMIN-1/accept"},
		{name: "reject", method: http.MethodPost, path: "/invoices/INV-ADMIN-1/reject", body: `{"reason":"bad"}`},
		{name: "issue", method: http.MethodPost, path: "/invoices/INV-ADMIN-1/issue", body: `{"invoice_number":"FP-1","issued_at":"2026-05-25T00:00:00Z","file_id":1}`},
		{name: "admin note", method: http.MethodPatch, path: "/invoices/INV-ADMIN-1/admin-note", body: `{"admin_note":"note"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name+" requires action permission", func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			request.Header.Set("Content-Type", "application/json")
			lowPermissionRouter.ServeHTTP(recorder, request)

			require.Equal(t, http.StatusForbidden, recorder.Code)
			requireAdminInvoiceEnvelopeCode(t, recorder, 40301)
		})
	}

	noViewRouter := newAdminInvoicePermissionRouter(nil, []string{"invoice:update"})
	t.Run("view route rejects missing view permission", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/invoices", nil)
		noViewRouter.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusForbidden, recorder.Code)
		requireAdminInvoiceEnvelopeCode(t, recorder, 40301)
	})

	updateRouter := newAdminInvoicePermissionRouter(nil, []string{"invoice:reject"})
	t.Run("reject request validates required reason before service", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/invoices/INV-ADMIN-1/reject", strings.NewReader(`{"reason":""}`))
		request.Header.Set("Content-Type", "application/json")
		updateRouter.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		requireAdminInvoiceEnvelopeCode(t, recorder, 40001)
	})

	viewRouter := newAdminInvoicePermissionRouter(nil, []string{"invoice:view"})
	t.Run("list request validates status filter before service", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/invoices?status=bad-status", nil)
		viewRouter.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		requireAdminInvoiceEnvelopeCode(t, recorder, 40001)
	})
}

func TestAdminInvoiceDownloadAllowsViewPermissionAndMapsFileErrors(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db,
		adminInvoiceHandlerUsersSchema,
		adminInvoiceHandlerApplicationsSchema,
		adminInvoiceHandlerOrdersSchema,
		adminInvoiceHandlerFilesSchema,
		adminInvoiceHandlerFileReferencesSchema,
	)
	storageRoot := t.TempDir()
	seedAdminInvoiceHandlerUser(t, db, 30)
	seedAdminInvoiceHandlerPDF(t, db, storageRoot, 701, "invoices/admin-issued.pdf", "invoice-admin.pdf", []byte("%PDF-admin-issued"))
	seedAdminInvoiceHandlerApplication(t, db, 3001, "INV-ADMIN-OWN", 30, domaininvoice.StatusIssued, uint64Ptr(701))
	seedAdminInvoiceHandlerFileReference(t, db, 701, 3001)

	router := newAdminInvoicePermissionRouter(
		invoiceusecase.NewService(db, nil, config.StorageConfig{LocalPath: storageRoot}),
		[]string{"invoice:view"},
	)

	t.Run("invoice view permission can download pdf", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/invoices/INV-ADMIN-OWN/download", nil)
		router.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "application/pdf", recorder.Header().Get("Content-Type"))
		require.Contains(t, recorder.Header().Get("Content-Disposition"), "filename*=UTF-8''invoice-admin.pdf")
		require.Contains(t, recorder.Header().Get("Cache-Control"), "no-store")
		require.Equal(t, []byte("%PDF-admin-issued"), recorder.Body.Bytes())
	})

	t.Run("missing invoice maps to not found envelope", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/invoices/INV-ADMIN-MISSING/download", nil)
		router.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusNotFound, recorder.Code)
		requireAdminInvoiceEnvelopeCode(t, recorder, 40401)
	})
}

func newAdminInvoicePermissionRouter(service *invoiceusecase.Service, permissionCodes []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("admin_id", uint64(99))
		c.Set("admin_permission_codes", permissionCodes)
	})
	handler := NewHandler(service)
	router.GET("/invoices", adminmiddleware.AdminAnyPermission("page.invoices", "invoice:view"), handler.List)
	router.GET("/invoices/:invoice_no/download", adminmiddleware.AdminAnyPermission("page.invoices", "invoice:view"), handler.Download)
	router.POST("/invoices/:invoice_no/accept", adminmiddleware.AdminPermission("invoice:update"), handler.Accept)
	router.POST("/invoices/:invoice_no/reject", adminmiddleware.AdminPermission("invoice:reject"), handler.Reject)
	router.POST("/invoices/:invoice_no/issue", adminmiddleware.AdminPermission("invoice:issue"), handler.Issue)
	router.PATCH("/invoices/:invoice_no/admin-note", adminmiddleware.AdminPermission("invoice:update"), handler.UpdateAdminNote)
	return router
}

func seedAdminInvoiceHandlerUser(t *testing.T, db *gorm.DB, id uint64) {
	t.Helper()
	require.NoError(t, db.Exec(`INSERT INTO users (id, username, email, status) VALUES (?, ?, ?, 'active')`, id, "admin-invoice-user", "admin-invoice@example.com").Error)
}

func seedAdminInvoiceHandlerApplication(t *testing.T, db *gorm.DB, id uint64, invoiceNo string, userID uint64, status string, fileID *uint64) {
	t.Helper()
	now := time.Now().Truncate(time.Millisecond)
	require.NoError(t, db.Exec(`INSERT INTO invoice_applications (
id, invoice_no, user_id, client_token, invoice_type, title_type, title, tax_no,
amount_cents, currency, status, invoice_number, invoice_file_id, issued_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'CNY', ?, ?, ?, ?, ?, ?)`,
		id, invoiceNo, userID, "client-"+invoiceNo, domaininvoice.TypeElectronicNormal,
		domaininvoice.TitleTypeCompany, "测试企业", "913300000000000000", 2500,
		status, "FP-"+invoiceNo, fileID, now, now, now,
	).Error)
}

func seedAdminInvoiceHandlerPDF(t *testing.T, db *gorm.DB, storageRoot string, id uint64, storagePath string, originalName string, content []byte) {
	t.Helper()
	absolutePath := filepath.Join(storageRoot, storagePath)
	require.NoError(t, os.MkdirAll(filepath.Dir(absolutePath), 0o755))
	require.NoError(t, os.WriteFile(absolutePath, content, 0o644))
	require.NoError(t, db.Exec(`INSERT INTO file_attachments (
id, original_name, stored_name, mime_type, extension, size, storage_path, storage_driver, checksum, status
) VALUES (?, ?, ?, 'application/pdf', 'pdf', ?, ?, 'local', 'checksum', 'active')`,
		id, originalName, originalName, len(content), storagePath,
	).Error)
}

func seedAdminInvoiceHandlerFileReference(t *testing.T, db *gorm.DB, fileID uint64, invoiceID uint64) {
	t.Helper()
	require.NoError(t, db.Exec(`INSERT INTO file_attachment_references (file_id, ref_type, ref_id) VALUES (?, ?, ?)`, fileID, domaininvoice.FileRefType, strconv.FormatUint(invoiceID, 10)).Error)
}

func requireAdminInvoiceEnvelopeCode(t *testing.T, recorder *httptest.ResponseRecorder, code int) {
	t.Helper()
	var envelope response.Envelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
	require.Equal(t, code, envelope.Code)
	require.Nil(t, envelope.Data)
}

func uint64Ptr(value uint64) *uint64 { return &value }

const adminInvoiceHandlerUsersSchema = `
CREATE TABLE users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(64) NOT NULL,
  email VARCHAR(128) NOT NULL,
  display_name VARCHAR(64) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const adminInvoiceHandlerApplicationsSchema = `
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

const adminInvoiceHandlerOrdersSchema = `
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

const adminInvoiceHandlerFilesSchema = `
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

const adminInvoiceHandlerFileReferencesSchema = `
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
