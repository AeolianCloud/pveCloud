package payment

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	domainorder "github.com/AeolianCloud/pveCloud/server/internal/domain/order"
	domainpayment "github.com/AeolianCloud/pveCloud/server/internal/domain/payment"
	integrationpayment "github.com/AeolianCloud/pveCloud/server/internal/integration/payment"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
	"github.com/AeolianCloud/pveCloud/server/internal/usecase/paymentalert"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
)

func TestCreatePaymentIsIdempotent(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, paymentSystemConfigsSchema, paymentOrdersSchema, paymentTransactionsSchema, paymentInstancesSchema, paymentAsyncTasksSchema, paymentEffectsSchema)
	seedPaymentConfigs(t, db)
	seedOrder(t, db, 11, "ORD-pay-1", domainorder.TypePurchase, nil, domainorder.StatusPending, domainorder.PaymentStatusUnpaid)

	service := NewService(db, config.InstanceLifecycleConfig{}, fakePaymentRegistry())
	req := webdto.PaymentCreateRequest{Provider: domainpayment.ProviderWechat, Method: domainpayment.MethodWechatNative, ClientToken: "pay-token-1"}
	first, err := service.Create(context.Background(), 11, "ORD-pay-1", req)
	if err != nil {
		t.Fatalf("create payment first time: %v", err)
	}
	second, err := service.Create(context.Background(), 11, "ORD-pay-1", req)
	if err != nil {
		t.Fatalf("create payment second time: %v", err)
	}
	if first.PaymentNo != second.PaymentNo {
		t.Fatalf("same client token should return existing payment, got %s and %s", first.PaymentNo, second.PaymentNo)
	}
	if first.QRCodeURL == nil || *first.QRCodeURL == "" {
		t.Fatalf("wechat native payment should expose qr code url: %#v", first)
	}
	var count int64
	if err := db.Table("payment_transactions").Count(&count).Error; err != nil {
		t.Fatalf("count payments: %v", err)
	}
	if count != 1 {
		t.Fatalf("idempotent create should persist one payment, got %d", count)
	}
}

func TestCallbackPaidRenewalExtendsInstanceAndCreatesEffect(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, paymentSystemConfigsSchema, paymentOrdersSchema, paymentTransactionsSchema, paymentInstancesSchema, paymentAsyncTasksSchema, paymentEffectsSchema)
	seedPaymentConfigs(t, db)
	instanceNo := "INS-renew-pay-1"
	before := time.Now().AddDate(0, 1, 0).Truncate(time.Millisecond)
	seedOrder(t, db, 22, "ORD-renew-pay-1", domainorder.TypeRenewal, &instanceNo, domainorder.StatusPending, domainorder.PaymentStatusUnpaid)
	seedInstance(t, db, 22, instanceNo, before)
	seedPayment(t, db, 22, "PAY-renew-1", "ORD-renew-pay-1", domainpayment.ProviderAlipay, domainpayment.MethodAlipayPage, domainpayment.StatusPending)

	service := NewService(db, config.InstanceLifecycleConfig{AutoReleaseEnabled: true, ExpireNoticeBeforeSeconds: 86400, ExpireReleaseAfterSeconds: 3600}, fakePaymentRegistry())
	err := service.HandleCallback(context.Background(), domainpayment.ProviderAlipay, httptest.NewRequest("POST", "/api/payment-callbacks/alipay", nil))
	if err != nil {
		t.Fatalf("handle paid callback: %v", err)
	}

	var order struct {
		Status        string
		PaymentStatus string `gorm:"column:payment_status"`
	}
	if err := db.Table("orders").Select("status, payment_status").Where("order_no = ?", "ORD-renew-pay-1").Take(&order).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if order.Status != domainorder.StatusFulfilled || order.PaymentStatus != domainorder.PaymentStatusPaid {
		t.Fatalf("renewal payment should fulfill order and mark paid, got %#v", order)
	}

	var instance struct {
		ExpiresAt time.Time `gorm:"column:expires_at"`
	}
	if err := db.Table("instances").Select("expires_at").Where("instance_no = ?", instanceNo).Take(&instance).Error; err != nil {
		t.Fatalf("load instance: %v", err)
	}
	if !instance.ExpiresAt.Equal(before.AddDate(0, 3, 0)) {
		t.Fatalf("renewal should extend from existing expiry, got %s want %s", instance.ExpiresAt, before.AddDate(0, 3, 0))
	}

	var effectCount int64
	if err := db.Table("payment_effects").Where("payment_no = ? AND status = ?", "PAY-renew-1", domainpayment.EffectStatusActive).Count(&effectCount).Error; err != nil {
		t.Fatalf("count effects: %v", err)
	}
	if effectCount != 1 {
		t.Fatalf("paid renewal should create one active effect, got %d", effectCount)
	}
}

