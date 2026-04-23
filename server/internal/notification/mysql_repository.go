package notification

import (
	"context"
	"database/sql"
	"time"
)

type MySQLRepository struct {
	db  *sql.DB
	now func() time.Time
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db, now: time.Now}
}

func (r *MySQLRepository) Create(ctx context.Context, n Notification) (Notification, error) {
	now := r.now().UTC()
	result, err := r.db.ExecContext(ctx, `
INSERT INTO notifications (user_id, title, body, type, is_read, created_at, updated_at)
VALUES (?, ?, ?, ?, 0, ?, ?)
`, n.UserID, n.Title, n.Body, n.Type, now, now)
	if err != nil {
		return Notification{}, err
	}
	id, _ := result.LastInsertId()
	n.ID = uint64(id)
	n.CreatedAt = now
	return n, nil
}

func (r *MySQLRepository) ListByUser(ctx context.Context, userID uint64, limit int) ([]Notification, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, user_id, title, body, type, is_read, created_at
FROM notifications
WHERE user_id = ?
ORDER BY created_at DESC
LIMIT ?
`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Notification
	for rows.Next() {
		var n Notification
		var isRead int
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Type, &isRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		n.IsRead = isRead == 1
		items = append(items, n)
	}
	return items, nil
}

func (r *MySQLRepository) MarkRead(ctx context.Context, id uint64, userID uint64) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE notifications SET is_read = 1, updated_at = ? WHERE id = ? AND user_id = ?
`, r.now().UTC(), id, userID)
	return err
}

func (r *MySQLRepository) CountUnread(ctx context.Context, userID uint64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0
`, userID).Scan(&count)
	return count, err
}
