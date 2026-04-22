package payment

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
)

// MySQLCallbackStore implements CallbackStore using a MySQL database.
type MySQLCallbackStore struct {
	db *sql.DB
}

// NewMySQLCallbackStore creates a new MySQLCallbackStore.
func NewMySQLCallbackStore(db *sql.DB) *MySQLCallbackStore {
	return &MySQLCallbackStore{db: db}
}

// WithTx wraps fn in a database transaction, providing a CallbackTxRepository
// backed by that transaction.
func (s *MySQLCallbackStore) WithTx(ctx context.Context, fn func(CallbackTxRepository) error) error {
	return database.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		txRepo := &mysqlCallbackTxRepo{tx: tx}
		return fn(txRepo)
	})
}

// mysqlCallbackTxRepo implements CallbackTxRepository backed by a *sql.Tx.
type mysqlCallbackTxRepo struct {
	tx *sql.Tx
}

// HasSuccessfulCallback checks whether a successful callback log already exists
// for the given payment order number.
func (r *mysqlCallbackTxRepo) HasSuccessfulCallback(ctx context.Context, paymentOrderNo string) bool {
	var count int
	err := r.tx.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM payment_callback_logs
WHERE payment_order_no = ? AND callback_status = 'success'
`, paymentOrderNo).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// InsertCallbackLog inserts a callback log record with status "success".
func (r *mysqlCallbackTxRepo) InsertCallbackLog(ctx context.Context, paymentOrderNo string, rawPayload []byte) error {
	_, err := r.tx.ExecContext(ctx, `
INSERT INTO payment_callback_logs (payment_order_no, callback_status, raw_payload, created_at)
VALUES (?, 'success', ?, NOW(3))
`, paymentOrderNo, rawPayload)
	return err
}

// MarkSuccessAndMoveOrderPaid marks the payment order as paid, updates the
// corresponding order status to "paid", and returns the order_id.
func (r *mysqlCallbackTxRepo) MarkSuccessAndMoveOrderPaid(ctx context.Context, paymentOrderNo string) (uint64, error) {
	_, err := r.tx.ExecContext(ctx, `
UPDATE payment_orders
SET pay_status = 'success', paid_at = NOW(3)
WHERE payment_order_no = ? AND pay_status = 'pending'
`, paymentOrderNo)
	if err != nil {
		return 0, err
	}

	var orderID uint64
	err = r.tx.QueryRowContext(ctx, `
SELECT order_id FROM payment_orders WHERE payment_order_no = ?
`, paymentOrderNo).Scan(&orderID)
	if err != nil {
		return 0, err
	}

	_, err = r.tx.ExecContext(ctx, `
UPDATE orders
SET order_status = 'paid', paid_at = NOW(3)
WHERE id = ? AND order_status = 'pending_payment'
`, orderID)
	if err != nil {
		return 0, err
	}

	return orderID, nil
}

// InsertPendingProvisionTask inserts a pending async task for instance provisioning.
// It uses ON DUPLICATE KEY UPDATE to handle uniqueness on (task_type, business_type, business_id).
func (r *mysqlCallbackTxRepo) InsertPendingProvisionTask(ctx context.Context, orderID uint64) error {
	now := time.Now().UTC()
	taskNo := fmt.Sprintf("TASK-%d-%d", now.UnixNano(), orderID)
	_, err := r.tx.ExecContext(ctx, `
INSERT INTO async_tasks (task_no, task_type, business_type, business_id, status, created_at, updated_at, next_run_at)
VALUES (?, 'create_instance', 'order', ?, 'pending', ?, ?, ?)
ON DUPLICATE KEY UPDATE updated_at = VALUES(updated_at)
`, taskNo, orderID, now, now, now)
	return err
}