func TestCreatePaymentFailureWritesAlertEvent(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, paymentSystemConfigsSchema, paymentOrdersSchema, paymentTransactionsSchema, paymentBackendRuntimeLogsSchema)
	seedPaymentConfigs(t, db)
	seedOrder(t, db, 44, "ORD-pay-alert-1", domainorder.TypePurchase, nil, domainorder.StatusPending, domainorder.PaymentStatusUnpaid)

	service := NewService(db, config.InstanceLifecycleConfig{}, integrationpayment.StaticRegistry{
		domainpayment.ProviderWechat: integrationpayment.FakeAdapter{CreatePaymentFunc: func(ctx context.Context, cfg integrationpayment.Config, req integrationpayment.CreatePaymentRequest) (integrationpayment.CreatePaymentResult, error) {
			return integrationpayment.CreatePaymentResult{}, errors.New("gateway rejected api_v3_key=secret-value")
		}},
	}).SetAlertRecorder(testPaymentAlertRecorder(db))
	_, err := service.Create(context.Background(), 44, "ORD-pay-alert-1", webdto.PaymentCreateRequest{Provider: domainpayment.ProviderWechat, Method: domainpayment.MethodWechatNative, ClientToken: "pay-alert-token"})
	if err == nil {
		t.Fatalf("create payment should fail")
	}
	detail := requirePaymentAlertDetail(t, db, paymentalert.EventPaymentCreateFailed)
	if !strings.Contains(detail, `"payment_no":"PAY-`) || !strings.Contains(detail, `"order_no":"ORD-pay-alert-1"`) {
		t.Fatalf("alert should include payment and order anchors, got %s", detail)
	}
	if strings.Contains(detail, "secret-value") {
		t.Fatalf("alert detail should redact sensitive values, got %s", detail)
	}
}

func TestCallbackInvalidSignatureWritesAlertEvent(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, paymentSystemConfigsSchema, paymentBackendRuntimeLogsSchema)
	seedPaymentConfigs(t, db)

	service := NewService(db, config.InstanceLifecycleConfig{}, integrationpayment.StaticRegistry{
		domainpayment.ProviderAlipay: integrationpayment.FakeAdapter{},
	}).SetAlertRecorder(testPaymentAlertRecorder(db))
	err := service.HandleCallback(context.Background(), domainpayment.ProviderAlipay, httptest.NewRequest("POST", "/api/payment-callbacks/alipay", nil))
	if err == nil {
		t.Fatalf("invalid signature callback should fail")
	}
	detail := requirePaymentAlertDetail(t, db, paymentalert.EventPaymentCallbackSignatureFailed)
	if !strings.Contains(detail, `"provider":"alipay"`) || !strings.Contains(detail, `"error_code":"CALLBACK_SIGNATURE_FAILED"`) {
		t.Fatalf("alert should include provider and local error code, got %s", detail)
	}
}

