package bootstrap_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap"
	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap/config"
	"github.com/stretchr/testify/require"
)

func TestNewPublicHTTPHandlerExposesHealthz(t *testing.T) {
	app, err := bootstrap.NewPublicApp(config.Config{
		AppEnv:        "test",
		PublicAPIAddr: ":8080",
		MySQLDSN:      "root:root@tcp(localhost:3306)/pvecloud?parseTime=true&loc=Local",
		RedisAddr:     "127.0.0.1:6379",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"status":"ok"`)
	require.Contains(t, rec.Body.String(), `"service":"public-api"`)
}
