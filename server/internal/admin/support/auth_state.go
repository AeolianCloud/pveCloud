package support

import (
	"context"

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

func VisibleAdminMenus(permissionCodes []string) []dto.MenuItem {
	visible := make([]dto.MenuItem, 0, 3)
	if rbac.HasPermissionCode(permissionCodes, "page.dashboard") {
		visible = append(visible, menuItem("dashboard", "控制台", "/dashboard", "layout-dashboard", "page.dashboard"))
	}
	if rbac.HasPermissionCode(permissionCodes, "page.system-settings.config") {
		visible = append(visible, menuItem("system_configs", "系统配置", "/system/settings", "settings", "page.system-settings.config"))
	}
	if permissionCode := firstAdminSettingsPermission(permissionCodes); permissionCode != "" {
		visible = append(visible, menuItem("admin_settings", "管理员设置", "/system/admin-users", "users", permissionCode))
	}
	return visible
}

func menuItem(key string, title string, path string, icon string, permissionCode string) dto.MenuItem {
	return dto.MenuItem{
		Key:            key,
		Title:          title,
		Path:           path,
		Icon:           &icon,
		PermissionCode: &permissionCode,
	}
}

func firstAdminSettingsPermission(permissionCodes []string) string {
	candidates := []string{
		"page.system-settings.admin-users",
		"page.system-settings.admin-roles",
		"page.system-settings.admin-sessions",
	}
	for _, code := range candidates {
		if rbac.HasPermissionCode(permissionCodes, code) {
			return code
		}
	}
	return ""
}
