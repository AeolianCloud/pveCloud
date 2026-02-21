// internal/service/oplog/oplog.go
// 操作日志查询业务逻辑。
package oplog

import (
	"gorm.io/gorm"
	"pvecloud/backend/internal/model"
	"pvecloud/backend/pkg/pagination"
)

// Service 操作日志服务。
type Service struct {
	db *gorm.DB
}

// New 创建操作日志服务实例。
func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ListReq 列表查询参数。
type ListReq struct {
	pagination.Page
	// 按操作人用户名过滤（模糊匹配）
	Username string `form:"username"`
	// 按模块过滤（精确）
	Module string `form:"module"`
	// 按动作过滤（精确）
	Action string `form:"action"`
}

// List 分页查询操作日志，支持多条件过滤，按时间倒序。
func (s *Service) List(req *ListReq) ([]*model.AdminOpLog, int64, error) {
	req.Normalize()

	query := s.db.Model(&model.AdminOpLog{})
	if req.Username != "" {
		query = query.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.Module != "" {
		query = query.Where("module = ?", req.Module)
	}
	if req.Action != "" {
		query = query.Where("action = ?", req.Action)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []*model.AdminOpLog
	err := query.Order("id DESC").
		Offset(req.Offset()).
		Limit(req.PageSize).
		Find(&logs).Error

	return logs, total, err
}
