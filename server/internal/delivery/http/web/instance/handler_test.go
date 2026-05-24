package instance

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
)

func TestCreateRenewalOrderRequiresCurrentUser(t *testing.T) {
	router := newRenewalOrderHandlerRouter(nil)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/instances/INS-1/renewal-orders", strings.NewReader(`{"billing_cycle":"monthly","client_token":"token-1"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	requireEnvelopeCode(t, recorder, 40101)
}

func TestCreateRenewalOrderValidatesJSONRequest(t *testing.T) {
	router := newRenewalOrderHandlerRouter(func(c *gin.Context) {
		c.Set("web_user_id", uint64(11))
	})

	tests := []struct {
		name string
		body string
	}{
		{name: "malformed json", body: `{"billing_cycle":`},
		{name: "invalid billing cycle", body: `{"billing_cycle":"weekly","client_token":"token-1"}`},
		{name: "sql-like billing cycle", body: `{"billing_cycle":"monthly' OR '1'='1","client_token":"token-1"}`},
		{name: "missing client token", body: `{"billing_cycle":"monthly"}`},
		{name: "overlong client token", body: `{"billing_cycle":"monthly","client_token":"` + strings.Repeat("a", 129) + `"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/instances/INS-1/renewal-orders", strings.NewReader(tt.body))
			request.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(recorder, request)

			require.Equal(t, http.StatusBadRequest, recorder.Code)
			requireEnvelopeCode(t, recorder, 40001)
		})
	}
}

func newRenewalOrderHandlerRouter(injectUser gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHandler(nil)
	if injectUser != nil {
		router.Use(injectUser)
	}
	router.POST("/instances/:instance_no/renewal-orders", handler.CreateRenewalOrder)
	return router
}

func requireEnvelopeCode(t *testing.T, recorder *httptest.ResponseRecorder, code int) {
	t.Helper()
	var envelope response.Envelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
	require.Equal(t, code, envelope.Code)
	require.Nil(t, envelope.Data)
}
