package payment

import (
	"context"
	"database/sql"
	"fmt"
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

func (r *MySQLRepository) CreatePendingPayment(ctx context.Context, q database.Querier, orderID uint64, payableAmount int64) (PaymentOrder, error) {
	querier := r.querier(q)
	now := r.now().UTC()
	paymentOrderNo := fmt.Sprintf("P%d", now.UnixNano())

	result, err := querier.ExecContext(ctx, `
INSERT INTO payment_orders (
	payment_order_no, order_id, pay_status, payable_amount, created_at, updated_at
)
VALUES (?, ?, 'pending', ?, ?, ?)
`, paymentOrderNo, orderID, payableAmount, now, now)
	if err != nil {
		return PaymentOrder{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return PaymentOrder{}, err
	}

	return PaymentOrder{
		ID:             uint64(id),
		PaymentOrderNo: paymentOrderNo,
		OrderID:        orderID,
		PayStatus:      "pending",
		PayableAmount:  payableAmount,
	}, nil
}

func (r *MySQLRepository) GetByPaymentOrderNo(ctx context.Context, paymentOrderNo string) (PaymentOrder, error) {
	var row PaymentOrder
	var paidAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
SELECT id, payment_order_no, order_id, pay_status, payable_amount, paid_at
FROM payment_orders
WHERE payment_order_no = ?
`, paymentOrderNo).Scan(
		&row.ID,
		&row.PaymentOrderNo,
		&row.OrderID,
		&row.PayStatus,
		&row.PayableAmount,
		&paidAt,
	)
	if err != nil {
		return PaymentOrder{}, err
	}
	if paidAt.Valid {
		row.PaidAt = &paidAt.Time
	}
	return row, nil
}

func (r *MySQLRepository) querier(q database.Querier) database.Querier {
	if q != nil {
		return q
	}
	return r.db
}
