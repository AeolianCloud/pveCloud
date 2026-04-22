package catalog_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/catalog"
	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
)

type fakeCatalogRepo struct {
	node        catalog.ResourceNode
	reservation catalog.Reservation
}

func (f *fakeCatalogRepo) FindSaleableNode(ctx context.Context, q database.Querier, skuID, regionID uint64) (catalog.ResourceNode, error) {
	return f.node, nil
}

func (f *fakeCatalogRepo) CreateReservation(ctx context.Context, q database.Querier, nodeID, userID, skuID, regionID uint64, expiresAt time.Time) (catalog.Reservation, error) {
	f.reservation.NodeID = nodeID
	f.reservation.UserID = userID
	f.reservation.SKUID = skuID
	f.reservation.RegionID = regionID
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

	reservation, err := svc.ReserveCapacityWithQuerier(context.Background(), fakeQuerier{}, catalog.ReserveInput{
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

func TestReserveCapacityRejectsNonTransactionalUsage(t *testing.T) {
	repo := &fakeCatalogRepo{
		node: catalog.ResourceNode{ID: 4001, RegionID: 3001},
	}
	svc := catalog.NewService(repo, 15*time.Minute)

	_, err := svc.ReserveCapacity(context.Background(), catalog.ReserveInput{
		UserID:   1001,
		SKUID:    2001,
		RegionID: 3001,
	})
	if err != errorsx.ErrInternal {
		t.Fatalf("expected ErrInternal for non-transactional reservation, got %v", err)
	}
}

type fakeQuerier struct{}

func (fakeQuerier) ExecContext(context.Context, string, ...any) (sql.Result, error) { return nil, nil }
func (fakeQuerier) QueryContext(context.Context, string, ...any) (*sql.Rows, error) { return nil, nil }
func (fakeQuerier) QueryRowContext(context.Context, string, ...any) *sql.Row        { return nil }
