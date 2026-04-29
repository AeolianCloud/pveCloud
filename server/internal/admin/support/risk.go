package support

import (
	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/sets"
)

const (
	AdminUserRoleUpdateAction       = "admin.user.role_update"
	AdminUserDisableAction          = "admin.user.disable"
	AdminRolePermissionUpdateAction = "admin.role.permission_update"
	AdminRoleDisableAction          = "admin.role.disable"
)

func AdminUserUpdateRisk(before models.AdminUser, after models.AdminUser, beforeRoleIDs []uint64, afterRoleIDs []uint64) (string, string, string) {
	if !sets.SameUint64Set(beforeRoleIDs, afterRoleIDs) {
		return AdminUserRoleUpdateAction, "high", "修改管理员角色"
	}
	if before.Status == AdminStatusActive && after.Status != AdminStatusActive {
		return AdminUserDisableAction, "high", "禁用管理员账号"
	}
	return "admin.user.update", "", ""
}

func AdminRoleUpdateRisk(before models.AdminRole, after models.AdminRole, beforeCodes []string, afterCodes []string) (string, string, string) {
	if !sets.SameStringSet(beforeCodes, afterCodes) {
		return AdminRolePermissionUpdateAction, "high", "修改角色权限"
	}
	if before.Status == AdminStatusActive && after.Status != AdminStatusActive {
		return AdminRoleDisableAction, "high", "禁用管理端角色"
	}
	return "admin.role.update", "", ""
}
