package paymentalert

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
)

func TestRecordSkipsUnknownEvent(t *testing.T) {
	db := openPaymentAlertDB(t)
	recorder := testRecorder(db)

	recorder.Record(context.Background(), Event{Event: "unexpected_event", PaymentNo: "PAY-unknown"})

	var count int64
	if err := db.Table("backend_runtime_logs").Where("message = ?", alertMessage).Count(&count).Error; err != nil {
		t.Fatalf("count alert rows: %v", err)
	}
	if count != 0 {
		t.Fatalf("unknown payment alert event must not be persisted, got %d rows", count)
	}
}

func TestRecordPersistsKnownEventWithSanitizedError(t *testing.T) {
	db := openPaymentAlertDB(t)
	recorder := testRecorder(db)

	recorder.Record(context.Background(), Event{
		Event:        EventPaymentCreateFailed,
		PaymentNo:    "PAY-alert-1",
		OrderNo:      "ORD-alert-1",
		Provider:     "wechat",
		Method:       "wechat_native",
		Status:       "failed",
		ErrorCode:    "CHANNEL_CREATE_FAILED",
		ErrorMessage: "gateway rejected api_v3_key=secret-value",
	})

	detail := requireAlertDetail(t, db, EventPaymentCreateFailed)
	if !strings.Contains(detail, `"payment_no":"PAY-alert-1"`) || !strings.Contains(detail, `"order_no":"ORD-alert-1"`) {
		t.Fatalf("alert should include business anchors, got %s", detail)
	}
	if strings.Contains(detail, "secret-value") || !strings.Contains(detail, "[已脱敏]") {
		t.Fatalf("alert should redact sensitive error fragments, got %s", detail)
	}
}

func openPaymentAlertDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, paymentAlertBackendRuntimeLogsSchema)
	return db
}

func testRecorder(db *gorm.DB) *Recorder {
	return New(db, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func requireAlertDetail(t *testing.T, db *gorm.DB, event string) string {
	t.Helper()
	var row struct {
		Level    string
		Category string
		Message  string
		Detail   *string
	}
	if err := db.Table("backend_runtime_logs").Where("message = ? AND detail LIKE ?", alertMessage, "%"+event+"%").Take(&row).Error; err != nil {
		t.Fatalf("load payment alert %s: %v", event, err)
	}
	if row.Level != "error" || row.Category != "runtime" || row.Message != alertMessage || row.Detail == nil {
		t.Fatalf("unexpected alert row: %#v", row)
	}
	return *row.Detail
}

const paymentAlertBackendRuntimeLogsSchema = `
CREATE TABLE backend_runtime_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  level VARCHAR(16) NOT NULL,
  category VARCHAR(32) NOT NULL,
  request_id VARCHAR(64) NULL,
  request_method VARCHAR(16) NULL,
  request_path VARCHAR(255) NULL,
  status INT NULL,
  latency_ms BIGINT NULL,
  client_ip VARCHAR(64) NULL,
  message VARCHAR(255) NOT NULL,
  detail JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
