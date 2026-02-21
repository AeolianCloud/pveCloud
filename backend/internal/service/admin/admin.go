// internal/service/admin/admin.go
// 管理员账号 CRUD 业务逻辑。
package admin

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"pvecloud/backend/internal/model"
	"pvecloud/backend/pkg/pagination"
)

// Service 管理员业务服务。
type Service struct {
	db *gorm.DB
}

// New 创建管理员服务实例。
func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ListReq 列表查询请求参数。
type ListReq struct {
	pagination.Page
	// 关键词，匹配用户名或昵称
	Keyword string `form:"keyword"`
}

// CreateReq 新建管理员请求体。
type CreateReq struct {
	Username string  `json:"username" binding:"required,min=3,max=64"`
	Password string  `json:"password" binding:"required,min=6"`
	Nickname string  `json:"nickname" binding:"max=64"`
	Email    string  `json:"email"    binding:"omitempty,email"`
	// 角色 ID 列表，至少指定一个
	RoleIDs  []uint  `json:"role_ids" binding:"required,min=1"`
}

// UpdateReq 更新管理员请求体。
type UpdateReq struct {
	Nickname string  `json:"nickname" binding:"max=64"`
	Email    string  `json:"email"    binding:"omitempty,email"`
	// 密码可选，非空时重置密码（至少 6 位）
	Password string  `json:"password" binding:"omitempty,min=6"`
	RoleIDs  []uint  `json:"role_ids"`
}

// List 分页查询管理员列表，支持关键词搜索。
func (s *Service) List(req *ListReq) ([]*model.AdminUser, int64, error) {
	req.Normalize()

	query := s.db.Model(&model.AdminUser{})
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		query = query.Where("username LIKE ? OR nickname LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []*model.AdminUser
	err := query.Preload("Roles").
		Order("id DESC").
		Offset(req.Offset()).
		Limit(req.PageSize).
		Find(&users).Error

	return users, total, err
}

// Create 创建管理员账号。
func (s *Service) Create(req *CreateReq) (*model.AdminUser, error) {
	// 检查用户名是否已存在
	var count int64
	s.db.Model(&model.AdminUser{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return nil, ErrUsernameExists
	}

	// bcrypt 加密密码
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 查询角色列表
	var roles []model.AdminRole
	if err := s.db.Find(&roles, req.RoleIDs).Error; err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return nil, ErrRoleNotFound
	}

	// 将空字符串 email 转为 nil，避免多条记录违反唯一约束
	var emailPtr *string
	if req.Email != "" {
		emailPtr = &req.Email
	}

	user := &model.AdminUser{
		Username: req.Username,
		Password: string(hashed),
		Nickname: req.Nickname,
		Email:    emailPtr,
		Status:   1,
		Roles:    roles,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// Update 更新管理员信息（昵称、邮箱、角色）。
func (s *Service) Update(id uint, req *UpdateReq) (*model.AdminUser, error) {
	var user model.AdminUser
	if err := s.db.First(&user, id).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}

	// 更新基本字段（email 为空时存 nil，避免唯一约束冲突）
	var emailPtr *string
	if req.Email != "" {
		emailPtr = &req.Email
	}
	user.Nickname = req.Nickname
	user.Email = emailPtr

	// 非空时重置密码
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashed)
	}

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	// 更新角色关联
	if len(req.RoleIDs) > 0 {
		var roles []model.AdminRole
		if err := s.db.Find(&roles, req.RoleIDs).Error; err != nil {
			return nil, err
		}
		if err := s.db.Model(&user).Association("Roles").Replace(roles); err != nil {
			return nil, err
		}
		user.Roles = roles
	}

	return &user, nil
}

// SetStatus 启用或禁用管理员账号（status: 1 启用, 0 禁用）。
func (s *Service) SetStatus(id uint, status int8) error {
	result := s.db.Model(&model.AdminUser{}).Where("id = ?", id).Update("status", status)
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return result.Error
}

// Delete 软删除管理员账号。
func (s *Service) Delete(id uint) error {
	result := s.db.Delete(&model.AdminUser{}, id)
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return result.Error
}

// 业务错误定义
var (
	ErrUserNotFound   = errors.New("管理员不存在")
	ErrUsernameExists = errors.New("用户名已存在")
	ErrRoleNotFound   = errors.New("指定角色不存在")
)
