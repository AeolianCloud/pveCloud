package admin

/**
 * LoginRequest 表示管理员登录请求。
 */
type LoginRequest struct {
	Username string `json:"username" validate:"required,max=191"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

/**
 * AdminSummary 表示登录后返回给管理后台的管理员基础资料。
 */
type AdminSummary struct {
	ID          uint64  `json:"id"`
	Username    string  `json:"username"`
	Email       *string `json:"email"`
	DisplayName string  `json:"display_name"`
	Status      string  `json:"status"`
}

/**
 * LoginResponse 表示管理员登录成功响应数据。
 */
type LoginResponse struct {
	AccessToken     string       `json:"access_token"`
	TokenType       string       `json:"token_type"`
	ExpiresIn       int64        `json:"expires_in"`
	Admin           AdminSummary `json:"admin"`
	RoleIDs         []uint64     `json:"role_ids"`
	PermissionCodes []string     `json:"permission_codes"`
}
