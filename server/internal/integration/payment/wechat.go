package payment

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	paymenth5 "github.com/wechatpay-apiv3/wechatpay-go/services/payments/h5"
	paymentnative "github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	wechatutils "github.com/wechatpay-apiv3/wechatpay-go/utils"
)

type WechatAdapter struct {
	httpClient *http.Client
}

func NewWechatAdapter() *WechatAdapter { return &WechatAdapter{} }

func NewWechatAdapterWithHTTPClient(client *http.Client) *WechatAdapter {
	return &WechatAdapter{httpClient: client}
}

func newWechatClient(ctx context.Context, cfg Config, opts ...core.ClientOption) (*core.Client, error) {
	privateKey, publicKey, err := wechatKeys(cfg)
	if err != nil {
		return nil, err
	}
	clientOpts := []core.ClientOption{option.WithWechatPayPublicKeyAuthCipher(cfg.Value("payment.wechat.mch_id"), cfg.Value("payment.wechat.mch_certificate_serial_no"), privateKey, cfg.Value("payment.wechat.platform_public_key_id"), publicKey)}
	clientOpts = append(clientOpts, opts...)
	return core.NewClient(ctx, clientOpts...)
}

func (a *WechatAdapter) CreatePayment(ctx context.Context, cfg Config, req CreatePaymentRequest) (CreatePaymentResult, error) {
	if err := ValidateProviderConfig(cfg, req.Method); err != nil {
		return CreatePaymentResult{}, err
	}
	client, err := a.client(ctx, cfg)
	if err != nil {
		return CreatePaymentResult{}, err
	}
	amount := int64(req.AmountCents)
	switch req.Method {
	case MethodWechatNative:
		svc := paymentnative.NativeApiService{Client: client}
		resp, _, err := svc.Prepay(ctx, paymentnative.PrepayRequest{
			Appid:       core.String(cfg.Value("payment.wechat.app_id")),
			Mchid:       core.String(cfg.Value("payment.wechat.mch_id")),
			Description: core.String(firstNonEmpty(req.Subject, req.OrderNo)),
			OutTradeNo:  core.String(req.PaymentNo),
			TimeExpire:  &req.ExpiresAt,
			NotifyUrl:   core.String(cfg.Value("payment.wechat.notify_url")),
			Amount:      &paymentnative.Amount{Total: core.Int64(amount), Currency: core.String(req.Currency)},
			SceneInfo:   &paymentnative.SceneInfo{PayerClientIp: core.String(firstNonEmpty(req.ClientIP, "127.0.0.1"))},
		})
		if err != nil {
			return CreatePaymentResult{}, err
		}
		return CreatePaymentResult{QRCodeURL: valueOf(resp.CodeUrl), Summary: Summary(map[string]any{"provider": ProviderWechat, "method": req.Method, "out_trade_no": req.PaymentNo})}, nil
	case MethodWechatH5:
		scene, err := h5SceneInfo(cfg.Value("payment.wechat.h5_scene_info"), req.ClientIP)
		if err != nil {
			return CreatePaymentResult{}, err
		}
		svc := paymenth5.H5ApiService{Client: client}
		resp, _, err := svc.Prepay(ctx, paymenth5.PrepayRequest{
			Appid:       core.String(cfg.Value("payment.wechat.app_id")),
			Mchid:       core.String(cfg.Value("payment.wechat.mch_id")),
			Description: core.String(firstNonEmpty(req.Subject, req.OrderNo)),
			OutTradeNo:  core.String(req.PaymentNo),
			TimeExpire:  &req.ExpiresAt,
			NotifyUrl:   core.String(cfg.Value("payment.wechat.notify_url")),
			Amount:      &paymenth5.Amount{Total: core.Int64(amount), Currency: core.String(req.Currency)},
			SceneInfo:   scene,
		})
		if err != nil {
			return CreatePaymentResult{}, err
		}
		return CreatePaymentResult{RedirectURL: valueOf(resp.H5Url), Summary: Summary(map[string]any{"provider": ProviderWechat, "method": req.Method, "out_trade_no": req.PaymentNo})}, nil
	default:
		return CreatePaymentResult{}, ErrUnsupportedProvider
	}
}

