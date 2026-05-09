package user

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUserSessionAndResetTokenPolicy(t *testing.T) {
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)

	require.True(t, IsActive("active"))
	require.False(t, IsActive("disabled"))
	require.True(t, IsSessionActiveAt("active", now.Add(time.Second), now))
	require.False(t, IsSessionActiveAt("active", now, now))
	require.False(t, IsSessionActiveAt("revoked", now.Add(time.Hour), now))
	require.True(t, ShouldExpireSession("active", now.Add(-time.Second), now))
	require.True(t, IsPasswordResetTokenUsable("active", now.Add(time.Minute), now))
	require.False(t, IsPasswordResetTokenUsable("used", now.Add(time.Minute), now))
	require.False(t, IsPasswordResetTokenUsable("active", now, now))
}
