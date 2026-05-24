package instance

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	domaininstance "github.com/AeolianCloud/pveCloud/server/internal/domain/instance"
	domainorder "github.com/AeolianCloud/pveCloud/server/internal/domain/order"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
)

func TestCreateRenewalOrderPersistsSnapshotAndBusinessLog(t *testing.T) {
	db := openRenewalOrderDB(t)
	seedRenewalCatalog(t, db)
	seedRenewalUserAndInstance(t, db, 11, "INS-renew-1", domaininstance.StatusRunning)

	service := NewService(db, nil)
	detail, err := service.CreateRenewalOrder(context.Background(), 11, "INS-renew-1", webdto.RenewalOrderCreateRequest{BillingCycle: "quarterly", ClientToken: "renew-token-1"})
	if err != nil {
		t.Fatalf("create renewal order: %v", err)
	}
	if detail.OrderType != domainorder.TypeRenewal || detail.Status != domainorder.StatusPending || detail.PaymentStatus != domainorder.PaymentStatusUnpaid {
		t.Fatalf("renewal order should be pending and unpaid: %#v", detail.OrderItem)
	}
	if detail.RelatedInstanceNo == nil || *detail.RelatedInstanceNo != "INS-renew-1" {
		t.Fatalf("renewal order should bind instance, got %#v", detail.RelatedInstanceNo)
	}
	if detail.TotalAmountCents != 3000 || detail.PriceCents != 3000 || detail.BillingCycle != "quarterly" {
		t.Fatalf("renewal price snapshot should come from current catalog, got amount=%d price=%d cycle=%s", detail.TotalAmountCents, detail.PriceCents, detail.BillingCycle)
	}
	if detail.PlanNo != "PLAN-1" || detail.RegionNo != "REG-1" || detail.TemplateNo != "TPL-1" || detail.NetworkTypeNo != "NET-1" {
		t.Fatalf("renewal order should copy catalog and instance selection snapshot: %#v", detail)
	}

	var orderCount int64
	if err := db.Table("orders").Where("order_type = ? AND related_instance_no = ?", domainorder.TypeRenewal, "INS-renew-1").Count(&orderCount).Error; err != nil {
		t.Fatalf("count renewal orders: %v", err)
	}
	if orderCount != 1 {
		t.Fatalf("expected exactly one renewal order, got %d", orderCount)
	}
	var instanceCount int64
	if err := db.Table("instances").Count(&instanceCount).Error; err != nil {
		t.Fatalf("count instances: %v", err)
	}
	if instanceCount != 1 {
		t.Fatalf("renewal order creation must not create instances, got %d rows", instanceCount)
	}
	var logCount int64
	if err := db.Table("user_business_logs").Where("user_id = ? AND action = ? AND object_type = ?", 11, "order.renewal.create", "order").Count(&logCount).Error; err != nil {
		t.Fatalf("count business logs: %v", err)
	}
	if logCount != 1 {
		t.Fatalf("renewal order creation should write one user business log, got %d", logCount)
	}
}

func TestCreateRenewalOrderIsIdempotentByUserClientToken(t *testing.T) {
	db := openRenewalOrderDB(t)
	seedRenewalCatalog(t, db)
	seedRenewalUserAndInstance(t, db, 11, "INS-renew-2", domaininstance.StatusRunning)

	service := NewService(db, nil)
	req := webdto.RenewalOrderCreateRequest{BillingCycle: "monthly", ClientToken: "renew-token-2"}
	first, err := service.CreateRenewalOrder(context.Background(), 11, "INS-renew-2", req)
	if err != nil {
		t.Fatalf("first renewal order create: %v", err)
	}
	second, err := service.CreateRenewalOrder(context.Background(), 11, "INS-renew-2", req)
	if err != nil {
		t.Fatalf("second renewal order create: %v", err)
	}
	if first.OrderNo != second.OrderNo {
		t.Fatalf("duplicate client token should return existing order, got %s and %s", first.OrderNo, second.OrderNo)
	}

	var orderCount int64
	if err := db.Table("orders").Where("user_id = ? AND client_token = ?", 11, "renew-token-2").Count(&orderCount).Error; err != nil {
		t.Fatalf("count idempotent renewal orders: %v", err)
	}
	if orderCount != 1 {
		t.Fatalf("duplicate client token must not create another order, got %d", orderCount)
	}
}

