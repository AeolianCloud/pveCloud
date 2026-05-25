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
 * DashboardBusinessMetric 表示管理端首页业务待办和异常指标。
 */
type DashboardBusinessMetric struct {
	Key              string  `json:"key"`
	Title            string  `json:"title"`
	Value            int64   `json:"value"`
	Unit             *string `json:"unit"`
	Description      string  `json:"description"`
	TargetPath       *string `json:"target_path"`
	TargetPermission *string `json:"target_permission"`
	Severity         string  `json:"severity"`
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
	Metrics         []DashboardMetric         `json:"metrics"`
	BusinessMetrics []DashboardBusinessMetric `json:"business_metrics"`
}
