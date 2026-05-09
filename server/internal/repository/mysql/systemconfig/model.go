package systemconfig

import "time"

/**
 * SystemConfig 映射 system_configs 系统配置表。
 */
type SystemConfig struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	ConfigKey   string    `gorm:"column:config_key"`
	ConfigValue *string   `gorm:"column:config_value"`
	ValueType   string    `gorm:"column:value_type"`
	GroupName   string    `gorm:"column:group_name"`
	IsSecret    bool      `gorm:"column:is_secret"`
	Description *string   `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

/**
 * TableName 返回系统配置表名。
 *
 * @return string 表名
 */
func (SystemConfig) TableName() string {
	return "system_configs"
}
