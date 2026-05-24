package payment

import (
	"context"
	"net/http"
)

type FakeAdapter struct {
	CreatePaymentFunc     func(context.Context, Config, CreatePaymentRequest) (CreatePaymentResult, error)
	ParseNotificationFunc func(context.Context, Config, *http.Request) (NotificationResult, error)
	QueryPaymentFunc      func(context.Context, Config, QueryPaymentRequest) (QueryPaymentResult, error)
	CreateRefundFunc      func(context.Context, Config, CreateRefundRequest) (RefundResult, error)
	QueryRefundFunc       func(context.Context, Config, QueryRefundRequest) (RefundResult, error)
}

func (f FakeAdapter) CreatePayment(ctx context.Context, cfg Config, req CreatePaymentRequest) (CreatePaymentResult, error) {
	if f.CreatePaymentFunc != nil {
		return f.CreatePaymentFunc(ctx, cfg, req)
	}
	return CreatePaymentResult{}, nil
}

func (f FakeAdapter) ParseNotification(ctx context.Context, cfg Config, req *http.Request) (NotificationResult, error) {
	if f.ParseNotificationFunc != nil {
		return f.ParseNotificationFunc(ctx, cfg, req)
	}
	return NotificationResult{}, ErrInvalidSignature
}

func (f FakeAdapter) QueryPayment(ctx context.Context, cfg Config, req QueryPaymentRequest) (QueryPaymentResult, error) {
	if f.QueryPaymentFunc != nil {
		return f.QueryPaymentFunc(ctx, cfg, req)
	}
	return QueryPaymentResult{PaymentNo: req.PaymentNo, UpstreamTradeNo: req.UpstreamTradeNo, Status: StatusPending}, nil
}

func (f FakeAdapter) CreateRefund(ctx context.Context, cfg Config, req CreateRefundRequest) (RefundResult, error) {
	if f.CreateRefundFunc != nil {
		return f.CreateRefundFunc(ctx, cfg, req)
	}
	return RefundResult{RefundNo: req.RefundNo, UpstreamTradeNo: req.UpstreamTradeNo, AmountCents: req.AmountCents, Currency: req.Currency, Status: RefundStatusSucceeded}, nil
}

func (f FakeAdapter) QueryRefund(ctx context.Context, cfg Config, req QueryRefundRequest) (RefundResult, error) {
	if f.QueryRefundFunc != nil {
		return f.QueryRefundFunc(ctx, cfg, req)
	}
	return RefundResult{RefundNo: req.RefundNo, UpstreamRefundNo: req.UpstreamRefundNo, Status: RefundStatusPending}, nil
}
