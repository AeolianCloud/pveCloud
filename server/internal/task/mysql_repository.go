package task

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type MySQLRepository struct {
	db  *sql.DB
	now func() time.Time
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{
		db:  db,
		now: time.Now,
	}
}

func (r *MySQLRepository) FindByBusinessKey(ctx context.Context, taskType, businessType string, businessID uint64) (Task, bool, error) {
	row, err := r.scanTask(r.db.QueryRowContext(ctx, `
SELECT id, task_no, task_type, business_type, business_id, status, payload, next_run_at,
	retry_count, max_retry_count, locked_by, locked_at, created_at, updated_at
FROM async_tasks
WHERE task_type = ? AND business_type = ? AND business_id = ?
`, taskType, businessType, businessID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Task{}, false, nil
		}
		return Task{}, false, err
	}
	return row, true, nil
}

func (r *MySQLRepository) CreateTask(ctx context.Context, in CreateTaskParams) (Task, error) {
	now := r.now().UTC()
	taskNo := NewTaskNo(now)
	result, err := r.db.ExecContext(ctx, `
INSERT INTO async_tasks (
	task_no, task_type, business_type, business_id, status, payload, next_run_at,
	retry_count, max_retry_count, created_at, updated_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?, ?, ?)
`, taskNo, in.TaskType, in.BusinessType, in.BusinessID, in.Status, in.Payload, in.NextRunAt, in.MaxRetryCount, now, now)
	if err != nil {
		return Task{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Task{}, err
	}

	return Task{
		ID:            uint64(id),
		TaskNo:        taskNo,
		TaskType:      in.TaskType,
		BusinessType:  in.BusinessType,
		BusinessID:    in.BusinessID,
		Status:        in.Status,
		Payload:       in.Payload,
		NextRunAt:     in.NextRunAt,
		RetryCount:    0,
		MaxRetryCount: in.MaxRetryCount,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (r *MySQLRepository) ListTasks(ctx context.Context, limit int) ([]Task, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT id, task_no, task_type, business_type, business_id, status, payload, next_run_at,
	retry_count, max_retry_count, locked_by, locked_at, created_at, updated_at
FROM async_tasks
ORDER BY created_at DESC, id DESC
LIMIT ?
`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Task
	for rows.Next() {
		row, err := scanTaskFromRows(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, row)
	}

	return items, rows.Err()
}

func (r *MySQLRepository) ClaimPendingTask(ctx context.Context, now time.Time, workerName string) (*Task, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	row, err := r.scanTask(tx.QueryRowContext(ctx, `
SELECT id, task_no, task_type, business_type, business_id, status, payload, next_run_at,
	retry_count, max_retry_count, locked_by, locked_at, created_at, updated_at
FROM async_tasks
WHERE status IN ('pending', 'retrying') AND next_run_at <= ?
ORDER BY next_run_at ASC, id ASC
LIMIT 1
FOR UPDATE
`, now.UTC()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err := tx.Commit(); err != nil {
				return nil, err
			}
			return nil, nil
		}
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE async_tasks
SET status = 'processing', locked_by = ?, locked_at = ?, updated_at = ?
WHERE id = ?
`, workerName, now.UTC(), now.UTC(), row.ID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	row.Status = "processing"
	row.LockedBy = workerName
	row.LockedAt = now.UTC()
	row.UpdatedAt = now.UTC()
	return &row, nil
}

func (r *MySQLRepository) MarkTaskSuccess(ctx context.Context, taskID uint64, now time.Time) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE async_tasks
SET status = 'success', locked_by = NULL, locked_at = NULL, updated_at = ?
WHERE id = ?
`, now.UTC(), taskID)
	return err
}

func (r *MySQLRepository) MarkTaskRetry(ctx context.Context, taskID uint64, nextRunAt time.Time, now time.Time) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE async_tasks
SET status = 'retrying', retry_count = retry_count + 1, next_run_at = ?, locked_by = NULL, locked_at = NULL, updated_at = ?
WHERE id = ?
`, nextRunAt.UTC(), now.UTC(), taskID)
	return err
}

func (r *MySQLRepository) MarkTaskFailed(ctx context.Context, taskID uint64, now time.Time) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE async_tasks
SET status = 'failed', locked_by = NULL, locked_at = NULL, updated_at = ?
WHERE id = ?
`, now.UTC(), taskID)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func (r *MySQLRepository) scanTask(row scanner) (Task, error) {
	var item Task
	var payload []byte
	var lockedBy sql.NullString
	var lockedAt sql.NullTime
	err := row.Scan(
		&item.ID,
		&item.TaskNo,
		&item.TaskType,
		&item.BusinessType,
		&item.BusinessID,
		&item.Status,
		&payload,
		&item.NextRunAt,
		&item.RetryCount,
		&item.MaxRetryCount,
		&lockedBy,
		&lockedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return Task{}, err
	}
	item.Payload = payload
	if lockedBy.Valid {
		item.LockedBy = lockedBy.String
	}
	if lockedAt.Valid {
		item.LockedAt = lockedAt.Time
	}
	return item, nil
}

func scanTaskFromRows(rows *sql.Rows) (Task, error) {
	repo := &MySQLRepository{}
	return repo.scanTask(rows)
}
