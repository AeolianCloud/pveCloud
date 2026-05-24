package dto

import "time"

type WalletListQuery struct {
	Page        int    `form:"page" validate:"omitempty,min=1"`
	PerPage     int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	WalletNo    string `form:"wallet_no" validate:"omitempty,max=64"`
	Status      string `form:"status" validate:"omitempty,oneof=active disabled"`
	UserKeyword string `form:"user_keyword" validate:"omitempty,max=128"`
}

type WalletLedgerListQuery struct {
	Page        int    `form:"page" validate:"omitempty,min=1"`
	PerPage     int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	WalletNo    string `form:"wallet_no" validate:"omitempty,max=64"`
	UserKeyword string `form:"user_keyword" validate:"omitempty,max=128"`
	Direction   string `form:"direction" validate:"omitempty,oneof=credit debit"`
	EntryType   string `form:"entry_type" validate:"omitempty,oneof=recharge payment refund"`
	RelatedNo   string `form:"related_no" validate:"omitempty,max=64"`
	DateFrom    string `form:"date_from" validate:"omitempty,max=32"`
	DateTo      string `form:"date_to" validate:"omitempty,max=32"`
}

type WalletRechargeListQuery struct {
	Page        int    `form:"page" validate:"omitempty,min=1"`
	PerPage     int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	WalletNo    string `form:"wallet_no" validate:"omitempty,max=64"`
	UserKeyword string `form:"user_keyword" validate:"omitempty,max=128"`
	Provider    string `form:"provider" validate:"omitempty,oneof=alipay wechat"`
	Method      string `form:"method" validate:"omitempty,oneof=alipay_page alipay_wap wechat_native wechat_h5"`
	Status      string `form:"status" validate:"omitempty,oneof=pending paid closed failed"`
	RechargeNo  string `form:"recharge_no" validate:"omitempty,max=64"`
	DateFrom    string `form:"date_from" validate:"omitempty,max=32"`
	DateTo      string `form:"date_to" validate:"omitempty,max=32"`
}

type WalletUserSummary struct {
	ID          uint64  `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
}

type WalletItem struct {
	WalletNo              string            `json:"wallet_no"`
	User                  WalletUserSummary `json:"user"`
	Currency              string            `json:"currency"`
	Status                string            `json:"status"`
	AvailableBalanceCents uint64            `json:"available_balance_cents"`
	TotalRechargedCents   uint64            `json:"total_recharged_cents"`
	TotalSpentCents       uint64            `json:"total_spent_cents"`
	TotalRefundedCents    uint64            `json:"total_refunded_cents"`
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`
}

type WalletLedgerItem struct {
	EntryNo            string            `json:"entry_no"`
	WalletNo           string            `json:"wallet_no"`
	User               WalletUserSummary `json:"user"`
	Direction          string            `json:"direction"`
	EntryType          string            `json:"entry_type"`
	AmountCents        uint64            `json:"amount_cents"`
	BalanceBeforeCents uint64            `json:"balance_before_cents"`
	BalanceAfterCents  uint64            `json:"balance_after_cents"`
	Currency           string            `json:"currency"`
	RelatedType        string            `json:"related_type"`
	RelatedNo          string            `json:"related_no"`
	Summary            *string           `json:"summary"`
	CreatedAt          time.Time         `json:"created_at"`
}

type WalletRechargeItem struct {
	RechargeNo  string            `json:"recharge_no"`
	WalletNo    string            `json:"wallet_no"`
	User        WalletUserSummary `json:"user"`
	Provider    string            `json:"provider"`
	Method      string            `json:"method"`
	Status      string            `json:"status"`
	AmountCents uint64            `json:"amount_cents"`
	Currency    string            `json:"currency"`
	ExpiresAt   time.Time         `json:"expires_at"`
	PaidAt      *time.Time        `json:"paid_at"`
	ClosedAt    *time.Time        `json:"closed_at"`
	FailedAt    *time.Time        `json:"failed_at"`
	CreatedAt   time.Time         `json:"created_at"`
}

type WalletDetail struct {
	WalletItem
	RecentLedger    []WalletLedgerItem   `json:"recent_ledger"`
	RecentRecharges []WalletRechargeItem `json:"recent_recharges"`
}
