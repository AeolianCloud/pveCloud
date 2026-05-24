package order

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	domainorder "github.com/AeolianCloud/pveCloud/server/internal/domain/order"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
	orderusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/order"
)

func TestCreateOrderIgnoresOverpostedBusinessFields(t *testing.T) {
	db := openOrderHandlerDB(t)
	seedOrderHandlerCatalog(t, db)

	router := newCreateOrderRouter(db, 11)
	body := `{
		"plan_no":"PLAN-ORDER-1",
		"billing_cycle":"monthly",
		"region_no":"REG-ORDER-1",
		"template_no":"TPL-ORDER-1",
		"network_type_no":"NET-ORDER-1",
		"quantity":1,
		"client_token":"order-overpost-token",
		"user_id":999,
		"status":"fulfilled",
		"order_type":"renewal",
		"related_instance_no":"INS-ATTACK",
		"payment_status":"manual_confirmed",
		"price_cents":1,
		"total_amount_cents":1,
		"product_name":"Tampered Product",
		"plan_name":"Tampered Plan",
		"region_name":"Tampered Region",
		"network_type_name":"Tampered Network",
		"template_name":"Tampered Template",
		"cpu_cores":128,
		"memory_mb":1048576
	}`

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var envelope response.Envelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
	require.Equal(t, 0, envelope.Code)

	var row struct {
		UserID            uint64  `gorm:"column:user_id"`
		Status            string  `gorm:"column:status"`
		OrderType         string  `gorm:"column:order_type"`
		RelatedInstanceNo *string `gorm:"column:related_instance_no"`
		PaymentStatus     string  `gorm:"column:payment_status"`
		ProductName       string  `gorm:"column:product_name"`
		PlanName          string  `gorm:"column:plan_name"`
		RegionName        string  `gorm:"column:region_name"`
		NetworkTypeName   string  `gorm:"column:network_type_name"`
		TemplateName      string  `gorm:"column:template_name"`
		CPUCores          int     `gorm:"column:cpu_cores"`
		MemoryMB          int     `gorm:"column:memory_mb"`
		PriceCents        uint64  `gorm:"column:price_cents"`
		TotalAmountCents  uint64  `gorm:"column:total_amount_cents"`
	}
	require.NoError(t, db.Table("orders").Where("client_token = ?", "order-overpost-token").Take(&row).Error)
	require.Equal(t, uint64(11), row.UserID)
	require.Equal(t, domainorder.StatusPending, row.Status)
	require.Equal(t, domainorder.TypePurchase, row.OrderType)
	require.Nil(t, row.RelatedInstanceNo)
	require.Equal(t, domainorder.PaymentStatusUnpaid, row.PaymentStatus)
	require.Equal(t, "Server", row.ProductName)
	require.Equal(t, "Basic", row.PlanName)
	require.Equal(t, "China", row.RegionName)
	require.Equal(t, "Classic", row.NetworkTypeName)
	require.Equal(t, "Ubuntu", row.TemplateName)
	require.Equal(t, 2, row.CPUCores)
	require.Equal(t, 4096, row.MemoryMB)
	require.Equal(t, uint64(1200), row.PriceCents)
	require.Equal(t, uint64(1200), row.TotalAmountCents)
}

