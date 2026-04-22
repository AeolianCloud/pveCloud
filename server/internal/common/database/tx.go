package database

import (
	"context"
	"database/sql"
	"errors"
)

func WithTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) (err error) {
	if db == nil {
		return errors.New("database is required")
	}
	if fn == nil {
		return errors.New("transaction function is required")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
