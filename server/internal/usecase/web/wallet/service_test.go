package wallet

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	domainpayment "github.com/AeolianCloud/pveCloud/server/internal/domain/payment"
	domainwallet "github.com/AeolianCloud/pveCloud/server/internal/domain/wallet"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
)

func TestRechargeNotificationCreditsWalletOnce(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, walletUsersSchema, walletAccountsSchema, walletLedgerSchema, walletRechargesSchema)
	seedWalletUser(t, db, 11)
	seedWalletAccount(t, db, 1001, "WAL-test-1", 11, 1000)
	seedWalletRecharge(t, db, 2001, "RCH-test-1", 1001, "WAL-test-1", 11, 2500)

	service := NewService(db)
	req := webdto.PaymentCallbackRequest{PaymentNo: "RCH-test-1", Provider: domainpayment.ProviderAlipay, UpstreamTradeNo: "ALI-RCH-1", AmountCents: 2500, Status: domainpayment.StatusPaid, Summary: `{"trade":"ALI-RCH-1"}`}
	for i := 0; i < 2; i++ {
		if err := mysqltx.NewManager(db).WithinContext(context.Background(), func(tx *gorm.DB) error {
			return service.ApplyRechargeNotification(context.Background(), tx, req)
		}); err != nil {
			t.Fatalf("apply recharge notification #%d: %v", i+1, err)
		}
	}

	var account struct {
		Balance   uint64 `gorm:"column:available_balance_cents"`
		Recharged uint64 `gorm:"column:total_recharged_cents"`
	}
	if err := db.Table("wallet_accounts").Select("available_balance_cents, total_recharged_cents").Where("wallet_no = ?", "WAL-test-1").Take(&account).Error; err != nil {
		t.Fatalf("load wallet: %v", err)
	}
	if account.Balance != 3500 || account.Recharged != 2500 {
		t.Fatalf("recharge should credit once, got balance=%d recharged=%d", account.Balance, account.Recharged)
	}
	var ledgerCount int64
	if err := db.Table("wallet_ledger_entries").Where("wallet_no = ?", "WAL-test-1").Count(&ledgerCount).Error; err != nil {
		t.Fatalf("count ledger: %v", err)
	}
	if ledgerCount != 1 {
		t.Fatalf("replayed callback should create one ledger entry, got %d", ledgerCount)
	}
}

func seedWalletUser(t *testing.T, db *gorm.DB, userID uint64) {
	t.Helper()
	if err := db.Exec(`INSERT INTO users (id, username, email, password_hash, status) VALUES (?, ?, ?, 'hash', 'active')`, userID, "wallet-user", "wallet@example.com").Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func seedWalletAccount(t *testing.T, db *gorm.DB, id uint64, walletNo string, userID uint64, balance uint64) {
	t.Helper()
	if err := db.Exec(`INSERT INTO wallet_accounts (id, wallet_no, user_id, currency, status, available_balance_cents) VALUES (?, ?, ?, ?, ?, ?)`, id, walletNo, userID, domainwallet.CurrencyCNY, domainwallet.AccountStatusActive, balance).Error; err != nil {
		t.Fatalf("seed wallet account: %v", err)
	}
}

func seedWalletRecharge(t *testing.T, db *gorm.DB, id uint64, rechargeNo string, walletID uint64, walletNo string, userID uint64, amount uint64) {
	t.Helper()
	if err := db.Exec(`INSERT INTO wallet_recharges (id, recharge_no, wallet_id, wallet_no, user_id, provider, method, status, client_token, amount_cents, currency, expires_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, id, rechargeNo, walletID, walletNo, userID, domainpayment.ProviderAlipay, domainpayment.MethodAlipayPage, domainwallet.RechargeStatusPending, "rch-token", amount, domainwallet.CurrencyCNY, time.Now().Add(30*time.Minute)).Error; err != nil {
		t.Fatalf("seed wallet recharge: %v", err)
	}
}

const walletUsersSchema = `
CREATE TABLE users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(64) NOT NULL,
  email VARCHAR(255) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  status VARCHAR(32) NOT NULL,
  display_name VARCHAR(128) NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const walletAccountsSchema = `
CREATE TABLE wallet_accounts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  wallet_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  available_balance_cents BIGINT UNSIGNED NOT NULL DEFAULT 0,
  total_recharged_cents BIGINT UNSIGNED NOT NULL DEFAULT 0,
  total_spent_cents BIGINT UNSIGNED NOT NULL DEFAULT 0,
  total_refunded_cents BIGINT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_wallet_accounts_wallet_no (wallet_no),
  UNIQUE KEY uk_wallet_accounts_user_currency (user_id, currency)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const walletLedgerSchema = `
CREATE TABLE wallet_ledger_entries (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  entry_no VARCHAR(64) NOT NULL,
  wallet_id BIGINT UNSIGNED NOT NULL,
  wallet_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  direction VARCHAR(16) NOT NULL,
  entry_type VARCHAR(32) NOT NULL,
  amount_cents BIGINT UNSIGNED NOT NULL,
  balance_before_cents BIGINT UNSIGNED NOT NULL,
  balance_after_cents BIGINT UNSIGNED NOT NULL,
  currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
  related_type VARCHAR(32) NOT NULL,
  related_no VARCHAR(64) NOT NULL,
  idempotency_key VARCHAR(160) NOT NULL,
  summary JSON NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_wallet_ledger_entries_entry_no (entry_no),
  UNIQUE KEY uk_wallet_ledger_entries_idempotency (wallet_id, idempotency_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const walletRechargesSchema = `
CREATE TABLE wallet_recharges (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  recharge_no VARCHAR(64) NOT NULL,
  wallet_id BIGINT UNSIGNED NOT NULL,
  wallet_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  provider VARCHAR(32) NOT NULL,
  method VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  client_token VARCHAR(128) NOT NULL,
  amount_cents BIGINT UNSIGNED NOT NULL,
  currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
  upstream_trade_no VARCHAR(128) NULL,
  upstream_prepay_id VARCHAR(128) NULL,
  qr_code_url VARCHAR(1000) NULL,
  redirect_url VARCHAR(1000) NULL,
  callback_summary JSON NULL,
  query_summary JSON NULL,
  last_error_code VARCHAR(64) NULL,
  last_error_message VARCHAR(500) NULL,
  expires_at DATETIME(3) NOT NULL,
  paid_at DATETIME(3) NULL,
  closed_at DATETIME(3) NULL,
  failed_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_wallet_recharges_recharge_no (recharge_no),
  UNIQUE KEY uk_wallet_recharges_idempotency (wallet_id, provider, method, client_token),
  UNIQUE KEY uk_wallet_recharges_upstream_trade (provider, upstream_trade_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
