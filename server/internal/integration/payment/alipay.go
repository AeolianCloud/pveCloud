package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	alipay "github.com/smartwalle/alipay/v3"
)

type AlipayAdapter struct{}

func NewAlipayAdapter() *AlipayAdapter { return &AlipayAdapter{} }

func newAlipayClient(cfg Config) (*alipay.Client, error) {
	return (&AlipayAdapter{}).client(cfg)
}

func (a *AlipayAdapter) CreatePayment(ctx context.Context, cfg Config, req CreatePaymentRequest) (CreatePaymentResult, error) {
	if err := ValidateProviderConfig(cfg, req.Method); err != nil {
		return CreatePaymentResult{}, err
	}
	client, err := a.client(cfg)
	if err != nil {
		return CreatePaymentResult{}, err
	}
	trade := alipay.Trade{
		Subject:        firstNonEmpty(req.Subject, req.OrderNo),
		OutTradeNo:     req.PaymentNo,
		TotalAmount:    centsToYuan(req.AmountCents),
		ProductCode:    alipayProductCode(req.Method),
		NotifyURL:      cfg.Value("payment.alipay.notify_url"),
		ReturnURL:      cfg.Value("payment.alipay.return_url"),
		TimeoutExpress: timeoutExpress(req.ExpiresAt),
	}
	var payURL string
	switch req.Method {
	case MethodAlipayPage:
		u, err := client.TradePagePay(alipay.TradePagePay{Trade: trade})
		if err != nil {
			return CreatePaymentResult{}, err
		}
		payURL = u.String()
	case MethodAlipayWap:
		u, err := client.TradeWapPay(alipay.TradeWapPay{Trade: trade, TimeExpire: req.ExpiresAt.Format("2006-01-02 15:04:05")})
		if err != nil {
			return CreatePaymentResult{}, err
		}
		payURL = u.String()
	default:
		return CreatePaymentResult{}, ErrUnsupportedProvider
	}
	return CreatePaymentResult{
		RedirectURL: payURL,
		Summary:     Summary(map[string]any{"provider": ProviderAlipay, "method": req.Method, "out_trade_no": req.PaymentNo}),
	}, nil
}

func (a *AlipayAdapter) ParseNotification(ctx context.Context, cfg Config, req *http.Request) (NotificationResult, error) {
	if err := ValidateProviderConfig(cfg, ""); err != nil {
		return NotificationResult{}, err
	}
	client, err := a.client(cfg)
	if err != nil {
		return NotificationResult{}, err
	}
	if err := req.ParseForm(); err != nil {
		return NotificationResult{}, err
	}
	notification, err := client.DecodeNotification(ctx, req.Form)
	if err != nil {
		return NotificationResult{}, fmt.Errorf("%w: %v", ErrInvalidSignature, err)
	}
	amount, err := yuanToCents(notification.TotalAmount)
	if err != nil {
		return NotificationResult{}, err
	}
	return NotificationResult{
		PaymentNo:       notification.OutTradeNo,
		Provider:        ProviderAlipay,
		UpstreamTradeNo: notification.TradeNo,
		AmountCents:     amount,
		Currency:        "CNY",
		Status:          alipayTradeStatus(notification.TradeStatus),
		Summary:         alipayNotificationSummary(notification),
	}, nil
}

func (a *AlipayAdapter) QueryPayment(ctx context.Context, cfg Config, req QueryPaymentRequest) (QueryPaymentResult, error) {
	if err := ValidateProviderConfig(cfg, req.Method); err != nil {
		return QueryPaymentResult{}, err
	}
	client, err := a.client(cfg)
	if err != nil {
		return QueryPaymentResult{}, err
	}
	rsp, err := client.TradeQuery(ctx, alipay.TradeQuery{OutTradeNo: req.PaymentNo, TradeNo: req.UpstreamTradeNo})
	if err != nil {
		return QueryPaymentResult{}, err
	}
	if !rsp.IsSuccess() {
		return QueryPaymentResult{}, fmt.Errorf("alipay query failed: %s", rsp.Error.Error())
	}
	amount, _ := yuanToCents(rsp.TotalAmount)
	return QueryPaymentResult{
		PaymentNo:       firstNonEmpty(rsp.OutTradeNo, req.PaymentNo),
		UpstreamTradeNo: rsp.TradeNo,
		AmountCents:     amount,
		Currency:        "CNY",
		Status:          alipayTradeStatus(rsp.TradeStatus),
		Summary:         Summary(map[string]any{"trade_no": rsp.TradeNo, "trade_status": rsp.TradeStatus, "total_amount": rsp.TotalAmount}),
	}, nil
}

