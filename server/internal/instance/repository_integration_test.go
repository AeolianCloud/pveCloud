package instance_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/AeolianCloud/pveCloud/server/internal/audit"
	"github.com/AeolianCloud/pveCloud/server/internal/instance"
	"github.com/AeolianCloud/pveCloud/server/internal/notification"
	"github.com/AeolianCloud/pveCloud/server/internal/resource"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil"
)

func TestCreateInstanceTxIsIdempotentAcrossRetries(t *testing.T) {
	db := testutil.OpenMariaDB(t)
	testutil.SeedUser(t, db, 31001)
	seed := testutil.SeedSaleableCatalogWithBase(t, db, 32000)
	seedPaidOrderForProvision(t, db, seed)

	svc := instance.NewService(
		instance.NewMySQLRepository(db),
		resource.NewMockClient(),
		audit.NewService(audit.NewMySQLRepository(db)),
		notification.NewService(notification.NewMySQLRepository(db)),
	)

	first, err := svc.HandleCreateInstanceTask(context.Background(), 5001)
	if err != nil {
		t.Fatalf("first provision: %v", err)
	}
	second, err := svc.HandleCreateInstanceTask(context.Background(), 5001)
	if err != nil {
		t.Fatalf("second provision: %v", err)
	}
	if first.Instance.InstanceNo != second.Instance.InstanceNo {
		t.Fatalf("expected idempotent provision result, got %s and %s", first.Instance.InstanceNo, second.Instance.InstanceNo)
	}

	assertRowCount(t, db, "instances", 1)
	assertRowCount(t, db, "instance_services", 1)

	var orderStatus string
	if err := db.QueryRow(`SELECT order_status FROM orders WHERE id = 5001`).Scan(&orderStatus); err != nil {
		t.Fatalf("load order status: %v", err)
	}
	if orderStatus != "active" {
		t.Fatalf("expected active order status, got %s", orderStatus)
	}
}

func seedPaidOrderForProvision(t *testing.T, db *sql.DB, seed testutil.CatalogSeed) {
	t.Helper()

	statements := []string{
		fmt.Sprintf(`UPDATE resource_nodes SET reserved_instances = 1 WHERE id = %d`, seed.NodeID),
		fmt.Sprintf(`INSERT INTO resource_reservations (id, reservation_no, user_id, sku_id, region_id, node_id, status, expires_at, created_at, updated_at) VALUES (6001, 'R6001', 31001, %d, %d, %d, 'reserved', NOW(3), NOW(3), NOW(3))`, seed.SKUID, seed.RegionID, seed.NodeID),
		fmt.Sprintf(`INSERT INTO orders (id, order_no, user_id, sku_id, region_id, reservation_id, order_status, cycle_unit, original_amount, discount_amount, payable_amount, paid_at, created_at, updated_at) VALUES (5001, 'O5001', 31001, %d, %d, 6001, 'paid', 'month', 10000, 0, 10000, NOW(3), NOW(3), NOW(3))`, seed.SKUID, seed.RegionID),
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("seed paid order: %v", err)
		}
	}
}

func assertRowCount(t *testing.T, db *sql.DB, table string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&got); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if got != want {
		t.Fatalf("expected %d rows in %s, got %d", want, table, got)
	}
}
