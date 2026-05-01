package support

import (
	"context"
	"sort"

	"gorm.io/gorm"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/rbac"
)

func RoleIDs(ctx context.Context, db *gorm.DB, adminID uint64) ([]uint64, error) {
	var roleIDs []uint64
	err := db.WithContext(ctx).
		Table("admin_user_roles").
		Select("admin_roles.id").
		Joins("JOIN admin_roles ON admin_roles.id = admin_user_roles.role_id").
		Where("admin_user_roles.admin_id = ?", adminID).
		Where("admin_roles.status = ?", AdminStatusActive).
		Order("admin_roles.id ASC").
		Scan(&roleIDs).Error
	return roleIDs, err
}

func PermissionCodes(ctx context.Context, db *gorm.DB, adminID uint64) ([]string, error) {
	var codes []string
	err := db.WithContext(ctx).
		Table("admin_user_roles").
		Distinct("admin_permissions.code").
		Joins("JOIN admin_roles ON admin_roles.id = admin_user_roles.role_id").
		Joins("JOIN admin_role_permissions ON admin_role_permissions.role_id = admin_roles.id").
		Joins("JOIN admin_permissions ON admin_permissions.id = admin_role_permissions.permission_id").
		Where("admin_user_roles.admin_id = ?", adminID).
		Where("admin_roles.status = ?", AdminStatusActive).
		Order("admin_permissions.code ASC").
		Scan(&codes).Error
	return codes, err
}

func AdminSummary(admin models.AdminUser) dto.AdminSummary {
	return dto.AdminSummary{
		ID:          admin.ID,
		Username:    admin.Username,
		Email:       admin.Email,
		DisplayName: admin.DisplayName,
		Status:      admin.Status,
	}
}

func SessionSummary(session models.AdminSession) dto.SessionSummary {
	return dto.SessionSummary{
		SessionID: session.SessionID,
		IssuedAt:  session.IssuedAt,
		ExpiresAt: session.ExpiresAt,
	}
}

func VisibleAdminMenus(ctx context.Context, db *gorm.DB, permissionCodes []string) ([]dto.MenuItem, error) {
	var permissions []models.AdminPermission
	if err := db.WithContext(ctx).
		Where("type = ?", "menu").
		Where("visible_in_menu = ?", true).
		Order("sort_order ASC, id ASC").
		Find(&permissions).Error; err != nil {
		return nil, err
	}
	return BuildVisibleAdminMenus(permissions, permissionCodes), nil
}

func BuildVisibleAdminMenus(permissions []models.AdminPermission, permissionCodes []string) []dto.MenuItem {
	menuByParent := make(map[string][]models.AdminPermission)
	for _, permission := range permissions {
		if !permission.VisibleInMenu || permission.Type != "menu" || permission.Path == nil {
			continue
		}
		if !rbac.HasPermissionCode(permissionCodes, permission.Code) {
			continue
		}
		parent := ""
		if permission.ParentCode != nil && rbac.HasPermissionCode(permissionCodes, *permission.ParentCode) {
			parent = *permission.ParentCode
		}
		menuByParent[parent] = append(menuByParent[parent], permission)
	}

	for parent := range menuByParent {
		sort.SliceStable(menuByParent[parent], func(left, right int) bool {
			if menuByParent[parent][left].SortOrder != menuByParent[parent][right].SortOrder {
				return menuByParent[parent][left].SortOrder < menuByParent[parent][right].SortOrder
			}
			return menuByParent[parent][left].ID < menuByParent[parent][right].ID
		})
	}

	return buildMenuItems("", menuByParent)
}

func buildMenuItems(parent string, menuByParent map[string][]models.AdminPermission) []dto.MenuItem {
	items := make([]dto.MenuItem, 0, len(menuByParent[parent]))
	for _, permission := range menuByParent[parent] {
		item := dto.MenuItem{
			Key:            permission.Code,
			Title:          permission.Name,
			Path:           *permission.Path,
			Icon:           permission.Icon,
			PermissionCode: &permission.Code,
			Children:       buildMenuItems(permission.Code, menuByParent),
		}
		items = append(items, item)
	}
	return items
}
