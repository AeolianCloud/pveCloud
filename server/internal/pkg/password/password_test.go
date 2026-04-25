package password

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashAndVerify(t *testing.T) {
	hash, err := Hash("secret")
	require.NoError(t, err)
	require.NotEqual(t, "secret", hash)
	require.True(t, Verify(hash, "secret"))
	require.False(t, Verify(hash, "wrong"))
}
