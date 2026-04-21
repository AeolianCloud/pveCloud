package config_test

import (
	"testing"

	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap/config"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigReadsRepositoryBaselineFields(t *testing.T) {
	t.Setenv("APP_ENV", "local")
	t.Setenv("PUBLIC_API_ADDR", ":8080")
	t.Setenv("ADMIN_API_ADDR", ":8081")
	t.Setenv("WORKER_ADDR", ":8082")
	t.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/pvecloud?parseTime=true&loc=Local")
	t.Setenv("REDIS_ADDR", "127.0.0.1:6379")
	t.Setenv("JWT_WEB_SECRET", "web-secret")
	t.Setenv("JWT_ADMIN_SECRET", "admin-secret")

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, "local", cfg.AppEnv)
	require.Equal(t, ":8080", cfg.PublicAPIAddr)
	require.Equal(t, ":8081", cfg.AdminAPIAddr)
	require.Equal(t, ":8082", cfg.WorkerAddr)
	require.Equal(t, "root:root@tcp(localhost:3306)/pvecloud?parseTime=true&loc=Local", cfg.MySQLDSN)
	require.Equal(t, "127.0.0.1:6379", cfg.RedisAddr)
	require.Equal(t, "web-secret", cfg.JWTWebSecret)
	require.Equal(t, "admin-secret", cfg.JWTAdminSecret)
}

func TestLoadConfigRequiresSecretsAndConnections(t *testing.T) {
	t.Setenv("APP_ENV", "test")

	_, err := config.Load()
	require.EqualError(t, err, "MYSQL_DSN is required")
}
