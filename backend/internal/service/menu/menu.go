// internal/service/menu/menu.go
// 菜单业务逻辑：树构建、按权限裁剪、菜单 CRUD。
package menu

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"pvecloud/backend/internal/model"
)

// Service 菜单业务服务。
type Service struct {
	db *gorm.DB
}

// New 创建菜单服务实例。
func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

// MenuNode 菜单树节点（用于接口返回）。
//
// 注意：
// - 返回结构刻意与 DB 表解耦：避免未来字段调整导致 API 频繁变更。
// - children 只在树结构接口中返回。
type MenuNode struct {
	ID             uint       `json:"id"`
	ParentID       uint       `json:"parent_id"`
	Title          string     `json:"title"`
	Path           *string    `json:"path"`
	Permission     *string    `json:"permission"`
	SuperAdminOnly int8       `json:"super_admin_only"`
	Icon           *string    `json:"icon"`
	Sort           int        `json:"sort"`
	Visible        int8       `json:"visible"`
	Children       []*MenuNode `json:"children,omitempty"`
}

// CreateReq 创建菜单请求体。
type CreateReq struct {
	ParentID       uint   `json:"parent_id"`
	Title          string `json:"title" binding:"required,min=1,max=64"`
	Path           string `json:"path"`
	Permission     string `json:"permission"`
	SuperAdminOnly int8   `json:"super_admin_only" binding:"oneof=0 1"`
	Icon           string `json:"icon"`
	Sort           int    `json:"sort"`
	Visible        int8   `json:"visible" binding:"oneof=0 1"`
}

// UpdateReq 更新菜单请求体。
type UpdateReq struct {
	ParentID       uint   `json:"parent_id"`
	Title          string `json:"title" binding:"required,min=1,max=64"`
	Path           string `json:"path"`
	Permission     string `json:"permission"`
	SuperAdminOnly int8   `json:"super_admin_only" binding:"oneof=0 1"`
	Icon           string `json:"icon"`
	Sort           int    `json:"sort"`
	Visible        int8   `json:"visible" binding:"oneof=0 1"`
}

// ListTreeAll 返回完整菜单树（用于菜单管理）。
func (s *Service) ListTreeAll() ([]*MenuNode, error) {
	menus, err := s.listMenus(false)
	if err != nil {
		return nil, err
	}
	return buildTree(menus), nil
}

// ListTreeForUser 返回当前用户可见菜单树（用于侧边栏渲染）。
//
// 裁剪规则：
// - visible=0 → 永远不下发
// - super_admin_only=1 → 仅 super_admin 可见
// - permission 为空 → 可见
// - permission 非空 → 需要用户拥有该权限
// - 目录节点：若被裁剪后 children 为空且自身无 path，则不下发（避免空目录）
func (s *Service) ListTreeForUser(userID uint) ([]*MenuNode, error) {
	isSuper, err := s.isSuperAdmin(userID)
	if err != nil {
		return nil, err
	}

	menus, err := s.listMenus(true)
	if err != nil {
		return nil, err
	}

	permSet, err := s.loadPermissionSet(userID)
	if err != nil {
		return nil, err
	}

	filtered := make([]*model.AdminMenu, 0, len(menus))
	for _, m := range menus {
		if m.Visible == 0 {
			continue
		}
		if m.SuperAdminOnly == 1 && !isSuper {
			continue
		}
		if m.Permission == nil || strings.TrimSpace(*m.Permission) == "" {
			filtered = append(filtered, m)
			continue
		}
		if isSuper || permSet[*m.Permission] {
			filtered = append(filtered, m)
			continue
		}
	}

	tree := buildTree(filtered)
	return pruneEmptyDirs(tree), nil
}

// Create 创建菜单。
func (s *Service) Create(req *CreateReq) (*model.AdminMenu, error) {
	if err := s.validateParent(req.ParentID); err != nil {
		return nil, err
	}
	pathPtr, err := normalizePath(req.Path)
	if err != nil {
		return nil, err
	}
	permPtr := normalizeOptional(req.Permission)
	iconPtr := normalizeOptional(req.Icon)

	menu := &model.AdminMenu{
		ParentID:       req.ParentID,
		Title:          req.Title,
		Path:           pathPtr,
		Permission:     permPtr,
		SuperAdminOnly: req.SuperAdminOnly,
		Icon:           iconPtr,
		Sort:           req.Sort,
		Visible:        req.Visible,
	}
	if err := s.db.Create(menu).Error; err != nil {
		return nil, err
	}
	return menu, nil
}

// Update 更新菜单。
func (s *Service) Update(id uint, req *UpdateReq) (*model.AdminMenu, error) {
	if id == 0 {
		return nil, ErrInvalidID
	}
	if err := s.validateParent(req.ParentID); err != nil {
		return nil, err
	}
	if req.ParentID == id {
		return nil, ErrInvalidParent
	}

	var menu model.AdminMenu
	if err := s.db.First(&menu, id).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMenuNotFound
	} else if err != nil {
		return nil, err
	}

	pathPtr, err := normalizePath(req.Path)
	if err != nil {
		return nil, err
	}
	menu.ParentID = req.ParentID
	menu.Title = req.Title
	menu.Path = pathPtr
	menu.Permission = normalizeOptional(req.Permission)
	menu.SuperAdminOnly = req.SuperAdminOnly
	menu.Icon = normalizeOptional(req.Icon)
	menu.Sort = req.Sort
	menu.Visible = req.Visible

	if err := s.db.Save(&menu).Error; err != nil {
		return nil, err
	}
	return &menu, nil
}

