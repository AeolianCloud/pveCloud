// internal/model/admin_menu.go
// 管理后台菜单表模型（动态下发 + 可配置）。
//
// 设计说明（尽量写清楚，避免后续维护时“以为自己懂了”）：
// 1) 菜单是一套全局结构，不按角色/租户分多套；对不同用户的差异通过后端“按权限裁剪”实现。
// 2) 使用 parent_id 表达树结构，使用 sort 表达兄弟节点排序（值越小越靠前）。
// 3) visible 控制是否显示；super_admin_only 控制是否仅超级管理员可见。
// 4) permission 仅用于“可见性裁剪”，不等同于接口鉴权：
//    - 可见性：前端侧边栏是否展示该入口（由后端裁剪后下发）。
//    - 鉴权：接口是否允许访问（仍由后端中间件/路由保护，避免“能看到就能用”的误解）。
package model

// AdminMenu 菜单节点。
//
// 注意：
// - 本表需要软删除（Model），因为菜单是可配置数据，误删后需要可恢复。
// - path 为 NULL 表示目录节点（仅用于分组/展开），目录节点不直接跳转路由。
type AdminMenu struct {
	Model

	// ParentID 父菜单 ID，0 表示顶级菜单。
	ParentID uint `gorm:"index;not null;default:0" json:"parent_id"`

	// Title 菜单标题（侧边栏显示）。
	Title string `gorm:"size:64;not null" json:"title"`

	// Path 前端路由路径（以 / 开头）。目录节点可为 NULL。
	// 说明：本项目菜单是“路由驱动”，path 变化会影响前端跳转与高亮逻辑。
	Path *string `gorm:"size:128;uniqueIndex" json:"path"`

	// Permission 可见权限标识（如 admin:list / role:list）。
	// 为空表示任何已登录用户都可见（仍需后端鉴权保证接口安全）。
	Permission *string `gorm:"size:64" json:"permission"`

	// SuperAdminOnly 是否仅超级管理员可见：1 是 0 否。
	SuperAdminOnly int8 `gorm:"not null;default:0" json:"super_admin_only"`

	// Icon 菜单图标标识（前端按约定映射到具体 icon 组件）。
	Icon *string `gorm:"size:64" json:"icon"`

	// Sort 排序权重：值越小越靠前。
	Sort int `gorm:"not null;default:0" json:"sort"`

	// Visible 是否显示：1 显示 0 隐藏。
	Visible int8 `gorm:"not null;default:1" json:"visible"`
}

// TableName 显式指定表名。
func (AdminMenu) TableName() string { return "admin_menus" }

