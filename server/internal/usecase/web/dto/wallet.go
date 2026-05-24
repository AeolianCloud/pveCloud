package dto

import "time"

type WalletSummary struct {
	WalletNo              string    `json:"wallet_no"`
	Currency              string    `json:"currency"`
	Status                string    `json:"status"`
	AvailableBalanceCents uint64    `json:"available_balance_cents"`
	TotalRechargedCents   uint64    `json:"total_recharged_cents"`
	TotalSpentCents       uint64    `json:"total_spent_cents"`
	TotalRefundedCents    uint64    `json:"total_refunded_cents"`
	CreatedAt             time.Time `json:"created_at"`
}

type WalletLedgerQuery struct {
	Page      int    `form:"page" validate:"omitempty,min=1"`
	PerPage   int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Direction string `form:"direction" validate:"omitempty,oneof=credit debit"`
	EntryType string `form:"entry_type" validate:"omitempty,oneof=recharge payment refund"`
	RelatedNo string `form:"related_no" validate:"omitempty,max=64"`
	DateFrom  string `form:"date_from" validate:"omitempty,max=32"`
	DateTo    string `form:"date_to" validate:"omitempty,max=32"`
}

type WalletLedgerItem struct {
	EntryNo            string    `json:"entry_no"`
	WalletNo           string    `json:"wallet_no"`
	Direction          string    `json:"direction"`
	EntryType          string    `json:"entry_type"`
	AmountCents        uint64    `json:"amount_cents"`
	BalanceBeforeCents uint64    `json:"balance_before_cents"`
	BalanceAfterCents  uint64    `json:"balance_after_cents"`
	Currency           string    `json:"currency"`
	RelatedType        string    `json:"related_type"`
	RelatedNo          string    `json:"related_no"`
	Summary            *string   `json:"summary"`
	CreatedAt          time.Time `json:"created_at"`
}

type WalletRechargeCreateRequest struct {
	Provider    string `json:"provider" validate:"required,oneof=alipay wechat"`
	Method      string `json:"method" validate:"required,oneof=alipay_page alipay_wap wechat_native wechat_h5"`
	AmountCents uint64 `json:"amount_cents" validate:"required,min=1"`
	ClientToken string `json:"client_token" validate:"required,max=128"`
}

type WalletRechargeStatus struct {
	RechargeNo       string     `json:"recharge_no"`
	WalletNo         string     `json:"wallet_no"`
	Provider         string     `json:"provider"`
	Method           string     `json:"method"`
	Status           string     `json:"status"`
	AmountCents      uint64     `json:"amount_cents"`
	Currency         string     `json:"currency"`
	ExpiresAt        time.Time  `json:"expires_at"`
	PaidAt           *time.Time `json:"paid_at"`
	ClosedAt         *time.Time `json:"closed_at"`
	FailedAt         *time.Time `json:"failed_at"`
	RedirectURL      *string    `json:"redirect_url"`
	QRCodeURL        *string    `json:"qr_code_url"`
	LastErrorMessage *string    `json:"last_error_message"`
	CreatedAt        time.Time  `json:"created_at"`
}
