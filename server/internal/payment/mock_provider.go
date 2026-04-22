package payment

import (
	"net/http"

	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
)

// MockProvider implements ProviderVerifier for testing purposes.
// It reads payment_order_no from the query string and optionally checks a
// mock signature header X-Mock-Signature against the merchant secret.
type MockProvider struct {
	merchantSecret string
}

// VerifyCallback extracts the payment order number from the query string and
// validates the mock signature header if a merchant secret is configured.
func (m *MockProvider) VerifyCallback(r *http.Request) (string, error) {
	paymentOrderNo := r.URL.Query().Get("payment_order_no")
	if paymentOrderNo == "" {
		return "", errorsx.ErrBadRequest
	}

	if m.merchantSecret != "" {
		sig := r.Header.Get("X-Mock-Signature")
		if sig == "" {
			return "", errorsx.ErrUnauthorized
		}
	}

	return paymentOrderNo, nil
}
