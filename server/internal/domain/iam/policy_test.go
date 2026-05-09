package iam

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPermissionWildcardPolicy(t *testing.T) {
	codes := []string{"admin-user:*", " page.dashboard "}

	require.True(t, HasPermission(codes, "admin-user:create"))
	require.True(t, HasPermission(codes, "admin-user:password-reset"))
	require.True(t, HasPermission(codes, "page.dashboard"))
	require.False(t, HasPermission(codes, "admin-role:create"))
	require.True(t, HasAllPermissions(codes, "admin-user:update", "page.dashboard"))
	require.True(t, HasAnyPermission(codes, "admin-role:create", "admin-user:update"))
}

func TestAdminAndSessionStatusPolicy(t *testing.T) {
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)

	require.True(t, IsAdminActive("active"))
	require.False(t, IsAdminActive("disabled"))
	require.True(t, IsSessionActiveAt("active", now.Add(time.Minute), now))
	require.False(t, IsSessionActiveAt("active", now, now))
	require.False(t, IsSessionActiveAt("revoked", now.Add(time.Minute), now))
	require.True(t, ShouldExpireSession("active", now, now))
	require.False(t, ShouldExpireSession("revoked", now.Add(time.Minute), now))
}
