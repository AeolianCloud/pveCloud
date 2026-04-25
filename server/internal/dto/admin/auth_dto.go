package admin

import "time"

/**
 * LoginRequest 表示管理员登录请求。
 */
type LoginRequest struct {
	Username    string `json:"username" validate:"required,max=191"`
	Password    string `json:"password" validate:"required,min=6,max=72"`
	CaptchaID   string `json:"captcha_id" validate:"required,min=16,max=128"`
	CaptchaCode string `json:"captcha_code" validate:"required,min=4,max=8"`
}

/**
 * LoginCaptchaResponse 表示管理员登录验证码响应数据。
 */
type LoginCaptchaResponse struct {
	CaptchaID string `json:"captcha_id"`
	Image     string `json:"image"`
	ExpiresIn int64  `json:"expires_in"`
}

/**
 * AdminSummary 表示登录后返回给管理后台的管理员基础资料。
 */
type AdminSummary struct {
	ID          uint64  `json:"id"`
	Username    string  `json:"username"`
	Email       *string `json:"email"`
	DisplayName string  `json:"display_name"`
	Status      string  `json:"status"`
}

/**
 * SessionSummary 表示当前管理端登录会话摘要。
 */
type SessionSummary struct {
	SessionID string    `json:"session_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

/**
 * AuthStateResponse 表示当前管理员认证态响应数据。
 */
type AuthStateResponse struct {
	Admin           AdminSummary   `json:"admin"`
	RoleIDs         []uint64       `json:"role_ids"`
	PermissionCodes []string       `json:"permission_codes"`
	Menus           []MenuItem     `json:"menus"`
	Session         SessionSummary `json:"session"`
}

/**
 * LoginResponse 表示管理员登录成功响应数据。
 */
type LoginResponse struct {
	AccessToken     string         `json:"access_token"`
	TokenType       string         `json:"token_type"`
	ExpiresIn       int64          `json:"expires_in"`
	Admin           AdminSummary   `json:"admin"`
	RoleIDs         []uint64       `json:"role_ids"`
	PermissionCodes []string       `json:"permission_codes"`
	Session         SessionSummary `json:"session"`
}