// Delete 软删除菜单。
func (s *Service) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	result := s.db.Delete(&model.AdminMenu{}, id)
	if result.RowsAffected == 0 {
		return ErrMenuNotFound
	}
	return result.Error
}

// ---- 内部实现：查询、树构建、权限加载与校验 -----------------------------

// listMenus 查询菜单列表。
//
// visibleOnly=true 时只返回 visible=1 的菜单（用于 /menus/my）。
func (s *Service) listMenus(visibleOnly bool) ([]*model.AdminMenu, error) {
	query := s.db.Model(&model.AdminMenu{}).Order("sort ASC, id ASC")
	if visibleOnly {
		query = query.Where("visible = 1")
	}
	var menus []*model.AdminMenu
	if err := query.Find(&menus).Error; err != nil {
		return nil, err
	}
	return menus, nil
}

// validateParent 校验 parent_id 是否存在（0 表示顶级，允许）。
func (s *Service) validateParent(parentID uint) error {
	if parentID == 0 {
		return nil
	}
	var count int64
	s.db.Model(&model.AdminMenu{}).Where("id = ?", parentID).Count(&count)
	if count == 0 {
		return ErrParentNotFound
	}
	return nil
}

// isSuperAdmin 判断是否为 super_admin。
func (s *Service) isSuperAdmin(userID uint) (bool, error) {
	var superCount int64
	err := s.db.Table("admin_user_roles ur").
		Joins("JOIN admin_roles r ON r.id = ur.admin_role_id").
		Where("ur.admin_user_id = ? AND r.name = 'super_admin' AND r.deleted_at IS NULL", userID).
		Count(&superCount).Error
	return superCount > 0, err
}

// loadPermissionSet 加载当前用户的权限集合（name -> true）。
func (s *Service) loadPermissionSet(userID uint) (map[string]bool, error) {
	// 通过 用户→角色→权限 三表联查，拉取权限 name 列表即可。
	// 注意：这里只做“可见性裁剪”，不是接口鉴权；即便可见也必须走后端鉴权中间件。
	type row struct{ Name string }
	var rows []row
	err := s.db.Table("admin_permissions p").
		Select("p.name AS name").
		Joins("JOIN admin_role_permissions rp ON rp.admin_permission_id = p.id").
		Joins("JOIN admin_user_roles ur ON ur.admin_role_id = rp.admin_role_id").
		Where("ur.admin_user_id = ?", userID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	set := make(map[string]bool, len(rows))
	for _, r := range rows {
		if r.Name != "" {
			set[r.Name] = true
		}
	}
	return set, nil
}

// normalizeOptional 将可选字符串规范化：trim 后为空返回 nil，否则返回指针。
func normalizeOptional(s string) *string {
	v := strings.TrimSpace(s)
	if v == "" {
		return nil
	}
	return &v
}

// normalizePath 规范化 path：
// - 空字符串或仅空白：返回 nil（表示目录节点）
// - 非空：必须以 / 开头，返回 trim 后的指针
func normalizePath(path string) (*string, error) {
	v := strings.TrimSpace(path)
	if v == "" {
		return nil, nil
	}
	if !strings.HasPrefix(v, "/") {
		return nil, ErrInvalidPath
	}
	return &v, nil
}

// buildTree 将扁平菜单列表构建成树。
func buildTree(menus []*model.AdminMenu) []*MenuNode {
	nodes := make(map[uint]*MenuNode, len(menus))
	roots := make([]*MenuNode, 0)

	for _, m := range menus {
		nodes[m.ID] = &MenuNode{
			ID:             m.ID,
			ParentID:       m.ParentID,
			Title:          m.Title,
			Path:           m.Path,
			Permission:     m.Permission,
			SuperAdminOnly: m.SuperAdminOnly,
			Icon:           m.Icon,
			Sort:           m.Sort,
			Visible:        m.Visible,
			Children:       []*MenuNode{},
		}
	}

	// 第二遍按 menus 的顺序挂载父子关系，保证返回结果的顺序稳定（sort ASC, id ASC）。
	for _, m := range menus {
		n := nodes[m.ID]
		if n.ParentID == 0 {
			roots = append(roots, n)
			continue
		}
		if p, ok := nodes[n.ParentID]; ok {
			p.Children = append(p.Children, n)
		} else {
			// 父节点不存在时，退化为顶级节点，避免整棵树丢失（数据异常也能看到并修复）。
			roots = append(roots, n)
		}
	}

	return roots
}

// pruneEmptyDirs 裁剪空目录：
// - 目录节点判定：path == nil
// - 如果目录节点最终 children 为空，则移除该节点
func pruneEmptyDirs(nodes []*MenuNode) []*MenuNode {
	out := make([]*MenuNode, 0, len(nodes))
	for _, n := range nodes {
		if len(n.Children) > 0 {
			n.Children = pruneEmptyDirs(n.Children)
		}
		if n.Path == nil && len(n.Children) == 0 {
			continue
		}
		out = append(out, n)
	}
	return out
}

// 业务错误定义（service 层只返回 error，由 handler 统一映射到 errcode）。
var (
	ErrMenuNotFound   = errors.New("菜单不存在")
	ErrParentNotFound = errors.New("父菜单不存在")
	ErrInvalidPath    = errors.New("path 必须以 / 开头或为空（目录节点）")
	ErrInvalidID      = errors.New("invalid id")
	ErrInvalidParent  = errors.New("invalid parent_id")
)