func (a *AlipayAdapter) CreateRefund(ctx context.Context, cfg Config, req CreateRefundRequest) (RefundResult, error) {
	if err := ValidateProviderConfig(cfg, ""); err != nil {
		return RefundResult{}, err
	}
	client, err := a.client(cfg)
	if err != nil {
		return RefundResult{}, err
	}
	rsp, err := client.TradeRefund(ctx, alipay.TradeRefund{OutTradeNo: req.PaymentNo, TradeNo: req.UpstreamTradeNo, RefundAmount: centsToYuan(req.AmountCents), RefundReason: req.Reason, OutRequestNo: req.RefundNo})
	if err != nil {
		return RefundResult{}, err
	}
	if !rsp.IsSuccess() {
		return RefundResult{}, fmt.Errorf("alipay refund failed: %s", rsp.Error.Error())
	}
	amount, _ := yuanToCents(rsp.RefundFee)
	return RefundResult{
		RefundNo:        req.RefundNo,
		UpstreamTradeNo: firstNonEmpty(rsp.TradeNo, req.UpstreamTradeNo),
		AmountCents:     amount,
		Currency:        req.Currency,
		Status:          RefundStatusSucceeded,
		Summary:         Summary(map[string]any{"trade_no": rsp.TradeNo, "out_trade_no": rsp.OutTradeNo, "refund_fee": rsp.RefundFee}),
	}, nil
}

func (a *AlipayAdapter) QueryRefund(ctx context.Context, cfg Config, req QueryRefundRequest) (RefundResult, error) {
	return RefundResult{RefundNo: req.RefundNo, UpstreamRefundNo: req.UpstreamRefundNo, Status: RefundStatusPending}, nil
}

func (a *AlipayAdapter) client(cfg Config) (*alipay.Client, error) {
	production := !strings.Contains(strings.ToLower(cfg.Value("payment.alipay.gateway_url")), "sandbox") && !strings.Contains(strings.ToLower(cfg.Value("payment.alipay.gateway_url")), "alipaydev")
	client, err := alipay.New(cfg.Value("payment.alipay.app_id"), cfg.Value("payment.alipay.app_private_key"), production, alipay.WithProductionGateway(cfg.Value("payment.alipay.gateway_url")), alipay.WithSandboxGateway(cfg.Value("payment.alipay.gateway_url")))
	if err != nil {
		return nil, err
	}
	if err := client.LoadAliPayPublicKey(cfg.Value("payment.alipay.alipay_public_key")); err != nil {
		return nil, err
	}
	return client, nil
}

func alipayProductCode(method string) string {
	if method == MethodAlipayWap {
		return "QUICK_WAP_WAY"
	}
	return "FAST_INSTANT_TRADE_PAY"
}

func timeoutExpress(expiresAt time.Time) string {
	minutes := int(time.Until(expiresAt).Minutes())
	if minutes < 1 {
		minutes = 1
	}
	return fmt.Sprintf("%dm", minutes)
}

func alipayTradeStatus(status alipay.TradeStatus) string {
	switch status {
	case alipay.TradeStatusSuccess, alipay.TradeStatusFinished:
		return StatusPaid
	case alipay.TradeStatusClosed:
		return StatusClosed
	default:
		return StatusPending
	}
}

func alipayNotificationSummary(notification *alipay.Notification) string {
	data, _ := json.Marshal(map[string]any{
		"notify_type":    notification.NotifyType,
		"trade_no":       notification.TradeNo,
		"out_trade_no":   notification.OutTradeNo,
		"trade_status":   notification.TradeStatus,
		"total_amount":   notification.TotalAmount,
		"out_request_no": notification.OutRequestNo,
	})
	return string(data)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
