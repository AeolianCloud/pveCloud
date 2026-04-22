package billing

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

func TestMySQLRepositoryCreateRecordPersistsSnapshot(t *testing.T) {
	state := &sqlState{
		execResults:  []driver.Result{driver.RowsAffected(1)},
		lastInsertID: []int64{81},
	}
	db := openSQLTestDB(t, state)
	repo := NewMySQLRepository(db)
	fixed := time.Date(2026, 4, 22, 8, 30, 0, 0, time.UTC)
	repo.now = func() time.Time { return fixed }

	record, err := repo.CreateRecord(context.Background(), nil, CreateRecordInput{
		OrderID:        5001,
		BillingType:    "create",
		Cycle:          "month",
		OriginalAmount: 10000,
		DiscountAmount: 500,
		PayableAmount:  9500,
	})
	if err != nil {
		t.Fatalf("create record: %v", err)
	}
	if record.ID != 81 || record.OrderID != 5001 || record.PayableAmount != 9500 {
		t.Fatalf("unexpected record: %+v", record)
	}
	if len(state.execQueries) != 1 || !strings.Contains(state.execQueries[0], "INSERT INTO billing_records") {
		t.Fatalf("expected billing insert query, got %#v", state.execQueries)
	}
	args := state.execArgs[0]
	if got := args[0].Value; got != int64(5001) {
		t.Fatalf("expected order id arg 5001, got %#v", got)
	}
	if got := args[1].Value; got != "create" {
		t.Fatalf("expected billing type create, got %#v", got)
	}
	if got := args[5].Value; got != int64(9500) {
		t.Fatalf("expected payable amount 9500, got %#v", got)
	}
}

type sqlState struct {
	execQueries  []string
	execArgs     [][]driver.NamedValue
	queryQueries []string
	queryArgs    [][]driver.NamedValue
	queryRows    []*rowsResult
	execResults  []driver.Result
	lastInsertID []int64
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
func (c *sqlConn) Begin() (driver.Tx, error)                 { return nil, fmt.Errorf("not supported") }

func (c *sqlConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.state.execQueries = append(c.state.execQueries, query)
	c.state.execArgs = append(c.state.execArgs, args)
	idx := len(c.state.execQueries) - 1
	if idx < len(c.state.lastInsertID) {
		return sqlResult{rowsAffected: 1, lastInsertID: c.state.lastInsertID[idx]}, nil
	}
	if idx < len(c.state.execResults) {
		return c.state.execResults[idx], nil
	}
	return sqlResult{rowsAffected: 1}, nil
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
	name := fmt.Sprintf("billing-sql-%s", strings.NewReplacer("/", "-", " ", "-").Replace(t.Name()))
	sql.Register(name, &sqlDriver{state: state})
	db, err := sql.Open(name, "test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
