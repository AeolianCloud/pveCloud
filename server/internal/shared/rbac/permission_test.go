package rbac

import "testing"

func TestHasPermissionCodeSupportsResourceWildcard(t *testing.T) {
	if !HasPermissionCode([]string{"admin-user:*"}, "admin-user:password-reset") {
		t.Fatal("expected resource wildcard to cover child permission")
	}
}

func TestHasPermissionCodeDoesNotWildcardPagePermissions(t *testing.T) {
	if HasPermissionCode([]string{"page:*"}, "page.dashboard") {
		t.Fatal("expected page permissions not to support wildcard")
	}
}

func TestHasAnyPermissionCode(t *testing.T) {
	if !HasAnyPermissionCode([]string{"system-config:view"}, "dashboard:view", "system-config:view") {
		t.Fatal("expected any permission to match")
	}
}
