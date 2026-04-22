package order

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestMySQLRepositoryCreateOrderMapsFieldsAndDefaultStatus(t *testing.T) {
	state := &orderSQLState{lastInsertID: []int64{81}}
	db := openOrderSQLTestDB(t, state)
	repo := NewMySQLRepository(db)
	fixed := time.Date(2026, 4, 22, 11, 0, 0, 0, time.UTC)
	repo.now = func() time.Time { return fixed }

	row, err := repo.CreateOrder(context.Background(), nil, CreateOrderParams{
		UserID:         1001,
		SKUID:          2001,
		RegionID:       3001,
		Cycle:          "month",
		OriginalAmount: 10000,
		DiscountAmount: 500,
		PayableAmount:  9500,
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if row.ID != 81 || row.UserID != 1001 || row.SKUID != 2001 || row.RegionID != 3001 {
		t.Fatalf("unexpected order identity fields: %+v", row)
	}
	if row.Status != "pending_payment" || row.Cycle != "month" {
		t.Fatalf("unexpected order state fields: %+v", row)
	}
	if row.OriginalAmount != 10000 || row.DiscountAmount != 500 || row.PayableAmount != 9500 {
		t.Fatalf("unexpected order amount fields: %+v", row)
	}
	if len(state.execQueries) != 1 || !strings.Contains(state.execQueries[0], "INSERT INTO orders") {
		t.Fatalf("expected order insert query, got %#v", state.execQueries)
	}
}

func TestMySQLRepositoryBindReservationUpdatesOrderReservation(t *testing.T) {
	state := &orderSQLState{}
	db := openOrderSQLTestDB(t, state)
	repo := NewMySQLRepository(db)
	fixed := time.Date(2026, 4, 22, 11, 5, 0, 0, time.UTC)
	repo.now = func() time.Time { return fixed }

	if err := repo.BindReservation(context.Background(), nil, 5001, 6001); err != nil {
		t.Fatalf("bind reservation: %v", err)
	}
	if len(state.execQueries) != 1 || !strings.Contains(state.execQueries[0], "UPDATE orders") {
		t.Fatalf("expected order update query, got %#v", state.execQueries)
	}
	args := state.execArgs[0]
	if len(args) != 3 {
		t.Fatalf("expected 3 bind args, got %d", len(args))
	}
	if got := args[0].Value; got != int64(6001) {
		t.Fatalf("expected reservation id 6001, got %#v", got)
	}
	if got := args[1].Value; got != fixed.UTC() {
		t.Fatalf("expected updated_at %v, got %#v", fixed.UTC(), got)
	}
	if got := args[2].Value; got != int64(5001) {
		t.Fatalf("expected order id 5001, got %#v", got)
	}
}

type orderSQLState struct {
	execQueries  []string
	execArgs     [][]driver.NamedValue
	lastInsertID []int64
}

type orderSQLDriver struct {
	state *orderSQLState
}

func (d *orderSQLDriver) Open(name string) (driver.Conn, error) {
	return &orderSQLConn{state: d.state}, nil
}

type orderSQLConn struct {
	state *orderSQLState
}

func (c *orderSQLConn) Prepare(query string) (driver.Stmt, error) {
	return nil, fmt.Errorf("not supported")
}
func (c *orderSQLConn) Close() error              { return nil }
func (c *orderSQLConn) Begin() (driver.Tx, error) { return nil, fmt.Errorf("not supported") }

func (c *orderSQLConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.state.execQueries = append(c.state.execQueries, query)
	c.state.execArgs = append(c.state.execArgs, args)
	idx := len(c.state.execQueries) - 1
	lastInsertID := int64(0)
	if idx < len(c.state.lastInsertID) {
		lastInsertID = c.state.lastInsertID[idx]
	}
	return orderSQLResult{lastInsertID: lastInsertID, rowsAffected: 1}, nil
}

type orderSQLResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (r orderSQLResult) LastInsertId() (int64, error) { return r.lastInsertID, nil }
func (r orderSQLResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

func openOrderSQLTestDB(t *testing.T, state *orderSQLState) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("order-sql-%s", strings.NewReplacer("/", "-", " ", "-").Replace(t.Name()))
	sql.Register(name, &orderSQLDriver{state: state})
	db, err := sql.Open(name, "test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
