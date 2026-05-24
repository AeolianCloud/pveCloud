package payment

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAlipayAdapterBuildsSignedPagePayURL(t *testing.T) {
	adapter := NewAlipayAdapter()
	result, err := adapter.CreatePayment(context.Background(), alipayTestConfig(t), CreatePaymentRequest{
		PaymentNo:   "PAY-ALI-SIGN-1",
		OrderNo:     "ORD-ALI-SIGN-1",
		Subject:     "Server",
		AmountCents: 1234,
		Currency:    "CNY",
		Method:      MethodAlipayPage,
		ExpiresAt:   time.Now().Add(30 * time.Minute),
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.RedirectURL)
	require.Contains(t, result.RedirectURL, "sign=")
	require.Contains(t, result.RedirectURL, "method=alipay.trade.page.pay")
	require.Contains(t, result.RedirectURL, "app_id=9021000122689420")
	require.Contains(t, result.RedirectURL, "biz_content=")
}

func TestAlipayAdapterRejectsInvalidNotificationSignature(t *testing.T) {
	adapter := NewAlipayAdapter()
	req := httptest.NewRequest("POST", "/api/payment-callbacks/alipay", strings.NewReader("out_trade_no=PAY-1&trade_no=ALI-1&total_amount=12.34&trade_status=TRADE_SUCCESS&sign=invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, err := adapter.ParseNotification(context.Background(), alipayTestConfig(t), req)
	require.ErrorIs(t, err, ErrInvalidSignature)
}

func alipayTestConfig(t *testing.T) Config {
	t.Helper()
	privateKey, publicKey := alipayTestKeys(t)
	return Config{Provider: ProviderAlipay, Values: map[string]string{
		"payment.alipay.app_id":            "9021000122689420",
		"payment.alipay.gateway_url":       "https://openapi.alipay.com/gateway.do",
		"payment.alipay.app_private_key":   privateKey,
		"payment.alipay.alipay_public_key": publicKey,
		"payment.alipay.notify_url":        "https://example.com/api/payment-callbacks/alipay",
		"payment.alipay.return_url":        "https://example.com/payments/return",
	}}
}

func alipayTestKeys(t *testing.T) (string, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	publicDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})
	return string(privatePEM), string(publicPEM)
}
