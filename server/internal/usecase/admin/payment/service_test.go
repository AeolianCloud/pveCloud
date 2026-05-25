package payment

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	domaininvoice "github.com/AeolianCloud/pveCloud/server/internal/domain/invoice"
	domainorder "github.com/AeolianCloud/pveCloud/server/internal/domain/order"
	domainpayment "github.com/AeolianCloud/pveCloud/server/internal/domain/payment"
	integrationpayment "github.com/AeolianCloud/pveCloud/server/internal/integration/payment"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/usecase/paymentalert"
)

func TestCreateRefundForRenewalRollsBackEffectAndOrder(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, adminPaymentSystemConfigsSchema, adminPaymentOrdersSchema, adminPaymentTransactionsSchema, adminRefundTransactionsSchema, adminPaymentInvoiceOrdersSchema, adminPaymentEffectsSchema, adminPaymentInstancesSchema, adminPaymentAuditLogsSchema)
	seedAdminPaymentConfigs(t, db)

	instanceNo := "INS-refund-renew-1"
	before := time.Now().AddDate(0, 1, 0).Truncate(time.Millisecond)
	after := before.AddDate(0, 3, 0).Truncate(time.Millisecond)
	seedAdminPaymentOrder(t, db, 33, "ORD-refund-renew-1", domainorder.TypeRenewal, &instanceNo, domainorder.StatusFulfilled, domainorder.PaymentStatusPaid)
	seedAdminPaymentInstance(t, db, 33, instanceNo, after)
	seedAdminPayment(t, db, 33, "PAY-refund-renew-1", "ORD-refund-renew-1", domainpayment.StatusPaid)
	seedAdminPaymentEffect(t, db, "EFF-refund-renew-1", "PAY-refund-renew-1", "ORD-refund-renew-1", instanceNo, before, after)

	service := NewService(db, nil, nil, integrationpayment.StaticRegistry{
		domainpayment.ProviderAlipay: integrationpayment.FakeAdapter{CreateRefundFunc: func(ctx context.Context, cfg integrationpayment.Config, req integrationpayment.CreateRefundRequest) (integrationpayment.RefundResult, error) {
			return integrationpayment.RefundResult{RefundNo: req.RefundNo, UpstreamTradeNo: req.UpstreamTradeNo, AmountCents: req.AmountCents, Currency: req.Currency, Status: domainpayment.RefundStatusSucceeded, Summary: `{"fake":"refund"}`}, nil
		}},
	})
	refund, err := service.CreateRefund(context.Background(), 99, "PAY-refund-renew-1", admindto.RefundCreateRequest{Reason: "用户申请退款"})
	if err != nil {
		t.Fatalf("create refund: %v", err)
	}
	if refund.Status != domainpayment.RefundStatusSucceeded {
		t.Fatalf("refund should complete in local mock flow, got %#v", refund)
	}

	var instance struct {
		ExpiresAt time.Time `gorm:"column:expires_at"`
	}
	if err := db.Table("instances").Select("expires_at").Where("instance_no = ?", instanceNo).Take(&instance).Error; err != nil {
		t.Fatalf("load instance: %v", err)
	}
	if !instance.ExpiresAt.Equal(before) {
		t.Fatalf("refund should roll expiry back to before value, got %s want %s", instance.ExpiresAt, before)
	}

	var order struct {
		Status        string
		PaymentStatus string `gorm:"column:payment_status"`
	}
	if err := db.Table("orders").Select("status, payment_status").Where("order_no = ?", "ORD-refund-renew-1").Take(&order).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if order.Status != domainorder.StatusClosed || order.PaymentStatus != domainorder.PaymentStatusRefunded {
		t.Fatalf("refund should close order and mark refunded, got %#v", order)
	}

	var effect struct {
		Status   string
		RefundNo *string `gorm:"column:refund_no"`
	}
	if err := db.Table("payment_effects").Select("status, refund_no").Where("payment_no = ?", "PAY-refund-renew-1").Take(&effect).Error; err != nil {
		t.Fatalf("load effect: %v", err)
	}
	if effect.Status != domainpayment.EffectStatusReverted || effect.RefundNo == nil || *effect.RefundNo != refund.RefundNo {
		t.Fatalf("refund should revert effect with refund number, got %#v", effect)
	}

	var auditCount int64
	if err := db.Table("admin_audit_logs").Where("admin_id = ? AND action = ? AND object_id = ?", 99, "payment.refund.create", refund.RefundNo).Count(&auditCount).Error; err != nil {
		t.Fatalf("count audit: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("refund should write audit log, got %d", auditCount)
	}
}

