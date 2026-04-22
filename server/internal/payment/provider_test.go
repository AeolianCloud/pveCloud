package payment

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap/config"
)

func TestMockProviderVerifyCallbackExtractsPaymentOrderNo(t *testing.T) {
	p := &MockProvider{}
	req := httptest.NewRequest(http.MethodPost, "/callback?payment_order_no=P12345", nil)

	got, err := p.VerifyCallback(req)
	if err != nil {
		t.Fatalf("verify callback: %v", err)
	}
	if got != "P12345" {
		t.Fatalf("expected payment_order_no P12345, got %s", got)
	}
}

func TestMockProviderVerifyCallbackRejectsEmptyPaymentOrderNo(t *testing.T) {
	p := &MockProvider{}
	req := httptest.NewRequest(http.MethodPost, "/callback", nil)

	_, err := p.VerifyCallback(req)
	if err == nil {
		t.Fatal("expected error for empty payment_order_no")
	}
}

func TestMockProviderVerifyCallbackRequiresSignatureWhenSecretSet(t *testing.T) {
	p := &MockProvider{merchantSecret: "secret123"}
	req := httptest.NewRequest(http.MethodPost, "/callback?payment_order_no=P12345", nil)

	_, err := p.VerifyCallback(req)
	if err == nil {
		t.Fatal("expected error when signature header is missing and secret is set")
	}
}

func TestMockProviderVerifyCallbackPassesWithSignatureWhenSecretSet(t *testing.T) {
	p := &MockProvider{merchantSecret: "secret123"}
	req := httptest.NewRequest(http.MethodPost, "/callback?payment_order_no=P12345", nil)
	req.Header.Set("X-Mock-Signature", "any-non-empty-value")

	got, err := p.VerifyCallback(req)
	if err != nil {
		t.Fatalf("verify callback: %v", err)
	}
	if got != "P12345" {
		t.Fatalf("expected payment_order_no P12345, got %s", got)
	}
}

func TestNewProviderVerifierReturnsMockForMockProvider(t *testing.T) {
	cfg := config.PaymentConfig{Provider: "mock", MerchantSecret: "s"}
	v := NewProviderVerifier(cfg)
	if _, ok := v.(*MockProvider); !ok {
		t.Fatalf("expected MockProvider, got %T", v)
	}
}

func TestNewProviderVerifierReturnsMockForUnknownProvider(t *testing.T) {
	cfg := config.PaymentConfig{Provider: "alipay", MerchantSecret: "s"}
	v := NewProviderVerifier(cfg)
	if _, ok := v.(*MockProvider); !ok {
		t.Fatalf("expected MockProvider for unknown provider, got %T", v)
	}
}
