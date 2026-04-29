package support

import (
	"testing"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
)

func TestAdminUserUpdateRiskDetectsRoleChange(t *testing.T) {
	action, level, reason := AdminUserUpdateRisk(
		models.AdminUser{Status: AdminStatusActive},
		models.AdminUser{Status: AdminStatusActive},
		[]uint64{1, 2},
		[]uint64{2, 3},
	)
	if action != AdminUserRoleUpdateAction || level != "high" || reason == "" {
		t.Fatalf("expected role update high risk, got %s %s %s", action, level, reason)
	}
}

func TestAdminUserUpdateRiskDetectsDisable(t *testing.T) {
	action, level, reason := AdminUserUpdateRisk(
		models.AdminUser{Status: AdminStatusActive},
		models.AdminUser{Status: "disabled"},
		[]uint64{1},
		[]uint64{1},
	)
	if action != AdminUserDisableAction || level != "high" || reason == "" {
		t.Fatalf("expected disable high risk, got %s %s %s", action, level, reason)
	}
}

func TestAdminRoleUpdateRiskDetectsPermissionChange(t *testing.T) {
	action, level, reason := AdminRoleUpdateRisk(
		models.AdminRole{Status: AdminStatusActive},
		models.AdminRole{Status: AdminStatusActive},
		[]string{"dashboard:view"},
		[]string{"dashboard:view", "audit:view"},
	)
	if action != AdminRolePermissionUpdateAction || level != "high" || reason == "" {
		t.Fatalf("expected permission update high risk, got %s %s %s", action, level, reason)
	}
}

func TestAdminRoleUpdateRiskDetectsDisable(t *testing.T) {
	action, level, reason := AdminRoleUpdateRisk(
		models.AdminRole{Status: AdminStatusActive},
		models.AdminRole{Status: "disabled"},
		[]string{"dashboard:view"},
		[]string{"dashboard:view"},
	)
	if action != AdminRoleDisableAction || level != "high" || reason == "" {
		t.Fatalf("expected role disable high risk, got %s %s %s", action, level, reason)
	}
}