func TestCreateRefundPendingWritesAlertEvent(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, adminPaymentSystemConfigsSchema, adminPaymentOrdersSchema, adminPaymentTransactionsSchema, adminRefundTransactionsSchema, adminPaymentInvoiceOrdersSchema, adminPaymentEffectsSchema, adminPaymentInstancesSchema, adminPaymentAuditLogsSchema, adminPaymentBackendRuntimeLogsSchema)
	seedAdminPaymentConfigs(t, db)
	seedAdminPaymentOrder(t, db, 44, "ORD-refund-pending-1", domainorder.TypePurchase, nil, domainorder.StatusFulfilled, domainorder.PaymentStatusPaid)
	seedAdminPayment(t, db, 44, "PAY-refund-pending-1", "ORD-refund-pending-1", domainpayment.StatusPaid)

	service := NewService(db, nil, nil, integrationpayment.StaticRegistry{
		domainpayment.ProviderAlipay: integrationpayment.FakeAdapter{CreateRefundFunc: func(ctx context.Context, cfg integrationpayment.Config, req integrationpayment.CreateRefundRequest) (integrationpayment.RefundResult, error) {
			return integrationpayment.RefundResult{RefundNo: req.RefundNo, UpstreamTradeNo: req.UpstreamTradeNo, AmountCents: req.AmountCents, Currency: req.Currency, Status: domainpayment.RefundStatusPending, Summary: `{"fake":"pending"}`}, nil
		}},
	}).SetAlertRecorder(testAdminPaymentAlertRecorder(db))
	refund, err := service.CreateRefund(context.Background(), 99, "PAY-refund-pending-1", admindto.RefundCreateRequest{Reason: "渠道异步退款"})
	if err != nil {
		t.Fatalf("create refund pending: %v", err)
	}
	if refund.Status != domainpayment.RefundStatusPending {
		t.Fatalf("refund should remain pending, got %#v", refund)
	}
	detail := requireAdminPaymentAlertDetail(t, db, paymentalert.EventRefundPending)
	if !strings.Contains(detail, `"payment_no":"PAY-refund-pending-1"`) || !strings.Contains(detail, `"status":"pending"`) {
		t.Fatalf("pending alert should include anchors and status, got %s", detail)
	}
}

func TestCreateRefundFailureWritesAlertEvent(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, adminPaymentSystemConfigsSchema, adminPaymentOrdersSchema, adminPaymentTransactionsSchema, adminRefundTransactionsSchema, adminPaymentInvoiceOrdersSchema, adminPaymentEffectsSchema, adminPaymentInstancesSchema, adminPaymentAuditLogsSchema, adminPaymentBackendRuntimeLogsSchema)
	seedAdminPaymentConfigs(t, db)
	seedAdminPaymentOrder(t, db, 55, "ORD-refund-failed-1", domainorder.TypePurchase, nil, domainorder.StatusFulfilled, domainorder.PaymentStatusPaid)
	seedAdminPayment(t, db, 55, "PAY-refund-failed-1", "ORD-refund-failed-1", domainpayment.StatusPaid)

	service := NewService(db, nil, nil, integrationpayment.StaticRegistry{
		domainpayment.ProviderAlipay: integrationpayment.FakeAdapter{CreateRefundFunc: func(ctx context.Context, cfg integrationpayment.Config, req integrationpayment.CreateRefundRequest) (integrationpayment.RefundResult, error) {
			return integrationpayment.RefundResult{}, errors.New("refund rejected app_private_key=secret-value")
		}},
	}).SetAlertRecorder(testAdminPaymentAlertRecorder(db))
	_, err := service.CreateRefund(context.Background(), 99, "PAY-refund-failed-1", admindto.RefundCreateRequest{Reason: "渠道失败"})
	if err == nil {
		t.Fatalf("create refund should fail")
	}
	detail := requireAdminPaymentAlertDetail(t, db, paymentalert.EventRefundFailed)
	if !strings.Contains(detail, `"payment_no":"PAY-refund-failed-1"`) || !strings.Contains(detail, `"error_code":"CHANNEL_REFUND_FAILED"`) {
		t.Fatalf("failed alert should include anchors and error code, got %s", detail)
	}
	if strings.Contains(detail, "secret-value") {
		t.Fatalf("alert detail should redact sensitive values, got %s", detail)
	}
}

