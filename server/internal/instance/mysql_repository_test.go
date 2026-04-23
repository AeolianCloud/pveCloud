package instance

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/catalog"
	"github.com/AeolianCloud/pveCloud/server/internal/resource"
)

func TestMySQLRepositoryLoadPaidOrderForProvisionMapsOrderAndReservation(t *testing.T) {
	now := time.Date(2026, 4, 23, 11, 0, 0, 0, time.UTC)
	state := &instanceSQLState{
		queryRows: []*instanceRowsResult{{
			columns: []string{
				"id", "order_no", "user_id", "sku_id", "region_id", "cycle_unit", "payable_amount",
				"id", "reservation_no", "user_id", "sku_id", "region_id", "node_id", "status", "expires_at", "created_at", "updated_at",
			},
			values: [][]driver.Value{{
				uint64(5001), "O5001", uint64(1001), uint64(2001), uint64(3001), "month", int64(10000),
				uint64(6001), "R6001", uint64(1001), uint64(2001), uint64(3001), uint64(4001), "reserved", now, now, now,
			}},
		}},
	}
	db := openInstanceSQLTestDB(t, state)
	repo := NewMySQLRepository(db)

	orderRow, reservation, err := repo.LoadPaidOrderForProvision(context.Background(), 5001)
	if err != nil {
		t.Fatalf("load paid order: %v", err)
	}
	if orderRow.OrderNo != "O5001" || reservation.ReservationNo != "R6001" || reservation.NodeID != 4001 {
		t.Fatalf("unexpected load result: %+v %+v", orderRow, reservation)
	}
}

func TestMySQLRepositoryFindProvisionResultByOrderMapsExistingInstance(t *testing.T) {
	now := time.Date(2026, 4, 23, 11, 2, 0, 0, time.UTC)
	state := &instanceSQLState{
		queryRows: []*instanceRowsResult{{
			columns: []string{
				"id", "instance_no", "user_id", "order_id", "node_id", "instance_status", "instance_ref", "created_at", "updated_at",
				"id", "instance_id", "current_period_start_at", "current_period_end_at", "billing_status", "created_at", "updated_at",
			},
			values: [][]driver.Value{{
				uint64(9001), "I9001", uint64(1001), uint64(5001), uint64(4001), "running", "mock-vm-5001", now, now,
				uint64(9101), uint64(9001), now, now.Add(30 * 24 * time.Hour), "active", now, now,
			}},
		}},
	}
	db := openInstanceSQLTestDB(t, state)
	repo := NewMySQLRepository(db)

	result, found, err := repo.FindProvisionResultByOrder(context.Background(), 5001)
	if err != nil {
		t.Fatalf("find provision result: %v", err)
	}
	if !found || result.Instance.InstanceNo != "I9001" || result.Service.InstanceID != 9001 {
		t.Fatalf("unexpected existing result: %+v found=%v", result, found)
	}
}

func TestMySQLRepositoryCreateInstanceAndActivateOrderPersistsProvisioningFacts(t *testing.T) {
	state := &instanceSQLState{lastInsertID: []int64{0, 0, 0, 91, 92, 0}}
	db := openInstanceSQLTestDB(t, state)
	repo := NewMySQLRepository(db)
	fixed := time.Date(2026, 4, 23, 11, 5, 0, 0, time.UTC)
	repo.now = func() time.Time { return fixed }

	result, err := repo.CreateInstanceAndActivateOrder(context.Background(),
		PaidOrder{ID: 5001, UserID: 1001, Cycle: "month"},
		catalog.Reservation{
			ID:     6001,
			NodeID: 4001,
			Status: "reserved",
		},
		resource.CreateVMResponse{InstanceRef: "mock-vm-5001", Status: "running"},
	)
	if err != nil {
		t.Fatalf("create instance and activate order: %v", err)
	}
	if result.Instance.ID != 91 || result.Service.ID != 92 {
		t.Fatalf("unexpected inserted ids: %+v", result)
	}
	if len(state.execQueries) != 6 {
		t.Fatalf("expected 6 exec queries, got %d", len(state.execQueries))
	}
	if !strings.Contains(state.execQueries[3], "INSERT INTO instances") {
		t.Fatalf("expected instance insert query, got %#v", state.execQueries)
	}
	if !strings.Contains(state.execQueries[4], "INSERT INTO instance_services") {
		t.Fatalf("expected instance_services insert query, got %#v", state.execQueries)
	}
}