func TestCreateRenewalOrderRejectsReleasedOrReleasingInstance(t *testing.T) {
	tests := []struct {
		name   string
		status string
	}{
		{name: "released", status: domaininstance.StatusReleased},
		{name: "releasing", status: domaininstance.StatusReleasing},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openRenewalOrderDB(t)
			seedRenewalCatalog(t, db)
			seedRenewalUserAndInstance(t, db, 11, "INS-"+tt.name, tt.status)

			service := NewService(db, nil)
			_, err := service.CreateRenewalOrder(context.Background(), 11, "INS-"+tt.name, webdto.RenewalOrderCreateRequest{BillingCycle: "monthly", ClientToken: "renew-token-" + tt.name})
			assertAppErrorCode(t, err, apperrors.ErrConflict.Code)

			var orderCount int64
			if countErr := db.Table("orders").Where("order_type = ?", domainorder.TypeRenewal).Count(&orderCount).Error; countErr != nil {
				t.Fatalf("count rejected renewal orders: %v", countErr)
			}
			if orderCount != 0 {
				t.Fatalf("rejected %s instance should not create renewal order, got %d", tt.status, orderCount)
			}
		})
	}
}

func TestCreateRenewalOrderRejectsCrossUserInstance(t *testing.T) {
	db := openRenewalOrderDB(t)
	seedRenewalCatalog(t, db)
	seedRenewalUserAndInstance(t, db, 11, "INS-owner-only", domaininstance.StatusRunning)
	if err := db.Exec(`INSERT INTO users (id, username, email, password_hash, status) VALUES (?, ?, ?, ?, ?)`, 12, "other-user", "other@example.com", "hash", "active").Error; err != nil {
		t.Fatalf("insert other user: %v", err)
	}

	service := NewService(db, nil)
	_, err := service.CreateRenewalOrder(context.Background(), 12, "INS-owner-only", webdto.RenewalOrderCreateRequest{BillingCycle: "monthly", ClientToken: "cross-user-token"})
	assertAppErrorCode(t, err, apperrors.ErrNotFound.Code)

	var orderCount int64
	if err := db.Table("orders").Where("order_type = ?", domainorder.TypeRenewal).Count(&orderCount).Error; err != nil {
		t.Fatalf("count cross-user renewal orders: %v", err)
	}
	if orderCount != 0 {
		t.Fatalf("cross-user request should not create renewal order, got %d", orderCount)
	}
}

func TestCreateRenewalOrderRejectsInjectedInstanceNo(t *testing.T) {
	db := openRenewalOrderDB(t)
	seedRenewalCatalog(t, db)
	seedRenewalUserAndInstance(t, db, 11, "INS-owner-safe", domaininstance.StatusRunning)

	service := NewService(db, nil)
	_, err := service.CreateRenewalOrder(context.Background(), 11, "INS-owner-safe' OR '1'='1", webdto.RenewalOrderCreateRequest{BillingCycle: "monthly", ClientToken: "injected-instance-token"})
	assertAppErrorCode(t, err, apperrors.ErrNotFound.Code)

	var orderCount int64
	if err := db.Table("orders").Where("order_type = ?", domainorder.TypeRenewal).Count(&orderCount).Error; err != nil {
		t.Fatalf("count injected renewal orders: %v", err)
	}
	if orderCount != 0 {
		t.Fatalf("injected instance number should not create renewal order, got %d", orderCount)
	}
}

func TestCreateRenewalOrderRejectsClientTokenUsedByOtherOrder(t *testing.T) {
	db := openRenewalOrderDB(t)
	seedRenewalCatalog(t, db)
	seedRenewalUserAndInstance(t, db, 11, "INS-token-conflict", domaininstance.StatusRunning)

	service := NewService(db, nil)
	_, err := service.CreateRenewalOrder(context.Background(), 11, "INS-token-conflict", webdto.RenewalOrderCreateRequest{BillingCycle: "monthly", ClientToken: "purchase-token-INS-token-conflict"})
	assertAppErrorCode(t, err, apperrors.ErrConflict.Code)

	var orderCount int64
	if err := db.Table("orders").Where("order_type = ?", domainorder.TypeRenewal).Count(&orderCount).Error; err != nil {
		t.Fatalf("count token-conflict renewal orders: %v", err)
	}
	if orderCount != 0 {
		t.Fatalf("client token already used by purchase order should not create renewal order, got %d", orderCount)
	}
}

func openRenewalOrderDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db,
		renewalOrderUsersSchema,
		renewalOrderProductsSchema,
		renewalOrderProductPlansSchema,
		renewalOrderPlanPricesSchema,
		renewalOrderSalesRegionsSchema,
		renewalOrderTemplatesSchema,
		renewalOrderPlanRegionsSchema,
		renewalOrderPlanTemplatesSchema,
		renewalOrderNetworkTypesSchema,
		renewalOrderPlanNetworkTypesSchema,
		renewalOrderOrdersSchema,
		renewalOrderInstancesSchema,
		renewalOrderUserBusinessLogsSchema,
	)
	return db
}

func seedRenewalCatalog(t *testing.T, db *gorm.DB) {
	t.Helper()
	statements := []string{
		`INSERT INTO products (id, product_no, type, slug, name, summary, status, visible) VALUES (1, 'PROD-1', 'server', 'server-basic', 'Server', 'Server summary', 'active', 1)`,
		`INSERT INTO product_plans (id, plan_no, product_id, code, name, summary, cpu_cores, memory_mb, system_disk_gb, data_disk_gb, bandwidth_mbps, public_ip_count, virtualization, architecture, status, visible) VALUES (1, 'PLAN-1', 1, 'basic', 'Basic', 'Basic summary', 2, 4096, 40, 0, 100, 1, 'kvm', 'x86_64', 'active', 1)`,
		`INSERT INTO plan_prices (plan_id, billing_cycle, price_cents, original_price_cents, currency, status) VALUES (1, 'monthly', 1200, 1500, 'CNY', 'active'), (1, 'quarterly', 3000, 3600, 'CNY', 'active')`,
		`INSERT INTO sales_regions (id, region_no, code, name, status, visible) VALUES (1, 'REG-1', 'cn', 'China', 'active', 1)`,
		`INSERT INTO server_os_templates (id, template_no, code, name, os_family, distribution, version, architecture, status, visible) VALUES (1, 'TPL-1', 'ubuntu', 'Ubuntu', 'linux', 'ubuntu', '22.04', 'x86_64', 'active', 1)`,
		`INSERT INTO plan_regions (plan_id, region_id, status) VALUES (1, 1, 'active')`,
		`INSERT INTO plan_os_templates (plan_id, template_id, status) VALUES (1, 1, 'active')`,
		`INSERT INTO network_types (id, network_type_no, code, name, status, visible) VALUES (1, 'NET-1', 'classic', 'Classic', 'active', 1)`,
		`INSERT INTO plan_network_types (plan_id, network_type_id, status) VALUES (1, 1, 'active')`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("seed renewal catalog with %q: %v", statement, err)
		}
	}
}

