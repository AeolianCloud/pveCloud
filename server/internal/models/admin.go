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
