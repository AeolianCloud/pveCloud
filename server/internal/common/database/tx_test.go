package database_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
)

func TestWithTxCommitsOnSuccess(t *testing.T) {
	state := &txState{}
	db := openTestDB(t, state)

	if err := database.WithTx(context.Background(), db, func(tx *sql.Tx) error {
		return nil
	}); err != nil {
		t.Fatalf("with tx: %v", err)
	}

	if state.beginCount != 1 {
		t.Fatalf("expected one begin, got %d", state.beginCount)
	}
	if state.commitCount != 1 {
		t.Fatalf("expected one commit, got %d", state.commitCount)
	}
	if state.rollbackCount != 0 {
		t.Fatalf("expected no rollback, got %d", state.rollbackCount)
	}
}

func TestWithTxRollsBackOnError(t *testing.T) {
	state := &txState{}
	db := openTestDB(t, state)

	wantErr := errors.New("boom")
	err := database.WithTx(context.Background(), db, func(tx *sql.Tx) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected original error, got %v", err)
	}

	if state.beginCount != 1 {
		t.Fatalf("expected one begin, got %d", state.beginCount)
	}
	if state.commitCount != 0 {
		t.Fatalf("expected no commit, got %d", state.commitCount)
	}
	if state.rollbackCount != 1 {
		t.Fatalf("expected one rollback, got %d", state.rollbackCount)
	}
}

func TestWithTxRollsBackOnPanic(t *testing.T) {
	state := &txState{}
	db := openTestDB(t, state)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic to propagate")
		}
		if state.rollbackCount != 1 {
			t.Fatalf("expected one rollback after panic, got %d", state.rollbackCount)
		}
	}()

	_ = database.WithTx(context.Background(), db, func(tx *sql.Tx) error {
		panic("boom")
	})
}

func TestWithTxJoinsRollbackError(t *testing.T) {
	state := &txState{rollbackErr: errors.New("rollback failed")}
	db := openTestDB(t, state)

	wantErr := errors.New("boom")
	err := database.WithTx(context.Background(), db, func(tx *sql.Tx) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected original error, got %v", err)
	}
	if !errors.Is(err, state.rollbackErr) {
		t.Fatalf("expected rollback error, got %v", err)
	}
}

type txState struct {
	mu            sync.Mutex
	beginCount    int
	commitCount   int
	rollbackCount int
	rollbackErr   error
}

type txDriver struct {
	state *txState
}

func (d *txDriver) Open(name string) (driver.Conn, error) {
	return &txConn{state: d.state}, nil
}

type txConn struct {
	state *txState
}

func (c *txConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported")
}

func (c *txConn) Close() error {
	return nil
}

func (c *txConn) Begin() (driver.Tx, error) {
	c.state.mu.Lock()
	c.state.beginCount++
	c.state.mu.Unlock()
	return &txHandle{state: c.state}, nil
}

func (c *txConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return c.Begin()
}

type txHandle struct {
	state *txState
}

func (t *txHandle) Commit() error {
	t.state.mu.Lock()
	t.state.commitCount++
	t.state.mu.Unlock()
	return nil
}

func (t *txHandle) Rollback() error {
	t.state.mu.Lock()
	t.state.rollbackCount++
	err := t.state.rollbackErr
	t.state.mu.Unlock()
	return err
}

func openTestDB(t *testing.T, state *txState) *sql.DB {
	t.Helper()

	name := fmt.Sprintf("tx-test-%s", strings.NewReplacer("/", "-", " ", "-").Replace(t.Name()))
	sql.Register(name, &txDriver{state: state})

	db, err := sql.Open(name, "test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}
