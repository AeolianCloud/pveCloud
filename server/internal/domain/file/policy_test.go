package file

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateUpload(t *testing.T) {
	allowed := []string{"image/jpeg", "application/pdf"}

	require.NoError(t, ValidateUpload("invoice.pdf", "application/pdf", "application/pdf", allowed))
	require.ErrorIs(t, ValidateUpload("shell.php", "application/octet-stream", "application/octet-stream", allowed), ErrDangerousExtension)
	require.ErrorIs(t, ValidateUpload("avatar.png", "image/png", "image/png", allowed), ErrTypeDisabled)
	require.ErrorIs(t, ValidateUpload("avatar.jpg", "image/png", "image/jpeg", allowed), ErrUnsupportedDeclaredMIME)
	require.ErrorIs(t, ValidateUpload("avatar.jpg", "image/jpeg", "application/octet-stream", allowed), ErrContentMismatch)
}

func TestPathAndDeletePolicy(t *testing.T) {
	require.True(t, IsSafeRelativeStoragePath("2026/05/10/a.pdf"))
	require.False(t, IsSafeRelativeStoragePath("../a.pdf"))
	require.False(t, IsSafeRelativeStoragePath("/tmp/a.pdf"))
	require.False(t, IsSafeRelativeStoragePath(""))

	require.True(t, CanDelete(0))
	require.False(t, CanDelete(1))
}
