package wallet

import "time"

type Account struct {
	ID                    uint64    `gorm:"column:id;primaryKey"`
	WalletNo              string    `gorm:"column:wallet_no"`
	UserID                uint64    `gorm:"column:user_id"`
	Currency              string    `gorm:"column:currency"`
	Status                string    `gorm:"column:status"`
	AvailableBalanceCents uint64    `gorm:"column:available_balance_cents"`
	TotalRechargedCents   uint64    `gorm:"column:total_recharged_cents"`
	TotalSpentCents       uint64    `gorm:"column:total_spent_cents"`
	TotalRefundedCents    uint64    `gorm:"column:total_refunded_cents"`
	CreatedAt             time.Time `gorm:"column:created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at"`
}

func (Account) TableName() string { return "wallet_accounts" }

type LedgerEntry struct {
	ID                 uint64    `gorm:"column:id;primaryKey"`
	EntryNo            string    `gorm:"column:entry_no"`
	WalletID           uint64    `gorm:"column:wallet_id"`
	WalletNo           string    `gorm:"column:wallet_no"`
	UserID             uint64    `gorm:"column:user_id"`
	Direction          string    `gorm:"column:direction"`
	EntryType          string    `gorm:"column:entry_type"`
	AmountCents        uint64    `gorm:"column:amount_cents"`
	BalanceBeforeCents uint64    `gorm:"column:balance_before_cents"`
	BalanceAfterCents  uint64    `gorm:"column:balance_after_cents"`
	Currency           string    `gorm:"column:currency"`
	RelatedType        string    `gorm:"column:related_type"`
	RelatedNo          string    `gorm:"column:related_no"`
	IdempotencyKey     string    `gorm:"column:idempotency_key"`
	Summary            *string   `gorm:"column:summary"`
	CreatedAt          time.Time `gorm:"column:created_at"`
}

func (LedgerEntry) TableName() string { return "wallet_ledger_entries" }

type Recharge struct {
	ID               uint64     `gorm:"column:id;primaryKey"`
	RechargeNo       string     `gorm:"column:recharge_no"`
	WalletID         uint64     `gorm:"column:wallet_id"`
	WalletNo         string     `gorm:"column:wallet_no"`
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

func (Recharge) TableName() string { return "wallet_recharges" }

type AccountRow struct {
	Account
	Username    string
	Email       string
	DisplayName *string
}

type LedgerRow struct {
	LedgerEntry
	Username    string
	Email       string
	DisplayName *string
}

type RechargeRow struct {
	Recharge
	Username    string
	Email       string
	DisplayName *string
}
