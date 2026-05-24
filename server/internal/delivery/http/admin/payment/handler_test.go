package payment

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	adminmiddleware "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/middleware"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
)

func TestPaymentWriteActionsRequireSpecificPermissions(t *testing.T) {
	router := newLowPermissionPaymentRouter()

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "refund", method: http.MethodPost, path: "/payments/PAY-1/refunds", body: `{"reason":"test"}`},
		{name: "sync", method: http.MethodPost, path: "/payments/PAY-1/sync"},
		{name: "retry provision", method: http.MethodPost, path: "/payments/PAY-1/retry-provision"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			request.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(recorder, request)

			require.Equal(t, http.StatusForbidden, recorder.Code)
			var envelope response.Envelope
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
			require.Equal(t, 40301, envelope.Code)
		})
	}
}

func newLowPermissionPaymentRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("admin_id", uint64(99))
		c.Set("admin_permission_codes", []string{"page.payments"})
	})
	handler := NewHandler(nil)
	router.POST("/payments/:payment_no/refunds", adminmiddleware.AdminPermission("payment:refund"), handler.CreateRefund)
	router.POST("/payments/:payment_no/sync", adminmiddleware.AdminPermission("payment:sync"), handler.Sync)
	router.POST("/payments/:payment_no/retry-provision", adminmiddleware.AdminPermission("payment:retry-provision"), handler.RetryProvision)
	return router
}