func TestWalletBalancePaymentDebitsWalletAndMarksOrderPaid(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, paymentSystemConfigsSchema, paymentOrdersSchema, paymentTransactionsSchema, paymentInstancesSchema, paymentAsyncTasksSchema, paymentEffectsSchema, paymentWalletAccountsSchema, paymentWalletLedgerSchema)
	seedPaymentConfigs(t, db)
	seedWalletPaymentConfig(t, db, true)
	seedOrder(t, db, 66, "ORD-wallet-pay-1", domainorder.TypePurchase, nil, domainorder.StatusPending, domainorder.PaymentStatusUnpaid)
	seedWalletPaymentAccount(t, db, 6601, "WAL-pay-1", 66, 5000)

	service := NewService(db, config.InstanceLifecycleConfig{}, fakePaymentRegistry())
	result, err := service.Create(context.Background(), 66, "ORD-wallet-pay-1", webdto.PaymentCreateRequest{Provider: domainpayment.ProviderWallet, Method: domainpayment.MethodWalletBalance, ClientToken: "wallet-pay-token"})
	if err != nil {
		t.Fatalf("create wallet payment: %v", err)
	}
	if result.Provider != domainpayment.ProviderWallet || result.Method != domainpayment.MethodWalletBalance || result.Status != domainpayment.StatusPaid {
		t.Fatalf("wallet payment should return paid wallet transaction, got %#v", result)
	}

	var account struct {
		Balance uint64 `gorm:"column:available_balance_cents"`
		Spent   uint64 `gorm:"column:total_spent_cents"`
	}
	if err := db.Table("wallet_accounts").Select("available_balance_cents, total_spent_cents").Where("wallet_no = ?", "WAL-pay-1").Take(&account).Error; err != nil {
		t.Fatalf("load wallet: %v", err)
	}
	if account.Balance != 2000 || account.Spent != 3000 {
		t.Fatalf("wallet payment should debit once, got balance=%d spent=%d", account.Balance, account.Spent)
	}

	var order struct {
		PaymentStatus string  `gorm:"column:payment_status"`
		Provider      *string `gorm:"column:payment_provider"`
	}
	if err := db.Table("orders").Select("payment_status, payment_provider").Where("order_no = ?", "ORD-wallet-pay-1").Take(&order).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if order.PaymentStatus != domainorder.PaymentStatusPaid || order.Provider == nil || *order.Provider != domainpayment.ProviderWallet {
		t.Fatalf("wallet payment should mark order paid by wallet, got %#v", order)
	}

	var ledgerCount int64
	if err := db.Table("wallet_ledger_entries").Where("wallet_no = ? AND entry_type = ?", "WAL-pay-1", "payment").Count(&ledgerCount).Error; err != nil {
		t.Fatalf("count wallet ledger: %v", err)
	}
	if ledgerCount != 1 {
		t.Fatalf("wallet payment should write one debit ledger, got %d", ledgerCount)
	}
}

func TestWalletBalancePaymentRejectsInsufficientBalance(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, paymentSystemConfigsSchema, paymentOrdersSchema, paymentTransactionsSchema, paymentInstancesSchema, paymentAsyncTasksSchema, paymentEffectsSchema, paymentWalletAccountsSchema, paymentWalletLedgerSchema)
	seedPaymentConfigs(t, db)
	seedWalletPaymentConfig(t, db, true)
	seedOrder(t, db, 77, "ORD-wallet-insufficient-1", domainorder.TypePurchase, nil, domainorder.StatusPending, domainorder.PaymentStatusUnpaid)
	seedWalletPaymentAccount(t, db, 7701, "WAL-insufficient-1", 77, 2999)

	service := NewService(db, config.InstanceLifecycleConfig{}, fakePaymentRegistry())
	_, err := service.Create(context.Background(), 77, "ORD-wallet-insufficient-1", webdto.PaymentCreateRequest{Provider: domainpayment.ProviderWallet, Method: domainpayment.MethodWalletBalance, ClientToken: "wallet-insufficient-token"})
	if err == nil || !strings.Contains(err.Error(), "钱包余额不足") {
		t.Fatalf("insufficient wallet balance should fail with conflict message, got %v", err)
	}
	var paymentCount int64
	if err := db.Table("payment_transactions").Count(&paymentCount).Error; err != nil {
		t.Fatalf("count payments: %v", err)
	}
	if paymentCount != 0 {
		t.Fatalf("failed wallet payment should not create payment, got %d", paymentCount)
	}
	var account struct {
		Balance uint64 `gorm:"column:available_balance_cents"`
		Spent   uint64 `gorm:"column:total_spent_cents"`
	}
	if err := db.Table("wallet_accounts").Select("available_balance_cents, total_spent_cents").Where("wallet_no = ?", "WAL-insufficient-1").Take(&account).Error; err != nil {
		t.Fatalf("load wallet: %v", err)
	}
	if account.Balance != 2999 || account.Spent != 0 {
		t.Fatalf("failed wallet payment should not change balance, got balance=%d spent=%d", account.Balance, account.Spent)
	}
}

