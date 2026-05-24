package router

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/AeolianCloud/pveCloud/server/internal/app/api"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
)

func TestPublicAdminPing(t *testing.T) {
	router := newTestRouter(t)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/admin-api/ping", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var envelope response.Envelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
	require.Equal(t, 0, envelope.Code)
}

func TestProtectedRoutesRequireAuthentication(t *testing.T) {
	router := newTestRouter(t)

	for _, route := range router.Routes() {
		route := route
		if isPublicRoute(route.Method, route.Path) {
			continue
		}

		t.Run(route.Method+" "+route.Path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(route.Method, samplePath(route.Path), nil)
			router.ServeHTTP(recorder, request)

			require.Equal(t, http.StatusUnauthorized, recorder.Code)
			var envelope response.Envelope
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
			require.Equal(t, 40101, envelope.Code)
		})
	}
}

func TestRegisteredRoutesAreClassified(t *testing.T) {
	router := newTestRouter(t)
	registered := make(map[string]struct{}, len(router.Routes()))
	for _, route := range router.Routes() {
		registered[routeKey(route.Method, route.Path)] = struct{}{}
	}

	for key := range publicRoutes {
		if _, ok := registered[key]; !ok {
			t.Fatalf("public route classification is stale, route not registered: %s", key)
		}
	}
	for key := range registered {
		if _, ok := publicRoutes[key]; ok {
			continue
		}
		// 未列入公开路由清单的接口都必须经过认证中间件；上一个测试会逐个请求验证 401。
	}
}

func TestWebRenewalOrderRouteIsProtectedUserRoute(t *testing.T) {
	router := newTestRouter(t)
	key := routeKey(http.MethodPost, "/api/instances/:instance_no/renewal-orders")
	registered := make(map[string]struct{}, len(router.Routes()))
	for _, route := range router.Routes() {
		registered[routeKey(route.Method, route.Path)] = struct{}{}
		if strings.Contains(route.Path, "/renewal-orders") && strings.HasPrefix(route.Path, "/admin-api/") {
			t.Fatalf("renewal order route must not be registered under admin-api: %s %s", route.Method, route.Path)
		}
	}

	if _, ok := registered[key]; !ok {
		t.Fatalf("renewal order route is not registered: %s", key)
	}
	if _, ok := publicRoutes[key]; ok {
		t.Fatalf("renewal order route must require user authentication: %s", key)
	}
}

func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := mysqltest.Open(t)
	app := &api.App{
		Config: &config.Config{
			App: config.AppConfig{
				Env: "test",
			},
			JWT: config.JWTConfig{
				UserSecret:         "test_user_jwt_secret_32_bytes_long",
				UserIssuer:         "pvecloud-user-test",
				UserExpireMinutes:  480,
				AdminSecret:        "test_admin_jwt_secret_32_bytes_long",
				AdminIssuer:        "pvecloud-admin-test",
				AdminExpireMinutes: 480,
			},
			Storage: config.StorageConfig{
				Driver:       "local",
				LocalPath:    t.TempDir(),
				MaxSize:      1024 * 1024,
				AllowedTypes: []string{"image/png", "application/pdf"},
			},
			InstanceLifecycle: config.InstanceLifecycleConfig{
				ExpireNoticeBeforeSeconds: 86400,
				ExpireReleaseAfterSeconds: 3600,
			},
		},
		DB:     db,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	app.Routes = api.NewRouteSets(app)
	return NewRouter(app)
}

var publicRoutes = map[string]struct{}{
	routeKey(http.MethodGet, "/healthz"):                                     {},
	routeKey(http.MethodGet, "/admin-api/ping"):                              {},
	routeKey(http.MethodGet, "/admin-api/auth/captcha"):                      {},
	routeKey(http.MethodPost, "/admin-api/auth/login"):                       {},
	routeKey(http.MethodGet, "/api/site-config"):                             {},
	routeKey(http.MethodGet, "/api/site-logo/:id"):                           {},
	routeKey(http.MethodPost, "/api/real-name/provider-callbacks/:provider"): {},
	routeKey(http.MethodPost, "/api/payment-callbacks/:provider"):            {},
	routeKey(http.MethodGet, "/api/server-catalog"):                          {},
	routeKey(http.MethodGet, "/api/auth/login-captcha"):                      {},
	routeKey(http.MethodGet, "/api/auth/register-captcha"):                   {},
	routeKey(http.MethodGet, "/api/auth/password-reset-request-captcha"):     {},
	routeKey(http.MethodGet, "/api/auth/password-reset-confirm-captcha"):     {},
	routeKey(http.MethodPost, "/api/auth/login"):                             {},
	routeKey(http.MethodPost, "/api/auth/register"):                          {},
	routeKey(http.MethodPost, "/api/auth/password-reset/request"):            {},
	routeKey(http.MethodPost, "/api/auth/password-reset/confirm"):            {},
	routeKey(http.MethodPost, "/api/client-logs/errors"):                     {},
}

func isPublicRoute(method string, path string) bool {
	_, ok := publicRoutes[routeKey(method, path)]
	return ok
}

func routeKey(method string, path string) string {
	return method + " " + path
}

func samplePath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "1"
		}
	}
	return strings.Join(parts, "/")
}
