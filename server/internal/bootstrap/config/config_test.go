package config_test

import (
	"testing"

	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap/config"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigReadsRequiredFields(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/pvecloud")
	t.Setenv("REDIS_ADDR", "127.0.0.1:6379")

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, "test", cfg.AppEnv)
	require.Equal(t, "root:root@tcp(localhost:3306)/pvecloud", cfg.MySQLDSN)
	require.Equal(t, "127.0.0.1:6379", cfg.RedisAddr)
}
