package openapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	spec, err := Load(context.Background(), "../../../docs/server/api/openapi.yaml")
	require.NoError(t, err)
	require.NotNil(t, spec)
}
