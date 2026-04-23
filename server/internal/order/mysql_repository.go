package order

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

func (r *MySQLRepository) CreateOrder(ctx context.Context, q database.Querier, in CreateOrderParams) (Order, error) {
	querier := r.querier(q)
	now := r.now().UTC()
	orderNo := fmt.Sprintf("O%d", now.UnixNano())

	result, err := querier.ExecContext(ctx, `
INSERT INTO orders (
	order_no, user_id, sku_id, region_id, order_status, cycle_unit, original_amount, discount_amount, payable_amount, created_at, updated_at
)
VALUES (?, ?, ?, ?, 'pending_payment', ?, ?, ?, ?, ?, ?)
`, orderNo, in.UserID, in.SKUID, in.RegionID, in.Cycle, in.OriginalAmount, in.DiscountAmount, in.PayableAmount, now, now)
	if err != nil {
		return Order{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Order{}, err
	}

	return Order{
		ID:             uint64(id),
		OrderNo:        orderNo,
		UserID:         in.UserID,
		SKUID:          in.SKUID,
		RegionID:       in.RegionID,
		Status:         "pending_payment",
		Cycle:          in.Cycle,
		OriginalAmount: in.OriginalAmount,
		DiscountAmount: in.DiscountAmount,
		PayableAmount:  in.PayableAmount,
	}, nil
}

func (r *MySQLRepository) BindReservation(ctx context.Context, q database.Querier, orderID, reservationID uint64) error {
	querier := r.querier(q)
	_, err := querier.ExecContext(ctx, `
UPDATE orders
SET reservation_id = ?, updated_at = ?
WHERE id = ?
`, reservationID, r.now().UTC(), orderID)
	return err
}

func (r *MySQLRepository) GetOrderByID(ctx context.Context, orderID uint64) (Order, error) {
	var row Order
	var paidAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
SELECT id, order_no, user_id, sku_id, region_id, IFNULL(reservation_id, 0), order_status, cycle_unit, original_amount, discount_amount, payable_amount
FROM orders
WHERE id = ?
`, orderID).Scan(
		&row.ID, &row.OrderNo, &row.UserID, &row.SKUID, &row.RegionID, &row.ReservationID,
		&row.Status, &row.Cycle, &row.OriginalAmount, &row.DiscountAmount, &row.PayableAmount,
	)
	if err != nil {
		return Order{}, err
	}
	_ = paidAt
	return row, nil
}

func (r *MySQLRepository) ListOrdersByUser(ctx context.Context, userID uint64) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, order_no, user_id, sku_id, region_id, IFNULL(reservation_id, 0), order_status, cycle_unit, original_amount, discount_amount, payable_amount
FROM orders
WHERE user_id = ?
ORDER BY created_at DESC
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.UserID, &o.SKUID, &o.RegionID, &o.ReservationID,
			&o.Status, &o.Cycle, &o.OriginalAmount, &o.DiscountAmount, &o.PayableAmount); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *MySQLRepository) ListAllOrders(ctx context.Context) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, order_no, user_id, sku_id, region_id, IFNULL(reservation_id, 0), order_status, cycle_unit, original_amount, discount_amount, payable_amount
FROM orders
ORDER BY created_at DESC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.UserID, &o.SKUID, &o.RegionID, &o.ReservationID,
			&o.Status, &o.Cycle, &o.OriginalAmount, &o.DiscountAmount, &o.PayableAmount); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *MySQLRepository) querier(q database.Querier) database.Querier {
	if q != nil {
		return q
	}
	return r.db
}
