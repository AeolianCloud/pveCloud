package systemconfig

import (
	"testing"

	mysqlsystemconfig "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/systemconfig"
	"github.com/stretchr/testify/require"
)

func TestSecretConfigItemAndAuditSnapshot(t *testing.T) {
	value := "secret"
	config := mysqlsystemconfig.SystemConfig{
		ID:          1,
		ConfigKey:   "real_name.identity_digest_secret",
		ConfigValue: &value,
		ValueType:   "string",
		GroupName:   "real_name",
		IsSecret:    true,
	}

	item := systemConfigItem(config)
	require.Nil(t, item.ConfigValue)
	require.True(t, item.HasValue)

	snapshot := systemConfigAuditSnapshot(config)
	require.Equal(t, adminAuditMaskedValue, *(snapshot["config_value"].(*string)))
	require.True(t, snapshot["has_value"].(bool))
}
