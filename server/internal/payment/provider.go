package payment

import (
	"net/http"

	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap/config"
)

// ProviderVerifier extracts and validates the payment order number from an
// incoming payment provider callback request.
type ProviderVerifier interface {
	VerifyCallback(r *http.Request) (paymentOrderNo string, err error)
}

// NewProviderVerifier returns the appropriate ProviderVerifier based on the
// payment configuration. Currently only "mock" provider is supported.
func NewProviderVerifier(cfg config.PaymentConfig) ProviderVerifier {
	switch cfg.Provider {
	case "mock":
		return &MockProvider{merchantSecret: cfg.MerchantSecret}
	default:
		return &MockProvider{merchantSecret: cfg.MerchantSecret}
	}
}