func TestCreateOrderRejectsMaliciousRequestShape(t *testing.T) {
	db := openOrderHandlerDB(t)
	seedOrderHandlerCatalog(t, db)
	router := newCreateOrderRouter(db, 11)

	tests := []struct {
		name string
		body string
	}{
		{name: "over quantity", body: `{"plan_no":"PLAN-ORDER-1","billing_cycle":"monthly","region_no":"REG-ORDER-1","template_no":"TPL-ORDER-1","network_type_no":"NET-ORDER-1","quantity":2,"client_token":"bad-quantity"}`},
		{name: "sql-like cycle", body: `{"plan_no":"PLAN-ORDER-1","billing_cycle":"monthly' OR '1'='1","region_no":"REG-ORDER-1","template_no":"TPL-ORDER-1","network_type_no":"NET-ORDER-1","quantity":1,"client_token":"bad-cycle"}`},
		{name: "overlong token", body: `{"plan_no":"PLAN-ORDER-1","billing_cycle":"monthly","region_no":"REG-ORDER-1","template_no":"TPL-ORDER-1","network_type_no":"NET-ORDER-1","quantity":1,"client_token":"` + strings.Repeat("a", 129) + `"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(tt.body))
			request.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(recorder, request)

			require.Equal(t, http.StatusBadRequest, recorder.Code)
			var envelope response.Envelope
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
			require.Equal(t, 40001, envelope.Code)
			require.Nil(t, envelope.Data)
		})
	}
}

func TestCancelOrderEnforcesOwnershipStatusAndIgnoresOverpostedFields(t *testing.T) {
	db := openOrderHandlerDB(t)
	insertOrderHandlerOrder(t, db, orderHandlerOrderSeed{OrderNo: "ORD-CANCEL-OWN", UserID: 11, ClientToken: "cancel-own-token", Status: domainorder.StatusPending})
	insertOrderHandlerOrder(t, db, orderHandlerOrderSeed{OrderNo: "ORD-CANCEL-OTHER", UserID: 12, ClientToken: "cancel-other-token", Status: domainorder.StatusPending})
	insertOrderHandlerOrder(t, db, orderHandlerOrderSeed{OrderNo: "ORD-CANCEL-FULFILLED", UserID: 11, ClientToken: "cancel-fulfilled-token", Status: domainorder.StatusFulfilled})

	router := newOrderRouter(db, 11)

	t.Run("cross user order is invisible", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/orders/ORD-CANCEL-OTHER/cancel", strings.NewReader(`{"reason":"try cancel other","status":"cancelled","user_id":11}`))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusNotFound, recorder.Code)
		requireEnvelope(t, recorder, 40401)
		requireOrderStatus(t, db, "ORD-CANCEL-OTHER", domainorder.StatusPending)
	})

	t.Run("fulfilled order cannot be cancelled", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/orders/ORD-CANCEL-FULFILLED/cancel", strings.NewReader(`{"reason":"try cancel fulfilled","status":"cancelled"}`))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusConflict, recorder.Code)
		requireEnvelope(t, recorder, 40901)
		requireOrderStatus(t, db, "ORD-CANCEL-FULFILLED", domainorder.StatusFulfilled)
	})

	t.Run("pending order cancellation ignores forged fields", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/orders/ORD-CANCEL-OWN/cancel", strings.NewReader(`{"reason":"user requested","status":"fulfilled","user_id":999,"payment_status":"manual_confirmed"}`))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusOK, recorder.Code)
		requireEnvelope(t, recorder, 0)

		var row struct {
			UserID        uint64 `gorm:"column:user_id"`
			Status        string `gorm:"column:status"`
			PaymentStatus string `gorm:"column:payment_status"`
			CancelReason  string `gorm:"column:cancel_reason"`
		}
		require.NoError(t, db.Table("orders").Where("order_no = ?", "ORD-CANCEL-OWN").Take(&row).Error)
		require.Equal(t, uint64(11), row.UserID)
		require.Equal(t, domainorder.StatusCancelled, row.Status)
		require.Equal(t, domainorder.PaymentStatusUnpaid, row.PaymentStatus)
		require.Equal(t, "user requested", row.CancelReason)
	})
}

func newCreateOrderRouter(db *gorm.DB, userID uint64) *gin.Engine {
	router := newOrderRouter(db, userID)
	return router
}

func newOrderRouter(db *gorm.DB, userID uint64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("web_user_id", userID)
	})
	handler := NewHandler(orderusecase.NewService(db, nil))
	router.POST("/orders", handler.Create)
	router.POST("/orders/:order_no/cancel", handler.Cancel)
	return router
}

func openOrderHandlerDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db,
		orderHandlerProductsSchema,
		orderHandlerProductPlansSchema,
		orderHandlerPlanPricesSchema,
		orderHandlerSalesRegionsSchema,
		orderHandlerTemplatesSchema,
		orderHandlerPlanRegionsSchema,
		orderHandlerPlanTemplatesSchema,
		orderHandlerNetworkTypesSchema,
		orderHandlerPlanNetworkTypesSchema,
		orderHandlerOrdersSchema,
		orderHandlerUserBusinessLogsSchema,
	)
	return db
}

func seedOrderHandlerCatalog(t *testing.T, db *gorm.DB) {
	t.Helper()
	statements := []string{
		`INSERT INTO products (id, product_no, type, slug, name, summary, status, visible) VALUES (1, 'PROD-ORDER-1', 'server', 'server-order-basic', 'Server', 'Server summary', 'active', 1)`,
		`INSERT INTO product_plans (id, plan_no, product_id, code, name, summary, cpu_cores, memory_mb, system_disk_gb, data_disk_gb, bandwidth_mbps, public_ip_count, virtualization, architecture, status, visible) VALUES (1, 'PLAN-ORDER-1', 1, 'basic', 'Basic', 'Basic summary', 2, 4096, 40, 0, 100, 1, 'kvm', 'x86_64', 'active', 1)`,
		`INSERT INTO plan_prices (plan_id, billing_cycle, price_cents, original_price_cents, currency, status) VALUES (1, 'monthly', 1200, 1500, 'CNY', 'active')`,
		`INSERT INTO sales_regions (id, region_no, code, name, status, visible) VALUES (1, 'REG-ORDER-1', 'cn', 'China', 'active', 1)`,
		`INSERT INTO server_os_templates (id, template_no, code, name, os_family, distribution, version, architecture, status, visible) VALUES (1, 'TPL-ORDER-1', 'ubuntu', 'Ubuntu', 'linux', 'ubuntu', '22.04', 'x86_64', 'active', 1)`,
		`INSERT INTO plan_regions (plan_id, region_id, status) VALUES (1, 1, 'active')`,
		`INSERT INTO plan_os_templates (plan_id, template_id, status) VALUES (1, 1, 'active')`,
		`INSERT INTO network_types (id, network_type_no, code, name, status, visible) VALUES (1, 'NET-ORDER-1', 'classic', 'Classic', 'active', 1)`,
		`INSERT INTO plan_network_types (plan_id, network_type_id, status) VALUES (1, 1, 'active')`,
	}
	for _, statement := range statements {
		require.NoError(t, db.Exec(statement).Error)
	}
}

type orderHandlerOrderSeed struct {
	OrderNo     string
	UserID      uint64
	ClientToken string
	Status      string
}

func insertOrderHandlerOrder(t *testing.T, db *gorm.DB, seed orderHandlerOrderSeed) {
	t.Helper()
	if seed.Status == "" {
		seed.Status = domainorder.StatusPending
	}
	require.NoError(t, db.Exec(`
INSERT INTO orders (
  order_no, user_id, client_token, status, order_type,
  product_no, product_type, product_name, plan_no, plan_code, plan_name,
  cpu_cores, memory_mb, system_disk_gb, data_disk_gb, bandwidth_mbps,
  public_ip_count, virtualization, architecture, billing_cycle, price_cents,
  original_price_cents, currency, quantity, total_amount_cents, payment_status,
  region_no, region_code, region_name, network_type_no, network_type_code,
  network_type_name, template_no, template_code, template_name, os_family,
  os_distribution, os_version, os_architecture
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		seed.OrderNo, seed.UserID, seed.ClientToken, seed.Status, domainorder.TypePurchase,
		"PROD-ORDER-1", "server", "Server", "PLAN-ORDER-1", "basic", "Basic",
		2, 4096, 40, 0, 100, 1, "kvm", "x86_64", "monthly", 1200,
		1500, "CNY", 1, 1200, domainorder.PaymentStatusUnpaid,
		"REG-ORDER-1", "cn", "China", "NET-ORDER-1", "classic", "Classic",
		"TPL-ORDER-1", "ubuntu", "Ubuntu", "linux", "ubuntu", "22.04", "x86_64",
	).Error)
}

func requireOrderStatus(t *testing.T, db *gorm.DB, orderNo string, status string) {
	t.Helper()
	var got string
	require.NoError(t, db.Table("orders").Select("status").Where("order_no = ?", orderNo).Take(&got).Error)
	require.Equal(t, status, got)
}

func requireEnvelope(t *testing.T, recorder *httptest.ResponseRecorder, code int) {
	t.Helper()
	var envelope response.Envelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
	require.Equal(t, code, envelope.Code)
}

const orderHandlerProductsSchema = `
CREATE TABLE products (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  product_no VARCHAR(64) NOT NULL,
  type VARCHAR(32) NOT NULL,
  slug VARCHAR(96) NOT NULL,
  name VARCHAR(128) NOT NULL,
  summary VARCHAR(255) NULL,
  description TEXT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'draft',
  visible TINYINT(1) NOT NULL DEFAULT 0,
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_products_product_no (product_no),
  UNIQUE KEY uk_products_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const orderHandlerProductPlansSchema = `
CREATE TABLE product_plans (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  plan_no VARCHAR(64) NOT NULL,
  product_id BIGINT UNSIGNED NOT NULL,
  code VARCHAR(96) NOT NULL,
  name VARCHAR(128) NOT NULL,
  summary VARCHAR(255) NULL,
  cpu_cores INT NOT NULL,
  memory_mb INT NOT NULL,
  system_disk_gb INT NOT NULL,
  data_disk_gb INT NOT NULL DEFAULT 0,
  bandwidth_mbps INT NOT NULL,
  traffic_gb INT NULL,
  public_ip_count INT NOT NULL DEFAULT 1,
  virtualization VARCHAR(32) NOT NULL DEFAULT 'kvm',
  architecture VARCHAR(32) NOT NULL DEFAULT 'x86_64',
  is_featured TINYINT(1) NOT NULL DEFAULT 0,
  status VARCHAR(32) NOT NULL DEFAULT 'draft',
  visible TINYINT(1) NOT NULL DEFAULT 0,
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_product_plans_plan_no (plan_no),
  UNIQUE KEY uk_product_plans_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const orderHandlerPlanPricesSchema = `
CREATE TABLE plan_prices (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  plan_id BIGINT UNSIGNED NOT NULL,
  billing_cycle VARCHAR(32) NOT NULL,
  price_cents BIGINT UNSIGNED NOT NULL,
  original_price_cents BIGINT UNSIGNED NULL,
  currency VARCHAR(16) NOT NULL DEFAULT 'CNY',
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_plan_prices_plan_cycle (plan_id, billing_cycle)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const orderHandlerSalesRegionsSchema = `
CREATE TABLE sales_regions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  region_no VARCHAR(64) NOT NULL,
  code VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  country VARCHAR(64) NULL,
  city VARCHAR(64) NULL,
  summary VARCHAR(255) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  visible TINYINT(1) NOT NULL DEFAULT 1,
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_sales_regions_region_no (region_no),
  UNIQUE KEY uk_sales_regions_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const orderHandlerTemplatesSchema = `
CREATE TABLE server_os_templates (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  template_no VARCHAR(64) NOT NULL,
  code VARCHAR(96) NOT NULL,
  name VARCHAR(128) NOT NULL,
  os_family VARCHAR(32) NOT NULL,
  distribution VARCHAR(64) NOT NULL,
  version VARCHAR(64) NOT NULL,
  architecture VARCHAR(32) NOT NULL DEFAULT 'x86_64',
  summary VARCHAR(255) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  visible TINYINT(1) NOT NULL DEFAULT 1,
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_server_os_templates_template_no (template_no),
  UNIQUE KEY uk_server_os_templates_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const orderHandlerPlanRegionsSchema = `
CREATE TABLE plan_regions (
  plan_id BIGINT UNSIGNED NOT NULL,
  region_id BIGINT UNSIGNED NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (plan_id, region_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const orderHandlerPlanTemplatesSchema = `
CREATE TABLE plan_os_templates (
  plan_id BIGINT UNSIGNED NOT NULL,
  template_id BIGINT UNSIGNED NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (plan_id, template_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const orderHandlerNetworkTypesSchema = `
CREATE TABLE network_types (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  network_type_no VARCHAR(64) NOT NULL,
  code VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  summary VARCHAR(255) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  visible TINYINT(1) NOT NULL DEFAULT 1,
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uk_network_types_network_type_no (network_type_no),
  UNIQUE KEY uk_network_types_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const orderHandlerPlanNetworkTypesSchema = `
CREATE TABLE plan_network_types (
  plan_id BIGINT UNSIGNED NOT NULL,
  network_type_id BIGINT UNSIGNED NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (plan_id, network_type_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const orderHandlerOrdersSchema = `
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
  product_summary VARCHAR(255) NULL,
  plan_no VARCHAR(64) NOT NULL,
  plan_code VARCHAR(64) NOT NULL,
  plan_name VARCHAR(128) NOT NULL,
  plan_summary VARCHAR(255) NULL,
  cpu_cores INT NOT NULL,
  memory_mb INT NOT NULL,
  system_disk_gb INT NOT NULL,
  data_disk_gb INT NOT NULL DEFAULT 0,
  bandwidth_mbps INT NOT NULL,
  traffic_gb INT NULL,
  public_ip_count INT NOT NULL DEFAULT 1,
  virtualization VARCHAR(32) NOT NULL,
  architecture VARCHAR(32) NOT NULL,
  billing_cycle VARCHAR(32) NOT NULL,
  price_cents BIGINT UNSIGNED NOT NULL,
  original_price_cents BIGINT UNSIGNED NULL,
  currency VARCHAR(16) NOT NULL,
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
  network_type_no VARCHAR(64) NOT NULL DEFAULT '',
  network_type_code VARCHAR(64) NOT NULL DEFAULT '',
  network_type_name VARCHAR(128) NOT NULL DEFAULT '',
  template_no VARCHAR(64) NOT NULL,
  template_code VARCHAR(96) NOT NULL,
  template_name VARCHAR(128) NOT NULL,
  os_family VARCHAR(32) NOT NULL,
  os_distribution VARCHAR(64) NOT NULL,
  os_version VARCHAR(64) NOT NULL,
  os_architecture VARCHAR(32) NOT NULL,
  user_note VARCHAR(500) NULL,
  admin_note VARCHAR(1000) NULL,
  cancel_reason VARCHAR(500) NULL,
  closed_reason VARCHAR(500) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  cancelled_at DATETIME(3) NULL,
  closed_at DATETIME(3) NULL,
  UNIQUE KEY uk_orders_order_no (order_no),
  UNIQUE KEY uk_orders_user_client_token (user_id, client_token)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const orderHandlerUserBusinessLogsSchema = `
CREATE TABLE user_business_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  username VARCHAR(64) NULL,
  email VARCHAR(191) NULL,
  request_id VARCHAR(64) NULL,
  request_method VARCHAR(16) NULL,
  request_path VARCHAR(255) NULL,
  module VARCHAR(64) NOT NULL,
  action VARCHAR(96) NOT NULL,
  object_type VARCHAR(64) NOT NULL,
  object_id VARCHAR(128) NULL,
  summary VARCHAR(500) NULL,
  ip VARCHAR(64) NULL,
  user_agent VARCHAR(500) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
