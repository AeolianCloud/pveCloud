package dto

import "time"

type PaymentCreateRequest struct {
	Provider    string `json:"provider" validate:"required,oneof=alipay wechat"`
	Method      string `json:"method" validate:"required,oneof=alipay_page alipay_wap wechat_native wechat_h5"`
	ClientToken string `json:"client_token" validate:"required,max=128"`
}

type PaymentStatus struct {
	PaymentNo          string     `json:"payment_no"`
	OrderNo            string     `json:"order_no"`
	Provider           string     `json:"provider"`
	Method             string     `json:"method"`
	AmountCents        uint64     `json:"amount_cents"`
	Currency           string     `json:"currency"`
	Status             string     `json:"status"`
	ExpiresAt          time.Time  `json:"expires_at"`
	RedirectURL        *string    `json:"redirect_url"`
	QRCodeURL          *string    `json:"qr_code_url"`
	PaidAt             *time.Time `json:"paid_at"`
	OrderStatus        string     `json:"order_status"`
	OrderPaymentStatus string     `json:"order_payment_status"`
	RelatedInstanceNo  *string    `json:"related_instance_no"`
	LastErrorMessage   *string    `json:"last_error_message"`
}

type PaymentCallbackRequest struct {
	PaymentNo       string `json:"payment_no" validate:"omitempty,max=64"`
	Provider        string `json:"provider" validate:"omitempty,oneof=alipay wechat"`
	UpstreamTradeNo string `json:"upstream_trade_no" validate:"omitempty,max=128"`
	AmountCents     uint64 `json:"amount_cents" validate:"required,min=1"`
	Status          string `json:"status" validate:"required,oneof=paid closed failed refunded"`
	Summary         string `json:"-" validate:"-"`
}
