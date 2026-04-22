package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
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

func (r *MySQLRepository) FindSaleableNode(ctx context.Context, q database.Querier, skuID, regionID uint64) (ResourceNode, error) {
	querier := r.querier(q)

	query := `
SELECT rn.id, rn.node_no, rn.region_id, rn.node_name, rn.total_instances, rn.used_instances, rn.reserved_instances, rn.status, rn.created_at, rn.updated_at
FROM sku_region_node_bindings AS srb
INNER JOIN resource_nodes AS rn ON rn.id = srb.node_id
INNER JOIN product_skus AS sku ON sku.id = srb.sku_id
INNER JOIN products AS p ON p.id = sku.product_id
INNER JOIN regions AS rg ON rg.id = srb.region_id
WHERE srb.sku_id = ?
  AND srb.region_id = ?
  AND srb.sale_status = 'saleable'
  AND sku.status = 'active'
  AND p.status = 'active'
  AND rg.status = 'active'
  AND rn.status = 'active'
  AND rn.total_instances > rn.used_instances + rn.reserved_instances
ORDER BY rn.reserved_instances ASC, rn.used_instances ASC, rn.id ASC
LIMIT 1`
	if _, ok := q.(*sql.Tx); ok {
		query += " FOR UPDATE"
	}

	var node ResourceNode
	err := querier.QueryRowContext(ctx, query, skuID, regionID).Scan(
		&node.ID,
		&node.NodeNo,
		&node.RegionID,
		&node.NodeName,
		&node.TotalInstances,
		&node.UsedInstances,
		&node.ReservedInstances,
		&node.Status,
		&node.CreatedAt,
		&node.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return ResourceNode{}, errorsx.ErrConflict
	}
	if err != nil {
		return ResourceNode{}, err
	}

	return node, nil
}

func (r *MySQLRepository) CreateReservation(ctx context.Context, q database.Querier, nodeID, userID, skuID, regionID uint64, expiresAt time.Time) (Reservation, error) {
	querier := r.querier(q)
	now := r.now().UTC()

	result, err := querier.ExecContext(ctx, `
UPDATE resource_nodes
SET reserved_instances = reserved_instances + 1,
    updated_at = ?
WHERE id = ?
  AND status = 'active'
  AND total_instances > used_instances + reserved_instances
`, now, nodeID)
	if err != nil {
		return Reservation{}, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return Reservation{}, err
	}
	if rows == 0 {
		return Reservation{}, errorsx.ErrConflict
	}

	reservationNo := fmt.Sprintf("R%d", now.UnixNano())
	result, err = querier.ExecContext(ctx, `
INSERT INTO resource_reservations (
	reservation_no, user_id, sku_id, region_id, node_id, status, expires_at, created_at, updated_at
)
VALUES (?, ?, ?, ?, ?, 'reserved', ?, ?, ?)
`, reservationNo, userID, skuID, regionID, nodeID, expiresAt.UTC(), now, now)
	if err != nil {
		return Reservation{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Reservation{}, err
	}

	return Reservation{
		ID:            uint64(id),
		ReservationNo: reservationNo,
		UserID:        userID,
		SKUID:         skuID,
		RegionID:      regionID,
		NodeID:        nodeID,
		Status:        "reserved",
		ExpiresAt:     expiresAt.UTC(),
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (r *MySQLRepository) ListSaleableProducts(ctx context.Context) ([]SaleableProduct, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT DISTINCT
	p.id, p.product_no, p.product_name, p.product_type, p.status, p.created_at, p.updated_at,
	sku.id, sku.sku_no, sku.product_id, sku.sku_name, sku.cpu_cores, sku.memory_mb, sku.disk_gb, sku.bandwidth_mbps, sku.status, sku.created_at, sku.updated_at
FROM products AS p
INNER JOIN product_skus AS sku ON sku.product_id = p.id
INNER JOIN sku_region_node_bindings AS srb ON srb.sku_id = sku.id
INNER JOIN regions AS rg ON rg.id = srb.region_id
INNER JOIN resource_nodes AS rn ON rn.id = srb.node_id
WHERE p.status = 'active'
  AND sku.status = 'active'
  AND srb.sale_status = 'saleable'
  AND rg.status = 'active'
  AND rn.status = 'active'
  AND rn.total_instances > rn.used_instances + rn.reserved_instances
ORDER BY p.id ASC, sku.id ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]SaleableProduct, 0)
	indexByProductID := make(map[uint64]int)
	for rows.Next() {
		var product Product
		var sku SKU
		if err := rows.Scan(
			&product.ID,
			&product.ProductNo,
			&product.ProductName,
			&product.ProductType,
			&product.Status,
			&product.CreatedAt,
			&product.UpdatedAt,
			&sku.ID,
			&sku.SKUNo,
			&sku.ProductID,
			&sku.SKUName,
			&sku.CPUCores,
			&sku.MemoryMB,
			&sku.DiskGB,
			&sku.BandwidthMbps,
			&sku.Status,
			&sku.CreatedAt,
			&sku.UpdatedAt,
		); err != nil {
			return nil, err
		}

		idx, ok := indexByProductID[product.ID]
		if !ok {
			products = append(products, SaleableProduct{Product: product})
			idx = len(products) - 1
			indexByProductID[product.ID] = idx
		}
		products[idx].SKUs = append(products[idx].SKUs, sku)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *MySQLRepository) CreateSKU(ctx context.Context, productID uint64, in CreateSKUInput) (SKU, error) {
	now := r.now().UTC()
	skuNo := fmt.Sprintf("SKU%d", now.UnixNano())
	status := in.Status
	if status == "" {
		status = "draft"
	}

	result, err := r.db.ExecContext(ctx, `
INSERT INTO product_skus (
	sku_no, product_id, sku_name, cpu_cores, memory_mb, disk_gb, bandwidth_mbps, status, created_at, updated_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, skuNo, productID, in.SKUName, in.CPUCores, in.MemoryMB, in.DiskGB, in.BandwidthMbps, status, now, now)
	if err != nil {
		return SKU{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return SKU{}, err
	}

	return SKU{
		ID:            uint64(id),
		SKUNo:         skuNo,
		ProductID:     productID,
		SKUName:       in.SKUName,
		CPUCores:      in.CPUCores,
		MemoryMB:      in.MemoryMB,
		DiskGB:        in.DiskGB,
		BandwidthMbps: in.BandwidthMbps,
		Status:        status,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (r *MySQLRepository) querier(q database.Querier) database.Querier {
	if q != nil {
		return q
	}
	return r.db
}
