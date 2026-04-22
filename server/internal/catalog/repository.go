package catalog

import (
	"context"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
)

type Repository interface {
	FindSaleableNode(ctx context.Context, q database.Querier, skuID, regionID uint64) (ResourceNode, error)
	CreateReservation(ctx context.Context, q database.Querier, nodeID, userID, skuID, regionID uint64, expiresAt time.Time) (Reservation, error)
	ListSaleableProducts(ctx context.Context) ([]SaleableProduct, error)
	CreateSKU(ctx context.Context, productID uint64, in CreateSKUInput) (SKU, error)
}
