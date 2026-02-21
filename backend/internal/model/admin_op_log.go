// internal/model/admin_op_log.go
// 管理员操作日志表，记录所有增删改操作的执行人、目标和结果。
package model

import "time"

// AdminOpLog 操作日志，不做软删除，保留完整审计历史。
type AdminOpLog struct {
	ID uint `gorm:"primarykey" json:"id"`
	// 操作人 ID（账号删除后日志仍保留）
	AdminUserID uint `gorm:"index;not null" json:"admin_user_id"`
	// 冗余存储操作人用户名
	Username string `gorm:"size:64;not null" json:"username"`
	// 操作模块，如 admin / role / permission
	Module string `gorm:"size:64;not null;index" json:"module"`
	// 操作动作，如 create / update / delete / set_status / assign_permissions
	Action string `gorm:"size:64;not null" json:"action"`
	// 操作目标 ID（如被修改的管理员 ID）
	TargetID uint `gorm:"default:0" json:"target_id"`
	// 操作目标描述（冗余记录，如被删除管理员的用户名）
	TargetLabel string `gorm:"size:128" json:"target_label"`
	// 操作结果：1 成功  0 失败
	Status int8 `gorm:"not null;default:1" json:"status"`
	// 操作来源 IP
	IP string `gorm:"size:64" json:"ip"`
	// 操作时间
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

// TableName 显式指定表名。
func (AdminOpLog) TableName() string { return "admin_op_logs" }
