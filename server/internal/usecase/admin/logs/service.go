package logs

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"

	mysqllogs "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/logs"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	adminsupport "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/support"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
)

const (
	logSourceAdmin = "admin"
	logSourceWeb   = "web"
)

type Service struct {
	db      *gorm.DB
	logs    *mysqllogs.Repository
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db:    db,
		logs:  mysqllogs.NewRepository(db),
	}
}

func (s *Service) UserSecurityLogs(ctx context.Context, query admindto.UserSecurityLogQuery) (admindto.PageResponse[admindto.UserSecurityLogItem], error) {
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	filters := mysqllogs.UserSecurityFilters{
		UserID:    query.UserID,
		Username:  query.Username,
		Action:    query.Action,
		Result:    query.Result,
		RequestID: query.RequestID,
		IP:        query.IP,
	}
	if err := fillTimeRange(&filters.DateFrom, &filters.DateTo, query.DateFrom, query.DateTo, true); err != nil {
		return admindto.PageResponse[admindto.UserSecurityLogItem]{}, err
	}
	rows, total, err := s.logs.UserSecurityLogs(ctx, filters, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.UserSecurityLogItem]{}, err
	}
	items := make([]admindto.UserSecurityLogItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, userSecurityItem(row))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func (s *Service) UserBusinessLogs(ctx context.Context, query admindto.UserBusinessLogQuery) (admindto.PageResponse[admindto.UserBusinessLogItem], error) {
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	filters := mysqllogs.UserBusinessFilters{
		UserID:     query.UserID,
		Module:     query.Module,
		Action:     query.Action,
		ObjectType: query.ObjectType,
		ObjectID:   query.ObjectID,
		RequestID:  query.RequestID,
	}
	if err := fillTimeRange(&filters.DateFrom, &filters.DateTo, query.DateFrom, query.DateTo, true); err != nil {
		return admindto.PageResponse[admindto.UserBusinessLogItem]{}, err
	}
	rows, total, err := s.logs.UserBusinessLogs(ctx, filters, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.UserBusinessLogItem]{}, err
	}
	items := make([]admindto.UserBusinessLogItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, userBusinessItem(row))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func (s *Service) FrontendErrorLogs(ctx context.Context, query admindto.FrontendErrorLogQuery) (admindto.PageResponse[admindto.FrontendErrorLogItem], error) {
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	filters := mysqllogs.FrontendErrorFilters{
		SourceApp:  query.SourceApp,
		PagePath:   query.PagePath,
		ErrorType:  query.ErrorType,
		APIPath:    query.APIPath,
		HTTPStatus: query.HTTPStatus,
		RequestID:  query.RequestID,
	}
	if err := fillTimeRange(&filters.DateFrom, &filters.DateTo, query.DateFrom, query.DateTo, true); err != nil {
		return admindto.PageResponse[admindto.FrontendErrorLogItem]{}, err
	}
	rows, total, err := s.logs.FrontendErrorLogs(ctx, filters, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.FrontendErrorLogItem]{}, err
	}
	items := make([]admindto.FrontendErrorLogItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, frontendErrorItem(row))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func (s *Service) BackendRuntimeLogs(ctx context.Context, query admindto.BackendRuntimeLogQuery) (admindto.PageResponse[admindto.BackendRuntimeLogItem], error) {
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	filters := mysqllogs.BackendRuntimeFilters{
		Level:       query.Level,
		Category:    query.Category,
		Status:      query.Status,
		RequestID:   query.RequestID,
		RequestPath: query.RequestPath,
	}
	if err := fillTimeRange(&filters.DateFrom, &filters.DateTo, query.DateFrom, query.DateTo, true); err != nil {
		return admindto.PageResponse[admindto.BackendRuntimeLogItem]{}, err
	}
	rows, total, err := s.logs.BackendRuntimeLogs(ctx, filters, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.BackendRuntimeLogItem]{}, err
	}
	items := make([]admindto.BackendRuntimeLogItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, backendRuntimeItem(row))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func (s *Service) RecordUserSecurityLog(ctx context.Context, db *gorm.DB, input mysqllogs.UserSecurityLog) error {
	return s.logs.CreateUserSecurityLog(ctx, db, &input)
}

func (s *Service) RecordUserBusinessLog(ctx context.Context, db *gorm.DB, input mysqllogs.UserBusinessLog) error {
	return s.logs.CreateUserBusinessLog(ctx, db, &input)
}

func (s *Service) RecordFrontendErrorLog(ctx context.Context, db *gorm.DB, input mysqllogs.FrontendErrorLog) error {
	return s.logs.CreateFrontendErrorLog(ctx, db, &input)
}

func (s *Service) RecordBackendRuntimeLog(ctx context.Context, db *gorm.DB, input mysqllogs.BackendRuntimeLog) error {
	return s.logs.CreateBackendRuntimeLog(ctx, db, &input)
}

