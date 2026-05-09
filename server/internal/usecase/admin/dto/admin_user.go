package dto

import "time"

/**
 * AdminUserListQuery 表示管理员账号列表查询参数。
 */
type AdminUserListQuery struct {
	Page    int    `form:"page" validate:"omitempty,min=1"`
	PerPage int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Keyword string `form:"keyword" validate:"omitempty,max=96"`
	Status  string `form:"status" validate:"omitempty,oneof=active disabled"`
	RoleID  uint64 `form:"role_id" validate:"omitempty,min=1"`
}

/**
 * AdminUserCreateRequest 表示创建管理员账号请求。
 */
type AdminUserCreateRequest struct {
	Username    string   `json:"username" validate:"required,min=3,max=64"`
	Email       *string  `json:"email" validate:"omitempty,email,max=191"`
	DisplayName string   `json:"display_name" validate:"required,min=1,max=64"`
	Password    string   `json:"password" validate:"required,min=6,max=72"`
	Status      string   `json:"status" validate:"required,oneof=active disabled"`
	RoleIDs     []uint64 `json:"role_ids" validate:"omitempty,dive,min=1"`
}

/**
 * AdminUserUpdateRequest 表示更新管理员账号请求。
 */
type AdminUserUpdateRequest struct {
	Email       *string  `json:"email" validate:"omitempty,email,max=191"`
	DisplayName *string  `json:"display_name" validate:"omitempty,min=1,max=64"`
	Status      *string  `json:"status" validate:"omitempty,oneof=active disabled"`
	RoleIDs     []uint64 `json:"role_ids" validate:"omitempty,dive,min=1"`
}

/**
 * AdminUserPasswordRequest 表示重置管理员密码请求。
 */
type AdminUserPasswordRequest struct {
	Password string `json:"password" validate:"required,min=6,max=72"`
}

/**
 * AdminRoleSummary 表示管理员账号关联角色摘要。
 */
type AdminRoleSummary struct {
	ID   uint64 `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

/**
 * AdminUserItem 表示管理员账号列表项。
 */
type AdminUserItem struct {
	ID          uint64             `json:"id"`
	Username    string             `json:"username"`
	Email       *string            `json:"email"`
	DisplayName string             `json:"display_name"`
	Status      string             `json:"status"`
	RoleIDs     []uint64           `json:"role_ids"`
	Roles       []AdminRoleSummary `json:"roles"`
	LastLoginAt *time.Time         `json:"last_login_at"`
	LastLoginIP *string            `json:"last_login_ip"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

/**
 * AdminUserDetail 表示管理员账号详情。
 */
type AdminUserDetail struct {
	AdminUserItem
	PermissionCodes []string         `json:"permission_codes"`
	Sessions        []SessionSummary `json:"sessions"`
}