func (a *WechatAdapter) ParseNotification(ctx context.Context, cfg Config, req *http.Request) (NotificationResult, error) {
	if err := ValidateProviderConfig(cfg, ""); err != nil {
		return NotificationResult{}, err
	}
	handler, err := a.notifyHandler(cfg)
	if err != nil {
		return NotificationResult{}, err
	}
	transaction := new(payments.Transaction)
	if _, err := handler.ParseNotifyRequest(ctx, req, transaction); err != nil {
		return NotificationResult{}, fmt.Errorf("%w: %v", ErrInvalidSignature, err)
	}
	amount := uint64(0)
	currency := "CNY"
	if transaction.Amount != nil {
		if transaction.Amount.Total != nil && *transaction.Amount.Total > 0 {
			amount = uint64(*transaction.Amount.Total)
		}
		if transaction.Amount.Currency != nil && *transaction.Amount.Currency != "" {
			currency = *transaction.Amount.Currency
		}
	}
	return NotificationResult{
		PaymentNo:       valueOf(transaction.OutTradeNo),
		Provider:        ProviderWechat,
		UpstreamTradeNo: valueOf(transaction.TransactionId),
		AmountCents:     amount,
		Currency:        currency,
		Status:          wechatTradeStatus(valueOf(transaction.TradeState)),
		Summary:         wechatTransactionSummary(transaction),
	}, nil
}

func (a *WechatAdapter) QueryPayment(ctx context.Context, cfg Config, req QueryPaymentRequest) (QueryPaymentResult, error) {
	if err := ValidateProviderConfig(cfg, req.Method); err != nil {
		return QueryPaymentResult{}, err
	}
	client, err := a.client(ctx, cfg)
	if err != nil {
		return QueryPaymentResult{}, err
	}
	var tx *payments.Transaction
	switch req.Method {
	case MethodWechatH5:
		svc := paymenth5.H5ApiService{Client: client}
		tx, _, err = svc.QueryOrderByOutTradeNo(ctx, paymenth5.QueryOrderByOutTradeNoRequest{OutTradeNo: core.String(req.PaymentNo), Mchid: core.String(cfg.Value("payment.wechat.mch_id"))})
	default:
		svc := paymentnative.NativeApiService{Client: client}
		tx, _, err = svc.QueryOrderByOutTradeNo(ctx, paymentnative.QueryOrderByOutTradeNoRequest{OutTradeNo: core.String(req.PaymentNo), Mchid: core.String(cfg.Value("payment.wechat.mch_id"))})
	}
	if err != nil {
		return QueryPaymentResult{}, err
	}
	amount := uint64(0)
	currency := "CNY"
	if tx.Amount != nil {
		if tx.Amount.Total != nil && *tx.Amount.Total > 0 {
			amount = uint64(*tx.Amount.Total)
		}
		if tx.Amount.Currency != nil && *tx.Amount.Currency != "" {
			currency = *tx.Amount.Currency
		}
	}
	return QueryPaymentResult{PaymentNo: valueOf(tx.OutTradeNo), UpstreamTradeNo: valueOf(tx.TransactionId), AmountCents: amount, Currency: currency, Status: wechatTradeStatus(valueOf(tx.TradeState)), Summary: wechatTransactionSummary(tx)}, nil
}

func (a *WechatAdapter) CreateRefund(ctx context.Context, cfg Config, req CreateRefundRequest) (RefundResult, error) {
	if err := ValidateProviderConfig(cfg, ""); err != nil {
		return RefundResult{}, err
	}
	client, err := a.client(ctx, cfg)
	if err != nil {
		return RefundResult{}, err
	}
	amount := int64(req.AmountCents)
	svc := refunddomestic.RefundsApiService{Client: client}
	refund, _, err := svc.Create(ctx, refunddomestic.CreateRequest{
		TransactionId: core.String(req.UpstreamTradeNo),
		OutTradeNo:    core.String(req.PaymentNo),
		OutRefundNo:   core.String(req.RefundNo),
		Reason:        core.String(req.Reason),
		Amount:        &refunddomestic.AmountReq{Refund: core.Int64(amount), Total: core.Int64(amount), Currency: core.String(req.Currency)},
	})
	if err != nil {
		return RefundResult{}, err
	}
	return wechatRefundResult(refund, req.RefundNo, req.AmountCents, req.Currency), nil
}

func (a *WechatAdapter) QueryRefund(ctx context.Context, cfg Config, req QueryRefundRequest) (RefundResult, error) {
	if err := ValidateProviderConfig(cfg, ""); err != nil {
		return RefundResult{}, err
	}
	client, err := a.client(ctx, cfg)
	if err != nil {
		return RefundResult{}, err
	}
	svc := refunddomestic.RefundsApiService{Client: client}
	refund, _, err := svc.QueryByOutRefundNo(ctx, refunddomestic.QueryByOutRefundNoRequest{OutRefundNo: core.String(req.RefundNo)})
	if err != nil {
		return RefundResult{}, err
	}
	return wechatRefundResult(refund, req.RefundNo, 0, "CNY"), nil
}

