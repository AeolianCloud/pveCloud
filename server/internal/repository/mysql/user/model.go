package user

import "time"

/**
 * User 映射 users 用户端账号表。
 */
type User struct {
	ID           uint64    `gorm:"column:id;primaryKey"`
	Username     string    `gorm:"column:username"`
	Email        string    `gorm:"column:email"`
	PasswordHash string    `gorm:"column:password_hash"`
	DisplayName  *string   `gorm:"column:display_name"`
	Status       string    `gorm:"column:status"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

/**
 * TableName 返回用户端账号表名。
 */
func (User) TableName() string {
	return "users"
}

/**
 * UserSession 映射 user_sessions 用户端登录会话表。
 */
type UserSession struct {
	ID           uint64     `gorm:"column:id;primaryKey"`
	UserID       uint64     `gorm:"column:user_id"`
	SessionID    string     `gorm:"column:session_id"`
	Status       string     `gorm:"column:status"`
	IssuedAt     time.Time  `gorm:"column:issued_at"`
	ExpiresAt    time.Time  `gorm:"column:expires_at"`
	RevokedAt    *time.Time `gorm:"column:revoked_at"`
	RevokeReason *string    `gorm:"column:revoke_reason"`
	LastSeenAt   *time.Time `gorm:"column:last_seen_at"`
	LastSeenIP   *string    `gorm:"column:last_seen_ip"`
	UserAgent    *string    `gorm:"column:user_agent"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

/**
 * TableName 返回用户端登录会话表名。
 */
func (UserSession) TableName() string {
	return "user_sessions"
}

/**
 * UserPasswordResetToken 映射 user_password_reset_tokens 用户端密码重置 Token 表。
 */
type UserPasswordResetToken struct {
	ID          uint64     `gorm:"column:id;primaryKey"`
	UserID      uint64     `gorm:"column:user_id"`
	TokenHash   string     `gorm:"column:token_hash"`
	Status      string     `gorm:"column:status"`
	ExpiresAt   time.Time  `gorm:"column:expires_at"`
	UsedAt      *time.Time `gorm:"column:used_at"`
	RequestedIP *string    `gorm:"column:requested_ip"`
	UserAgent   *string    `gorm:"column:user_agent"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
}

/**
 * TableName 返回用户端密码重置 Token 表名。
 */
func (UserPasswordResetToken) TableName() string {
	return "user_password_reset_tokens"
}
