package dto

import "time"

/**
 * LoginRequest 表示用户端登录请求。
 */
type LoginRequest struct {
	Account     string `json:"account" validate:"required,max=191"`
	Password    string `json:"password" validate:"required,min=6,max=72"`
	CaptchaID   string `json:"captcha_id" validate:"omitempty,min=16,max=128"`
	CaptchaCode string `json:"captcha_code" validate:"omitempty,min=4,max=8"`
}

/**
 * RegisterRequest 表示用户端注册请求。
 */
type RegisterRequest struct {
	Username    string  `json:"username" validate:"required,min=3,max=64"`
	Email       string  `json:"email" validate:"required,email,max=191"`
	Password    string  `json:"password" validate:"required,min=6,max=72"`
	DisplayName *string `json:"display_name" validate:"omitempty,max=64"`
	CaptchaID   string  `json:"captcha_id" validate:"omitempty,min=16,max=128"`
	CaptchaCode string  `json:"captcha_code" validate:"omitempty,min=4,max=8"`
}

/**
 * PasswordResetRequest 表示用户端密码找回申请请求。
 */
type PasswordResetRequest struct {
	Email       string `json:"email" validate:"required,email,max=191"`
	CaptchaID   string `json:"captcha_id" validate:"omitempty,min=16,max=128"`
	CaptchaCode string `json:"captcha_code" validate:"omitempty,min=4,max=8"`
}

/**
 * PasswordResetConfirmRequest 表示用户端密码重置确认请求。
 */
type PasswordResetConfirmRequest struct {
	Token       string `json:"token" validate:"required,max=256"`
	Password    string `json:"password" validate:"required,min=6,max=72"`
	CaptchaID   string `json:"captcha_id" validate:"omitempty,min=16,max=128"`
	CaptchaCode string `json:"captcha_code" validate:"omitempty,min=4,max=8"`
}

/**
 * CaptchaResponse 表示用户端认证验证码响应数据。
 */
type CaptchaResponse struct {
	CaptchaID string `json:"captcha_id"`
	Image     string `json:"image"`
	ExpiresIn int64  `json:"expires_in"`
}

/**
 * UpdateProfileRequest 表示当前用户资料编辑请求。
 */
type UpdateProfileRequest struct {
	Email       string  `json:"email" validate:"required,email,max=191"`
	DisplayName *string `json:"display_name" validate:"omitempty,max=64"`
}

/**
 * ChangePasswordRequest 表示当前用户修改密码请求。
 */
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=6,max=72"`
	Password        string `json:"password" validate:"required,min=6,max=72"`
}

/**
 * UserSummary 表示用户端账号摘要。
 */
type UserSummary struct {
	ID          uint64  `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
	Status      string  `json:"status"`
}

/**
 * SessionSummary 表示用户端登录会话摘要。
 */
type SessionSummary struct {
	SessionID string    `json:"session_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

/**
 * AuthStateResponse 表示用户端当前认证态。
 */
type AuthStateResponse struct {
	User    UserSummary    `json:"user"`
	Session SessionSummary `json:"session"`
}

/**
 * LoginResponse 表示用户端登录或刷新响应。
 */
type LoginResponse struct {
	AccessToken string         `json:"access_token"`
	TokenType   string         `json:"token_type"`
	ExpiresIn   int64          `json:"expires_in"`
	User        UserSummary    `json:"user"`
	Session     SessionSummary `json:"session"`
}