func seedRenewalUserAndInstance(t *testing.T, db *gorm.DB, userID uint64, instanceNo string, status string) {
	t.Helper()
	if err := db.Exec(`INSERT INTO users (id, username, email, password_hash, status) VALUES (?, ?, ?, ?, ?)`, userID, "renew-user", "renew@example.com", "hash", "active").Error; err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if err := db.Exec(`
INSERT INTO orders (
  id, order_no, user_id, client_token, status, order_type,
  product_no, product_type, product_name, plan_no, plan_code, plan_name,
  cpu_cores, memory_mb, system_disk_gb, data_disk_gb, bandwidth_mbps,
  public_ip_count, virtualization, architecture, billing_cycle, price_cents,
  original_price_cents, currency, quantity, total_amount_cents, payment_status,
  region_no, region_code, region_name, network_type_no, network_type_code,
  network_type_name, template_no, template_code, template_name, os_family,
  os_distribution, os_version, os_architecture
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		100+userID, "ORD-purchase-"+instanceNo, userID, "purchase-token-"+instanceNo, domainorder.StatusFulfilled, domainorder.TypePurchase,
		"PROD-1", "server", "Server", "PLAN-1", "basic", "Basic",
		2, 4096, 40, 0, 100, 1, "kvm", "x86_64", "monthly", 1200,
		1500, "CNY", 1, 1200, domainorder.PaymentStatusUnpaid,
		"REG-1", "cn", "China", "NET-1", "classic", "Classic",
		"TPL-1", "ubuntu", "Ubuntu", "linux", "ubuntu", "22.04", "x86_64",
	).Error; err != nil {
		t.Fatalf("insert purchase order: %v", err)
	}
	networkTypeNo := "NET-1"
	networkTypeName := "Classic"
	if err := db.Exec(`
INSERT INTO instances (
  instance_no, user_id, order_id, order_no, status, product_no, product_name,
  plan_no, plan_name, cpu_cores, memory_mb, system_disk_gb, data_disk_gb,
  bandwidth_mbps, region_no, region_name, network_type_no, network_type_name,
  template_no, template_name, os_family, os_distribution, os_version,
  external_node, external_vmid
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		instanceNo, userID, 100+userID, "ORD-purchase-"+instanceNo, status, "PROD-1", "Server",
		"PLAN-1", "Basic", 2, 4096, 40, 0, 100, "REG-1", "China", networkTypeNo, networkTypeName,
		"TPL-1", "Ubuntu", "linux", "ubuntu", "22.04", "node-a", 1000+userID,
	).Error; err != nil {
		t.Fatalf("insert instance: %v", err)
	}
}

func assertAppErrorCode(t *testing.T, err error, code int) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected app error code %d, got nil", code)
	}
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error code %d, got %T: %v", code, err, err)
	}
	if appErr.Code != code {
		t.Fatalf("expected app error code %d, got %d (%s)", code, appErr.Code, appErr.Message)
	}
}

const renewalOrderUsersSchema = `
CREATE TABLE users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(64) NOT NULL,
  email VARCHAR(191) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  display_name VARCHAR(64) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  UNIQUE KEY uk_users_username (username),
  UNIQUE KEY uk_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const renewalOrderProductsSchema = `
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

const renewalOrderProductPlansSchema = `
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

const renewalOrderPlanPricesSchema = `
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

const renewalOrderSalesRegionsSchema = `
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

const renewalOrderTemplatesSchema = `
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

const renewalOrderPlanRegionsSchema = `
CREATE TABLE plan_regions (
  plan_id BIGINT UNSIGNED NOT NULL,
  region_id BIGINT UNSIGNED NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (plan_id, region_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const renewalOrderPlanTemplatesSchema = `
CREATE TABLE plan_os_templates (
  plan_id BIGINT UNSIGNED NOT NULL,
  template_id BIGINT UNSIGNED NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (plan_id, template_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const renewalOrderNetworkTypesSchema = `
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

const renewalOrderPlanNetworkTypesSchema = `
CREATE TABLE plan_network_types (
  plan_id BIGINT UNSIGNED NOT NULL,
  network_type_id BIGINT UNSIGNED NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (plan_id, network_type_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const renewalOrderOrdersSchema = `
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

const renewalOrderInstancesSchema = `
CREATE TABLE instances (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  instance_no VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  order_id BIGINT UNSIGNED NOT NULL,
  order_no VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
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
  external_resource_location VARCHAR(255) NULL,
  last_error_code VARCHAR(64) NULL,
  last_error_message VARCHAR(500) NULL,
  service_started_at DATETIME(3) NULL,
  expires_at DATETIME(3) NULL,
  expire_notice_sent_at DATETIME(3) NULL,
  expire_release_scheduled_at DATETIME(3) NULL,
  expire_released_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  released_at DATETIME(3) NULL,
  UNIQUE KEY uk_instances_instance_no (instance_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const renewalOrderUserBusinessLogsSchema = `
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
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  KEY idx_user_business_logs_user_time (user_id, created_at),
  KEY idx_user_business_logs_object (object_type, object_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
