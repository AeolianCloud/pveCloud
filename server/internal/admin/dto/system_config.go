package dto

import "time"

/**
 * SystemConfigListQuery 表示系统配置查询参数。
 */
type SystemConfigListQuery struct {
	GroupName string `form:"group_name" validate:"omitempty,max=64"`
}

/**
 * SystemConfigUpdateRequest 表示系统配置更新请求。
 */
type SystemConfigUpdateRequest struct {
	ConfigValue string `json:"config_value" validate:"max=20000"`
}

/**
 * SystemConfigItem 表示系统配置项。
 */
type SystemConfigItem struct {
	ID          uint64    `json:"id"`
	ConfigKey   string    `json:"config_key"`
	ConfigValue *string   `json:"config_value"`
	ValueType   string    `json:"value_type"`
	GroupName   string    `json:"group_name"`
	IsSecret    bool      `json:"is_secret"`
	HasValue    bool      `json:"has_value"`
	Description *string   `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}

/**
 * SystemConfigGroup 表示系统配置分组。
 */
type SystemConfigGroup struct {
	GroupName string             `json:"group_name"`
	Items     []SystemConfigItem `json:"items"`
}
