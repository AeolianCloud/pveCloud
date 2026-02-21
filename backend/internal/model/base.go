// internal/model/base.go
// 所有数据表的公共基础字段，业务 Model 嵌入此结构体即可继承。
package model

import (
	"time"

	"gorm.io/gorm"
)

// Model 基础模型，包含主键、时间戳、软删除。
// 适用于需要软删除的业务表（admin_users、admin_roles 等）。
type Model struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `                  json:"created_at"`
	UpdatedAt time.Time      `                  json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"      json:"-"` // 软删除，不暴露给前端
}

// BaseModel 不含软删除的基础模型。
// 适用于系统常量类表（admin_permissions 等），这类数据只增不删，不需要软删除。
type BaseModel struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `                  json:"created_at"`
	UpdatedAt time.Time `                  json:"updated_at"`
}
