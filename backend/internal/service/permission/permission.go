// internal/service/permission/permission.go
// 权限查询业务逻辑，按 group 分组返回，供前端权限分配使用。
package permission

import (
	"gorm.io/gorm"
	"pvecloud/backend/internal/model"
)

// Service 权限业务服务。
type Service struct {
	db *gorm.DB
}

// New 创建权限服务实例。
func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GroupedPermissions 按分组聚合的权限结构，用于前端树形展示。
type GroupedPermissions struct {
	// 分组标识，如 admin / role
	Group string `json:"group"`
	// 分组下的权限列表
	Permissions []model.AdminPermission `json:"permissions"`
}

// ListGrouped 查询全部权限并按 group 分组返回。
func (s *Service) ListGrouped() ([]*GroupedPermissions, error) {
	var perms []model.AdminPermission
	if err := s.db.Order("`group` ASC, id ASC").Find(&perms).Error; err != nil {
		return nil, err
	}

	// 按 group 聚合
	groupMap := make(map[string]*GroupedPermissions)
	var order []string // 保持原始顺序
	for _, p := range perms {
		if _, ok := groupMap[p.Group]; !ok {
			groupMap[p.Group] = &GroupedPermissions{
				Group:       p.Group,
				Permissions: []model.AdminPermission{},
			}
			order = append(order, p.Group)
		}
		groupMap[p.Group].Permissions = append(groupMap[p.Group].Permissions, p)
	}

	result := make([]*GroupedPermissions, 0, len(order))
	for _, g := range order {
		result = append(result, groupMap[g])
	}
	return result, nil
}
