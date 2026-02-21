// internal/session/store.go
// 管理后台会话存储抽象。
//
// 为什么需要 Store：
// - Access Token（JWT）本身是无状态的，服务端无法“立即撤销”
// - 通过在服务端保存 session，可以实现：
//   1) 退出登录立即失效（服务端撤销 session）
//   2) Refresh Token 刷新（并做 refresh rotation，防止 refresh 泄露后被长期复用）
//
// 本项目按你的要求使用 Redis 作为 session 存储。
package session

import "time"

// LoginMeta 登录/刷新时附带的元信息，写入 session 便于审计与排障。
type LoginMeta struct {
	IP        string
	UserAgent string
}

// Session 会话数据（从存储中读取）。
type Session struct {
	ID         uint
	AdminUserID uint
	RefreshJTI string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
}

// Store 会话存储接口。
type Store interface {
	// Create 创建一个新会话并返回 session_id、refresh_jti、expires_at。
	Create(userID uint, expiresAt time.Time, meta LoginMeta) (sessionID uint, refreshJTI string, err error)

	// Get 获取会话数据；不存在返回 ErrNotFound。
	Get(sessionID uint) (*Session, error)

	// RotateRefreshJTI 原子旋转 refresh_jti：
	// - 只有当当前 refresh_jti 与 expected 一致时才更新为 newJTI
	// - 用于 refresh rotation，避免旧 refresh token 被重复使用
	RotateRefreshJTI(sessionID uint, expected, newJTI string, meta LoginMeta) error

	// Revoke 撤销会话（退出登录/强制下线）；不存在也应返回 ErrNotFound 或视为成功由上层决定。
	Revoke(sessionID uint) error
}

