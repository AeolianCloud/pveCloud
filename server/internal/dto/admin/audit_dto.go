package admin

import "time"

/**
 * AuditAdminSummary 表示审计记录中展示的管理员摘要。
 */
type AuditAdminSummary struct {
	ID          uint64  `json:"id"`
	Username    string  `json:"username"`
	DisplayName string  `json:"display_name"`
	Email       *string `json:"email"`
}

/**
 * AuditLogListQuery 表示审计日志列表查询参数。
 */
type AuditLogListQuery struct {
	Page       int    `form:"page" validate:"omitempty,min=1"`
	PerPage    int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	AdminID    uint64 `form:"admin_id" validate:"omitempty,min=1"`
	Action     string `form:"action" validate:"omitempty,max=96"`
	ObjectType string `form:"object_type" validate:"omitempty,max=64"`
	ObjectID   string `form:"object_id" validate:"omitempty,max=64"`
	DateFrom   string `form:"date_from" validate:"omitempty,max=32"`
	DateTo     string `form:"date_to" validate:"omitempty,max=32"`
}

/**
 * RiskLogListQuery 表示高危操作日志列表查询参数。
 */
type RiskLogListQuery struct {
	Page       int    `form:"page" validate:"omitempty,min=1"`
	PerPage    int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	AdminID    uint64 `form:"admin_id" validate:"omitempty,min=1"`
	RiskLevel  string `form:"risk_level" validate:"omitempty,oneof=medium high critical"`
	Action     string `form:"action" validate:"omitempty,max=96"`
	ObjectType string `form:"object_type" validate:"omitempty,max=64"`
	ObjectID   string `form:"object_id" validate:"omitempty,max=64"`
	DateFrom   string `form:"date_from" validate:"omitempty,max=32"`
	DateTo     string `form:"date_to" validate:"omitempty,max=32"`
}

/**
 * AuditLogItem 表示普通审计日志列表项。
 */
type AuditLogItem struct {
	ID         uint64             `json:"id"`
	Admin      *AuditAdminSummary `json:"admin"`
	Action     string             `json:"action"`
	ObjectType string             `json:"object_type"`
	ObjectID   *string            `json:"object_id"`
	BeforeData *string            `json:"before_data"`
	AfterData  *string            `json:"after_data"`
	IP         *string            `json:"ip"`
	UserAgent  *string            `json:"user_agent"`
	Remark     *string            `json:"remark"`
	CreatedAt  time.Time          `json:"created_at"`
}

/**
 * RiskLogItem 表示高危操作日志列表项。
 */
type RiskLogItem struct {
	ID         uint64             `json:"id"`
	AuditLogID *uint64            `json:"audit_log_id"`
	Admin      *AuditAdminSummary `json:"admin"`
	RiskLevel  string             `json:"risk_level"`
	Action     string             `json:"action"`
	ObjectType string             `json:"object_type"`
	ObjectID   *string            `json:"object_id"`
	RiskReason string             `json:"risk_reason"`
	BeforeData *string            `json:"before_data"`
	AfterData  *string            `json:"after_data"`
	IP         *string            `json:"ip"`
	UserAgent  *string            `json:"user_agent"`
	Remark     *string            `json:"remark"`
	CreatedAt  time.Time          `json:"created_at"`
}

/**
 * PageResponse 表示通用分页响应。
 */
type PageResponse[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PerPage  int   `json:"per_page"`
	LastPage int   `json:"last_page"`
}
