package instance

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/catalog"
	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
	"github.com/AeolianCloud/pveCloud/server/internal/resource"
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

func (r *MySQLRepository) FindProvisionResultByOrder(ctx context.Context, orderID uint64) (ProvisionResult, bool, error) {
	return r.findProvisionResultByOrder(ctx, r.db, orderID)
}

func (r *MySQLRepository) LoadPaidOrderForProvision(ctx context.Context, orderID uint64) (PaidOrder, catalog.Reservation, error) {
	var orderRow PaidOrder
	var reservation catalog.Reservation

	err := r.db.QueryRowContext(ctx, `
SELECT o.id, o.order_no, o.user_id, o.sku_id, o.region_id, o.cycle_unit, o.payable_amount,
	rr.id, rr.reservation_no, rr.user_id, rr.sku_id, rr.region_id, rr.node_id, rr.status, rr.expires_at, rr.created_at, rr.updated_at
FROM orders AS o
INNER JOIN resource_reservations AS rr ON rr.id = o.reservation_id
WHERE o.id = ? AND o.order_status IN ('paid', 'provisioning', 'active') AND rr.status IN ('reserved', 'consumed')
`, orderID).Scan(
		&orderRow.ID,
		&orderRow.OrderNo,
		&orderRow.UserID,
		&orderRow.SKUID,
		&orderRow.RegionID,
		&orderRow.Cycle,
		&orderRow.PayableAmount,
		&reservation.ID,
		&reservation.ReservationNo,
		&reservation.UserID,
		&reservation.SKUID,
		&reservation.RegionID,
		&reservation.NodeID,
		&reservation.Status,
		&reservation.ExpiresAt,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
	)
	if err != nil {
		return PaidOrder{}, catalog.Reservation{}, err
	}

	return orderRow, reservation, nil
}

func (r *MySQLRepository) CreateInstanceAndActivateOrder(ctx context.Context, orderRow PaidOrder, reservation catalog.Reservation, vmResp resource.CreateVMResponse) (ProvisionResult, error) {
	var result ProvisionResult

	err := database.WithTx(ctx, r.db, func(tx *sql.Tx) error {
		existing, found, err := r.findProvisionResultByOrder(ctx, tx, orderRow.ID)
		if err != nil {
			return err
		}
		if found {
			result = existing
			return nil
		}

		now := r.now().UTC()

		if _, err := tx.ExecContext(ctx, `
UPDATE orders
SET order_status = 'provisioning', updated_at = ?
WHERE id = ? AND order_status IN ('paid', 'provisioning', 'active')
`, now, orderRow.ID); err != nil {
			return err
		}

		if reservation.Status == "reserved" {
			if _, err := tx.ExecContext(ctx, `
UPDATE resource_reservations
SET status = 'consumed', updated_at = ?
WHERE id = ? AND status = 'reserved'
`, now, reservation.ID); err != nil {
				return err
			}
		}

		if _, err := tx.ExecContext(ctx, `
UPDATE resource_nodes
SET used_instances = used_instances + 1,
	reserved_instances = CASE WHEN reserved_instances > 0 THEN reserved_instances - 1 ELSE 0 END,
	updated_at = ?
WHERE id = ?
`, now, reservation.NodeID); err != nil {
			return err
		}

		instanceNo := fmt.Sprintf("I%d", now.UnixNano())
		instanceStatus := vmResp.Status
		if instanceStatus == "" {
			instanceStatus = "running"
		}

		row, err := tx.ExecContext(ctx, `
INSERT INTO instances (
	instance_no, user_id, order_id, node_id, instance_status, instance_ref, created_at, updated_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`, instanceNo, orderRow.UserID, orderRow.ID, reservation.NodeID, instanceStatus, vmResp.InstanceRef, now, now)
		if err != nil {
			return err
		}

		instanceID, err := row.LastInsertId()
		if err != nil {
			return err
		}

		startAt := now
		endAt := calculatePeriodEnd(orderRow.Cycle, startAt)
		serviceRow, err := tx.ExecContext(ctx, `
INSERT INTO instance_services (
	instance_id, current_period_start_at, current_period_end_at, billing_status, created_at, updated_at
)
VALUES (?, ?, ?, 'active', ?, ?)
`, instanceID, startAt, endAt, now, now)
		if err != nil {
			return err
		}

		serviceID, err := serviceRow.LastInsertId()
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `
UPDATE orders
SET order_status = 'active', updated_at = ?
WHERE id = ?
`, now, orderRow.ID); err != nil {
			return err
		}

		result = ProvisionResult{
			Instance: Instance{
				ID:          uint64(instanceID),
				InstanceNo:  instanceNo,
				UserID:      orderRow.UserID,
				OrderID:     orderRow.ID,
				NodeID:      reservation.NodeID,
				Status:      instanceStatus,
				InstanceRef: vmResp.InstanceRef,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			Service: ServiceFact{
				ID:                   uint64(serviceID),
				InstanceID:           uint64(instanceID),
				CurrentPeriodStartAt: startAt,
				CurrentPeriodEndAt:   endAt,
				BillingStatus:        "active",
				CreatedAt:            now,
				UpdatedAt:            now,
			},
		}
		return nil
	})
	if err != nil {
		return ProvisionResult{}, err
	}

	return result, nil
}

