package admin

import "time"

/**
 * AdminRoleListQuery 表示管理端角色列表查询参数。
 */
type AdminRoleListQuery struct {
	Page    int    `form:"page" validate:"omitempty,min=1"`
	PerPage int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Keyword string `form:"keyword" validate:"omitempty,max=96"`
	Status  string `form:"status" validate:"omitempty,oneof=active disabled"`
}

/**
 * AdminRoleCreateRequest 表示创建管理端角色请求。
 */
type AdminRoleCreateRequest struct {
	Code            string   `json:"code" validate:"required,min=2,max=64"`
	Name            string   `json:"name" validate:"required,min=1,max=64"`
	Description     *string  `json:"description" validate:"omitempty,max=255"`
	Status          string   `json:"status" validate:"required,oneof=active disabled"`
	PermissionCodes []string `json:"permission_codes" validate:"omitempty,dive,min=1,max=96"`
}

/**
 * AdminRoleUpdateRequest 表示更新管理端角色请求。
 */
type AdminRoleUpdateRequest struct {
	Name            *string  `json:"name" validate:"omitempty,min=1,max=64"`
	Description     *string  `json:"description" validate:"omitempty,max=255"`
	Status          *string  `json:"status" validate:"omitempty,oneof=active disabled"`
	PermissionCodes []string `json:"permission_codes" validate:"omitempty,dive,min=1,max=96"`
}

/**
 * AdminRoleItem 表示管理端角色列表项。
 */
type AdminRoleItem struct {
	ID              uint64    `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Description     *string   `json:"description"`
	Status          string    `json:"status"`
	PermissionCodes []string  `json:"permission_codes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

/**
 * AdminPermissionListQuery 表示管理端权限码查询参数。
 */
type AdminPermissionListQuery struct {
	GroupName string `form:"group_name" validate:"omitempty,max=64"`
}

/**
 * AdminPermissionItem 表示权限码列表项。
 */
type AdminPermissionItem struct {
	ID          uint64  `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	GroupName   string  `json:"group_name"`
	Description *string `json:"description"`
}

/**
 * AdminPermissionGroup 表示按分组返回的权限码。
 */
type AdminPermissionGroup struct {
	GroupName   string                `json:"group_name"`
	Permissions []AdminPermissionItem `json:"permissions"`
}