func TestMySQLRepositoryCreateInstanceAndActivateOrderReturnsExistingInstance(t *testing.T) {
	now := time.Date(2026, 4, 23, 11, 6, 0, 0, time.UTC)
	state := &instanceSQLState{
		queryRows: []*instanceRowsResult{{
			columns: []string{
				"id", "instance_no", "user_id", "order_id", "node_id", "instance_status", "instance_ref", "created_at", "updated_at",
				"id", "instance_id", "current_period_start_at", "current_period_end_at", "billing_status", "created_at", "updated_at",
			},
			values: [][]driver.Value{{
				uint64(9002), "I9002", uint64(1001), uint64(5002), uint64(4001), "running", "mock-vm-5002", now, now,
				uint64(9102), uint64(9002), now, now.Add(30 * 24 * time.Hour), "active", now, now,
			}},
		}},
	}
	db := openInstanceSQLTestDB(t, state)
	repo := NewMySQLRepository(db)

	result, err := repo.CreateInstanceAndActivateOrder(context.Background(),
		PaidOrder{ID: 5002, UserID: 1001, Cycle: "month"},
		catalog.Reservation{ID: 6002, NodeID: 4001, Status: "consumed"},
		resource.CreateVMResponse{InstanceRef: "mock-vm-5002", Status: "running"},
	)
	if err != nil {
		t.Fatalf("create instance and activate order: %v", err)
	}
	if result.Instance.InstanceNo != "I9002" || result.Service.InstanceID != 9002 {
		t.Fatalf("expected existing provision result, got %+v", result)
	}
	if len(state.execQueries) != 0 {
		t.Fatalf("expected no writes when instance already exists, got %#v", state.execQueries)
	}
}

type instanceSQLState struct {
	execQueries  []string
	execArgs     [][]driver.NamedValue
	queryQueries []string
	queryRows    []*instanceRowsResult
	lastInsertID []int64
}

type instanceRowsResult struct {
	columns []string
	values  [][]driver.Value
}

type instanceSQLDriver struct {
	state *instanceSQLState
}

func (d *instanceSQLDriver) Open(name string) (driver.Conn, error) {
	return &instanceSQLConn{state: d.state}, nil
}

type instanceSQLConn struct {
	state *instanceSQLState
}

func (c *instanceSQLConn) Prepare(query string) (driver.Stmt, error) {
	return nil, fmt.Errorf("not supported")
}
func (c *instanceSQLConn) Close() error              { return nil }
func (c *instanceSQLConn) Begin() (driver.Tx, error) { return instanceSQLTx{}, nil }
func (c *instanceSQLConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return instanceSQLTx{}, nil
}

func (c *instanceSQLConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.state.execQueries = append(c.state.execQueries, query)
	c.state.execArgs = append(c.state.execArgs, args)
	idx := len(c.state.execQueries) - 1
	lastInsertID := int64(0)
	if idx < len(c.state.lastInsertID) {
		lastInsertID = c.state.lastInsertID[idx]
	}
	return instanceSQLResult{lastInsertID: lastInsertID, rowsAffected: 1}, nil
}

func (c *instanceSQLConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.state.queryQueries = append(c.state.queryQueries, query)
	idx := len(c.state.queryQueries) - 1
	if idx >= len(c.state.queryRows) {
		return &instanceTestRows{}, nil
	}
	return &instanceTestRows{columns: c.state.queryRows[idx].columns, values: c.state.queryRows[idx].values}, nil
}

type instanceSQLTx struct{}

func (instanceSQLTx) Commit() error   { return nil }
func (instanceSQLTx) Rollback() error { return nil }

type instanceSQLResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (r instanceSQLResult) LastInsertId() (int64, error) { return r.lastInsertID, nil }
func (r instanceSQLResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

type instanceTestRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *instanceTestRows) Columns() []string { return r.columns }
func (r *instanceTestRows) Close() error      { return nil }
func (r *instanceTestRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func openInstanceSQLTestDB(t *testing.T, state *instanceSQLState) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("instance-sql-%s", strings.NewReplacer("/", "-", " ", "-").Replace(t.Name()))
	sql.Register(name, &instanceSQLDriver{state: state})
	db, err := sql.Open(name, "test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