func (r *MySQLRepository) findProvisionResultByOrder(ctx context.Context, q database.Querier, orderID uint64) (ProvisionResult, bool, error) {
	var result ProvisionResult

	err := q.QueryRowContext(ctx, `
SELECT i.id, i.instance_no, i.user_id, i.order_id, i.node_id, i.instance_status, i.instance_ref, i.created_at, i.updated_at,
	s.id, s.instance_id, s.current_period_start_at, s.current_period_end_at, s.billing_status, s.created_at, s.updated_at
FROM instances AS i
INNER JOIN instance_services AS s ON s.instance_id = i.id
WHERE i.order_id = ?
ORDER BY i.id DESC
LIMIT 1
`, orderID).Scan(
		&result.Instance.ID,
		&result.Instance.InstanceNo,
		&result.Instance.UserID,
		&result.Instance.OrderID,
		&result.Instance.NodeID,
		&result.Instance.Status,
		&result.Instance.InstanceRef,
		&result.Instance.CreatedAt,
		&result.Instance.UpdatedAt,
		&result.Service.ID,
		&result.Service.InstanceID,
		&result.Service.CurrentPeriodStartAt,
		&result.Service.CurrentPeriodEndAt,
		&result.Service.BillingStatus,
		&result.Service.CreatedAt,
		&result.Service.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return ProvisionResult{}, false, nil
		}
		return ProvisionResult{}, false, err
	}

	return result, true, nil
}

func (r *MySQLRepository) ListByUser(ctx context.Context, userID uint64) ([]Instance, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, instance_no, user_id, order_id, node_id, instance_status, instance_ref, created_at, updated_at
FROM instances
WHERE user_id = ?
ORDER BY created_at DESC, id DESC
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanInstances(rows)
}

func (r *MySQLRepository) ListAll(ctx context.Context) ([]Instance, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, instance_no, user_id, order_id, node_id, instance_status, instance_ref, created_at, updated_at
FROM instances
ORDER BY created_at DESC, id DESC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanInstances(rows)
}

func scanInstances(rows *sql.Rows) ([]Instance, error) {
	var items []Instance
	for rows.Next() {
		var item Instance
		if err := rows.Scan(
			&item.ID,
			&item.InstanceNo,
			&item.UserID,
			&item.OrderID,
			&item.NodeID,
			&item.Status,
			&item.InstanceRef,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func calculatePeriodEnd(cycle string, startAt time.Time) time.Time {
	switch cycle {
	case "year":
		return startAt.AddDate(1, 0, 0)
	case "quarter":
		return startAt.AddDate(0, 3, 0)
	default:
		return startAt.AddDate(0, 1, 0)
	}
}