func TestCreateRefundForWalletPaymentCreditsWallet(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, adminPaymentSystemConfigsSchema, adminPaymentOrdersSchema, adminPaymentTransactionsSchema, adminRefundTransactionsSchema, adminPaymentInvoiceOrdersSchema, adminPaymentEffectsSchema, adminPaymentInstancesSchema, adminPaymentAuditLogsSchema, adminPaymentWalletAccountsSchema, adminPaymentWalletLedgerSchema)
	seedAdminPaymentOrder(t, db, 66, "ORD-wallet-refund-1", domainorder.TypePurchase, nil, domainorder.StatusFulfilled, domainorder.PaymentStatusPaid)
	seedAdminPaymentWithChannel(t, db, 66, "PAY-wallet-refund-1", "ORD-wallet-refund-1", domainpayment.ProviderWallet, domainpayment.MethodWalletBalance, domainpayment.StatusPaid)
	seedAdminWalletAccount(t, db, 6601, "WAL-refund-1", 66, 1200)

	service := NewService(db, nil, nil)
	refund, err := service.CreateRefund(context.Background(), 99, "PAY-wallet-refund-1", admindto.RefundCreateRequest{Reason: "余额支付退款"})
	if err != nil {
		t.Fatalf("create wallet refund: %v", err)
	}
	if refund.Status != domainpayment.RefundStatusSucceeded {
		t.Fatalf("wallet refund should complete locally, got %#v", refund)
	}

	var account struct {
		Balance  uint64 `gorm:"column:available_balance_cents"`
		Refunded uint64 `gorm:"column:total_refunded_cents"`
	}
	if err := db.Table("wallet_accounts").Select("available_balance_cents, total_refunded_cents").Where("wallet_no = ?", "WAL-refund-1").Take(&account).Error; err != nil {
		t.Fatalf("load wallet: %v", err)
	}
	if account.Balance != 4200 || account.Refunded != 3000 {
		t.Fatalf("wallet refund should credit amount once, got balance=%d refunded=%d", account.Balance, account.Refunded)
	}

	var ledgerCount int64
	if err := db.Table("wallet_ledger_entries").Where("wallet_no = ? AND entry_type = ? AND related_no = ?", "WAL-refund-1", "refund", refund.RefundNo).Count(&ledgerCount).Error; err != nil {
		t.Fatalf("count refund ledger: %v", err)
	}
	if ledgerCount != 1 {
		t.Fatalf("wallet refund should write one credit ledger, got %d", ledgerCount)
	}

	var order struct {
		Status        string
		PaymentStatus string `gorm:"column:payment_status"`
	}
	if err := db.Table("orders").Select("status, payment_status").Where("order_no = ?", "ORD-wallet-refund-1").Take(&order).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if order.Status != domainorder.StatusClosed || order.PaymentStatus != domainorder.PaymentStatusRefunded {
		t.Fatalf("wallet refund should close and refund order, got %#v", order)
	}
}

func TestCreateRefundBlocksActiveInvoiceApplication(t *testing.T) {
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, adminPaymentSystemConfigsSchema, adminPaymentOrdersSchema, adminPaymentTransactionsSchema, adminRefundTransactionsSchema, adminPaymentInvoiceOrdersSchema, adminPaymentEffectsSchema, adminPaymentInstancesSchema, adminPaymentAuditLogsSchema)
	seedAdminPaymentConfigs(t, db)
	seedAdminPaymentOrder(t, db, 77, "ORD-invoice-refund-block", domainorder.TypePurchase, nil, domainorder.StatusFulfilled, domainorder.PaymentStatusPaid)
	seedAdminPayment(t, db, 77, "PAY-invoice-refund-block", "ORD-invoice-refund-block", domainpayment.StatusPaid)
	seedAdminInvoiceOrder(t, db, 77, "INV-refund-block", "ORD-invoice-refund-block", domaininvoice.StatusIssued)

	service := NewService(db, nil, nil, integrationpayment.StaticRegistry{
		domainpayment.ProviderAlipay: integrationpayment.FakeAdapter{CreateRefundFunc: func(ctx context.Context, cfg integrationpayment.Config, req integrationpayment.CreateRefundRequest) (integrationpayment.RefundResult, error) {
			t.Fatalf("channel refund must not be called when order has active invoice")
			return integrationpayment.RefundResult{}, nil
		}},
	})
	_, err := service.CreateRefund(context.Background(), 99, "PAY-invoice-refund-block", admindto.RefundCreateRequest{Reason: "申请退款"})
	if err == nil || !strings.Contains(err.Error(), "发票") {
		t.Fatalf("active invoice should block refund, got %v", err)
	}
	var refundCount int64
	if err := db.Table("refund_transactions").Where("payment_no = ?", "PAY-invoice-refund-block").Count(&refundCount).Error; err != nil {
		t.Fatalf("count refunds: %v", err)
	}
	if refundCount != 0 {
		t.Fatalf("refund row must not be created when invoice blocks refund, got %d", refundCount)
	}
}

