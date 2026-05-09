package adminsupport

import (
	"context"
	"sort"

	"gorm.io/gorm"

	mysqliam "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/iam"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/rbac"
	"github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
)

func RoleIDs(ctx context.Context, db *gorm.DB, adminID uint64) ([]uint64, error) {
	return mysqliam.NewRepository(db).AdminRoleIDs(ctx, db, adminID, AdminStatusActive)
}

func PermissionCodes(ctx context.Context, db *gorm.DB, adminID uint64) ([]string, error) {
	return mysqliam.NewRepository(db).AdminPermissionCodes(ctx, db, adminID, AdminStatusActive)
}

func AdminSummary(admin mysqliam.AdminUser) dto.AdminSummary {
	return dto.AdminSummary{
		ID:          admin.ID,
		Username:    admin.Username,
		Email:       admin.Email,
		DisplayName: admin.DisplayName,
		Status:      admin.Status,
	}
}

func SessionSummary(session mysqliam.AdminSession) dto.SessionSummary {
	return dto.SessionSummary{
		SessionID: session.SessionID,
		IssuedAt:  session.IssuedAt,
		ExpiresAt: session.ExpiresAt,
	}
}

func VisibleAdminMenus(ctx context.Context, db *gorm.DB, permissionCodes []string) ([]dto.MenuItem, error) {
	permissions, err := mysqliam.NewRepository(db).VisibleMenuPermissions(ctx, db)
	if err != nil {
		return nil, err
	}
	return BuildVisibleAdminMenus(permissions, permissionCodes), nil
}

func BuildVisibleAdminMenus(permissions []mysqliam.AdminPermission, permissionCodes []string) []dto.MenuItem {
	menuByParent := make(map[string][]mysqliam.AdminPermission)
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

func buildMenuItems(parent string, menuByParent map[string][]mysqliam.AdminPermission) []dto.MenuItem {
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
