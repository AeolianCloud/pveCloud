package adminrole

import (
	"context"
	"errors"
	"sort"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/support"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/rbac"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/sets"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
)

const (
	adminRoleObjectType             = "admin_role"
	adminRoleCreateAction           = "admin.role.create"
	adminRoleUpdateAction           = "admin.role.update"
	adminRoleDisableAction          = "admin.role.disable"
	adminRolePermissionUpdateAction = "admin.role.permission_update"
	superAdminRoleCode              = "super_admin"
)

/**
 * AdminRoleService 处理管理端角色和权限码管理。
 */
type AdminRoleService struct {
	db           *gorm.DB
	auditService *AdminAuditService
}

/**
 * NewAdminRoleService 创建管理端角色服务。
 *
 * @param db 数据库连接
 * @param auditService 后台审计服务
 * @return *AdminRoleService 管理端角色服务
 */
func NewAdminRoleService(db *gorm.DB, auditService *AdminAuditService) *AdminRoleService {
	if auditService == nil {
		auditService = NewAdminAuditService(db)
	}
	return &AdminRoleService{db: db, auditService: auditService}
}

/**
 * Roles 分页查询管理端角色。
 *
 * @param ctx 请求上下文
 * @param query 查询参数
 * @return admin.PageResponse[admin.AdminRoleItem] 分页结果
 * @return error 查询失败原因
 */
func (s *AdminRoleService) Roles(ctx context.Context, query admindto.AdminRoleListQuery) (admindto.PageResponse[admindto.AdminRoleItem], error) {
	page, perPage := support.NormalizePage(query.Page, query.PerPage)
	db := s.db.WithContext(ctx).Model(&models.AdminRole{})
	if query.Keyword != "" {
		keyword := "%" + strings.TrimSpace(query.Keyword) + "%"
		db = db.Where("code LIKE ? OR name LIKE ? OR description LIKE ?", keyword, keyword, keyword)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return admindto.PageResponse[admindto.AdminRoleItem]{}, err
	}

	var roles []models.AdminRole
	if err := db.Order("id ASC").Limit(perPage).Offset((page - 1) * perPage).Find(&roles).Error; err != nil {
		return admindto.PageResponse[admindto.AdminRoleItem]{}, err
	}
	permissionMap, err := s.permissionCodesByRoleIDs(ctx, adminRoleIDs(roles))
	if err != nil {
		return admindto.PageResponse[admindto.AdminRoleItem]{}, err
	}

	items := make([]admindto.AdminRoleItem, 0, len(roles))
	for _, role := range roles {
		items = append(items, adminRoleItem(role, permissionMap[role.ID]))
	}
	return support.PageResponse(items, total, page, perPage), nil
}

/**
 * CreateRole 创建管理端角色。
 *
 * @param ctx 请求上下文
 * @param operatorID 当前操作者管理员 ID
 * @param req 创建请求
 * @return admin.AdminRoleItem 新建角色
 * @return error 创建失败原因
 */
func (s *AdminRoleService) CreateRole(ctx context.Context, operatorID uint64, operatorPermissionCodes []string, req admindto.AdminRoleCreateRequest) (admindto.AdminRoleItem, error) {
	code := strings.TrimSpace(req.Code)
	if err := s.ensureRoleCodeUnique(ctx, 0, code); err != nil {
		return admindto.AdminRoleItem{}, err
	}
	permissionIDs, normalizedCodes, err := s.permissionIDsByCodes(ctx, req.PermissionCodes)
	if err != nil {
		return admindto.AdminRoleItem{}, err
	}
	if err := ensurePermissionCodesAssignable(operatorPermissionCodes, normalizedCodes); err != nil {
		return admindto.AdminRoleItem{}, err
	}

	var created models.AdminRole
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		created = models.AdminRole{
			Code:        code,
			Name:        strings.TrimSpace(req.Name),
			Description: textutil.NormalizeOptionalString(req.Description),
			Status:      strings.TrimSpace(req.Status),
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		if err := replaceAdminRolePermissions(ctx, tx, created.ID, permissionIDs); err != nil {
			return err
		}
		return s.auditService.Record(ctx, tx, AdminAuditWriteInput{
			AdminID:    &operatorID,
			Action:     adminRoleCreateAction,
			ObjectType: adminRoleObjectType,
			ObjectID:   textutil.Uint64String(created.ID),
			AfterData:  adminRoleAuditSnapshot(created, normalizedCodes),
			Remark:     "创建管理端角色",
		})
	}); err != nil {
		return admindto.AdminRoleItem{}, err
	}
	return adminRoleItem(created, normalizedCodes), nil
}

