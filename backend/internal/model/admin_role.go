// internal/model/admin_role.go
// 角色表模型，角色拥有多个权限。
package model

// AdminRole 角色，对应一组权限集合。
type AdminRole struct {
	Model
	// 角色标识，全局唯一，如 super_admin / admin
	Name string `gorm:"uniqueIndex;size:64;not null"         json:"name"`
	// 角色显示名，如 超级管理员
	Label string `gorm:"size:64;not null"                     json:"label"`
	// 角色描述
	Description string `gorm:"size:255"                       json:"description"`
	// 排序，越小越靠前
	Sort int `gorm:"default:0"                                json:"sort"`
	// 关联权限（多对多）
	Permissions []AdminPermission `gorm:"many2many:admin_role_permissions;" json:"permissions,omitempty"`
}

// TableName 显式指定表名。
func (AdminRole) TableName() string { return "admin_roles" }
