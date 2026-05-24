package payment

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	wechatutils "github.com/wechatpay-apiv3/wechatpay-go/utils"
)

func TestWechatAdapterQueryPaymentUsesSDKSignatureAndMapsPaidState(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.NotEmpty(t, req.Header.Get("Authorization"))
		require.Contains(t, req.URL.Path, "/v3/pay/transactions/out-trade-no/PAY-WX-1")
		return signedWechatResponse(t, req, `{"out_trade_no":"PAY-WX-1","transaction_id":"WX-TX-1","trade_state":"SUCCESS","trade_type":"NATIVE","amount":{"total":3000,"currency":"CNY"}}`), nil
	})}
	adapter := NewWechatAdapterWithHTTPClient(client)

	result, err := adapter.QueryPayment(context.Background(), wechatTestConfig(t), QueryPaymentRequest{PaymentNo: "PAY-WX-1", Method: MethodWechatNative})
	require.NoError(t, err)
	require.Equal(t, "PAY-WX-1", result.PaymentNo)
	require.Equal(t, "WX-TX-1", result.UpstreamTradeNo)
	require.Equal(t, uint64(3000), result.AmountCents)
	require.Equal(t, StatusPaid, result.Status)
}

func TestWechatAdapterQueryRefundMapsSuccess(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.NotEmpty(t, req.Header.Get("Authorization"))
		require.Contains(t, req.URL.Path, "/v3/refund/domestic/refunds/RF-WX-1")
		return signedWechatResponse(t, req, `{"refund_id":"WX-RF-1","out_refund_no":"RF-WX-1","transaction_id":"WX-TX-1","out_trade_no":"PAY-WX-1","channel":"ORIGINAL","user_received_account":"零钱","create_time":"2026-05-24T12:00:00+08:00","success_time":"2026-05-24T12:01:00+08:00","status":"SUCCESS","amount":{"total":3000,"refund":3000,"payer_total":3000,"payer_refund":3000,"settlement_refund":3000,"settlement_total":3000,"discount_refund":0,"currency":"CNY"}}`), nil
	})}
	adapter := NewWechatAdapterWithHTTPClient(client)

	result, err := adapter.QueryRefund(context.Background(), wechatTestConfig(t), QueryRefundRequest{RefundNo: "RF-WX-1"})
	require.NoError(t, err)
	require.Equal(t, "RF-WX-1", result.RefundNo)
	require.Equal(t, "WX-RF-1", result.UpstreamRefundNo)
	require.Equal(t, RefundStatusSucceeded, result.Status)
}

func TestWechatAdapterRejectsInvalidNotificationSignature(t *testing.T) {
	adapter := NewWechatAdapter()
	req, err := http.NewRequest("POST", "/api/payment-callbacks/wechat", strings.NewReader(`{"id":"notify-1","create_time":"2026-05-24T12:00:00+08:00","event_type":"TRANSACTION.SUCCESS","resource_type":"encrypt-resource","resource":{"algorithm":"AEAD_AES_256_GCM","ciphertext":"invalid","associated_data":"transaction","nonce":"nonce"}}`))
	require.NoError(t, err)
	req.Header.Set("Wechatpay-Signature", "invalid")
	req.Header.Set("Wechatpay-Serial", "PUB_KEY_ID_TEST")
	req.Header.Set("Wechatpay-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	req.Header.Set("Wechatpay-Nonce", "nonce")

	_, err = adapter.ParseNotification(context.Background(), wechatTestConfig(t), req)
	require.ErrorIs(t, err, ErrInvalidSignature)
}

func signedWechatResponse(t *testing.T, req *http.Request, body string) *http.Response {
	t.Helper()
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := "test-nonce"
	signature, err := wechatutils.SignSHA256WithRSA(fmt.Sprintf("%s\n%s\n%s\n", timestamp, nonce, body), wechatPlatformPrivateKey(t))
	require.NoError(t, err)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Request:    req,
	}
	resp.Header.Set("Content-Type", "application/json")
	resp.Header.Set("Request-Id", "test-request-id")
	resp.Header.Set("Wechatpay-Serial", "PUB_KEY_ID_TEST")
	resp.Header.Set("Wechatpay-Timestamp", timestamp)
	resp.Header.Set("Wechatpay-Nonce", nonce)
	resp.Header.Set("Wechatpay-Signature", signature)
	return resp
}

