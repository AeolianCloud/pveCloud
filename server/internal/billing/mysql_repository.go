package billing

import (
	"context"
	"database/sql"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
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

func (r *MySQLRepository) CreateRecord(ctx context.Context, q database.Querier, in CreateRecordInput) (Record, error) {
	querier := r.querier(q)
	now := r.now().UTC()

	result, err := querier.ExecContext(ctx, `
INSERT INTO billing_records (
	order_id, billing_type, cycle_unit, original_amount, discount_amount, payable_amount, created_at, updated_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`, in.OrderID, in.BillingType, in.Cycle, in.OriginalAmount, in.DiscountAmount, in.PayableAmount, now, now)
	if err != nil {
		return Record{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Record{}, err
	}

	return Record{
		ID:             uint64(id),
		OrderID:        in.OrderID,
		BillingType:    in.BillingType,
		Cycle:          in.Cycle,
		OriginalAmount: in.OriginalAmount,
		DiscountAmount: in.DiscountAmount,
		PayableAmount:  in.PayableAmount,
	}, nil
}

func (r *MySQLRepository) querier(q database.Querier) database.Querier {
	if q != nil {
		return q
	}
	return r.db
}
