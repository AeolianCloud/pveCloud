package payment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	domainorder "github.com/AeolianCloud/pveCloud/server/internal/domain/order"
	domainpayment "github.com/AeolianCloud/pveCloud/server/internal/domain/payment"
	integrationpayment "github.com/AeolianCloud/pveCloud/server/internal/integration/payment"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
	paymentusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/payment"
)

func TestShowPaymentRejectsCrossUserAccess(t *testing.T) {
	db := openPaymentHandlerDB(t)
	seedPaymentHandlerOrder(t, db, 11, "ORD-PAY-HANDLER-1")
	seedPaymentHandlerPayment(t, db, 11, "PAY-HANDLER-1", "ORD-PAY-HANDLER-1")
	router := newPaymentHandlerRouter(db, 12, fakePaymentRegistry())

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/payments/PAY-HANDLER-1", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNotFound, recorder.Code)
	requirePaymentEnvelopeCode(t, recorder, 40401)
}

func TestCallbackAllowsNoBearerButRejectsInvalidSignature(t *testing.T) {
	db := openPaymentHandlerDB(t)
	router := newPaymentHandlerRouter(db, 0, integrationpayment.StaticRegistry{
		domainpayment.ProviderAlipay: integrationpayment.FakeAdapter{ParseNotificationFunc: func(ctx context.Context, cfg integrationpayment.Config, req *http.Request) (integrationpayment.NotificationResult, error) {
			return integrationpayment.NotificationResult{}, integrationpayment.ErrInvalidSignature
		}},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/payment-callbacks/alipay", strings.NewReader("out_trade_no=PAY-1"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusBadGateway, recorder.Code)
	requirePaymentEnvelopeCode(t, recorder, 70002)
}

func newPaymentHandlerRouter(db *gorm.DB, userID uint64, registry integrationpayment.Registry) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	if userID > 0 {
		router.Use(func(c *gin.Context) {
			c.Set("web_user_id", userID)
		})
	}
	handler := NewHandler(paymentusecase.NewService(db, config.InstanceLifecycleConfig{}, registry))
	router.GET("/payments/:payment_no", handler.Show)
	router.POST("/payment-callbacks/:provider", handler.Callback)
	return router
}

func openPaymentHandlerDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db, paymentHandlerSystemConfigsSchema, paymentHandlerOrdersSchema, paymentHandlerTransactionsSchema)
	seedPaymentHandlerConfigs(t, db)
	return db
}

func seedPaymentHandlerConfigs(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`INSERT INTO system_configs (config_key, config_value, value_type, group_name, is_secret) VALUES
('payment.enabled', 'true', 'bool', '支付设置', 0),
('payment.default_expire_minutes', '30', 'int', '支付设置', 0),
('payment.alipay.enabled', 'true', 'bool', '支付设置', 0),
('payment.alipay.app_id', 'app-test', 'string', '支付设置', 0),
('payment.alipay.gateway_url', 'https://openapi.alipay.com/gateway.do', 'string', '支付设置', 0),
('payment.alipay.app_private_key', 'private-key', 'string', '支付设置', 1),
('payment.alipay.alipay_public_key', 'public-key', 'string', '支付设置', 1),
('payment.alipay.notify_url', 'https://example.com/api/payment-callbacks/alipay', 'string', '支付设置', 0),
('payment.alipay.return_url', 'https://example.com/payments/return', 'string', '支付设置', 0)`).Error)
}

func seedPaymentHandlerOrder(t *testing.T, db *gorm.DB, userID uint64, orderNo string) {
	t.Helper()
	require.NoError(t, db.Exec(`INSERT INTO orders (
id, order_no, user_id, client_token, status, order_type,
product_no, product_type, product_name, plan_no, plan_code, plan_name,
cpu_cores, memory_mb, system_disk_gb, data_disk_gb, bandwidth_mbps,
public_ip_count, virtualization, architecture, billing_cycle, price_cents,
currency, quantity, total_amount_cents, payment_status, region_no,
region_code, region_name, network_type_no, network_type_code, network_type_name,
template_no, template_code, template_name, os_family, os_distribution,
os_version, os_architecture
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID+1000, orderNo, userID, "client-"+orderNo, domainorder.StatusPending, domainorder.TypePurchase,
		"PROD-1", "server", "Server", "PLAN-1", "basic", "Basic",
		2, 4096, 40, 0, 100, 1, "kvm", "x86_64", "monthly", 3000,
		"CNY", 1, 3000, domainorder.PaymentStatusUnpaid, "REG-1",
		"cn", "China", "NET-1", "classic", "Classic",
		"TPL-1", "ubuntu", "Ubuntu", "linux", "ubuntu", "22.04", "x86_64",
	).Error)
}

func seedPaymentHandlerPayment(t *testing.T, db *gorm.DB, userID uint64, paymentNo, orderNo string) {
	t.Helper()
	require.NoError(t, db.Exec(`INSERT INTO payment_transactions (
payment_no, order_id, order_no, user_id, provider, method, status, client_token,
amount_cents, currency, expires_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		paymentNo, userID+1000, orderNo, userID, domainpayment.ProviderAlipay, domainpayment.MethodAlipayPage, domainpayment.StatusPending, "pay-"+paymentNo, 3000, "CNY", time.Now().Add(30*time.Minute),
	).Error)
}

func fakePaymentRegistry() integrationpayment.Registry {
	return integrationpayment.StaticRegistry{
		domainpayment.ProviderAlipay: integrationpayment.FakeAdapter{},
	}
}

func requirePaymentEnvelopeCode(t *testing.T, recorder *httptest.ResponseRecorder, code int) {
	t.Helper()
	var envelope response.Envelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
	require.Equal(t, code, envelope.Code)
}

const paymentHandlerSystemConfigsSchema = `
CREATE TABLE system_configs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  config_key VARCHAR(128) NOT NULL,
  config_value TEXT NULL,
  value_type VARCHAR(32) NOT NULL DEFAULT 'string',
  group_name VARCHAR(64) NOT NULL,
  is_secret TINYINT(1) NOT NULL DEFAULT 0,
  UNIQUE KEY uk_system_configs_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const paymentHandlerOrdersSchema = `
CREATE TABLE orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  order_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  client_token VARCHAR(128) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  order_type VARCHAR(32) NOT NULL DEFAULT 'purchase',
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
  UNIQUE KEY uk_orders_order_no (order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const paymentHandlerTransactionsSchema = `
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
  UNIQUE KEY uk_payment_transactions_payment_no (payment_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
