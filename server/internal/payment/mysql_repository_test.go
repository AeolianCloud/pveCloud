package payment

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

func TestMySQLRepositoryCreatePendingPaymentPersistsPendingRow(t *testing.T) {
	state := &sqlState{lastInsertID: []int64{71}}
	db := openSQLTestDB(t, state)
	repo := NewMySQLRepository(db)
	fixed := time.Date(2026, 4, 22, 9, 0, 0, 123, time.UTC)
	repo.now = func() time.Time { return fixed }

	row, err := repo.CreatePendingPayment(context.Background(), nil, 5001, 10000)
	if err != nil {
		t.Fatalf("create pending payment: %v", err)
	}
	if row.ID != 71 || row.PaymentOrderNo != fmt.Sprintf("P%d", fixed.UnixNano()) {
		t.Fatalf("unexpected payment row: %+v", row)
	}
	if len(state.execQueries) != 1 || !strings.Contains(state.execQueries[0], "INSERT INTO payment_orders") {
		t.Fatalf("expected payment insert query, got %#v", state.execQueries)
	}
	if got := state.execArgs[0][1].Value; got != int64(5001) {
		t.Fatalf("expected order id arg 5001, got %#v", got)
	}
	if got := state.execArgs[0][2].Value; got != int64(10000) {
		t.Fatalf("expected payable amount arg 10000, got %#v", got)
	}
}

func TestMySQLRepositoryGetByPaymentOrderNoMapsFields(t *testing.T) {
	paidAt := time.Date(2026, 4, 22, 9, 5, 0, 0, time.UTC)
	state := &sqlState{
		queryRows: []*rowsResult{{
			columns: []string{"id", "payment_order_no", "order_id", "pay_status", "payable_amount", "paid_at"},
			values:  [][]driver.Value{{uint64(71), "P71", uint64(5001), "success", int64(10000), paidAt}},
		}},
	}
	db := openSQLTestDB(t, state)
	repo := NewMySQLRepository(db)

	row, err := repo.GetByPaymentOrderNo(context.Background(), "P71")
	if err != nil {
		t.Fatalf("get payment order: %v", err)
	}
	if row.OrderID != 5001 || row.PayStatus != "success" || row.PaidAt == nil || !row.PaidAt.Equal(paidAt) {
		t.Fatalf("unexpected payment row: %+v", row)
	}
	if len(state.queryQueries) != 1 || !strings.Contains(state.queryQueries[0], "FROM payment_orders") {
		t.Fatalf("expected payment select query, got %#v", state.queryQueries)
	}
}

type sqlState struct {
	execQueries  []string
	execArgs     [][]driver.NamedValue
	queryQueries []string
	queryArgs    [][]driver.NamedValue
	queryRows    []*rowsResult
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
	name := fmt.Sprintf("payment-sql-%s", strings.NewReplacer("/", "-", " ", "-").Replace(t.Name()))
	sql.Register(name, &sqlDriver{state: state})
	db, err := sql.Open(name, "test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
