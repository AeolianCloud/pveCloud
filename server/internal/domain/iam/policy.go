package iam

import (
	"strings"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/shared/rbac"
)

const (
	AdminStatusActive     = "active"
	SessionStatusActive   = "active"
	SessionStatusExpired  = "expired"
	SessionStatusRevoked  = "revoked"
	RevokeReasonExpired   = "expired"
	RevokeReasonRefresh   = "refresh"
	RevokeReasonLogout    = "logout"
	RevokeReasonAdmin     = "admin_revoke"
	RevokeReasonDisabled  = "admin_disabled"
	RevokeReasonResetPass = "password_reset"
)

func IsAdminActive(status string) bool {
	return strings.TrimSpace(status) == AdminStatusActive
}

func IsSessionActiveAt(status string, expiresAt time.Time, now time.Time) bool {
	return strings.TrimSpace(status) == SessionStatusActive && expiresAt.After(now)
}

func ShouldExpireSession(status string, expiresAt time.Time, now time.Time) bool {
	return strings.TrimSpace(status) == SessionStatusActive && !expiresAt.After(now)
}

func HasPermission(permissionCodes []string, required string) bool {
	return rbac.HasPermissionCode(permissionCodes, required)
}

func HasAllPermissions(permissionCodes []string, requiredCodes ...string) bool {
	return rbac.HasAllPermissionCodes(permissionCodes, requiredCodes...)
}

func HasAnyPermission(permissionCodes []string, requiredCodes ...string) bool {
	return rbac.HasAnyPermissionCode(permissionCodes, requiredCodes...)
}