func (s *Service) RecordExport(ctx context.Context, db *gorm.DB, adminID uint64, logType string, filters any, rowCount int) error {
	raw, err := json.Marshal(filters)
	if err != nil {
		return err
	}
	return s.logs.CreateExportRecord(ctx, db, &mysqllogs.ExportRecord{
		AdminID:  adminID,
		LogType:  strings.TrimSpace(logType),
		Filters:  textutil.StringPtr(string(raw)),
		RowCount: rowCount,
	})
}

func (s *Service) Within(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return mysqltx.NewManager(s.db).WithinContext(ctx, fn)
}

func fillTimeRange(start **time.Time, end **time.Time, from string, to string, openEnd bool) error {
	if strings.TrimSpace(from) != "" {
		parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(from), time.Local)
		if err != nil {
			return apperrors.ErrValidation.WithMessage("开始日期格式错误")
		}
		*start = &parsed
	}
	if strings.TrimSpace(to) != "" {
		parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(to), time.Local)
		if err != nil {
			return apperrors.ErrValidation.WithMessage("结束日期格式错误")
		}
		if openEnd {
			parsed = parsed.Add(24 * time.Hour)
		}
		*end = &parsed
	}
	return nil
}

func userSecurityItem(row mysqllogs.UserSecurityLogRow) admindto.UserSecurityLogItem {
	return admindto.UserSecurityLogItem{
		ID:            row.ID,
		User:          buildLogUserSummary(row.UserID, row.UserUsername, row.UserEmail, row.UserDisplayName),
		SessionID:     row.SessionID,
		RequestID:     row.RequestID,
		RequestMethod: row.RequestMethod,
		RequestPath:   row.RequestPath,
		Action:        row.Action,
		Result:        row.Result,
		IP:            row.IP,
		UserAgent:     row.UserAgent,
		Remark:        row.Remark,
		CreatedAt:     row.CreatedAt,
	}
}

func userBusinessItem(row mysqllogs.UserBusinessLogRow) admindto.UserBusinessLogItem {
	return admindto.UserBusinessLogItem{
		ID:            row.ID,
		User:          derefLogUserSummary(buildLogUserSummary(&row.UserID, row.UserUsername, row.UserEmail, row.UserDisplayName)),
		RequestID:     row.RequestID,
		RequestMethod: row.RequestMethod,
		RequestPath:   row.RequestPath,
		Module:        row.Module,
		Action:        row.Action,
		ObjectType:    row.ObjectType,
		ObjectID:      row.ObjectID,
		Summary:       row.Summary,
		IP:            row.IP,
		UserAgent:     row.UserAgent,
		CreatedAt:     row.CreatedAt,
	}
}

func frontendErrorItem(row mysqllogs.FrontendErrorLog) admindto.FrontendErrorLogItem {
	return admindto.FrontendErrorLogItem{
		ID:           row.ID,
		SourceApp:    row.SourceApp,
		UserID:       row.UserID,
		AdminID:      row.AdminID,
		RequestID:    row.RequestID,
		PagePath:     row.PagePath,
		ErrorType:    row.ErrorType,
		Message:      row.Message,
		Stack:        row.Stack,
		APIPath:      row.APIPath,
		HTTPStatus:   row.HTTPStatus,
		BusinessCode: row.BusinessCode,
		Browser:      row.Browser,
		OS:           row.OS,
		AppVersion:   row.AppVersion,
		IP:           row.IP,
		UserAgent:    row.UserAgent,
		CreatedAt:    row.CreatedAt,
	}
}

func backendRuntimeItem(row mysqllogs.BackendRuntimeLog) admindto.BackendRuntimeLogItem {
	return admindto.BackendRuntimeLogItem{
		ID:            row.ID,
		Level:         row.Level,
		Category:      row.Category,
		RequestID:     row.RequestID,
		RequestMethod: row.RequestMethod,
		RequestPath:   row.RequestPath,
		Status:        row.Status,
		LatencyMS:     row.LatencyMS,
		ClientIP:      row.ClientIP,
		Message:       row.Message,
		Detail:        row.Detail,
		CreatedAt:     row.CreatedAt,
	}
}

func buildLogUserSummary(userID *uint64, username *string, email *string, displayName *string) *admindto.LogUserSummary {
	if userID == nil && username == nil && email == nil {
		return nil
	}
	return &admindto.LogUserSummary{
		ID:          valueOrZero(userID),
		Username:    valueOrEmpty(username),
		Email:       email,
		DisplayName: displayName,
	}
}

func derefLogUserSummary(summary *admindto.LogUserSummary) admindto.LogUserSummary {
	if summary == nil {
		return admindto.LogUserSummary{}
	}
	return *summary
}

func valueOrZero(v *uint64) uint64 {
	if v == nil {
		return 0
	}
	return *v
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
