package dto

import "time"

/**
 * AdminSessionListQuery 表示管理员会话列表查询参数。
 */
type AdminSessionListQuery struct {
	Page    int    `form:"page" validate:"omitempty,min=1"`
	PerPage int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Keyword string `form:"keyword" validate:"omitempty,max=128"`
	Status  string `form:"status" validate:"omitempty,oneof=active revoked expired"`
}

/**
 * AdminSessionUpdateRequest 表示管理员会话更新请求。
 */
type AdminSessionUpdateRequest struct {
	Status string `json:"status" validate:"required,oneof=revoked"`
}

/**
 * AdminSessionItem 表示管理员会话列表项。
 */
type AdminSessionItem struct {
	SessionID        string     `json:"session_id"`
	AdminID          uint64     `json:"admin_id"`
	AdminUsername    string     `json:"admin_username"`
	AdminDisplayName string     `json:"admin_display_name"`
	AdminEmail       *string    `json:"admin_email"`
	Status           string     `json:"status"`
	IssuedAt         time.Time  `json:"issued_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	LastSeenAt       *time.Time `json:"last_seen_at"`
	LastSeenIP       *string    `json:"last_seen_ip"`
	UserAgent        *string    `json:"user_agent"`
	RevokedAt        *time.Time `json:"revoked_at"`
	RevokeReason     *string    `json:"revoke_reason"`
	IsCurrent        bool       `json:"is_current"`
}
