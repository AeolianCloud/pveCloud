package productcatalog

import (
	"context"
	"testing"

	"gorm.io/gorm"

	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil/mysqltest"
)

func TestDeleteProductRejectsProductWithPlans(t *testing.T) {
	db := openProductCatalogDB(t)
	seedProductCatalog(t, db)

	service := NewProductCatalogService(db, nil)
	err := service.DeleteProduct(context.Background(), 7, 1)
	if apperrors.From(err).Code != apperrors.ErrConflict.Code {
		t.Fatalf("delete product with plans should conflict, got %v", err)
	}

	var productCount int64
	if err := db.Table("products").Where("id = ?", 1).Count(&productCount).Error; err != nil {
		t.Fatalf("count products: %v", err)
	}
	if productCount != 1 {
		t.Fatalf("product with plans must remain, got %d row(s)", productCount)
	}

	var auditCount int64
	if err := db.Table("admin_audit_logs").Where("action = ?", "product.delete").Count(&auditCount).Error; err != nil {
		t.Fatalf("count audit logs: %v", err)
	}
	if auditCount != 0 {
		t.Fatalf("rejected product delete must not write audit log, got %d", auditCount)
	}
}

func TestDeletePlanCascadesRelationsAndWritesAudit(t *testing.T) {
	db := openProductCatalogDB(t)
	seedProductCatalog(t, db)

	service := NewProductCatalogService(db, nil)
	if err := service.DeletePlan(context.Background(), 9, 10); err != nil {
		t.Fatalf("delete plan: %v", err)
	}

	for _, table := range []string{"product_plans", "plan_prices", "plan_regions", "plan_os_templates", "plan_network_types"} {
		var count int64
		if err := db.Table(table).Count(&count).Error; err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("delete plan should clear %s rows, got %d", table, count)
		}
	}

	var auditCount int64
	if err := db.Table("admin_audit_logs").Where("admin_id = ? AND action = ? AND object_type = ? AND object_id = ?", 9, "product_plan.delete", "product_catalog", "10").Count(&auditCount).Error; err != nil {
		t.Fatalf("count audit logs: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("delete plan should write one audit log, got %d", auditCount)
	}
}

func openProductCatalogDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := mysqltest.Open(t)
	mysqltest.Exec(t, db,
		productCatalogProductsSchema,
		productCatalogPlansSchema,
		productCatalogPricesSchema,
		productCatalogRegionsSchema,
		productCatalogTemplatesSchema,
		productCatalogNetworkTypesSchema,
		productCatalogPlanRegionsSchema,
		productCatalogPlanTemplatesSchema,
		productCatalogPlanNetworkTypesSchema,
		productCatalogAdminAuditLogsSchema,
	)
	return db
}

func seedProductCatalog(t *testing.T, db *gorm.DB) {
	t.Helper()
	statements := []string{
		`INSERT INTO products (id, product_no, type, slug, name, status, visible) VALUES (1, 'PROD-TEST-1', 'server', 'server-basic', 'Server', 'active', 1)`,
		`INSERT INTO product_plans (id, plan_no, product_id, code, name, cpu_cores, memory_mb, system_disk_gb, data_disk_gb, bandwidth_mbps, public_ip_count, virtualization, architecture, status, visible) VALUES (10, 'PLAN-TEST-1', 1, 'basic', 'Basic', 2, 4096, 40, 0, 100, 1, 'kvm', 'x86_64', 'active', 1)`,
		`INSERT INTO plan_prices (id, plan_id, billing_cycle, price_cents, currency, status) VALUES (20, 10, 'monthly', 1200, 'CNY', 'active')`,
		`INSERT INTO sales_regions (id, region_no, code, name, status, visible) VALUES (30, 'REG-TEST-1', 'cn', 'China', 'active', 1)`,
		`INSERT INTO server_os_templates (id, template_no, code, name, os_family, distribution, version, architecture, status, visible) VALUES (40, 'TPL-TEST-1', 'ubuntu', 'Ubuntu', 'linux', 'ubuntu', '22.04', 'x86_64', 'active', 1)`,
		`INSERT INTO network_types (id, network_type_no, code, name, status, visible) VALUES (50, 'NET-TEST-1', 'classic', 'Classic', 'active', 1)`,
		`INSERT INTO plan_regions (plan_id, region_id, status) VALUES (10, 30, 'active')`,
		`INSERT INTO plan_os_templates (plan_id, template_id, status) VALUES (10, 40, 'active')`,
		`INSERT INTO plan_network_types (plan_id, network_type_id, status) VALUES (10, 50, 'active')`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("seed product catalog with %q: %v", statement, err)
		}
	}
}

