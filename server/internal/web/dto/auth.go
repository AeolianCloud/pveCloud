package dto

import "time"

/**
 * LoginRequest 表示用户端登录请求。
 */
type LoginRequest struct {
	Account  string `json:"account" validate:"required,max=191"`
	Password string `json:"password" validate:"required,min=6,max=72"`
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
