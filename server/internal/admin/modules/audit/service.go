package audit

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/support"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
)

const (
	defaultPage    = 1
	defaultPerPage = 15
	maxPerPage     = 100

	adminAuditMaskedValue = "[已脱敏]"
)

/**
 * AdminAuditService 处理普通审计日志写入和查询。
 */
type AdminAuditService struct {
	db *gorm.DB
}

/**
 * NewAdminAuditService 创建后台审计服务。
 *
 * @param db 数据库连接
 * @return *AdminAuditService 后台审计服务
 */
func NewAdminAuditService(db *gorm.DB) *AdminAuditService {
	return &AdminAuditService{db: db}
}

/**
 * AdminAuditWriteInput 表示普通审计日志写入参数。
 */
type AdminAuditWriteInput struct {
	AdminID          *uint64
	AdminUsername    string
	AdminDisplayName string
	SessionID        string
	RequestID        string
	RequestMethod    string
	RequestPath      string
	Action           string
	ObjectType       string
	ObjectID         string
	BeforeData       any
	AfterData        any
	IP               string
	UserAgent        string
	Remark           string
}

/**
 * Record 写入普通后台审计日志。
 *
 * @param ctx 请求上下文
 * @param db 可选数据库连接或事务，传 nil 时使用服务默认连接
 * @param input 审计写入参数
 * @return error 写入失败原因
 */
func (s *AdminAuditService) Record(ctx context.Context, db *gorm.DB, input AdminAuditWriteInput) error {
	targetDB := s.auditDB(db)
	audit, err := buildAdminAuditLog(ctx, input)
	if err != nil {
		return err
	}
	return targetDB.WithContext(ctx).Create(&audit).Error
}

/**
 * AuditLogs 分页查询普通审计日志。
 *
 * @param ctx 请求上下文
 * @param query 查询参数
 * @return admin.PageResponse[admin.AuditLogItem] 分页结果
 * @return error 查询失败原因
 */
func (s *AdminAuditService) AuditLogs(ctx context.Context, query admindto.AuditLogListQuery, includeSensitive bool) (admindto.PageResponse[admindto.AuditLogItem], error) {
	page, perPage := support.NormalizePage(query.Page, query.PerPage)
	db := s.db.WithContext(ctx).Table("admin_audit_logs")
	db, err := applyLogFilters(db, query.AdminID, query.Action, query.ObjectType, query.ObjectID, query.DateFrom, query.DateTo)
	if err != nil {
		return admindto.PageResponse[admindto.AuditLogItem]{}, err
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return admindto.PageResponse[admindto.AuditLogItem]{}, err
	}

	var rows []auditLogRow
	if err := db.
		Select(`admin_audit_logs.*,
			COALESCE(admin_audit_logs.admin_username, admin_users.username) AS actor_username,
			COALESCE(admin_audit_logs.admin_display_name, admin_users.display_name) AS actor_display_name,
			admin_users.email AS admin_email`).
		Joins("LEFT JOIN admin_users ON admin_users.id = admin_audit_logs.admin_id").
		Order("admin_audit_logs.id DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&rows).Error; err != nil {
		return admindto.PageResponse[admindto.AuditLogItem]{}, err
	}

	items := make([]admindto.AuditLogItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.auditItem(includeSensitive))
	}
	return support.PageResponse(items, total, page, perPage), nil
}

func (s *AdminAuditService) auditDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return s.db
}

func buildAdminAuditLog(ctx context.Context, input AdminAuditWriteInput) (models.AdminAuditLog, error) {
	input = mergeAuditRequestContext(ctx, input)
	action := strings.TrimSpace(input.Action)
	objectType := strings.TrimSpace(input.ObjectType)
	if action == "" || objectType == "" {
		return models.AdminAuditLog{}, errors.New("审计日志动作和对象类型不能为空")
	}

	return models.AdminAuditLog{
		AdminID:          input.AdminID,
		AdminUsername:    textutil.StringPtr(input.AdminUsername),
		AdminDisplayName: textutil.StringPtr(input.AdminDisplayName),
		SessionID:        textutil.StringPtr(input.SessionID),
		RequestID:        textutil.StringPtr(input.RequestID),
		RequestMethod:    textutil.StringPtr(input.RequestMethod),
		RequestPath:      textutil.StringPtr(input.RequestPath),
		Action:           action,
		ObjectType:       objectType,
		ObjectID:         textutil.StringPtr(input.ObjectID),
		BeforeData:       auditJSONPtr(input.BeforeData),
		AfterData:        auditJSONPtr(input.AfterData),
		IP:               textutil.StringPtr(input.IP),
		UserAgent:        textutil.StringPtr(textutil.TrimTo(input.UserAgent, 500)),
		Remark:           textutil.StringPtr(input.Remark),
	}, nil
}

