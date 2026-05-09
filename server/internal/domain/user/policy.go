package user

import (
	"strings"
	"time"
)

const (
	StatusActive               = "active"
	SessionStatusActive        = "active"
	SessionStatusExpired       = "expired"
	SessionStatusRevoked       = "revoked"
	PasswordResetStatusActive  = "active"
	PasswordResetStatusUsed    = "used"
	PasswordResetStatusRevoked = "revoked"
)

func IsActive(status string) bool {
	return strings.TrimSpace(status) == StatusActive
}

func IsSessionActiveAt(status string, expiresAt time.Time, now time.Time) bool {
	return strings.TrimSpace(status) == SessionStatusActive && expiresAt.After(now)
}

func ShouldExpireSession(status string, expiresAt time.Time, now time.Time) bool {
	return strings.TrimSpace(status) == SessionStatusActive && !expiresAt.After(now)
}

func IsPasswordResetTokenUsable(status string, expiresAt time.Time, now time.Time) bool {
	return strings.TrimSpace(status) == PasswordResetStatusActive && expiresAt.After(now)
}