func seedPaymentConfigs(t *testing.T, db *gorm.DB) {
	if err := db.Exec(`INSERT INTO system_configs (config_key, config_value, value_type, group_name, is_secret) VALUES
('payment.enabled', 'true', 'bool', '支付设置', 0),
('payment.default_expire_minutes', '30', 'int', '支付设置', 0),
('payment.alipay.enabled', 'true', 'bool', '支付设置', 0),
('payment.alipay.app_id', 'app-test', 'string', '支付设置', 0),
('payment.alipay.gateway_url', 'https://openapi.alipay.com/gateway.do', 'string', '支付设置', 0),
('payment.alipay.app_private_key', 'private-key', 'string', '支付设置', 1),
('payment.alipay.alipay_public_key', 'public-key', 'string', '支付设置', 1),
('payment.alipay.notify_url', 'https://example.com/api/payment-callbacks/alipay', 'string', '支付设置', 0),
('payment.alipay.return_url', 'https://example.com/payments/return', 'string', '支付设置', 0),
('payment.wechat.enabled', 'true', 'bool', '支付设置', 0),
('payment.wechat.app_id', 'wx-test', 'string', '支付设置', 0),
('payment.wechat.mch_id', 'mch-test', 'string', '支付设置', 0),
('payment.wechat.api_v3_key', '12345678901234567890123456789012', 'string', '支付设置', 1),
('payment.wechat.mch_private_key', 'private-key', 'string', '支付设置', 1),
('payment.wechat.mch_certificate_serial_no', 'serial-test', 'string', '支付设置', 0),
('payment.wechat.platform_public_key_id', 'PUB_KEY_ID_TEST', 'string', '支付设置', 0),
('payment.wechat.platform_public_key', 'public-key', 'string', '支付设置', 1),
('payment.wechat.notify_url', 'https://example.com/api/payment-callbacks/wechat', 'string', '支付设置', 0),
('payment.wechat.h5_scene_info', '{"type":"Wap","app_name":"pveCloud","app_url":"https://example.com"}', 'string', '支付设置', 0)`).Error; err != nil {
		t.Fatalf("seed payment configs: %v", err)
	}
}

func seedWalletPaymentConfig(t *testing.T, db *gorm.DB, enabled bool) {
	t.Helper()
	value := "false"
	if enabled {
		value = "true"
	}
	if err := db.Exec(`INSERT INTO system_configs (config_key, config_value, value_type, group_name, is_secret) VALUES ('wallet.enabled', ?, 'bool', '钱包设置', 0)`, value).Error; err != nil {
		t.Fatalf("seed wallet config: %v", err)
	}
}

func seedWalletPaymentAccount(t *testing.T, db *gorm.DB, id uint64, walletNo string, userID uint64, balance uint64) {
	t.Helper()
	if err := db.Exec(`INSERT INTO wallet_accounts (id, wallet_no, user_id, currency, status, available_balance_cents) VALUES (?, ?, ?, 'CNY', 'active', ?)`, id, walletNo, userID, balance).Error; err != nil {
		t.Fatalf("seed wallet account: %v", err)
	}
}

