// internal/model/admin_user.go
// 管理后台账号表模型。
package model

import "time"

// AdminUser 管理后台账号，通过 admin_user_roles 关联多个角色。
type AdminUser struct {
	Model
	// 登录用户名，全局唯一
	Username string `gorm:"uniqueIndex;size:64;not null"  json:"username"`
	// bcrypt 哈希密码，不对外输出
	Password string `gorm:"size:128;not null"              json:"-"`
	// 显示昵称
	Nickname string `gorm:"size:64"                        json:"nickname"`
	// 头像 URL
	Avatar string `gorm:"size:255"                         json:"avatar"`
	// 邮箱，唯一，可为 NULL（不填时存 NULL，避免多条空字符串违反唯一约束）
	Email *string `gorm:"uniqueIndex;size:128"              json:"email"`
	// 状态：1 启用  0 禁用
	Status int8 `gorm:"default:1;not null"                  json:"status"`
	// 最后登录时间
	LastLoginAt *time.Time `gorm:"column:last_login_at"     json:"last_login_at"`
	// 关联角色（多对多）
	Roles []AdminRole `gorm:"many2many:admin_user_roles;"  json:"roles,omitempty"`
}

// TableName 显式指定表名。
func (AdminUser) TableName() string { return "admin_users" }
