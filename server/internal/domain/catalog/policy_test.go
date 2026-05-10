package catalog

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatalogVisibilityPolicy(t *testing.T) {
	require.True(t, IsPublicServerProduct("server", "active", true))
	require.False(t, IsPublicServerProduct("server", "draft", true))
	require.False(t, IsPublicServerProduct("storage", "active", true))
	require.False(t, IsPublicServerProduct("server", "active", false))

	require.True(t, IsPublicServerPlan("active", true))
	require.True(t, IsPublicServerPlan("sold_out", true))
	require.False(t, IsPublicServerPlan("inactive", true))
	require.False(t, IsPublicServerPlan("active", false))

	require.True(t, HasRenderablePlanParts(1, 1, 1, 1))
	require.False(t, HasRenderablePlanParts(0, 1, 1, 1))
	require.False(t, HasRenderablePlanParts(1, 0, 1, 1))
	require.False(t, HasRenderablePlanParts(1, 1, 0, 1))
	require.False(t, HasRenderablePlanParts(1, 1, 1, 0))
}
