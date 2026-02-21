// internal/model/admin_permission.go
// 权限表模型，权限是系统操作的最小粒度。
package model

// AdminPermission 权限，如 admin:create / admin:list 等。
// 权限属于系统常量，只增不删，使用 BaseModel（无软删除）。
type AdminPermission struct {
	BaseModel
	// 权限标识，全局唯一，格式：模块:操作（如 admin:create）
	Name string `gorm:"uniqueIndex;size:64;not null" json:"name"`
	// 权限显示名，如 创建管理员
	Label string `gorm:"size:64;not null"             json:"label"`
	// 所属分组，如 admin / order / user
	Group string `gorm:"size:64;not null"             json:"group"`
}

// TableName 显式指定表名。
func (AdminPermission) TableName() string { return "admin_permissions" }