/**
 * RoleDetail 查询管理端角色详情。
 *
 * @param ctx 请求上下文
 * @param id 角色 ID
 * @return admin.AdminRoleItem 角色详情
 * @return error 查询失败原因
 */
func (s *AdminRoleService) RoleDetail(ctx context.Context, id uint64) (admindto.AdminRoleItem, error) {
	role, err := s.findRole(ctx, s.db, id)
	if err != nil {
		return admindto.AdminRoleItem{}, err
	}
	permissionMap, err := s.permissionCodesByRoleIDs(ctx, []uint64{id})
	if err != nil {
		return admindto.AdminRoleItem{}, err
	}
	return adminRoleItem(role, permissionMap[id]), nil
}

/**
 * UpdateRole 更新管理端角色。
 *
 * @param ctx 请求上下文
 * @param operatorID 当前操作者管理员 ID
 * @param id 角色 ID
 * @param req 更新请求
 * @return admin.AdminRoleItem 更新后的角色
 * @return error 更新失败原因
 */
func (s *AdminRoleService) UpdateRole(ctx context.Context, operatorID uint64, operatorPermissionCodes []string, id uint64, req admindto.AdminRoleUpdateRequest) (admindto.AdminRoleItem, error) {
	if err := s.ensureBuiltInRoleCanUpdate(ctx, id, req); err != nil {
		return admindto.AdminRoleItem{}, err
	}

	var permissionIDs []uint64
	var err error
	if req.PermissionCodes != nil {
		var normalizedCodes []string
		permissionIDs, normalizedCodes, err = s.permissionIDsByCodes(ctx, req.PermissionCodes)
		if err != nil {
			return admindto.AdminRoleItem{}, err
		}
		if err := ensurePermissionCodesAssignable(operatorPermissionCodes, normalizedCodes); err != nil {
			return admindto.AdminRoleItem{}, err
		}
	}

	var updated models.AdminRole
	var afterPermissionCodes []string
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		current, err := s.findRoleForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		beforePermissionCodes, err := s.permissionCodesByRoleID(ctx, tx, id)
		if err != nil {
			return err
		}
		updates := map[string]interface{}{}
		if req.Name != nil {
			updates["name"] = strings.TrimSpace(*req.Name)
		}
		if req.Description != nil {
			updates["description"] = textutil.NormalizeOptionalString(req.Description)
		}
		if req.Status != nil {
			updates["status"] = strings.TrimSpace(*req.Status)
		}
		if len(updates) > 0 {
			if err := tx.Model(&models.AdminRole{}).Where("id = ?", id).Updates(updates).Error; err != nil {
				return err
			}
		}
		if req.PermissionCodes != nil {
			if err := replaceAdminRolePermissions(ctx, tx, id, permissionIDs); err != nil {
				return err
			}
		}
		if err := tx.Where("id = ?", id).First(&updated).Error; err != nil {
			return err
		}
		afterPermissionCodes, err = s.permissionCodesByRoleID(ctx, tx, id)
		if err != nil {
			return err
		}

		action := support.AdminRoleUpdateAuditAction(current, updated, beforePermissionCodes, afterPermissionCodes)
		input := AdminAuditWriteInput{
			AdminID:    &operatorID,
			Action:     action,
			ObjectType: adminRoleObjectType,
			ObjectID:   textutil.Uint64String(id),
			BeforeData: adminRoleAuditSnapshot(current, beforePermissionCodes),
			AfterData:  adminRoleAuditSnapshot(updated, afterPermissionCodes),
			Remark:     "更新管理端角色",
		}
		return s.auditService.Record(ctx, tx, input)
	}); err != nil {
		return admindto.AdminRoleItem{}, err
	}
	return adminRoleItem(updated, afterPermissionCodes), nil
}

/**
 * Permissions 查询系统权限目录树。
 *
 * @param ctx 请求上下文
 * @param query 查询参数
 * @return []admin.AdminPermissionItem 权限目录树
 * @return error 查询失败原因
 */
func (s *AdminRoleService) Permissions(ctx context.Context, query admindto.AdminPermissionListQuery) ([]admindto.AdminPermissionItem, error) {
	db := s.db.WithContext(ctx).Model(&models.AdminPermission{})
	if query.GroupName != "" {
		db = db.Where("group_name = ?", strings.TrimSpace(query.GroupName))
	}

	var permissions []models.AdminPermission
	if err := db.Order("sort_order ASC, id ASC").Find(&permissions).Error; err != nil {
		return nil, err
	}
	return buildPermissionTree(permissions), nil
}

