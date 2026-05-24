package payment

import "time"

type PaymentTransaction struct {
	ID               uint64     `gorm:"column:id;primaryKey"`
	PaymentNo        string     `gorm:"column:payment_no"`
	OrderID          uint64     `gorm:"column:order_id"`
	OrderNo          string     `gorm:"column:order_no"`
	UserID           uint64     `gorm:"column:user_id"`
	Provider         string     `gorm:"column:provider"`
	Method           string     `gorm:"column:method"`
	Status           string     `gorm:"column:status"`
	ClientToken      string     `gorm:"column:client_token"`
	AmountCents      uint64     `gorm:"column:amount_cents"`
	Currency         string     `gorm:"column:currency"`
	UpstreamTradeNo  *string    `gorm:"column:upstream_trade_no"`
	UpstreamPrepayID *string    `gorm:"column:upstream_prepay_id"`
	QRCodeURL        *string    `gorm:"column:qr_code_url"`
	RedirectURL      *string    `gorm:"column:redirect_url"`
	CallbackSummary  *string    `gorm:"column:callback_summary"`
	QuerySummary     *string    `gorm:"column:query_summary"`
	LastErrorCode    *string    `gorm:"column:last_error_code"`
	LastErrorMessage *string    `gorm:"column:last_error_message"`
	ExpiresAt        time.Time  `gorm:"column:expires_at"`
	PaidAt           *time.Time `gorm:"column:paid_at"`
	ClosedAt         *time.Time `gorm:"column:closed_at"`
	FailedAt         *time.Time `gorm:"column:failed_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (PaymentTransaction) TableName() string { return "payment_transactions" }

type RefundTransaction struct {
	ID                 uint64     `gorm:"column:id;primaryKey"`
	RefundNo           string     `gorm:"column:refund_no"`
	PaymentID          uint64     `gorm:"column:payment_id"`
	PaymentNo          string     `gorm:"column:payment_no"`
	OrderID            uint64     `gorm:"column:order_id"`
	OrderNo            string     `gorm:"column:order_no"`
	UserID             uint64     `gorm:"column:user_id"`
	Provider           string     `gorm:"column:provider"`
	Status             string     `gorm:"column:status"`
	AmountCents        uint64     `gorm:"column:amount_cents"`
	Currency           string     `gorm:"column:currency"`
	Reason             string     `gorm:"column:reason"`
	RequestedByAdminID uint64     `gorm:"column:requested_by_admin_id"`
	UpstreamRefundNo   *string    `gorm:"column:upstream_refund_no"`
	UpstreamTradeNo    *string    `gorm:"column:upstream_trade_no"`
	CallbackSummary    *string    `gorm:"column:callback_summary"`
	QuerySummary       *string    `gorm:"column:query_summary"`
	LastErrorCode      *string    `gorm:"column:last_error_code"`
	LastErrorMessage   *string    `gorm:"column:last_error_message"`
	ChannelConfirmedAt *time.Time `gorm:"column:channel_confirmed_at"`
	CompletedAt        *time.Time `gorm:"column:completed_at"`
	FailedAt           *time.Time `gorm:"column:failed_at"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at"`
}

func (RefundTransaction) TableName() string { return "refund_transactions" }

type PaymentEffect struct {
	ID              uint64     `gorm:"column:id;primaryKey"`
	EffectNo        string     `gorm:"column:effect_no"`
	PaymentID       uint64     `gorm:"column:payment_id"`
	PaymentNo       string     `gorm:"column:payment_no"`
	OrderID         uint64     `gorm:"column:order_id"`
	OrderNo         string     `gorm:"column:order_no"`
	OrderType       string     `gorm:"column:order_type"`
	EffectType      string     `gorm:"column:effect_type"`
	Status          string     `gorm:"column:status"`
	InstanceID      *uint64    `gorm:"column:instance_id"`
	InstanceNo      *string    `gorm:"column:instance_no"`
	BeforeExpiresAt *time.Time `gorm:"column:before_expires_at"`
	AfterExpiresAt  *time.Time `gorm:"column:after_expires_at"`
	RefundID        *uint64    `gorm:"column:refund_id"`
	RefundNo        *string    `gorm:"column:refund_no"`
	AppliedAt       time.Time  `gorm:"column:applied_at"`
	RevertedAt      *time.Time `gorm:"column:reverted_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

func (PaymentEffect) TableName() string { return "payment_effects" }

type PaymentRow struct {
	PaymentTransaction
	Username    string
	Email       string
	DisplayName *string
	OrderStatus string
	OrderType   string
}

type RefundRow struct {
	RefundTransaction
	Username    string
	Email       string
	DisplayName *string
}
