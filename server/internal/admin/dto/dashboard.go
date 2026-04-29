package dto

/**
 * DashboardMetric 表示管理端首页概览指标。
 */
type DashboardMetric struct {
	Key   string  `json:"key"`
	Title string  `json:"title"`
	Value int64   `json:"value"`
	Unit  *string `json:"unit"`
}

/**
 * MenuItem 表示管理端可见菜单项。
 */
type MenuItem struct {
	Key            string     `json:"key"`
	Title          string     `json:"title"`
	Path           string     `json:"path"`
	Icon           *string    `json:"icon"`
	PermissionCode *string    `json:"permission_code"`
	Children       []MenuItem `json:"children,omitempty"`
}

/**
 * DashboardResponse 表示管理端首页响应数据。
 */
type DashboardResponse struct {
	AuthStateResponse
	Metrics []DashboardMetric `json:"metrics"`
}