func (s *AdminRoleService) findRole(ctx context.Context, db *gorm.DB, id uint64) (models.AdminRole, error) {
	var role models.AdminRole
	err := db.WithContext(ctx).Where("id = ?", id).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.AdminRole{}, apperrors.ErrNotFound.WithMessage("角色不存在")
	}
	return role, err
}

func (s *AdminRoleService) findRoleForUpdate(ctx context.Context, db *gorm.DB, id uint64) (models.AdminRole, error) {
	var role models.AdminRole
	err := db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.AdminRole{}, apperrors.ErrNotFound.WithMessage("角色不存在")
	}
	return role, err
}

func (s *AdminRoleService) ensureRoleCodeUnique(ctx context.Context, excludeID uint64, code string) error {
	query := s.db.WithContext(ctx).Model(&models.AdminRole{}).Where("code = ?", code)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return apperrors.ErrConflict.WithMessage("角色编码已存在")
	}
	return nil
}

func (s *AdminRoleService) ensureBuiltInRoleCanUpdate(ctx context.Context, id uint64, req admindto.AdminRoleUpdateRequest) error {
	role, err := s.findRole(ctx, s.db, id)
	if err != nil {
		return err
	}
	if role.Code != superAdminRoleCode {
		return nil
	}
	if req.Status != nil && *req.Status != support.AdminStatusActive {
		return apperrors.ErrConflict.WithMessage("禁止禁用内置超级管理员角色")
	}
	if req.PermissionCodes != nil {
		return apperrors.ErrConflict.WithMessage("禁止修改内置超级管理员角色权限")
	}
	return nil
}

func (s *AdminRoleService) permissionIDsByCodes(ctx context.Context, codes []string) ([]uint64, []string, error) {
	codes = sets.UniqueStrings(codes)
	if len(codes) == 0 {
		return nil, nil, nil
	}
	normalizedCodes, err := s.normalizePermissionCodes(ctx, codes)
	if err != nil {
		return nil, nil, err
	}
	if len(normalizedCodes) == 0 {
		return nil, nil, nil
	}
	var rows []models.AdminPermission
	if err := s.db.WithContext(ctx).Where("code IN ?", normalizedCodes).Find(&rows).Error; err != nil {
		return nil, nil, err
	}
	if len(rows) != len(normalizedCodes) {
		return nil, nil, apperrors.ErrValidation.WithMessage("权限码不存在")
	}
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, normalizedCodes, nil
}

