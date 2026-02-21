// internal/service/role/role.go
// 角色 CRUD 业务逻辑，包含权限分配。
package role

import (
	"errors"

	"gorm.io/gorm"
	"pvecloud/backend/internal/model"
	"pvecloud/backend/pkg/pagination"
)

// Service 角色业务服务。
type Service struct {
	db *gorm.DB
}

// New 创建角色服务实例。
func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ListReq 列表查询参数。
type ListReq struct {
	pagination.Page
	Keyword string `form:"keyword"`
}

// CreateReq 创建角色请求体。
type CreateReq struct {
	Name        string `json:"name"        binding:"required,min=2,max=64"`
	Label       string `json:"label"       binding:"required,min=1,max=64"`
	Description string `json:"description" binding:"max=255"`
	Sort        int    `json:"sort"`
}

// UpdateReq 更新角色请求体。
type UpdateReq struct {
	Label       string `json:"label"       binding:"required,min=1,max=64"`
	Description string `json:"description" binding:"max=255"`
	Sort        int    `json:"sort"`
}

// AssignPermissionsReq 分配权限请求体。
type AssignPermissionsReq struct {
	// 权限 ID 列表，传空数组表示清空该角色的所有权限
	PermissionIDs []uint `json:"permission_ids"`
}

// List 分页查询角色列表（含权限预加载）。
func (s *Service) List(req *ListReq) ([]*model.AdminRole, int64, error) {
	req.Normalize()

	query := s.db.Model(&model.AdminRole{})
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		query = query.Where("name LIKE ? OR label LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var roles []*model.AdminRole
	err := query.Preload("Permissions").
		Order("sort ASC, id ASC").
		Offset(req.Offset()).
		Limit(req.PageSize).
		Find(&roles).Error

	return roles, total, err
}

// GetByID 根据 ID 查询角色（含权限）。
func (s *Service) GetByID(id uint) (*model.AdminRole, error) {
	var role model.AdminRole
	err := s.db.Preload("Permissions").First(&role, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRoleNotFound
	}
	return &role, err
}

// Create 创建角色。
func (s *Service) Create(req *CreateReq) (*model.AdminRole, error) {
	var count int64
	s.db.Model(&model.AdminRole{}).Where("name = ?", req.Name).Count(&count)
	if count > 0 {
		return nil, ErrRoleNameExists
	}

	role := &model.AdminRole{
		Name:        req.Name,
		Label:       req.Label,
		Description: req.Description,
		Sort:        req.Sort,
	}
	if err := s.db.Create(role).Error; err != nil {
		return nil, err
	}
	return role, nil
}

// Update 更新角色基本信息（不含权限）。
func (s *Service) Update(id uint, req *UpdateReq) (*model.AdminRole, error) {
	var role model.AdminRole
	if err := s.db.First(&role, id).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRoleNotFound
	} else if err != nil {
		return nil, err
	}

	role.Label = req.Label
	role.Description = req.Description
	role.Sort = req.Sort
	if err := s.db.Save(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// Delete 软删除角色。
func (s *Service) Delete(id uint) error {
	result := s.db.Delete(&model.AdminRole{}, id)
	if result.RowsAffected == 0 {
		return ErrRoleNotFound
	}
	return result.Error
}

// AssignPermissions 替换角色的权限列表（传空数组则清空）。
func (s *Service) AssignPermissions(id uint, req *AssignPermissionsReq) error {
	var role model.AdminRole
	if err := s.db.First(&role, id).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRoleNotFound
	} else if err != nil {
		return err
	}

	// 查询权限列表（空数组时 Replace 会清空关联）
	var perms []model.AdminPermission
	if len(req.PermissionIDs) > 0 {
		if err := s.db.Find(&perms, req.PermissionIDs).Error; err != nil {
			return err
		}
	}

	return s.db.Model(&role).Association("Permissions").Replace(perms)
}

// 业务错误定义
var (
	ErrRoleNotFound  = errors.New("角色不存在")
	ErrRoleNameExists = errors.New("角色标识已存在")
)
