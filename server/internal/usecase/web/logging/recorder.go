package logging

import (
	"context"
	"net/url"
	"regexp"
	"strings"

	"gorm.io/gorm"

	mysqllogs "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/logs"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/requestcontext"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
)

type Recorder struct {
	logs *mysqllogs.Repository
}

const frontendLogMaskedValue = "[已脱敏]"

var sensitiveFrontendTextPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)((?:cookie|set-cookie)\s*[:=]\s*)[^\r\n]+`),
	regexp.MustCompile(`(?i)(bearer\s+)[^\s"'<>]+`),
	regexp.MustCompile(`(?i)((?:access[_-]?token|refresh[_-]?token|token|password|passwd|pwd|secret|captcha|authorization|credential|cookie|set-cookie|config[_-]?value|api[_-]?key)\s*[:=]\s*)("[^"]*"|'[^']*'|[^\s&;,]+)`),
	regexp.MustCompile(`(?i)("(?:access[_-]?token|refresh[_-]?token|token|password|passwd|pwd|secret|captcha|authorization|credential|cookie|set-cookie|config[_-]?value|api[_-]?key)"\s*:\s*)("[^"]*"|[^,\}\s]+)`),
}

type FrontendErrorInput struct {
	SourceApp    string
	UserID       *uint64
	AdminID      *uint64
	RequestID    string
	PagePath     string
	ErrorType    string
	Message      string
	Stack        string
	APIPath      string
	HTTPStatus   *int
	BusinessCode *int
	Browser      string
	OS           string
	AppVersion   string
	IP           string
	UserAgent    string
}

type UserSnapshot struct {
	ID       *uint64
	Username *string
	Email    *string
}

func NewRecorder(db *gorm.DB) *Recorder {
	return &Recorder{logs: mysqllogs.NewRepository(db)}
}

func (r *Recorder) Security(ctx context.Context, tx *gorm.DB, user UserSnapshot, sessionID string, action string, result string, remark string) error {
	if r == nil {
		return nil
	}
	request := requestcontext.RequestContextFrom(ctx)
	sessionID = firstNonEmpty(sessionID, request.SessionID)
	return r.logs.CreateUserSecurityLog(ctx, tx, &mysqllogs.UserSecurityLog{
		UserID:        user.ID,
		Username:      trimPtr(user.Username, 64),
		Email:         trimPtr(user.Email, 191),
		SessionID:     stringPtr(sessionID),
		RequestID:     stringPtr(request.RequestID),
		RequestMethod: stringPtr(request.RequestMethod),
		RequestPath:   stringPtr(request.RequestPath),
		Action:        strings.TrimSpace(action),
		Result:        firstNonEmpty(result, "success"),
		IP:            stringPtr(request.IP),
		UserAgent:     stringPtr(textutil.TrimTo(request.UserAgent, 500)),
		Remark:        stringPtr(textutil.TrimTo(remark, 500)),
	})
}

func (r *Recorder) SecurityNoTx(ctx context.Context, user UserSnapshot, sessionID string, action string, result string, remark string) error {
	return r.Security(ctx, nil, user, sessionID, action, result, remark)
}

func (r *Recorder) Business(ctx context.Context, tx *gorm.DB, user UserSnapshot, module string, action string, objectType string, objectID string, summary string) error {
	if r == nil || user.ID == nil || *user.ID == 0 {
		return nil
	}
	request := requestcontext.RequestContextFrom(ctx)
	return r.logs.CreateUserBusinessLog(ctx, tx, &mysqllogs.UserBusinessLog{
		UserID:        *user.ID,
		Username:      trimPtr(user.Username, 64),
		Email:         trimPtr(user.Email, 191),
		RequestID:     stringPtr(request.RequestID),
		RequestMethod: stringPtr(request.RequestMethod),
		RequestPath:   stringPtr(request.RequestPath),
		Module:        strings.TrimSpace(module),
		Action:        strings.TrimSpace(action),
		ObjectType:    strings.TrimSpace(objectType),
		ObjectID:      stringPtr(objectID),
		Summary:       stringPtr(textutil.TrimTo(summary, 500)),
		IP:            stringPtr(request.IP),
		UserAgent:     stringPtr(textutil.TrimTo(request.UserAgent, 500)),
	})
}

func (r *Recorder) BusinessNoTx(ctx context.Context, user UserSnapshot, module string, action string, objectType string, objectID string, summary string) error {
	return r.Business(ctx, nil, user, module, action, objectType, objectID, summary)
}

