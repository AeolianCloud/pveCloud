package catalog_test

import (
	"context"
	"testing"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/catalog"
)

type fakeCatalogRepo struct {
	node        catalog.ResourceNode
	reservation catalog.Reservation
}

func (f *fakeCatalogRepo) FindSaleableNode(ctx context.Context, skuID, regionID uint64) (catalog.ResourceNode, error) {
	return f.node, nil
}

func (f *fakeCatalogRepo) CreateReservation(ctx context.Context, nodeID, userID, skuID uint64, expiresAt time.Time) (catalog.Reservation, error) {
	f.reservation.NodeID = nodeID
	f.reservation.UserID = userID
	f.reservation.SKUID = skuID
	f.reservation.ExpiresAt = expiresAt
	f.reservation.Status = "reserved"
	if f.reservation.RegionID == 0 {
		f.reservation.RegionID = 3001
	}
	return f.reservation, nil
}

func (f *fakeCatalogRepo) ListSaleableProducts(ctx context.Context) ([]catalog.SaleableProduct, error) {
	return nil, nil
}

func (f *fakeCatalogRepo) CreateSKU(ctx context.Context, productID uint64, in catalog.CreateSKUInput) (catalog.SKU, error) {
	return catalog.SKU{}, nil
}

func TestReserveCapacityCreatesExpiringReservation(t *testing.T) {
	repo := &fakeCatalogRepo{
		node: catalog.ResourceNode{ID: 4001, RegionID: 3001},
	}
	svc := catalog.NewService(repo, 15*time.Minute)

	reservation, err := svc.ReserveCapacity(context.Background(), catalog.ReserveInput{
		UserID:   1001,
		SKUID:    2001,
		RegionID: 3001,
	})
	if err != nil {
		t.Fatalf("reserve capacity: %v", err)
	}

	if reservation.Status != "reserved" {
		t.Fatalf("expected status reserved, got %s", reservation.Status)
	}
	if reservation.ExpiresAt.IsZero() {
		t.Fatalf("expected expires at to be set")
	}
	if reservation.NodeID != 4001 {
		t.Fatalf("expected node id 4001, got %d", reservation.NodeID)
	}
}
