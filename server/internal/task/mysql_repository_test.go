package task

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

func TestMySQLRepositoryCreateTaskPersistsPendingRow(t *testing.T) {
	state := &taskSQLState{lastInsertID: []int64{81}}
	db := openTaskSQLTestDB(t, state)
	repo := NewMySQLRepository(db)
	fixed := time.Date(2026, 4, 23, 8, 0, 0, 0, time.UTC)
	repo.now = func() time.Time { return fixed }

	row, err := repo.CreateTask(context.Background(), CreateTaskParams{
		TaskType:      "create_instance",
		BusinessType:  "order",
		BusinessID:    5001,
		Status:        "pending",
		Payload:       []byte(`{"order_id":5001}`),
		NextRunAt:     fixed,
		MaxRetryCount: 5,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if row.ID != 81 || row.TaskType != "create_instance" || row.BusinessID != 5001 {
		t.Fatalf("unexpected task row: %+v", row)
	}
	if len(state.execQueries) != 1 || !strings.Contains(state.execQueries[0], "INSERT INTO async_tasks") {
		t.Fatalf("expected async_tasks insert, got %#v", state.execQueries)
	}
}

func TestMySQLRepositoryClaimPendingTaskMovesTaskToProcessing(t *testing.T) {
	now := time.Date(2026, 4, 23, 8, 5, 0, 0, time.UTC)
	state := &taskSQLState{
		queryRows: []*taskRowsResult{{
			columns: []string{
				"id", "task_no", "task_type", "business_type", "business_id", "status", "payload", "next_run_at",
				"retry_count", "max_retry_count", "locked_by", "locked_at", "created_at", "updated_at",
			},
			values: [][]driver.Value{{
				uint64(81), "T81", "create_instance", "order", uint64(5001), "pending", []byte(`{"order_id":5001}`), now,
				int64(0), int64(5), nil, nil, now, now,
			}},
		}},
	}
	db := openTaskSQLTestDB(t, state)
	repo := NewMySQLRepository(db)

	row, err := repo.ClaimPendingTask(context.Background(), now, "worker-1")
	if err != nil {
		t.Fatalf("claim task: %v", err)
	}
	if row == nil || row.Status != "processing" || row.LockedBy != "worker-1" {
		t.Fatalf("unexpected claimed task: %+v", row)
	}
	if len(state.queryQueries) != 1 || !strings.Contains(state.queryQueries[0], "FOR UPDATE") {
		t.Fatalf("expected claim select with FOR UPDATE, got %#v", state.queryQueries)
	}
	if len(state.execQueries) != 1 || !strings.Contains(state.execQueries[0], "SET status = 'processing'") {
		t.Fatalf("expected processing update, got %#v", state.execQueries)
	}
}

func TestMySQLLogRepositoryAppendLogPersistsRow(t *testing.T) {
	state := &taskSQLState{}
	db := openTaskSQLTestDB(t, state)
	repo := NewMySQLLogRepository(db)
	fixed := time.Date(2026, 4, 23, 8, 10, 0, 0, time.UTC)
	repo.now = func() time.Time { return fixed }

	if err := repo.AppendLog(context.Background(), 81, "info", "start task T81"); err != nil {
		t.Fatalf("append log: %v", err)
	}
	if len(state.execQueries) != 1 || !strings.Contains(state.execQueries[0], "INSERT INTO async_task_logs") {
		t.Fatalf("expected async_task_logs insert, got %#v", state.execQueries)
	}
}

type taskSQLState struct {
	execQueries  []string
	execArgs     [][]driver.NamedValue
	queryQueries []string
	queryArgs    [][]driver.NamedValue
	queryRows    []*taskRowsResult
	lastInsertID []int64
}

type taskRowsResult struct {
	columns []string
	values  [][]driver.Value
}

type taskSQLDriver struct {
	state *taskSQLState
}

func (d *taskSQLDriver) Open(name string) (driver.Conn, error) {
	return &taskSQLConn{state: d.state}, nil
}

type taskSQLConn struct {
	state *taskSQLState
}

func (c *taskSQLConn) Prepare(query string) (driver.Stmt, error) {
	return nil, fmt.Errorf("not supported")
}
func (c *taskSQLConn) Close() error              { return nil }
func (c *taskSQLConn) Begin() (driver.Tx, error) { return taskSQLTx{}, nil }
func (c *taskSQLConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return taskSQLTx{}, nil
}

func (c *taskSQLConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.state.execQueries = append(c.state.execQueries, query)
	c.state.execArgs = append(c.state.execArgs, args)
	idx := len(c.state.execQueries) - 1
	lastInsertID := int64(0)
	if idx < len(c.state.lastInsertID) {
		lastInsertID = c.state.lastInsertID[idx]
	}
	return taskSQLResult{lastInsertID: lastInsertID, rowsAffected: 1}, nil
}

func (c *taskSQLConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.state.queryQueries = append(c.state.queryQueries, query)
	c.state.queryArgs = append(c.state.queryArgs, args)
	idx := len(c.state.queryQueries) - 1
	if idx >= len(c.state.queryRows) {
		return &taskTestRows{}, nil
	}
	return &taskTestRows{columns: c.state.queryRows[idx].columns, values: c.state.queryRows[idx].values}, nil
}

type taskSQLTx struct{}

func (taskSQLTx) Commit() error   { return nil }
func (taskSQLTx) Rollback() error { return nil }

type taskSQLResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (r taskSQLResult) LastInsertId() (int64, error) { return r.lastInsertID, nil }
func (r taskSQLResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

type taskTestRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *taskTestRows) Columns() []string { return r.columns }
func (r *taskTestRows) Close() error      { return nil }
func (r *taskTestRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func openTaskSQLTestDB(t *testing.T, state *taskSQLState) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("task-sql-%s", strings.NewReplacer("/", "-", " ", "-").Replace(t.Name()))
	sql.Register(name, &taskSQLDriver{state: state})
	db, err := sql.Open(name, "test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