func fakePaymentRegistry() integrationpayment.Registry {
	return integrationpayment.StaticRegistry{
		domainpayment.ProviderWechat: integrationpayment.FakeAdapter{CreatePaymentFunc: func(ctx context.Context, cfg integrationpayment.Config, req integrationpayment.CreatePaymentRequest) (integrationpayment.CreatePaymentResult, error) {
			return integrationpayment.CreatePaymentResult{QRCodeURL: "weixin://pay/" + req.PaymentNo, Summary: `{"fake":"wechat"}`}, nil
		}},
		domainpayment.ProviderAlipay: integrationpayment.FakeAdapter{ParseNotificationFunc: func(ctx context.Context, cfg integrationpayment.Config, req *http.Request) (integrationpayment.NotificationResult, error) {
			return integrationpayment.NotificationResult{PaymentNo: "PAY-renew-1", Provider: domainpayment.ProviderAlipay, UpstreamTradeNo: "ALI-TRADE-1", AmountCents: 3000, Currency: "CNY", Status: domainpayment.StatusPaid, Summary: `{"fake":"alipay"}`}, nil
		}},
	}
}

func seedOrder(t *testing.T, db *gorm.DB, userID uint64, orderNo, orderType string, relatedInstanceNo *string, status string, paymentStatus string) {
	if err := db.Exec(`INSERT INTO orders (
id, order_no, user_id, client_token, status, order_type, related_instance_no,
product_no, product_type, product_name, plan_no, plan_code, plan_name,
cpu_cores, memory_mb, system_disk_gb, data_disk_gb, bandwidth_mbps,
public_ip_count, virtualization, architecture, billing_cycle, price_cents,
currency, quantity, total_amount_cents, payment_status, region_no,
region_code, region_name, network_type_no, network_type_code, network_type_name,
template_no, template_code, template_name, os_family, os_distribution,
os_version, os_architecture
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID+1000, orderNo, userID, "client-"+orderNo, status, orderType, relatedInstanceNo,
		"PROD-1", "server", "Server", "PLAN-1", "basic", "Basic",
		2, 4096, 40, 0, 100, 1, "kvm", "x86_64", "quarterly", 3000,
		"CNY", 1, 3000, paymentStatus, "REG-1",
		"cn", "China", "NET-1", "classic", "Classic",
		"TPL-1", "ubuntu", "Ubuntu", "linux", "ubuntu", "22.04", "x86_64",
	).Error; err != nil {
		t.Fatalf("seed order: %v", err)
	}
}

func seedInstance(t *testing.T, db *gorm.DB, userID uint64, instanceNo string, expiresAt time.Time) {
	if err := db.Exec(`INSERT INTO instances (
id, instance_no, user_id, order_id, order_no, status, product_no, product_name,
plan_no, plan_name, cpu_cores, memory_mb, system_disk_gb, data_disk_gb,
bandwidth_mbps, region_no, region_name, network_type_no, network_type_name,
template_no, template_name, os_family, os_distribution, os_version,
external_node, external_vmid, expires_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID+2000, instanceNo, userID, userID+900, "ORD-original", "running", "PROD-1", "Server",
		"PLAN-1", "Basic", 2, 4096, 40, 0, 100, "REG-1", "China", "NET-1", "Classic",
		"TPL-1", "Ubuntu", "linux", "ubuntu", "22.04", "node-a", 1001, expiresAt,
	).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
}

func seedPayment(t *testing.T, db *gorm.DB, userID uint64, paymentNo, orderNo, provider, method, status string) {
	if err := db.Exec(`INSERT INTO payment_transactions (
payment_no, order_id, order_no, user_id, provider, method, status, client_token,
amount_cents, currency, expires_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		paymentNo, userID+1000, orderNo, userID, provider, method, status, "pay-"+paymentNo, 3000, "CNY", time.Now().Add(30*time.Minute),
	).Error; err != nil {
		t.Fatalf("seed payment: %v", err)
	}
}

const paymentSystemConfigsSchema = `
CREATE TABLE system_configs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  config_key VARCHAR(128) NOT NULL,
  config_value TEXT NULL,
  value_type VARCHAR(32) NOT NULL DEFAULT 'string',
  group_name VARCHAR(64) NOT NULL,
  is_secret TINYINT(1) NOT NULL DEFAULT 0,
  UNIQUE KEY uk_system_configs_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const paymentOrdersSchema = `
CREATE TABLE orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  order_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  client_token VARCHAR(128) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  order_type VARCHAR(32) NOT NULL DEFAULT 'purchase',
  related_instance_no VARCHAR(64) NULL,
  product_no VARCHAR(64) NOT NULL,
  product_type VARCHAR(32) NOT NULL,
  product_name VARCHAR(128) NOT NULL,
  plan_no VARCHAR(64) NOT NULL,
  plan_code VARCHAR(96) NOT NULL,
  plan_name VARCHAR(128) NOT NULL,
  cpu_cores INT NOT NULL,
  memory_mb INT NOT NULL,
  system_disk_gb INT NOT NULL,
  data_disk_gb INT NOT NULL DEFAULT 0,
  bandwidth_mbps INT NOT NULL,
  public_ip_count INT NOT NULL DEFAULT 1,
  virtualization VARCHAR(32) NOT NULL DEFAULT 'kvm',
  architecture VARCHAR(32) NOT NULL DEFAULT 'x86_64',
  billing_cycle VARCHAR(32) NOT NULL,
  price_cents BIGINT UNSIGNED NOT NULL,
  currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
  quantity INT NOT NULL DEFAULT 1,
  total_amount_cents BIGINT UNSIGNED NOT NULL,
  payment_status VARCHAR(32) NOT NULL DEFAULT 'unpaid',
  paid_at DATETIME(3) NULL,
  payment_provider VARCHAR(32) NULL,
  payment_trade_no VARCHAR(128) NULL,
  payment_callback_payload TEXT NULL,
  region_no VARCHAR(64) NOT NULL,
  region_code VARCHAR(64) NOT NULL,
  region_name VARCHAR(128) NOT NULL,
  network_type_no VARCHAR(64) NOT NULL,
  network_type_code VARCHAR(64) NOT NULL,
  network_type_name VARCHAR(128) NOT NULL,
  template_no VARCHAR(64) NOT NULL,
  template_code VARCHAR(96) NOT NULL,
  template_name VARCHAR(128) NOT NULL,
  os_family VARCHAR(32) NOT NULL,
  os_distribution VARCHAR(64) NOT NULL,
  os_version VARCHAR(64) NOT NULL,
  os_architecture VARCHAR(32) NOT NULL DEFAULT 'x86_64',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  cancelled_at DATETIME(3) NULL,
  closed_at DATETIME(3) NULL,
  UNIQUE KEY uk_orders_order_no (order_no),
  UNIQUE KEY uk_orders_user_client_token (user_id, client_token)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const paymentTransactionsSchema = `
CREATE TABLE payment_transactions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  payment_no VARCHAR(64) NOT NULL,
  order_id BIGINT UNSIGNED NOT NULL,
  order_no VARCHAR(64) NOT NULL,
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
  UNIQUE KEY uk_payment_transactions_payment_no (payment_no),
  UNIQUE KEY uk_payment_transactions_idempotency (order_id, provider, method, client_token),
  UNIQUE KEY uk_payment_transactions_upstream_trade (provider, upstream_trade_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const paymentInstancesSchema = `
CREATE TABLE instances (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  instance_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  order_id BIGINT UNSIGNED NOT NULL,
  order_no VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'running',
  product_no VARCHAR(64) NOT NULL,
  product_name VARCHAR(128) NOT NULL,
  plan_no VARCHAR(64) NOT NULL,
  plan_name VARCHAR(128) NOT NULL,
  cpu_cores INT NOT NULL,
  memory_mb INT NOT NULL,
  system_disk_gb INT NOT NULL,
  data_disk_gb INT NOT NULL DEFAULT 0,
  bandwidth_mbps INT NOT NULL,
  region_no VARCHAR(64) NOT NULL,
  region_name VARCHAR(128) NOT NULL,
  network_type_no VARCHAR(64) NULL,
  network_type_name VARCHAR(128) NULL,
  template_no VARCHAR(64) NOT NULL,
  template_name VARCHAR(128) NOT NULL,
  os_family VARCHAR(32) NOT NULL,
  os_distribution VARCHAR(64) NOT NULL,
  os_version VARCHAR(64) NOT NULL,
  external_node VARCHAR(128) NOT NULL,
  external_vmid INT UNSIGNED NOT NULL,
  expires_at DATETIME(3) NULL,
  expire_notice_sent_at DATETIME(3) NULL,
  expire_release_scheduled_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  released_at DATETIME(3) NULL,
  UNIQUE KEY uk_instances_instance_no (instance_no),
  UNIQUE KEY uk_instances_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const paymentAsyncTasksSchema = `
CREATE TABLE async_tasks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  task_no VARCHAR(64) NOT NULL,
  task_type VARCHAR(64) NOT NULL,
  idempotency_key VARCHAR(191) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  object_type VARCHAR(64) NULL,
  object_no VARCHAR(64) NULL,
  payload JSON NULL,
  result JSON NULL,
  attempts INT NOT NULL DEFAULT 0,
  max_attempts INT NOT NULL DEFAULT 10,
  scheduled_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  locked_by VARCHAR(128) NULL,
  locked_until DATETIME(3) NULL,
  last_error_code VARCHAR(64) NULL,
  last_error_message VARCHAR(500) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  completed_at DATETIME(3) NULL,
  UNIQUE KEY uk_async_tasks_task_no (task_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const paymentEffectsSchema = `
CREATE TABLE payment_effects (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  effect_no VARCHAR(64) NOT NULL,
  payment_id BIGINT UNSIGNED NOT NULL,
  payment_no VARCHAR(64) NOT NULL,
  order_id BIGINT UNSIGNED NOT NULL,
  order_no VARCHAR(64) NOT NULL,
  order_type VARCHAR(32) NOT NULL,
  effect_type VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  instance_id BIGINT UNSIGNED NULL,
  instance_no VARCHAR(64) NULL,
  before_expires_at DATETIME(3) NULL,
  after_expires_at DATETIME(3) NULL,
  refund_id BIGINT UNSIGNED NULL,
  refund_no VARCHAR(64) NULL,
  applied_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  reverted_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_payment_effects_effect_no (effect_no),
  UNIQUE KEY uk_payment_effects_payment (payment_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const paymentBackendRuntimeLogsSchema = `
CREATE TABLE backend_runtime_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  level VARCHAR(16) NOT NULL,
  category VARCHAR(32) NOT NULL,
  request_id VARCHAR(64) NULL,
  request_method VARCHAR(16) NULL,
  request_path VARCHAR(255) NULL,
  status INT NULL,
  latency_ms BIGINT NULL,
  client_ip VARCHAR(64) NULL,
  message VARCHAR(500) NOT NULL,
  detail TEXT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const paymentWalletAccountsSchema = `
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

const paymentWalletLedgerSchema = `
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

func testPaymentAlertRecorder(db *gorm.DB) *paymentalert.Recorder {
	return paymentalert.New(db, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func requirePaymentAlertDetail(t *testing.T, db *gorm.DB, event string) string {
	t.Helper()
	var row struct {
		Level    string
		Category string
		Message  string
		Detail   *string
	}
	if err := db.Table("backend_runtime_logs").Where("message = ? AND detail LIKE ?", "payment_alert", "%"+event+"%").Take(&row).Error; err != nil {
		t.Fatalf("load payment alert %s: %v", event, err)
	}
	if row.Level != "error" || row.Category != "runtime" || row.Message != "payment_alert" || row.Detail == nil {
		t.Fatalf("unexpected alert row: %#v", row)
	}
	return *row.Detail
}
