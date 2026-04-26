package admin

import "time"

/**
 * AdminSessionListQuery 表示管理端会话列表查询参数。
 */
type AdminSessionListQuery struct {
	Page    int    `form:"page" validate:"omitempty,min=1"`
	PerPage int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	AdminID uint64 `form:"admin_id" validate:"omitempty,min=1"`
	Status  string `form:"status" validate:"omitempty,oneof=active revoked expired"`
	Keyword string `form:"keyword" validate:"omitempty,max=96"`
}

/**
 * AdminSessionItem 表示管理端会话列表项。
 */
type AdminSessionItem struct {
	ID           uint64             `json:"id"`
	SessionID    string             `json:"session_id"`
	Admin        *AuditAdminSummary `json:"admin"`
	Status       string             `json:"status"`
	IssuedAt     time.Time          `json:"issued_at"`
	ExpiresAt    time.Time          `json:"expires_at"`
	LastSeenAt   *time.Time         `json:"last_seen_at"`
	LastSeenIP   *string            `json:"last_seen_ip"`
	UserAgent    *string            `json:"user_agent"`
	RevokedAt    *time.Time         `json:"revoked_at"`
	RevokeReason *string            `json:"revoke_reason"`
}