func seedAdminPaymentConfigs(t *testing.T, db *gorm.DB) {
	if err := db.Exec(`INSERT INTO system_configs (config_key, config_value, value_type, group_name, is_secret) VALUES
('payment.alipay.app_id', 'app-test', 'string', '支付设置', 0),
('payment.alipay.gateway_url', 'https://openapi.alipay.com/gateway.do', 'string', '支付设置', 0),
('payment.alipay.app_private_key', 'private-key', 'string', '支付设置', 1),
('payment.alipay.alipay_public_key', 'public-key', 'string', '支付设置', 1),
('payment.alipay.notify_url', 'https://example.com/api/payment-callbacks/alipay', 'string', '支付设置', 0),
('payment.alipay.return_url', 'https://example.com/payments/return', 'string', '支付设置', 0)`).Error; err != nil {
		t.Fatalf("seed payment configs: %v", err)
	}
}

const adminPaymentSystemConfigsSchema = `
CREATE TABLE system_configs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  config_key VARCHAR(128) NOT NULL,
  config_value TEXT NULL,
  value_type VARCHAR(32) NOT NULL DEFAULT 'string',
  group_name VARCHAR(64) NOT NULL,
  is_secret TINYINT(1) NOT NULL DEFAULT 0,
  UNIQUE KEY uk_system_configs_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

func seedAdminPaymentOrder(t *testing.T, db *gorm.DB, userID uint64, orderNo, orderType string, relatedInstanceNo *string, status string, paymentStatus string) {
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

func seedAdminPaymentInstance(t *testing.T, db *gorm.DB, userID uint64, instanceNo string, expiresAt time.Time) {
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

func seedAdminPayment(t *testing.T, db *gorm.DB, userID uint64, paymentNo, orderNo, status string) {
	seedAdminPaymentWithChannel(t, db, userID, paymentNo, orderNo, domainpayment.ProviderAlipay, domainpayment.MethodAlipayPage, status)
}

func seedAdminPaymentWithChannel(t *testing.T, db *gorm.DB, userID uint64, paymentNo, orderNo string, provider string, method string, status string) {
	if err := db.Exec(`INSERT INTO payment_transactions (
id, payment_no, order_id, order_no, user_id, provider, method, status, client_token,
amount_cents, currency, upstream_trade_no, expires_at, paid_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID+3000, paymentNo, userID+1000, orderNo, userID, provider, method, status, "pay-"+paymentNo, 3000, "CNY", "UP-"+paymentNo, time.Now().Add(30*time.Minute), time.Now(),
	).Error; err != nil {
		t.Fatalf("seed payment: %v", err)
	}
}

func seedAdminWalletAccount(t *testing.T, db *gorm.DB, id uint64, walletNo string, userID uint64, balance uint64) {
	t.Helper()
	if err := db.Exec(`INSERT INTO wallet_accounts (id, wallet_no, user_id, currency, status, available_balance_cents) VALUES (?, ?, ?, 'CNY', 'active', ?)`, id, walletNo, userID, balance).Error; err != nil {
		t.Fatalf("seed wallet account: %v", err)
	}
}

