// internal/model/admin_login_log.go
// 管理员登录日志表，记录每次登录的结果。
package model

import "time"

// AdminLoginLog 登录日志，不做软删除，保留完整历史。
type AdminLoginLog struct {
	// 使用独立 ID，不嵌入 Model（不需要 updated_at / deleted_at）
	ID uint `gorm:"primarykey"           json:"id"`
	// 关联的管理员 ID（管理员删除后日志仍保留，冗余 username）
	AdminUserID uint `gorm:"index;not null" json:"admin_user_id"`
	// 冗余存储用户名，防止账号删除后日志丢失上下文
	Username string `gorm:"size:64;not null" json:"username"`
	// 登录 IP
	IP string `gorm:"size:64"              json:"ip"`
	// 浏览器 User-Agent
	UserAgent string `gorm:"size:255"       json:"user_agent"`
	// 登录结果：1 成功  0 失败
	Status int8 `gorm:"not null"           json:"status"`
	// 失败备注，如 密码错误 / 账号禁用
	Remark string `gorm:"size:128"         json:"remark"`
	// 登录时间
	CreatedAt time.Time `gorm:"index"      json:"created_at"`
}

// TableName 显式指定表名。
func (AdminLoginLog) TableName() string { return "admin_login_logs" }
