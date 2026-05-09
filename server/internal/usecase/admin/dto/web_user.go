package dto

import "time"

/**
 * WebUserListQuery 表示用户端账号列表查询参数。
 */
type WebUserListQuery struct {
	Page    int    `form:"page" validate:"omitempty,min=1"`
	PerPage int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Keyword string `form:"keyword" validate:"omitempty,max=96"`
	Status  string `form:"status" validate:"omitempty,oneof=active disabled"`
}

/**
 * WebUserCreateRequest 表示创建用户端账号请求。
 */
type WebUserCreateRequest struct {
	Username    string  `json:"username" validate:"required,min=3,max=64"`
	Email       string  `json:"email" validate:"required,email,max=191"`
	DisplayName *string `json:"display_name" validate:"omitempty,min=1,max=64"`
	Password    string  `json:"password" validate:"required,min=6,max=72"`
	Status      string  `json:"status" validate:"required,oneof=active disabled"`
}

/**
 * WebUserUpdateRequest 表示更新用户端账号请求。
 */
type WebUserUpdateRequest struct {
	Email       *string `json:"email" validate:"omitempty,email,max=191"`
	DisplayName *string `json:"display_name" validate:"omitempty,min=1,max=64"`
	Status      *string `json:"status" validate:"omitempty,oneof=active disabled"`
}

/**
 * WebUserPasswordRequest 表示重置用户端账号密码请求。
 */
type WebUserPasswordRequest struct {
	Password string `json:"password" validate:"required,min=6,max=72"`
}

/**
 * WebUserItem 表示用户端账号列表项。
 */
type WebUserItem struct {
	ID          uint64    `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName *string   `json:"display_name"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

/**
 * WebUserSessionListQuery 表示用户端会话列表查询参数。
 */
type WebUserSessionListQuery struct {
	Page     int    `form:"page" validate:"omitempty,min=1"`
	PerPage  int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	UserID   uint64 `form:"user_id" validate:"omitempty,min=1"`
	Status   string `form:"status" validate:"omitempty,oneof=active revoked expired"`
	DateFrom string `form:"date_from" validate:"omitempty,max=32"`
	DateTo   string `form:"date_to" validate:"omitempty,max=32"`
}

/**
 * WebUserSessionUpdateRequest 表示用户端会话更新请求。
 */
type WebUserSessionUpdateRequest struct {
	Status string `json:"status" validate:"required,oneof=revoked"`
}

/**
 * WebUserSessionItem 表示用户端会话列表项。
 */
type WebUserSessionItem struct {
	SessionID    string      `json:"session_id"`
	User         WebUserItem `json:"user"`
	Status       string      `json:"status"`
	IssuedAt     time.Time   `json:"issued_at"`
	ExpiresAt    time.Time   `json:"expires_at"`
	LastSeenAt   *time.Time  `json:"last_seen_at"`
	LastSeenIP   *string     `json:"last_seen_ip"`
	UserAgent    *string     `json:"user_agent"`
	RevokedAt    *time.Time  `json:"revoked_at"`
	RevokeReason *string     `json:"revoke_reason"`
	CreatedAt    time.Time   `json:"created_at"`
}
