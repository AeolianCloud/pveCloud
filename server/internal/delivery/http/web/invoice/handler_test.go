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

	domaininvoice "github.com/AeolianCloud/pveCloud/server/internal/domain/invoice"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
	invoiceusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/invoice"
)

func TestWebInvoiceHandlersRequireCurrentUserAndValidateRequests(t *testing.T) {
	router := newWebInvoiceHandlerRouter(nil, 0, false)

	t.Run("list requires current user", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/invoices", nil)
		router.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		requireInvoiceEnvelopeCode(t, recorder, 40101)
	})

	t.Run("download requires current user", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/invoices/INV-WEB-AUTH/download", nil)
		router.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		requireInvoiceEnvelopeCode(t, recorder, 40101)
	})

	authenticated := newWebInvoiceHandlerRouter(nil, 10, true)
	tests := []struct {
		name string
		body string
	}{
		{name: "malformed json", body: `{"order_nos":`},
		{name: "missing order numbers", body: `{"title_type":"personal","title":"个人","client_token":"token-1"}`},
		{name: "invalid title type", body: `{"order_nos":["ORD-1"],"title_type":"special","title":"个人","client_token":"token-2"}`},
		{name: "overlong client token", body: `{"order_nos":["ORD-1"],"title_type":"personal","title":"个人","client_token":"` + strings.Repeat("a", 129) + `"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/invoices", strings.NewReader(tt.body))
			request.Header.Set("Content-Type", "application/json")
			authenticated.ServeHTTP(recorder, request)

			require.Equal(t, http.StatusBadRequest, recorder.Code)
			requireInvoiceEnvelopeCode(t, recorder, 40001)
		})
	}

	t.Run("invalid list status maps to validation error", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/invoices?status=bad-status", nil)
		authenticated.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		requireInvoiceEnvelopeCode(t, recorder, 40001)
	})
}

func TestWebInvoiceDownloadEnforcesOwnershipStateAndHeaders(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db,
		webInvoiceHandlerUsersSchema,
		webInvoiceHandlerApplicationsSchema,
		webInvoiceHandlerOrdersSchema,
		webInvoiceHandlerFilesSchema,
		webInvoiceHandlerFileReferencesSchema,
	)
	storageRoot := t.TempDir()
	seedWebInvoiceHandlerUser(t, db, 10, "owner")
	seedWebInvoiceHandlerUser(t, db, 20, "other")
	seedWebInvoiceHandlerPDF(t, db, storageRoot, 501, "invoices/web-issued.pdf", "invoice-web.pdf", []byte("%PDF-web-issued"))
	seedWebInvoiceHandlerApplication(t, db, 1001, "INV-WEB-OWN", 10, domaininvoice.StatusIssued, uintPtr(501))
	seedWebInvoiceHandlerApplication(t, db, 1002, "INV-WEB-PENDING", 10, domaininvoice.StatusPending, uintPtr(501))
	seedWebInvoiceHandlerApplication(t, db, 1003, "INV-WEB-OTHER", 20, domaininvoice.StatusIssued, uintPtr(501))
	seedWebInvoiceHandlerFileReference(t, db, 501, 1001)
	seedWebInvoiceHandlerFileReference(t, db, 501, 1003)

	router := newWebInvoiceHandlerRouter(invoiceusecase.NewService(db, config.StorageConfig{LocalPath: storageRoot}), 10, true)

	t.Run("own issued invoice downloads with protected headers", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/invoices/INV-WEB-OWN/download", nil)
		router.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "application/pdf", recorder.Header().Get("Content-Type"))
		require.Contains(t, recorder.Header().Get("Content-Disposition"), "filename*=UTF-8''invoice-web.pdf")
		require.Contains(t, recorder.Header().Get("Cache-Control"), "no-store")
		require.Contains(t, recorder.Header().Get("Cache-Control"), "private")
		require.Equal(t, []byte("%PDF-web-issued"), recorder.Body.Bytes())
	})

	t.Run("cross user invoice stays invisible", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/invoices/INV-WEB-OTHER/download", nil)
		router.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusNotFound, recorder.Code)
		requireInvoiceEnvelopeCode(t, recorder, 40401)
	})

	t.Run("non issued invoice maps to conflict", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/invoices/INV-WEB-PENDING/download", nil)
		router.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusConflict, recorder.Code)
		requireInvoiceEnvelopeCode(t, recorder, 40901)
	})
}

func newWebInvoiceHandlerRouter(service *invoiceusecase.Service, userID uint64, injectUser bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	if injectUser {
		router.Use(func(c *gin.Context) {
			c.Set("web_user_id", userID)
		})
	}
	handler := NewHandler(service)
	router.GET("/invoices", handler.List)
	router.POST("/invoices", handler.Create)
	router.GET("/invoices/:invoice_no/download", handler.Download)
	return router
}

func seedWebInvoiceHandlerUser(t *testing.T, db *gorm.DB, id uint64, username string) {
	t.Helper()
	require.NoError(t, db.Exec(`INSERT INTO users (id, username, email, status) VALUES (?, ?, ?, 'active')`, id, username, username+"@example.com").Error)
}

func seedWebInvoiceHandlerApplication(t *testing.T, db *gorm.DB, id uint64, invoiceNo string, userID uint64, status string, fileID *uint64) {
	t.Helper()
	now := time.Now().Truncate(time.Millisecond)
	require.NoError(t, db.Exec(`INSERT INTO invoice_applications (
id, invoice_no, user_id, client_token, invoice_type, title_type, title, tax_no,
amount_cents, currency, status, invoice_number, invoice_file_id, issued_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'CNY', ?, ?, ?, ?, ?, ?)`,
		id, invoiceNo, userID, "client-"+invoiceNo, domaininvoice.TypeElectronicNormal,
		domaininvoice.TitleTypeCompany, "测试企业", "913300000000000000", 1200,
		status, "FP-"+invoiceNo, fileID, now, now, now,
	).Error)
}

func seedWebInvoiceHandlerPDF(t *testing.T, db *gorm.DB, storageRoot string, id uint64, storagePath string, originalName string, content []byte) {
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

func seedWebInvoiceHandlerFileReference(t *testing.T, db *gorm.DB, fileID uint64, invoiceID uint64) {
	t.Helper()
	require.NoError(t, db.Exec(`INSERT INTO file_attachment_references (file_id, ref_type, ref_id) VALUES (?, ?, ?)`, fileID, domaininvoice.FileRefType, uint64ToString(invoiceID)).Error)
}

func requireInvoiceEnvelopeCode(t *testing.T, recorder *httptest.ResponseRecorder, code int) {
	t.Helper()
	var envelope response.Envelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
	require.Equal(t, code, envelope.Code)
	require.Nil(t, envelope.Data)
}

func uintPtr(value uint64) *uint64 { return &value }

func uint64ToString(value uint64) string {
	return strconv.FormatUint(value, 10)
}

const webInvoiceHandlerUsersSchema = `
CREATE TABLE users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(64) NOT NULL,
  email VARCHAR(128) NOT NULL,
  display_name VARCHAR(64) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const webInvoiceHandlerApplicationsSchema = `
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

const webInvoiceHandlerOrdersSchema = `
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

const webInvoiceHandlerFilesSchema = `
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

const webInvoiceHandlerFileReferencesSchema = `
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
