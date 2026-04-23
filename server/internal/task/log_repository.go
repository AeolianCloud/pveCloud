package task

import (
	"context"
	"database/sql"
	"time"
)

type LogRepository interface {
	AppendLog(ctx context.Context, taskID uint64, level, message string) error
}

type MySQLLogRepository struct {
	db  *sql.DB
	now func() time.Time
}

func NewMySQLLogRepository(db *sql.DB) *MySQLLogRepository {
	return &MySQLLogRepository{
		db:  db,
		now: time.Now,
	}
}

func (r *MySQLLogRepository) AppendLog(ctx context.Context, taskID uint64, level, message string) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO async_task_logs (task_id, log_level, message, created_at)
VALUES (?, ?, ?, ?)
`, taskID, level, message, r.now().UTC())
	return err
}