func seedAdminPaymentEffect(t *testing.T, db *gorm.DB, effectNo, paymentNo, orderNo, instanceNo string, before time.Time, after time.Time) {
	if err := db.Exec(`INSERT INTO payment_effects (
effect_no, payment_id, payment_no, order_id, order_no, order_type, effect_type, status,
instance_id, instance_no, before_expires_at, after_expires_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		effectNo, 3033, paymentNo, 1033, orderNo, domainorder.TypeRenewal, domainpayment.EffectTypeRenewalExtension, domainpayment.EffectStatusActive, 2033, instanceNo, before, after,
	).Error; err != nil {
		t.Fatalf("seed effect: %v", err)
	}
}

func seedAdminInvoiceOrder(t *testing.T, db *gorm.DB, userID uint64, invoiceNo string, orderNo string, status string) {
	t.Helper()
	if err := db.Exec(`INSERT INTO invoice_application_orders (
invoice_id, invoice_no, user_id, order_id, order_no, order_type, order_amount_cents,
currency, payment_status, status_snapshot
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID+4000, invoiceNo, userID, userID+1000, orderNo, domainorder.TypePurchase, 3000,
		"CNY", domainorder.PaymentStatusPaid, status,
	).Error; err != nil {
		t.Fatalf("seed invoice order: %v", err)
	}
}

const adminPaymentOrdersSchema = `
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
  closed_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_orders_order_no (order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const adminPaymentTransactionsSchema = `
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
  expires_at DATETIME(3) NOT NULL,
  paid_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_payment_transactions_payment_no (payment_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const adminRefundTransactionsSchema = `
CREATE TABLE refund_transactions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  refund_no VARCHAR(64) NOT NULL,
  payment_id BIGINT UNSIGNED NOT NULL,
  payment_no VARCHAR(64) NOT NULL,
  order_id BIGINT UNSIGNED NOT NULL,
  order_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  provider VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  amount_cents BIGINT UNSIGNED NOT NULL,
  currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
  reason VARCHAR(500) NOT NULL,
  requested_by_admin_id BIGINT UNSIGNED NOT NULL,
  upstream_refund_no VARCHAR(128) NULL,
  upstream_trade_no VARCHAR(128) NULL,
  callback_summary JSON NULL,
  query_summary JSON NULL,
  last_error_code VARCHAR(64) NULL,
  last_error_message VARCHAR(500) NULL,
  channel_confirmed_at DATETIME(3) NULL,
  completed_at DATETIME(3) NULL,
  failed_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_refund_transactions_refund_no (refund_no),
  UNIQUE KEY uk_refund_transactions_payment (payment_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const adminPaymentInvoiceOrdersSchema = `
CREATE TABLE invoice_application_orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  invoice_id BIGINT UNSIGNED NOT NULL,
  invoice_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  order_id BIGINT UNSIGNED NOT NULL,
  order_no VARCHAR(64) NOT NULL,
  order_type VARCHAR(32) NOT NULL,
  order_amount_cents BIGINT UNSIGNED NOT NULL,
  currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
  payment_status VARCHAR(32) NOT NULL,
  paid_at DATETIME(3) NULL,
  product_name VARCHAR(128) NULL,
  plan_name VARCHAR(128) NULL,
  status_snapshot VARCHAR(32) NOT NULL DEFAULT 'pending',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  KEY idx_invoice_application_orders_order (order_id, status_snapshot)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const adminPaymentEffectsSchema = `
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

const adminPaymentInstancesSchema = `
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
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  released_at DATETIME(3) NULL,
  UNIQUE KEY uk_instances_instance_no (instance_no),
  UNIQUE KEY uk_instances_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const adminPaymentAuditLogsSchema = `
CREATE TABLE admin_audit_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  admin_id BIGINT UNSIGNED NULL,
  admin_username VARCHAR(64) NULL,
  admin_display_name VARCHAR(64) NULL,
  session_id VARCHAR(128) NULL,
  action VARCHAR(128) NOT NULL,
  object_type VARCHAR(64) NOT NULL,
  object_id VARCHAR(128) NULL,
  before_data JSON NULL,
  after_data JSON NULL,
  remark VARCHAR(500) NULL,
  request_id VARCHAR(128) NULL,
  request_method VARCHAR(16) NULL,
  request_path VARCHAR(255) NULL,
  ip VARCHAR(64) NULL,
  user_agent VARCHAR(500) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const adminPaymentBackendRuntimeLogsSchema = `
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

const adminPaymentWalletAccountsSchema = `
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

const adminPaymentWalletLedgerSchema = `
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

func testAdminPaymentAlertRecorder(db *gorm.DB) *paymentalert.Recorder {
	return paymentalert.New(db, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func requireAdminPaymentAlertDetail(t *testing.T, db *gorm.DB, event string) string {
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