func (r *Recorder) FrontendError(ctx context.Context, tx *gorm.DB, input FrontendErrorInput) error {
	if r == nil {
		return nil
	}
	request := requestcontext.RequestContextFrom(ctx)
	sourceApp := strings.TrimSpace(input.SourceApp)
	if sourceApp == "" {
		sourceApp = "web"
	}
	return r.logs.CreateFrontendErrorLog(ctx, tx, &mysqllogs.FrontendErrorLog{
		SourceApp:    sourceApp,
		UserID:       input.UserID,
		AdminID:      input.AdminID,
		RequestID:    stringPtr(sanitizeFrontendLogText(firstNonEmpty(input.RequestID, request.RequestID), 64)),
		PagePath:     sanitizeFrontendLogURL(firstNonEmpty(input.PagePath, request.RequestPath), 255),
		ErrorType:    sanitizeFrontendLogText(firstNonEmpty(input.ErrorType, "unknown"), 64),
		Message:      sanitizeFrontendLogText(firstNonEmpty(input.Message, "前端错误"), 500),
		Stack:        stringPtr(sanitizeFrontendLogText(input.Stack, 5000)),
		APIPath:      stringPtr(sanitizeFrontendLogURL(input.APIPath, 255)),
		HTTPStatus:   input.HTTPStatus,
		BusinessCode: input.BusinessCode,
		Browser:      stringPtr(sanitizeFrontendLogText(input.Browser, 255)),
		OS:           stringPtr(sanitizeFrontendLogText(input.OS, 255)),
		AppVersion:   stringPtr(sanitizeFrontendLogText(input.AppVersion, 64)),
		IP:           stringPtr(firstNonEmpty(input.IP, request.IP)),
		UserAgent:    stringPtr(sanitizeFrontendLogText(firstNonEmpty(input.UserAgent, request.UserAgent), 500)),
	})
}

func (r *Recorder) BackendRuntime(ctx context.Context, tx *gorm.DB, input mysqllogs.BackendRuntimeLog) error {
	if r == nil {
		return nil
	}
	return r.logs.CreateBackendRuntimeLog(ctx, tx, &input)
}

func Snapshot(id uint64, username string, email string) UserSnapshot {
	return UserSnapshot{
		ID:       uint64Ptr(id),
		Username: stringPtr(username),
		Email:    stringPtr(email),
	}
}

func SnapshotPtr(id *uint64, username string, email string) UserSnapshot {
	return UserSnapshot{
		ID:       id,
		Username: stringPtr(username),
		Email:    stringPtr(email),
	}
}

func uint64Ptr(value uint64) *uint64 {
	if value == 0 {
		return nil
	}
	return &value
}

func stringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func trimPtr(value *string, limit int) *string {
	if value == nil {
		return nil
	}
	return stringPtr(textutil.TrimTo(*value, limit))
}

func sanitizeFrontendLogText(value string, limit int) string {
	return textutil.TrimTo(redactFrontendLogSensitive(value), limit)
}

func sanitizeFrontendLogURL(value string, limit int) string {
	redacted := redactFrontendLogURLValues(strings.TrimSpace(value))
	return sanitizeFrontendLogText(redacted, limit)
}

func redactFrontendLogSensitive(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	for _, pattern := range sensitiveFrontendTextPatterns {
		value = pattern.ReplaceAllString(value, "${1}"+frontendLogMaskedValue)
	}
	return value
}

func redactFrontendLogURLValues(value string) string {
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.RawQuery == "" {
		return value
	}
	query := parsed.Query()
	changed := false
	for key := range query {
		if isSensitiveFrontendLogKey(key) {
			query.Set(key, frontendLogMaskedValue)
			changed = true
		}
	}
	if !changed {
		return value
	}
	parsed.RawQuery = query.Encode()
	return strings.ReplaceAll(parsed.String(), url.QueryEscape(frontendLogMaskedValue), frontendLogMaskedValue)
}

func isSensitiveFrontendLogKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")
	if normalized == "" {
		return false
	}
	sensitiveParts := []string{
		"password",
		"passwd",
		"pwd",
		"token",
		"jwt",
		"secret",
		"captcha",
		"authorization",
		"credential",
		"cookie",
		"config_value",
		"configvalue",
		"api_key",
		"apikey",
		"access_key",
		"accesskey",
		"secret_key",
		"secretkey",
	}
	for _, part := range sensitiveParts {
		if strings.Contains(normalized, part) {
			return true
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