func (s *AdminRoleService) normalizePermissionCodes(ctx context.Context, requestedCodes []string) ([]string, error) {
	requestedCodes = sets.UniqueStrings(requestedCodes)
	if len(requestedCodes) == 0 {
		return nil, nil
	}

	var permissions []models.AdminPermission
	if err := s.db.WithContext(ctx).Find(&permissions).Error; err != nil {
		return nil, err
	}

	permissionByCode := make(map[string]models.AdminPermission, len(permissions))
	for _, permission := range permissions {
		permissionByCode[permission.Code] = permission
	}

	result := make(map[string]struct{}, len(requestedCodes))
	for _, code := range requestedCodes {
		permission, ok := permissionByCode[code]
		if !ok {
			return nil, apperrors.ErrValidation.WithMessage("权限码不存在")
		}
		result[permission.Code] = struct{}{}
		parent := strings.TrimSpace(parentCode(permission))
		for parent != "" {
			parentPermission, ok := permissionByCode[parent]
			if !ok {
				return nil, apperrors.ErrValidation.WithMessage("权限目录父节点不存在")
			}
			result[parentPermission.Code] = struct{}{}
			parent = strings.TrimSpace(parentCode(parentPermission))
		}
	}

	codes := make([]string, 0, len(result))
	for code := range result {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes, nil
}

func (s *AdminRoleService) permissionCodesByRoleIDs(ctx context.Context, roleIDs []uint64) (map[uint64][]string, error) {
	result := make(map[uint64][]string)
	if len(roleIDs) == 0 {
		return result, nil
	}
	var rows []adminRolePermissionRow
	err := s.db.WithContext(ctx).
		Table("admin_role_permissions").
		Select("admin_role_permissions.role_id, admin_permissions.code").
		Joins("JOIN admin_permissions ON admin_permissions.id = admin_role_permissions.permission_id").
		Where("admin_role_permissions.role_id IN ?", roleIDs).
		Order("admin_permissions.code ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.RoleID] = append(result[row.RoleID], row.Code)
	}
	return result, nil
}

func (s *AdminRoleService) permissionCodesByRoleID(ctx context.Context, db *gorm.DB, roleID uint64) ([]string, error) {
	var codes []string
	err := db.WithContext(ctx).
		Table("admin_role_permissions").
		Select("admin_permissions.code").
		Joins("JOIN admin_permissions ON admin_permissions.id = admin_role_permissions.permission_id").
		Where("admin_role_permissions.role_id = ?", roleID).
		Order("admin_permissions.code ASC").
		Scan(&codes).Error
	return codes, err
}

func replaceAdminRolePermissions(ctx context.Context, db *gorm.DB, roleID uint64, permissionIDs []uint64) error {
	if err := db.WithContext(ctx).Exec("DELETE FROM admin_role_permissions WHERE role_id = ?", roleID).Error; err != nil {
		return err
	}
	for _, permissionID := range sets.UniqueUint64s(permissionIDs) {
		if err := db.WithContext(ctx).Table("admin_role_permissions").Create(map[string]any{
			"role_id":       roleID,
			"permission_id": permissionID,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func ensurePermissionCodesAssignable(operatorCodes []string, targetCodes []string) error {
	for _, code := range sets.UniqueStrings(targetCodes) {
		if !rbac.HasPermissionCode(operatorCodes, code) {
			return apperrors.ErrForbidden.WithMessage("不能分配当前管理员未拥有的权限")
		}
	}
	return nil
}

type adminRolePermissionRow struct {
	RoleID uint64 `gorm:"column:role_id"`
	Code   string `gorm:"column:code"`
}

func adminRoleIDs(roles []models.AdminRole) []uint64 {
	ids := make([]uint64, 0, len(roles))
	for _, role := range roles {
		ids = append(ids, role.ID)
	}
	return ids
}

func adminRoleItem(role models.AdminRole, permissionCodes []string) admindto.AdminRoleItem {
	return admindto.AdminRoleItem{
		ID:              role.ID,
		Code:            role.Code,
		Name:            role.Name,
		Description:     role.Description,
		Status:          role.Status,
		PermissionCodes: sets.UniqueStrings(permissionCodes),
		CreatedAt:       role.CreatedAt,
		UpdatedAt:       role.UpdatedAt,
	}
}

func adminRoleAuditSnapshot(role models.AdminRole, permissionCodes []string) map[string]any {
	return map[string]any{
		"id":               role.ID,
		"code":             role.Code,
		"name":             role.Name,
		"description":      role.Description,
		"status":           role.Status,
		"permission_codes": sets.UniqueStrings(permissionCodes),
	}
}

func buildPermissionTree(permissions []models.AdminPermission) []admindto.AdminPermissionItem {
	childrenByParent := make(map[string][]models.AdminPermission)
	for _, permission := range permissions {
		childrenByParent[parentCode(permission)] = append(childrenByParent[parentCode(permission)], permission)
	}
	for parent := range childrenByParent {
		sort.SliceStable(childrenByParent[parent], func(left, right int) bool {
			if childrenByParent[parent][left].SortOrder != childrenByParent[parent][right].SortOrder {
				return childrenByParent[parent][left].SortOrder < childrenByParent[parent][right].SortOrder
			}
			return childrenByParent[parent][left].ID < childrenByParent[parent][right].ID
		})
	}
	return buildPermissionTreeItems("", childrenByParent)
}

func buildPermissionTreeItems(parent string, childrenByParent map[string][]models.AdminPermission) []admindto.AdminPermissionItem {
	items := make([]admindto.AdminPermissionItem, 0, len(childrenByParent[parent]))
	for _, permission := range childrenByParent[parent] {
		items = append(items, admindto.AdminPermissionItem{
			ID:          permission.ID,
			Code:        permission.Code,
			Name:        permission.Name,
			Type:        permission.Type,
			ParentCode:  permission.ParentCode,
			Path:        permission.Path,
			Icon:        permission.Icon,
			SortOrder:   permission.SortOrder,
			GroupName:   permission.GroupName,
			Description: permission.Description,
			Children:    buildPermissionTreeItems(permission.Code, childrenByParent),
		})
	}
	return items
}

func parentCode(permission models.AdminPermission) string {
	if permission.ParentCode == nil {
		return ""
	}
	return strings.TrimSpace(*permission.ParentCode)
}
