package catalog

import (
	"context"
	"time"
)

type Repository interface {
	FindSaleableNode(ctx context.Context, skuID, regionID uint64) (ResourceNode, error)
	CreateReservation(ctx context.Context, nodeID, userID, skuID uint64, expiresAt time.Time) (Reservation, error)
	ListSaleableProducts(ctx context.Context) ([]SaleableProduct, error)
	CreateSKU(ctx context.Context, productID uint64, in CreateSKUInput) (SKU, error)
}
