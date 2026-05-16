package logging

import (
	"context"
	"strings"

	"gorm.io/gorm"

	mysqllogs "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/logs"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/requestcontext"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
)

type Recorder struct {
	logs *mysqllogs.Repository
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
		RequestID:    stringPtr(firstNonEmpty(input.RequestID, request.RequestID)),
		PagePath:     firstNonEmpty(input.PagePath, request.RequestPath),
		ErrorType:    firstNonEmpty(input.ErrorType, "unknown"),
		Message:      textutil.TrimTo(firstNonEmpty(input.Message, "前端错误"), 500),
		Stack:        stringPtr(textutil.TrimTo(input.Stack, 5000)),
		APIPath:      stringPtr(textutil.TrimTo(input.APIPath, 255)),
		HTTPStatus:   input.HTTPStatus,
		BusinessCode: input.BusinessCode,
		Browser:      stringPtr(textutil.TrimTo(input.Browser, 255)),
		OS:           stringPtr(textutil.TrimTo(input.OS, 255)),
		AppVersion:   stringPtr(textutil.TrimTo(input.AppVersion, 64)),
		IP:           stringPtr(firstNonEmpty(input.IP, request.IP)),
		UserAgent:    stringPtr(textutil.TrimTo(firstNonEmpty(input.UserAgent, request.UserAgent), 500)),
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
