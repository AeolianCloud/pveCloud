package support

import (
	"testing"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
)

func TestAdminUserUpdateAuditActionDetectsRoleChange(t *testing.T) {
	action := AdminUserUpdateAuditAction(
		models.AdminUser{Status: AdminStatusActive},
		models.AdminUser{Status: AdminStatusActive},
		[]uint64{1, 2},
		[]uint64{2, 3},
	)
	if action != AdminUserRoleUpdateAction {
		t.Fatalf("expected role update action, got %s", action)
	}
}

func TestAdminUserUpdateAuditActionDetectsDisable(t *testing.T) {
	action := AdminUserUpdateAuditAction(
		models.AdminUser{Status: AdminStatusActive},
		models.AdminUser{Status: "disabled"},
		[]uint64{1},
		[]uint64{1},
	)
	if action != AdminUserDisableAction {
		t.Fatalf("expected disable action, got %s", action)
	}
}

func TestAdminRoleUpdateAuditActionDetectsPermissionChange(t *testing.T) {
	action := AdminRoleUpdateAuditAction(
		models.AdminRole{Status: AdminStatusActive},
		models.AdminRole{Status: AdminStatusActive},
		[]string{"page.dashboard"},
		[]string{"page.dashboard", "audit-log:sensitive-view"},
	)
	if action != AdminRolePermissionUpdateAction {
		t.Fatalf("expected permission update action, got %s", action)
	}
}

func TestAdminRoleUpdateAuditActionDetectsDisable(t *testing.T) {
	action := AdminRoleUpdateAuditAction(
		models.AdminRole{Status: AdminStatusActive},
		models.AdminRole{Status: "disabled"},
		[]string{"page.dashboard"},
		[]string{"page.dashboard"},
	)
	if action != AdminRoleDisableAction {
		t.Fatalf("expected role disable action, got %s", action)
	}
}

func TestBuildVisibleAdminMenusUsesOwnedMenuPermissions(t *testing.T) {
	systemParent := "page.system-settings"
	systemPath := "/system"
	systemIcon := "Setting"
	settingsPath := "/system/settings"
	auditPath := "/system/audit-logs"
	permissions := []models.AdminPermission{
		{
			ID:            1,
			Code:          "page.system-settings",
			Name:          "系统设置",
			Type:          "menu",
			Path:          &systemPath,
			Icon:          &systemIcon,
			SortOrder:     20,
			VisibleInMenu: true,
		},
		{
			ID:            2,
			Code:          "page.system-settings.config",
			Name:          "系统配置",
			Type:          "menu",
			ParentCode:    &systemParent,
			Path:          &settingsPath,
			SortOrder:     10,
			VisibleInMenu: true,
		},
		{
			ID:            3,
			Code:          "page.system-settings.audit-logs",
			Name:          "操作日志",
			Type:          "menu",
			ParentCode:    &systemParent,
			Path:          &auditPath,
			SortOrder:     30,
			VisibleInMenu: true,
		},
	}

	menus := BuildVisibleAdminMenus(permissions, []string{
		"page.system-settings",
		"page.system-settings.config",
	})

	if len(menus) != 1 {
		t.Fatalf("expected one root menu, got %d", len(menus))
	}
	if menus[0].Key != "page.system-settings" {
		t.Fatalf("expected system settings root, got %s", menus[0].Key)
	}
	if len(menus[0].Children) != 1 || menus[0].Children[0].Key != "page.system-settings.config" {
		t.Fatalf("expected only system config child, got %#v", menus[0].Children)
	}
}
