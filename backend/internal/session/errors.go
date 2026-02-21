// internal/session/errors.go
// 会话存储层错误定义。
package session

import "errors"

var (
	// ErrNotFound 会话不存在（可能已过期、已撤销或从未创建）。
	ErrNotFound = errors.New("session not found")
	// ErrRefreshJTINotMatch refresh_jti 不匹配（常见于 refresh token 重放/并发刷新）。
	ErrRefreshJTINotMatch = errors.New("refresh jti not match")
	// ErrRevoked 会话已撤销。
	ErrRevoked = errors.New("session revoked")
)