func wechatTestConfig(t *testing.T) Config {
	return Config{Provider: ProviderWechat, Values: map[string]string{
		"payment.wechat.app_id":                    "wx-test",
		"payment.wechat.mch_id":                    "example-mchid",
		"payment.wechat.api_v3_key":                "12345678901234567890123456789012",
		"payment.wechat.mch_private_key":           strings.ReplaceAll(wechatMerchantPrivateKey, "TESTING KEY", "PRIVATE KEY"),
		"payment.wechat.mch_certificate_serial_no": "example-sn",
		"payment.wechat.platform_public_key_id":    "PUB_KEY_ID_TEST",
		"payment.wechat.platform_public_key":       wechatPlatformPublicKeyPEM(t),
		"payment.wechat.notify_url":                "https://example.com/api/payment-callbacks/wechat",
		"payment.wechat.h5_scene_info":             `{"type":"Wap","app_name":"pveCloud","app_url":"https://example.com"}`,
	}}
}

func wechatPlatformPrivateKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := wechatutils.LoadPrivateKey(strings.ReplaceAll(wechatPlatformPrivateKeyPEM, "TESTING KEY", "PRIVATE KEY"))
	require.NoError(t, err)
	return key
}

func wechatPlatformPublicKeyPEM(t *testing.T) string {
	t.Helper()
	key := wechatPlatformPrivateKey(t)
	pub, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	return "-----BEGIN PUBLIC KEY-----\n" + pemLineWrap(pub) + "-----END PUBLIC KEY-----\n"
}

