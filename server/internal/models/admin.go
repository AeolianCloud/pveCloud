package models

import "time"

/**
 * AdminUser 映射 admin_users 管理员账号表。
 */
type AdminUser struct {
	ID           uint64     `gorm:"column:id;primaryKey"`
	Username     string     `gorm:"column:username"`
	Email        *string    `gorm:"column:email"`
	PasswordHash string     `gorm:"column:password_hash"`
	DisplayName  string     `gorm:"column:display_name"`
	Status       string     `gorm:"column:status"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
	LastLoginIP  *string    `gorm:"column:last_login_ip"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
}

/**
 * TableName 返回管理员账号表名。
 *
 * @return string 表名
 */
func (AdminUser) TableName() string {
	return "admin_users"
}

/**
 * AdminRole 映射 admin_roles 管理端角色表。
 */
type AdminRole struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	Code        string    `gorm:"column:code"`
	Name        string    `gorm:"column:name"`
	Description *string   `gorm:"column:description"`
	Status      string    `gorm:"column:status"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

/**
 * TableName 返回管理端角色表名。
 *
 * @return string 表名
 */
func (AdminRole) TableName() string {
	return "admin_roles"
}

/**
 * AdminPermission 映射 admin_permissions 管理端权限码表。
 */
type AdminPermission struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	Code        string    `gorm:"column:code"`
	Name        string    `gorm:"column:name"`
	GroupName   string    `gorm:"column:group_name"`
	Description *string   `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

/**
 * TableName 返回管理端权限码表名。
 *
 * @return string 表名
 */
func (AdminPermission) TableName() string {
	return "admin_permissions"
}

/**
 * AdminSession 映射 admin_sessions 管理端登录会话表。
 */
type AdminSession struct {
	ID           uint64     `gorm:"column:id;primaryKey"`
	SessionID    string     `gorm:"column:session_id"`
	AdminID      uint64     `gorm:"column:admin_id"`
	Status       string     `gorm:"column:status"`
	IssuedAt     time.Time  `gorm:"column:issued_at"`
	ExpiresAt    time.Time  `gorm:"column:expires_at"`
	LastSeenAt   *time.Time `gorm:"column:last_seen_at"`
	LastSeenIP   *string    `gorm:"column:last_seen_ip"`
	UserAgent    *string    `gorm:"column:user_agent"`
	RevokedAt    *time.Time `gorm:"column:revoked_at"`
	RevokeReason *string    `gorm:"column:revoke_reason"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

/**
 * TableName 返回管理端登录会话表名。
 *
 * @return string 表名
 */
func (AdminSession) TableName() string {
	return "admin_sessions"
}

/**
 * AdminAuditLog 映射 admin_audit_logs 后台操作审计表。
 */
type AdminAuditLog struct {
	ID         uint64    `gorm:"column:id;primaryKey"`
	AdminID    *uint64   `gorm:"column:admin_id"`
	Action     string    `gorm:"column:action"`
	ObjectType string    `gorm:"column:object_type"`
	ObjectID   *string   `gorm:"column:object_id"`
	BeforeData *string   `gorm:"column:before_data"`
	AfterData  *string   `gorm:"column:after_data"`
	IP         *string   `gorm:"column:ip"`
	UserAgent  *string   `gorm:"column:user_agent"`
	Remark     *string   `gorm:"column:remark"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

/**
 * TableName 返回后台操作审计表名。
 *
 * @return string 表名
 */
func (AdminAuditLog) TableName() string {
	return "admin_audit_logs"
}