func mergeAuditRequestContext(ctx context.Context, input AdminAuditWriteInput) AdminAuditWriteInput {
	request := RequestContextFrom(ctx)
	if input.AdminID == nil {
		input.AdminID = request.AdminID
	}
	if strings.TrimSpace(input.AdminUsername) == "" {
		input.AdminUsername = request.AdminUsername
	}
	if strings.TrimSpace(input.AdminDisplayName) == "" {
		input.AdminDisplayName = request.AdminDisplayName
	}
	if strings.TrimSpace(input.SessionID) == "" {
		input.SessionID = request.SessionID
	}
	if strings.TrimSpace(input.RequestID) == "" {
		input.RequestID = request.RequestID
	}
	if strings.TrimSpace(input.RequestMethod) == "" {
		input.RequestMethod = request.RequestMethod
	}
	if strings.TrimSpace(input.RequestPath) == "" {
		input.RequestPath = request.RequestPath
	}
	if strings.TrimSpace(input.IP) == "" {
		input.IP = request.IP
	}
	if strings.TrimSpace(input.UserAgent) == "" {
		input.UserAgent = request.UserAgent
	}
	return input
}

func auditJSONPtr(value any) *string {
	if isNilAuditValue(value) {
		return nil
	}
	if raw, ok := value.(*string); ok {
		if raw == nil {
			return nil
		}
		return maskedAuditJSONStringPtr(*raw)
	}
	if raw, ok := value.(string); ok {
		return maskedAuditJSONStringPtr(raw)
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return maskedAuditJSONStringPtr(string(encoded))
}

func maskedAuditJSONStringPtr(raw string) *string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var decoded any
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return textutil.StringPtr(raw)
	}
	masked := maskAuditValue(decoded, "")
	encoded, err := json.Marshal(masked)
	if err != nil {
		return nil
	}
	return textutil.StringPtr(string(encoded))
}

func maskAuditValue(value any, key string) any {
	if isSensitiveAuditKey(key) {
		return adminAuditMaskedValue
	}

	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for itemKey, itemValue := range typed {
			result[itemKey] = maskAuditValue(itemValue, itemKey)
		}
		return result
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, maskAuditValue(item, ""))
		}
		return result
	default:
		return typed
	}
}

func isSensitiveAuditKey(key string) bool {
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
		"passphrase",
		"token",
		"jwt",
		"secret",
		"private_key",
		"api_key",
		"access_key",
		"captcha",
		"config_value",
		"credential",
	}
	for _, part := range sensitiveParts {
		if strings.Contains(normalized, part) {
			return true
		}
	}
	return false
}

func applyLogFilters(db *gorm.DB, adminID uint64, action string, objectType string, objectID string, dateFrom string, dateTo string) (*gorm.DB, error) {
	if adminID > 0 {
		db = db.Where("admin_id = ?", adminID)
	}
	if action != "" {
		db = db.Where("action = ?", action)
	}
	if objectType != "" {
		db = db.Where("object_type = ?", objectType)
	}
	if objectID != "" {
		db = db.Where("object_id = ?", objectID)
	}
	if dateFrom != "" {
		from, err := parseLogTime(dateFrom)
		if err != nil {
			return nil, apperrors.ErrValidation.WithMessage("开始时间格式错误")
		}
		db = db.Where("created_at >= ?", from)
	}
	if dateTo != "" {
		to, wholeDay, err := parseLogTimeEnd(dateTo)
		if err != nil {
			return nil, apperrors.ErrValidation.WithMessage("结束时间格式错误")
		}
		if wholeDay {
			db = db.Where("created_at < ?", to)
		} else {
			db = db.Where("created_at <= ?", to)
		}
	}
	return db, nil
}

func parseLogTime(value string) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	return time.Parse("2006-01-02", value)
}

func parseLogTimeEnd(value string) (time.Time, bool, error) {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, false, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, false, err
	}
	return parsed.Add(24 * time.Hour), true, nil
}

type auditLogRow struct {
	models.AdminAuditLog
	ActorUsername    *string `gorm:"column:actor_username"`
	ActorDisplayName *string `gorm:"column:actor_display_name"`
	AdminEmail       *string `gorm:"column:admin_email"`
}

func (row auditLogRow) auditItem(includeSensitive bool) admindto.AuditLogItem {
	beforeData := auditJSONPtr(row.BeforeData)
	afterData := auditJSONPtr(row.AfterData)
	userAgent := row.UserAgent
	if !includeSensitive {
		beforeData = nil
		afterData = nil
		userAgent = nil
	}
	return admindto.AuditLogItem{
		ID:            row.ID,
		Admin:         row.adminSummary(),
		SessionID:     row.SessionID,
		RequestID:     row.RequestID,
		RequestMethod: row.RequestMethod,
		RequestPath:   row.RequestPath,
		Action:        row.Action,
		ObjectType:    row.ObjectType,
		ObjectID:      row.ObjectID,
		BeforeData:    beforeData,
		AfterData:     afterData,
		IP:            row.IP,
		UserAgent:     userAgent,
		Remark:        row.Remark,
		CreatedAt:     row.CreatedAt,
	}
}

func (row auditLogRow) adminSummary() *admindto.AuditAdminSummary {
	if row.AdminID == nil || row.ActorUsername == nil {
		return nil
	}
	return &admindto.AuditAdminSummary{
		ID:          *row.AdminID,
		Username:    *row.ActorUsername,
		DisplayName: valueOrEmpty(row.ActorDisplayName),
		Email:       row.AdminEmail,
	}
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func isNilAuditValue(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
