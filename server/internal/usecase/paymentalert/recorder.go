package paymentalert

import (
	"context"
	"encoding/json"
	"log/slog"
	"regexp"
	"strings"

	"gorm.io/gorm"

	mysqllogs "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/logs"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/requestcontext"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
)

const (
	EventPaymentCreateFailed            = "payment_create_failed"
	EventPaymentCallbackSignatureFailed = "payment_callback_signature_failed"
	EventRefundPending                  = "refund_pending"
	EventRefundFailed                   = "refund_failed"

	alertModule  = "payment"
	alertMessage = "payment_alert"
)

type Recorder struct {
	log  *slog.Logger
	logs *mysqllogs.Repository
}

type Event struct {
	Event        string
	PaymentNo    string
	RefundNo     string
	OrderNo      string
	Provider     string
	Method       string
	Status       string
	ErrorCode    string
	ErrorMessage string
}

var sensitiveAlertPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)((?:token|password|passwd|pwd|secret|private[_-]?key|api[_-]?key|api[_-]?v3[_-]?key|signature|sign)\s*[:=]\s*)("[^"]*"|'[^']*'|[^\s&;,]+)`),
	regexp.MustCompile(`(?i)("(?:token|password|passwd|pwd|secret|private[_-]?key|api[_-]?key|api[_-]?v3[_-]?key|signature|sign)"\s*:\s*)("[^"]*"|[^,\}\s]+)`),
}

func New(db *gorm.DB, log *slog.Logger) *Recorder {
	if log == nil {
		log = slog.Default()
	}
	return &Recorder{log: log, logs: mysqllogs.NewRepository(db)}
}

func (r *Recorder) Record(ctx context.Context, event Event) {
	if r == nil {
		return
	}
	detail := event.detail()
	attrs := []any{
		"module", alertModule,
		"event", detail.Event,
		"payment_no", detail.PaymentNo,
		"refund_no", detail.RefundNo,
		"order_no", detail.OrderNo,
		"provider", detail.Provider,
		"method", detail.Method,
		"status", detail.Status,
		"error_code", detail.ErrorCode,
	}
	if detail.ErrorMessage != "" {
		attrs = append(attrs, "error_message", detail.ErrorMessage)
	}
	r.log.Error(alertMessage, attrs...)
	if r.logs == nil {
		return
	}
	req := requestcontext.RequestContextFrom(ctx)
	detailJSON, _ := json.Marshal(detail)
	// 告警日志是运维事件源，不能影响支付主流程；写入失败时只保留 stdout 事件。
	_ = r.logs.CreateBackendRuntimeLog(ctx, nil, &mysqllogs.BackendRuntimeLog{
		Level:         "error",
		Category:      "runtime",
		RequestID:     stringPtr(req.RequestID),
		RequestMethod: stringPtr(req.RequestMethod),
		RequestPath:   stringPtr(req.RequestPath),
		ClientIP:      stringPtr(req.IP),
		Message:       alertMessage,
		Detail:        stringPtr(string(detailJSON)),
	})
}

type detail struct {
	Module       string `json:"module"`
	Event        string `json:"event"`
	PaymentNo    string `json:"payment_no,omitempty"`
	RefundNo     string `json:"refund_no,omitempty"`
	OrderNo      string `json:"order_no,omitempty"`
	Provider     string `json:"provider,omitempty"`
	Method       string `json:"method,omitempty"`
	Status       string `json:"status,omitempty"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

func (e Event) detail() detail {
	return detail{
		Module:       alertModule,
		Event:        strings.TrimSpace(e.Event),
		PaymentNo:    strings.TrimSpace(e.PaymentNo),
		RefundNo:     strings.TrimSpace(e.RefundNo),
		OrderNo:      strings.TrimSpace(e.OrderNo),
		Provider:     strings.TrimSpace(e.Provider),
		Method:       strings.TrimSpace(e.Method),
		Status:       strings.TrimSpace(e.Status),
		ErrorCode:    strings.TrimSpace(e.ErrorCode),
		ErrorMessage: sanitize(e.ErrorMessage),
	}
}

func sanitize(value string) string {
	value = textutil.TrimTo(strings.TrimSpace(value), 500)
	if value == "" {
		return ""
	}
	for _, pattern := range sensitiveAlertPatterns {
		value = pattern.ReplaceAllString(value, "${1}[已脱敏]")
	}
	return value
}

func stringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
