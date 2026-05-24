package dto

import "time"

type PaymentListQuery struct {
	Page        int    `form:"page" validate:"omitempty,min=1"`
	PerPage     int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Provider    string `form:"provider" validate:"omitempty,oneof=alipay wechat"`
	Method      string `form:"method" validate:"omitempty,oneof=alipay_page alipay_wap wechat_native wechat_h5"`
	Status      string `form:"status" validate:"omitempty,oneof=pending paid closed failed refunded"`
	OrderNo     string `form:"order_no" validate:"omitempty,max=64"`
	PaymentNo   string `form:"payment_no" validate:"omitempty,max=64"`
	UserKeyword string `form:"user_keyword" validate:"omitempty,max=128"`
	DateFrom    string `form:"date_from" validate:"omitempty,max=32"`
	DateTo      string `form:"date_to" validate:"omitempty,max=32"`
}

type RefundListQuery struct {
	Page      int    `form:"page" validate:"omitempty,min=1"`
	PerPage   int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Provider  string `form:"provider" validate:"omitempty,oneof=alipay wechat"`
	Status    string `form:"status" validate:"omitempty,oneof=pending succeeded failed"`
	OrderNo   string `form:"order_no" validate:"omitempty,max=64"`
	PaymentNo string `form:"payment_no" validate:"omitempty,max=64"`
	RefundNo  string `form:"refund_no" validate:"omitempty,max=64"`
	DateFrom  string `form:"date_from" validate:"omitempty,max=32"`
	DateTo    string `form:"date_to" validate:"omitempty,max=32"`
}

type PaymentItem struct {
	PaymentNo   string           `json:"payment_no"`
	OrderNo     string           `json:"order_no"`
	User        OrderUserSummary `json:"user"`
	Provider    string           `json:"provider"`
	Method      string           `json:"method"`
	Status      string           `json:"status"`
	AmountCents uint64           `json:"amount_cents"`
	Currency    string           `json:"currency"`
	ExpiresAt   time.Time        `json:"expires_at"`
	PaidAt      *time.Time       `json:"paid_at"`
	CreatedAt   time.Time        `json:"created_at"`
	OrderStatus string           `json:"order_status"`
	OrderType   string           `json:"order_type"`
}

type PaymentDetail struct {
	PaymentItem
	UpstreamTradeNo  *string     `json:"upstream_trade_no"`
	LastErrorMessage *string     `json:"last_error_message"`
	Refund           *RefundItem `json:"refund"`
}

type RefundItem struct {
	RefundNo    string           `json:"refund_no"`
	PaymentNo   string           `json:"payment_no"`
	OrderNo     string           `json:"order_no"`
	User        OrderUserSummary `json:"user"`
	Provider    string           `json:"provider"`
	Status      string           `json:"status"`
	AmountCents uint64           `json:"amount_cents"`
	Currency    string           `json:"currency"`
	Reason      string           `json:"reason"`
	CreatedAt   time.Time        `json:"created_at"`
	CompletedAt *time.Time       `json:"completed_at"`
	FailedAt    *time.Time       `json:"failed_at"`
}

type RefundCreateRequest struct {
	Reason string `json:"reason" validate:"required,max=500"`
}