func (a *WechatAdapter) client(ctx context.Context, cfg Config) (*core.Client, error) {
	var opts []core.ClientOption
	if a.httpClient != nil {
		opts = append(opts, option.WithHTTPClient(a.httpClient))
	}
	return newWechatClient(ctx, cfg, opts...)
}

func (a *WechatAdapter) notifyHandler(cfg Config) (*notify.Handler, error) {
	_, publicKey, err := wechatKeys(cfg)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher([]byte(cfg.Value("payment.wechat.api_v3_key")))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	handler := notify.NewEmptyHandler()
	handler.AddRSAWithAESGCM(verifiers.NewSHA256WithRSAPubkeyVerifier(cfg.Value("payment.wechat.platform_public_key_id"), *publicKey), gcm)
	return handler, nil
}

func wechatKeys(cfg Config) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := wechatutils.LoadPrivateKey(cfg.Value("payment.wechat.mch_private_key"))
	if err != nil {
		return nil, nil, err
	}
	publicKey, err := wechatutils.LoadPublicKey(cfg.Value("payment.wechat.platform_public_key"))
	if err != nil {
		return nil, nil, err
	}
	return privateKey, publicKey, nil
}

func h5SceneInfo(raw string, clientIP string) (*paymenth5.SceneInfo, error) {
	type sceneConfig struct {
		Type        string `json:"type"`
		AppName     string `json:"app_name"`
		AppURL      string `json:"app_url"`
		BundleID    string `json:"bundle_id"`
		PackageName string `json:"package_name"`
	}
	var cfg sceneConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, err
	}
	return &paymenth5.SceneInfo{
		PayerClientIp: core.String(firstNonEmpty(clientIP, "127.0.0.1")),
		H5Info: &paymenth5.H5Info{
			Type:        core.String(firstNonEmpty(cfg.Type, "Wap")),
			AppName:     optionalString(cfg.AppName),
			AppUrl:      optionalString(cfg.AppURL),
			BundleId:    optionalString(cfg.BundleID),
			PackageName: optionalString(cfg.PackageName),
		},
	}, nil
}

func wechatTradeStatus(state string) string {
	switch strings.ToUpper(strings.TrimSpace(state)) {
	case "SUCCESS":
		return StatusPaid
	case "CLOSED", "REVOKED":
		return StatusClosed
	case "PAYERROR":
		return StatusFailed
	default:
		return StatusPending
	}
}

func wechatRefundStatus(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "SUCCESS":
		return RefundStatusSucceeded
	case "CLOSED", "ABNORMAL":
		return RefundStatusFailed
	default:
		return RefundStatusPending
	}
}

func wechatRefundResult(refund *refunddomestic.Refund, fallbackRefundNo string, fallbackAmount uint64, fallbackCurrency string) RefundResult {
	amount := fallbackAmount
	currency := firstNonEmpty(fallbackCurrency, "CNY")
	if refund.Amount != nil {
		if refund.Amount.Refund != nil && *refund.Amount.Refund > 0 {
			amount = uint64(*refund.Amount.Refund)
		}
		if refund.Amount.Currency != nil && *refund.Amount.Currency != "" {
			currency = *refund.Amount.Currency
		}
	}
	return RefundResult{
		RefundNo:         firstNonEmpty(valueOf(refund.OutRefundNo), fallbackRefundNo),
		UpstreamRefundNo: valueOf(refund.RefundId),
		UpstreamTradeNo:  valueOf(refund.TransactionId),
		AmountCents:      amount,
		Currency:         currency,
		Status:           wechatRefundStatus(refundStatusValue(refund.Status)),
		Summary:          Summary(map[string]any{"refund_id": valueOf(refund.RefundId), "out_refund_no": valueOf(refund.OutRefundNo), "status": refundStatusValue(refund.Status)}),
	}
}

func wechatTransactionSummary(tx *payments.Transaction) string {
	return Summary(map[string]any{"transaction_id": valueOf(tx.TransactionId), "out_trade_no": valueOf(tx.OutTradeNo), "trade_state": valueOf(tx.TradeState), "trade_type": valueOf(tx.TradeType)})
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return core.String(strings.TrimSpace(value))
}

func valueOf(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func refundStatusValue(value *refunddomestic.Status) string {
	if value == nil {
		return ""
	}
	return string(*value)
}
