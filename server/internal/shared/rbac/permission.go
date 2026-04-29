package rbac

import "strings"

/**
 * HasPermissionCode 判断权限集合是否满足指定权限码。
 * 规则：
 * - 精确权限码直接命中
 * - `module:*` 可覆盖同模块下所有细粒度权限
 */
func HasPermissionCode(permissionCodes []string, required string) bool {
	required = strings.TrimSpace(required)
	if required == "" {
		return true
	}

	permissionSet := make(map[string]struct{}, len(permissionCodes))
	for _, code := range permissionCodes {
		trimmed := strings.TrimSpace(code)
		if trimmed == "" {
			continue
		}
		permissionSet[trimmed] = struct{}{}
	}

	if _, ok := permissionSet[required]; ok {
		return true
	}

	module, action, ok := strings.Cut(required, ":")
	if !ok || strings.TrimSpace(module) == "" || strings.TrimSpace(action) == "" || action == "*" {
		return false
	}

	_, ok = permissionSet[module+":*"]
	return ok
}

/**
 * HasAllPermissionCodes 判断权限集合是否满足全部指定权限码。
 */
func HasAllPermissionCodes(permissionCodes []string, requiredCodes ...string) bool {
	for _, code := range requiredCodes {
		if !HasPermissionCode(permissionCodes, code) {
			return false
		}
	}
	return true
}

/**
 * HasAnyPermissionCode 判断权限集合是否满足任一指定权限码。
 */
func HasAnyPermissionCode(permissionCodes []string, requiredCodes ...string) bool {
	if len(requiredCodes) == 0 {
		return true
	}
	for _, code := range requiredCodes {
		if HasPermissionCode(permissionCodes, code) {
			return true
		}
	}
	return false
}