const productCatalogProductsSchema = `
CREATE TABLE products (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  product_no VARCHAR(64) NOT NULL,
  type VARCHAR(32) NOT NULL,
  slug VARCHAR(96) NOT NULL,
  name VARCHAR(128) NOT NULL,
  summary VARCHAR(255) NULL,
  description TEXT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'draft',
  visible TINYINT(1) NOT NULL DEFAULT 1,
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const productCatalogPlansSchema = `
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
  virtualization VARCHAR(32) NOT NULL,
  architecture VARCHAR(32) NOT NULL,
  is_featured TINYINT(1) NOT NULL DEFAULT 0,
  status VARCHAR(32) NOT NULL DEFAULT 'draft',
  visible TINYINT(1) NOT NULL DEFAULT 1,
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const productCatalogPricesSchema = `
CREATE TABLE plan_prices (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  plan_id BIGINT UNSIGNED NOT NULL,
  billing_cycle VARCHAR(32) NOT NULL,
  price_cents BIGINT UNSIGNED NOT NULL,
  original_price_cents BIGINT UNSIGNED NULL,
  currency VARCHAR(16) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const productCatalogRegionsSchema = `
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
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const productCatalogTemplatesSchema = `
CREATE TABLE server_os_templates (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  template_no VARCHAR(64) NOT NULL,
  code VARCHAR(96) NOT NULL,
  name VARCHAR(128) NOT NULL,
  os_family VARCHAR(32) NOT NULL,
  distribution VARCHAR(64) NOT NULL,
  version VARCHAR(64) NOT NULL,
  architecture VARCHAR(32) NOT NULL,
  summary VARCHAR(255) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  visible TINYINT(1) NOT NULL DEFAULT 1,
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const productCatalogNetworkTypesSchema = `
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
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const productCatalogPlanRegionsSchema = `
CREATE TABLE plan_regions (
  plan_id BIGINT UNSIGNED NOT NULL,
  region_id BIGINT UNSIGNED NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (plan_id, region_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const productCatalogPlanTemplatesSchema = `
CREATE TABLE plan_os_templates (
  plan_id BIGINT UNSIGNED NOT NULL,
  template_id BIGINT UNSIGNED NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (plan_id, template_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const productCatalogPlanNetworkTypesSchema = `
CREATE TABLE plan_network_types (
  plan_id BIGINT UNSIGNED NOT NULL,
  network_type_id BIGINT UNSIGNED NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (plan_id, network_type_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const productCatalogAdminAuditLogsSchema = `
CREATE TABLE admin_audit_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  admin_id BIGINT UNSIGNED NULL,
  admin_username VARCHAR(64) NULL,
  admin_display_name VARCHAR(64) NULL,
  session_id VARCHAR(64) NULL,
  request_id VARCHAR(64) NULL,
  request_method VARCHAR(16) NULL,
  request_path VARCHAR(255) NULL,
  action VARCHAR(128) NOT NULL,
  object_type VARCHAR(64) NOT NULL,
  object_id VARCHAR(128) NULL,
  before_data JSON NULL,
  after_data JSON NULL,
  ip VARCHAR(64) NULL,
  user_agent VARCHAR(500) NULL,
  remark VARCHAR(500) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
