// internal/service/loginlog/loginlog.go
// 登录日志查询业务逻辑。
package loginlog

import (
	"gorm.io/gorm"
	"pvecloud/backend/internal/model"
	"pvecloud/backend/pkg/pagination"
)

// Service 登录日志服务。
type Service struct {
	db *gorm.DB
}

// New 创建登录日志服务实例。
func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ListReq 列表查询参数。
type ListReq struct {
	pagination.Page
	// 按用户名过滤（模糊匹配）
	Username string `form:"username"`
	// 按状态过滤：1 成功  0 失败  空字符串不过滤
	Status *int8 `form:"status"`
}

// List 分页查询登录日志，支持用户名和状态过滤。
func (s *Service) List(req *ListReq) ([]*model.AdminLoginLog, int64, error) {
	req.Normalize()

	query := s.db.Model(&model.AdminLoginLog{})
	if req.Username != "" {
		query = query.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []*model.AdminLoginLog
	err := query.Order("id DESC").
		Offset(req.Offset()).
		Limit(req.PageSize).
		Find(&logs).Error

	return logs, total, err
}
