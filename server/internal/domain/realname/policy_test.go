package realname

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRealNamePolicy(t *testing.T) {
	require.True(t, CanManualReview("manual", "pending"))
	require.False(t, CanManualReview("alipay", "pending"))
	require.True(t, ShouldRejectApprovedWithoutDigest(""))
	require.False(t, ShouldRejectApprovedWithoutDigest("digest"))
	require.Equal(t, "实名供应商核验失败：CODE", ProviderUserMessage("CODE", ""))
	require.Equal(t, "实名供应商核验失败", ProviderUserMessage("", ""))
	require.True(t, HasApprovedDigestConflict(1))
	require.False(t, HasApprovedDigestConflict(0))
	require.True(t, AllowCallbackReplay("", false))
	require.True(t, AllowCallbackReplay("key", true))
	require.False(t, AllowCallbackReplay("key", false))
}
