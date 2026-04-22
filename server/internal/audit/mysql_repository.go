package audit

import (
	"context"
	"database/sql"
	"time"
)

// MySQLRepository implements Repository backed by MariaDB.
type MySQLRepository struct {
	db  *sql.DB
	now func() time.Time
}

// NewMySQLRepository creates a new MySQLRepository.
func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{
		db:  db,
		now: time.Now,
	}
}

// Record persists an audit event into audit_logs.
func (r *MySQLRepository) Record(ctx context.Context, event string, businessType string, businessID uint64, payload []byte) error {
	now := r.now().UTC()
	_, err := r.db.ExecContext(ctx, `
INSERT INTO audit_logs (event, business_type, business_id, payload, created_at)
VALUES (?, ?, ?, ?, ?)
`, event, businessType, businessID, payload, now)
	return err
}

var _ Repository = (*MySQLRepository)(nil)
