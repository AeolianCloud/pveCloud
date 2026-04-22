package catalog

import (
	"context"
	"time"

	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
)

type Service struct {
	repo Repository
	ttl  time.Duration
	now  func() time.Time
}

func NewService(repo Repository, ttl time.Duration) *Service {
	return &Service{
		repo: repo,
		ttl:  ttl,
		now:  time.Now,
	}
}

func (s *Service) ReserveCapacity(ctx context.Context, in ReserveInput) (Reservation, error) {
	if in.UserID == 0 || in.SKUID == 0 || in.RegionID == 0 {
		return Reservation{}, errorsx.ErrBadRequest
	}

	node, err := s.repo.FindSaleableNode(ctx, in.SKUID, in.RegionID)
	if err != nil {
		return Reservation{}, err
	}

	reservation, err := s.repo.CreateReservation(ctx, node.ID, in.UserID, in.SKUID, s.now().Add(s.ttl))
	if err != nil {
		return Reservation{}, err
	}
	if reservation.RegionID == 0 {
		reservation.RegionID = in.RegionID
	}
	if reservation.NodeID == 0 {
		reservation.NodeID = node.ID
	}
	return reservation, nil
}

func (s *Service) ListSaleableProducts(ctx context.Context) ([]SaleableProduct, error) {
	return s.repo.ListSaleableProducts(ctx)
}

func (s *Service) CreateSKU(ctx context.Context, productID uint64, in CreateSKUInput) (SKU, error) {
	if productID == 0 || in.SKUName == "" {
		return SKU{}, errorsx.ErrBadRequest
	}
	return s.repo.CreateSKU(ctx, productID, in)
}
