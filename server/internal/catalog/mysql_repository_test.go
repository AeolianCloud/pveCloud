package catalog

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

func TestMySQLRepositoryFindSaleableNodeMapsRow(t *testing.T) {
	now := time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)
	state := &sqlState{
		queryRows: []*rowsResult{{
			columns: []string{"id", "node_no", "region_id", "node_name", "total_instances", "used_instances", "reserved_instances", "status", "created_at", "updated_at"},
			values:  [][]driver.Value{{uint64(4001), "N4001", uint64(3001), "node-a", int64(20), int64(5), int64(2), "active", now, now}},
		}},
	}
	db := openSQLTestDB(t, state)
	repo := NewMySQLRepository(db)

	node, err := repo.FindSaleableNode(context.Background(), nil, 2001, 3001)
	if err != nil {
		t.Fatalf("find saleable node: %v", err)
	}
	if node.ID != 4001 || node.RegionID != 3001 || node.NodeName != "node-a" {
		t.Fatalf("unexpected node: %+v", node)
	}
	if len(state.queryQueries) != 1 || !strings.Contains(state.queryQueries[0], "FROM sku_region_node_bindings") {
		t.Fatalf("expected saleable node query, got %#v", state.queryQueries)
	}
}

func TestMySQLRepositoryCreateReservationUpdatesCapacityAndInsertsReservation(t *testing.T) {
	state := &sqlState{lastInsertID: []int64{0, 61}}
	db := openSQLTestDB(t, state)
	repo := NewMySQLRepository(db)
	fixed := time.Date(2026, 4, 22, 10, 5, 0, 0, time.UTC)
	repo.now = func() time.Time { return fixed }

	reservation, err := repo.CreateReservation(context.Background(), nil, 4001, 1001, 2001, 3001, fixed.Add(15*time.Minute))
	if err != nil {
		t.Fatalf("create reservation: %v", err)
	}
	if reservation.ID != 61 || reservation.NodeID != 4001 || reservation.RegionID != 3001 {
		t.Fatalf("unexpected reservation: %+v", reservation)
	}
	if len(state.execQueries) != 2 {
		t.Fatalf("expected two exec queries, got %d", len(state.execQueries))
	}
	if !strings.Contains(state.execQueries[0], "UPDATE resource_nodes") {
		t.Fatalf("expected capacity update query, got %s", state.execQueries[0])
	}
	if !strings.Contains(state.execQueries[1], "INSERT INTO resource_reservations") {
		t.Fatalf("expected reservation insert query, got %s", state.execQueries[1])
	}
}

func TestMySQLRepositoryFindSaleableNodeUsesForUpdateWithinTransaction(t *testing.T) {
	now := time.Date(2026, 4, 22, 10, 10, 0, 0, time.UTC)
	state := &sqlState{
		queryRows: []*rowsResult{{
			columns: []string{"id", "node_no", "region_id", "node_name", "total_instances", "used_instances", "reserved_instances", "status", "created_at", "updated_at"},
			values:  [][]driver.Value{{uint64(4002), "N4002", uint64(3001), "node-b", int64(20), int64(4), int64(1), "active", now, now}},
		}},
	}
	db := openSQLTestDB(t, state)
	repo := NewMySQLRepository(db)

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	_, err = repo.FindSaleableNode(context.Background(), tx, 2001, 3001)
	if err != nil {
		t.Fatalf("find saleable node in tx: %v", err)
	}
	if len(state.queryQueries) != 1 {
		t.Fatalf("expected one query, got %d", len(state.queryQueries))
	}
	if !strings.Contains(state.queryQueries[0], "FOR UPDATE") {
		t.Fatalf("expected FOR UPDATE query, got %s", state.queryQueries[0])
	}
}

type sqlState struct {
	execQueries   []string
	execArgs      [][]driver.NamedValue
	queryQueries  []string
	queryArgs     [][]driver.NamedValue
	queryRows     []*rowsResult
	lastInsertID  []int64
	beginCount    int
	commitCount   int
	rollbackCount int
}

type rowsResult struct {
	columns []string
	values  [][]driver.Value
}

type sqlDriver struct {
	state *sqlState
}

func (d *sqlDriver) Open(name string) (driver.Conn, error) {
	return &sqlConn{state: d.state}, nil
}

type sqlConn struct {
	state *sqlState
}

func (c *sqlConn) Prepare(query string) (driver.Stmt, error) { return nil, fmt.Errorf("not supported") }
func (c *sqlConn) Close() error                              { return nil }
func (c *sqlConn) Begin() (driver.Tx, error) {
	c.state.beginCount++
	return &sqlTx{state: c.state}, nil
}
func (c *sqlConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return c.Begin()
}

func (c *sqlConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.state.execQueries = append(c.state.execQueries, query)
	c.state.execArgs = append(c.state.execArgs, args)
	idx := len(c.state.execQueries) - 1
	lastInsertID := int64(0)
	if idx < len(c.state.lastInsertID) {
		lastInsertID = c.state.lastInsertID[idx]
	}
	return sqlResult{lastInsertID: lastInsertID, rowsAffected: 1}, nil
}

func (c *sqlConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.state.queryQueries = append(c.state.queryQueries, query)
	c.state.queryArgs = append(c.state.queryArgs, args)
	idx := len(c.state.queryQueries) - 1
	if idx >= len(c.state.queryRows) {
		return &testRows{}, nil
	}
	return &testRows{columns: c.state.queryRows[idx].columns, values: c.state.queryRows[idx].values}, nil
}

type sqlTx struct {
	state *sqlState
}

func (t *sqlTx) Commit() error {
	t.state.commitCount++
	return nil
}

func (t *sqlTx) Rollback() error {
	t.state.rollbackCount++
	return nil
}

type sqlResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (r sqlResult) LastInsertId() (int64, error) { return r.lastInsertID, nil }
func (r sqlResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

type testRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *testRows) Columns() []string { return r.columns }
func (r *testRows) Close() error      { return nil }
func (r *testRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func openSQLTestDB(t *testing.T, state *sqlState) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("catalog-sql-%s", strings.NewReplacer("/", "-", " ", "-").Replace(t.Name()))
	sql.Register(name, &sqlDriver{state: state})
	db, err := sql.Open(name, "test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