func pemLineWrap(der []byte) string {
	encoded := base64.StdEncoding.EncodeToString(der)
	var builder strings.Builder
	for len(encoded) > 64 {
		builder.WriteString(encoded[:64] + "\n")
		encoded = encoded[64:]
	}
	builder.WriteString(encoded + "\n")
	return builder.String()
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

const wechatMerchantPrivateKey = `-----BEGIN TESTING KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCkxOav8p5RFFmN
7hjLGrNtXPYgCd0Zuvxabv+IWl1HVkWi/1iVqac+XwKH/ZeCYDURqZ6P0iiq8NBd
pygoeJiQM+qzaTV3alNXdLcpcaQSNqJ16a7Z2Co5LgOkBJHIF1qUWf+BpAtyLqo5
iGUQ47w3IWQHtfpW7RrECiNsI7kGKuSc4U2JX/gxrFG7ugpRA9Gp1eF0/wBMWSKV
mKveKERvAueaTKEujpN4lctoO1wsMW93nuFNH1gtHPYmkaZaJS88GEp0VYJIcOpX
2HlVPPiWx+0KAndcqMLVQ+qTk21tivpjuqPDstxcT9cXn/3CzSEkYrKkMLpjpSl/
In+amHddAgMBAAECggEADtQRlsAU82MLdDR7UrwCbdMx60w387raPyFCKflH78WZ
2sN0K3PrMzfFuItf+UHDROWo+XSGaGvntKX4fTvtLv0dICxVvXt6KKK+YSJzC5iT
Il13eO91TVQQy9AFdqZzZmp7DiW/SfVdKHRX9B8qryN4JyF/eBc6k23+JhtI6X8J
rPeXGcw/Muo2cvoZ6oarRvzKU7pDivZADavAsgGwi+QwZUtrpUhDbJxWxoasCrWz
6X24D+JbKndf/AeyHC2mqoAwTYdQgkkPHLJWGsqAt/GHtHLmZHIPc0dXfksP5omX
OIii34YNud1j2/X0ryxoRBKPGIolV5Tyyh9PX75UAQKBgQDa/oPFUNsBo+NkfE4r
Mry+mnAeBM06vZh65acbHu3prqzn7tzfQR5rF9nIEUjNkach+ZeTnUxKfX4TwfiL
ibbM1Epb+yhEpSlA9HhY1AGhdNKpFbq63oVzS+lwpmLcXFLHhOYw8GrnYH9lcE+E
YCqFcuk4t/y3rU/8Y5GhZKZ53QKBgQDAnKk6CEgnSkXfZLmzVtoM/5hnI0GEDOUo
V35alqvgdJtCiPs4C03snYLVHqHjAknGzLGONAQ6h9au2qwcHy1qH/noq151u92b
bCrKmghnm2SoIgCaZ7i2scWm6NM9Da60H662WxjaKcZMnUClm+G+Irl9m5cm1i3V
56ZU63nbgQKBgHyAlFO6mzg8f4via+J9TvciADngyvjpT2YXaECv/dyL9TtK/oFi
mTOTdLocsYJFm3piVv2SQQxcejArZ+2U1rtuufO/P25/Y4vNMRp3NZIgQ5/jfay9
06rv7oCf57aWOm26LdCG7pAquWLnTh3ZOnNyGAup9mBKhR3dUa8q9MZ1AoGATZGJ
0VYugKw3sXymEKRkkiGJJdgb9WsgCnwZ5a+SLoWnVUdHLM3YpvbUDrIUbhCo14ft
5Z/rKAs2mRp1f6nKp1eTVHFXTEDJQWNxZEBeLCN3iQKQjZ5B1EmJmOtgztCoz9+G
g+fx/UIfmxElTMyXP/RKEVzMpZZRxThSUxa174ECgYEAsviAmgBskM2ibtrA8tNW
jHdgut0xAtIJIIHlYmtksWgAWD4cPtCg7HPurXqBKxMogH/ZsZc6/5PpOIRYBNl0
EebH0MZ/yiEOrmgsFJ1gKWk3fx8/yLQBlhn32AIhj7wmcFwzi/4hcwihHRCjS7t5
QhVpKswxQyxqeIdQw0CgKZY=
-----END TESTING KEY-----`

const wechatPlatformPrivateKeyPEM = `-----BEGIN TESTING KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDzbMaiMbCzJ11sZ2r7/XisolGu1pVpWvnv1PVOAWZZAWr/WcNNuLni8ddJkIs8NdTz+iAHtb9XMNG4hj7d10cy8QE6QG8YUef1fGb47Wee3DjJRk8N9lWyPDAy9AW70yYWItl/05XgkGt2eJuoU5CcZ+Cy0u2nAXxEGs2Z4Fg/100Ylcyq4GrimjngIyUFLnowOcZltJUQSw/Iu63V0BEh9PMNnXhKkwS2xfkA2nZRzSgzpMvX74v8F7Zf+HkyrHHzwY/7YUWg3pSj3GD0xKsJOwzz0MLlS8uIdLj1lKGzzt1ROgwe0sM5LL5XMfmbjDhcVBmyxQI80WiNaC281tVhAgMBAAECggEBAJ134Wrs0Ayky2ej4u5OAvFSM5rxj0fPJV3DGkiy2R18sFWtII03kXBA1+7rxVZW0IJfbLbwGG3z08cVeLeTWqiWhR/ErNlDqtT/+7DOCrkWZtm1VNCIaNla3Ccp+keNiNbLBn4NRqg1ZH8H+FHEdQjonc+waTIe4N9Bo30GRrBMeMbAgN8mwQhZ+6R23j3GsrJOViFpPRgGhih4aEAORxU+DWl22vklxO8lnDzSwfnXvNDvapnPaA8VXekkThxABq3p/ggv5MI1QPYl8BU5PWd6AJlPs/u2nzFGmCgVufHMjdszWYd3hbj5EmhlL1VMs4uxhCC8OM+ypbnx0CmBB5ECgYEA+W3KJciw4qG6XqWgjK9hkgiZrO8z3tu6tQ6ge28f3BxYEbGUab8bKKzacbYODRiRCM5oKcrXfTqP6IbqUoESiuoz0CUPbShp7k00wXKd7BH8LoIECDbctz77NG0KBE31OGEJw4hm762M14V9nU0KDHGuud1CH1fqivTGG0g2HiUCgYEA+dZ+e5iSvin+omXkr3Vhwf3kutX+GNkKm5LrWZNmWybOKw/K1YFvESpfl1b6YgjA/qUXj0tbOumFQw6e/OLfxIvK/dtkd+7pbSC4T9w8rH7zgjYdJ030Nyv3UGfHDbES+z9MDMo5h+3RKsLN2bp1JcXp6vht9CiXDYm1df3O340CgYB118Qg08+WU1iM7O2MajPL3dpVFPJJwUBV2GJDzv2bbZzCR0baKxr2vau6+4tp7ohfQ718uUPT+34QGuXMMwUCsqHmHgxKw0RA/SMGnlM0PE8L3gtvohPnU481dqq72+UWTOpjAie35yPak0wErGgp9u/ZCkr6Kfw6yGhsbVJ8LQKBgEBLxS1FrK4n3JIqqtnE2a21C4JRxBzc7m/vNYZN+s+GgxRt8gNUViMSxpsKFVHZcuGV1yRXflkA8/y37I6kTHYmi80dAxQidgxRmV1kDnFOEpj2GDafRzRTqkgVDRMm+P2T4pyABqJGv8fDbnqUE8Xu0y5XVOS69XTUddCxyuWZAoGAbpF6JOh6B7OV4XRTDm98Z8OPYmYd9JQ4xt8bqsG9LdzvhU/PI4zwIaKDqZ8vzCI+r8TOrC6SBEfjAe6o2FEExmFWTjBAVCp+Qvnz+Pj7d+WP3kCX/B62IZckVhdV3a2frMTBPvAh8XdENdvsu4DsWJBCA54GLU3wdUa/FO0RUsA=
-----END TESTING KEY-----`
