package adminsupport

import (
	mysqliam "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/iam"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/sets"
)

const (
	AdminUserRoleUpdateAction       = "admin.user.role_update"
	AdminUserDisableAction          = "admin.user.disable"
	AdminRolePermissionUpdateAction = "admin.role.permission_update"
	AdminRoleDisableAction          = "admin.role.disable"
)

func AdminUserUpdateAuditAction(before mysqliam.AdminUser, after mysqliam.AdminUser, beforeRoleIDs []uint64, afterRoleIDs []uint64) string {
	if !sets.SameUint64Set(beforeRoleIDs, afterRoleIDs) {
		return AdminUserRoleUpdateAction
	}
	if before.Status == AdminStatusActive && after.Status != AdminStatusActive {
		return AdminUserDisableAction
	}
	return "admin.user.update"
}

func AdminRoleUpdateAuditAction(before mysqliam.AdminRole, after mysqliam.AdminRole, beforeCodes []string, afterCodes []string) string {
	if !sets.SameStringSet(beforeCodes, afterCodes) {
		return AdminRolePermissionUpdateAction
	}
	if before.Status == AdminStatusActive && after.Status != AdminStatusActive {
		return AdminRoleDisableAction
	}
	return "admin.role.update"
}
