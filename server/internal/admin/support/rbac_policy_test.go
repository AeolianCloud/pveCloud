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
		[]string{"dashboard:view"},
		[]string{"dashboard:view", "audit:view"},
	)
	if action != AdminRolePermissionUpdateAction {
		t.Fatalf("expected permission update action, got %s", action)
	}
}

func TestAdminRoleUpdateAuditActionDetectsDisable(t *testing.T) {
	action := AdminRoleUpdateAuditAction(
		models.AdminRole{Status: AdminStatusActive},
		models.AdminRole{Status: "disabled"},
		[]string{"dashboard:view"},
		[]string{"dashboard:view"},
	)
	if action != AdminRoleDisableAction {
		t.Fatalf("expected role disable action, got %s", action)
	}
}
