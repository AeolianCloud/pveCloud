package services

import (
	"testing"

	"github.com/AeolianCloud/pveCloud/server/internal/models"
)

func TestAdminUserUpdateRiskDetectsRoleChange(t *testing.T) {
	action, level, reason := adminUserUpdateRisk(
		models.AdminUser{Status: adminStatusActive},
		models.AdminUser{Status: adminStatusActive},
		[]uint64{1, 2},
		[]uint64{2, 3},
	)
	if action != adminUserRoleUpdateAction || level != "high" || reason == "" {
		t.Fatalf("expected role update high risk, got %s %s %s", action, level, reason)
	}
}

func TestAdminUserUpdateRiskDetectsDisable(t *testing.T) {
	action, level, reason := adminUserUpdateRisk(
		models.AdminUser{Status: adminStatusActive},
		models.AdminUser{Status: "disabled"},
		[]uint64{1},
		[]uint64{1},
	)
	if action != adminUserDisableAction || level != "high" || reason == "" {
		t.Fatalf("expected disable high risk, got %s %s %s", action, level, reason)
	}
}

func TestAdminRoleUpdateRiskDetectsPermissionChange(t *testing.T) {
	action, level, reason := adminRoleUpdateRisk(
		models.AdminRole{Status: adminStatusActive},
		models.AdminRole{Status: adminStatusActive},
		[]string{"dashboard:view"},
		[]string{"dashboard:view", "audit:view"},
	)
	if action != adminRolePermissionUpdateAction || level != "high" || reason == "" {
		t.Fatalf("expected permission update high risk, got %s %s %s", action, level, reason)
	}
}

func TestAdminRoleUpdateRiskDetectsDisable(t *testing.T) {
	action, level, reason := adminRoleUpdateRisk(
		models.AdminRole{Status: adminStatusActive},
		models.AdminRole{Status: "disabled"},
		[]string{"dashboard:view"},
		[]string{"dashboard:view"},
	)
	if action != adminRoleDisableAction || level != "high" || reason == "" {
		t.Fatalf("expected role disable high risk, got %s %s %s", action, level, reason)
	}
}
