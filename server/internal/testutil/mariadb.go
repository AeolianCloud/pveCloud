package testutil

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap/config"
	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
)

func OpenMariaDB(t *testing.T) *sql.DB {
	t.Helper()

	root := projectRoot(t)
	cfg, err := config.LoadFrom(filepath.Join(root, "config", "config.yaml"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	db, err := database.Open(cfg.MySQLDSN)
	if err != nil {
		t.Fatalf("open mariadb: %v", err)
	}

	if _, err := db.Exec(`SELECT GET_LOCK('pvecloud_integration_test', 30)`); err != nil {
		_ = db.Close()
		t.Fatalf("acquire integration test lock: %v", err)
	}

	if err := RunMigrations(db); err != nil {
		_ = db.Close()
		t.Fatalf("run migrations: %v", err)
	}
	if err := ResetTables(db); err != nil {
		_ = db.Close()
		t.Fatalf("reset tables: %v", err)
	}

	t.Cleanup(func() {
		_ = ResetTables(db)
		_, _ = db.Exec(`SELECT RELEASE_LOCK('pvecloud_integration_test')`)
		_ = db.Close()
	})

	return db
}

func projectRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller")
	}

	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	if root == "." || root == string(filepath.Separator) {
		t.Fatalf("unexpected project root: %s", root)
	}
	return root
}

func SeedSaleableCatalog(t *testing.T, db *sql.DB) {
	t.Helper()

	SeedSaleableCatalogWithBase(t, db, 1000)
}

type CatalogSeed struct {
	ProductID uint64
	SKUID     uint64
	RegionID  uint64
	NodeID    uint64
}

func SeedSaleableCatalogWithBase(t *testing.T, db *sql.DB, base uint64) CatalogSeed {
	t.Helper()

	seed := CatalogSeed{
		ProductID: base + 1,
		SKUID:     base + 2,
		RegionID:  base + 3,
		NodeID:    base + 4,
	}

	statements := []string{
		fmt.Sprintf(`INSERT INTO products (id, product_no, product_name, product_type, status, created_at, updated_at) VALUES (%d, 'P%d', 'Cloud Host', 'cloud_host', 'active', NOW(3), NOW(3))`, seed.ProductID, seed.ProductID),
		fmt.Sprintf(`INSERT INTO product_skus (id, sku_no, product_id, sku_name, cpu_cores, memory_mb, disk_gb, bandwidth_mbps, status, created_at, updated_at) VALUES (%d, 'SKU%d', %d, 'Starter', 2, 4096, 40, 100, 'active', NOW(3), NOW(3))`, seed.SKUID, seed.SKUID, seed.ProductID),
		fmt.Sprintf(`INSERT INTO regions (id, region_no, region_name, status, created_at, updated_at) VALUES (%d, 'R%d', 'Shanghai', 'active', NOW(3), NOW(3))`, seed.RegionID, seed.RegionID),
		fmt.Sprintf(`INSERT INTO resource_nodes (id, node_no, region_id, node_name, total_instances, used_instances, reserved_instances, status, created_at, updated_at) VALUES (%d, 'N%d', %d, 'node-1', 20, 0, 0, 'active', NOW(3), NOW(3))`, seed.NodeID, seed.NodeID, seed.RegionID),
		fmt.Sprintf(`INSERT INTO sku_region_node_bindings (id, sku_id, region_id, node_id, sale_status, created_at, updated_at) VALUES (%d, %d, %d, %d, 'saleable', NOW(3), NOW(3))`, base+5, seed.SKUID, seed.RegionID, seed.NodeID),
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("seed saleable catalog: %v", err)
		}
	}

	return seed
}

func SeedUser(t *testing.T, db *sql.DB, id uint64) {
	t.Helper()

	if _, err := db.Exec(`
INSERT INTO users (id, user_no, email, phone, password_hash, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, 'active', NOW(3), NOW(3))
`, id, fmt.Sprintf("U%d", id), fmt.Sprintf("user%d@example.com", id), fmt.Sprintf("138%08d", id%100000000), "hash"); err != nil {
		t.Fatalf("seed user: %v", err)
	}
}
