// Package payment contains payment channel adapters.
package payment

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
)

const (
	ProviderAlipay = "alipay"
	ProviderWechat = "wechat"

	MethodAlipayPage   = "alipay_page"
	MethodAlipayWap    = "alipay_wap"
	MethodWechatNative = "wechat_native"
	MethodWechatH5     = "wechat_h5"

	StatusPending  = "pending"
	StatusPaid     = "paid"
	StatusClosed   = "closed"
	StatusFailed   = "failed"
	StatusRefunded = "refunded"

	RefundStatusPending   = "pending"
	RefundStatusSucceeded = "succeeded"
	RefundStatusFailed    = "failed"
)

var (
	ErrUnsupportedProvider = errors.New("unsupported payment provider")
	ErrInvalidSignature    = errors.New("invalid payment signature")
	ErrIncompleteConfig    = errors.New("incomplete payment config")
)

// Adapter is the stable business-facing boundary for payment channels.
//
// SDK-specific request signing, response verification, callback signature
// checks, encryption, and protocol quirks are intentionally contained behind
// this interface so usecases only handle local business state transitions.
type Adapter interface {
	CreatePayment(ctx context.Context, cfg Config, req CreatePaymentRequest) (CreatePaymentResult, error)
	ParseNotification(ctx context.Context, cfg Config, req *http.Request) (NotificationResult, error)
	QueryPayment(ctx context.Context, cfg Config, req QueryPaymentRequest) (QueryPaymentResult, error)
	CreateRefund(ctx context.Context, cfg Config, req CreateRefundRequest) (RefundResult, error)
	QueryRefund(ctx context.Context, cfg Config, req QueryRefundRequest) (RefundResult, error)
}

type Registry interface {
	Adapter(provider string) (Adapter, error)
}

type StaticRegistry map[string]Adapter

func (r StaticRegistry) Adapter(provider string) (Adapter, error) {
	adapter, ok := r[strings.TrimSpace(provider)]
	if !ok || adapter == nil {
		return nil, ErrUnsupportedProvider
	}
	return adapter, nil
}

func NewSDKRegistry() Registry {
	return StaticRegistry{
		ProviderAlipay: NewAlipayAdapter(),
		ProviderWechat: NewWechatAdapter(),
	}
}

type Config struct {
	Provider string
	Values   map[string]string
}

func (c Config) Value(key string) string {
	return strings.TrimSpace(c.Values[key])
}

type CreatePaymentRequest struct {
	PaymentNo   string
	OrderNo     string
	Subject     string
	AmountCents uint64
	Currency    string
	Method      string
	ExpiresAt   time.Time
	ClientIP    string
}

type CreatePaymentResult struct {
	UpstreamTradeNo  string
	UpstreamPrepayID string
	RedirectURL      string
	QRCodeURL        string
	Summary          string
}

type NotificationResult struct {
	PaymentNo       string
	Provider        string
	UpstreamTradeNo string
	AmountCents     uint64
	Currency        string
	Status          string
	Summary         string
}

type QueryPaymentRequest struct {
	PaymentNo       string
	UpstreamTradeNo string
	Method          string
}

type QueryPaymentResult struct {
	PaymentNo       string
	UpstreamTradeNo string
	AmountCents     uint64
	Currency        string
	Status          string
	Summary         string
}

type CreateRefundRequest struct {
	RefundNo        string
	PaymentNo       string
	UpstreamTradeNo string
	AmountCents     uint64
	Currency        string
	Reason          string
}

type RefundResult struct {
	RefundNo         string
	UpstreamRefundNo string
	UpstreamTradeNo  string
	AmountCents      uint64
	Currency         string
	Status           string
	Summary          string
}

type QueryRefundRequest struct {
	RefundNo         string
	UpstreamRefundNo string
}
