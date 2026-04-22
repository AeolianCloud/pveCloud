package order

import (
	"context"

	"github.com/AeolianCloud/pveCloud/server/internal/catalog"
	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/payment"
)

type Repo interface {
	CreateOrder(ctx context.Context, in CreateOrderParams) (Order, error)
	BindReservation(ctx context.Context, orderID, reservationID uint64) error
}

type BillingService interface {
	Quote(ctx context.Context, skuID uint64, cycle string) (BillingQuote, error)
}

type PaymentService interface {
	CreatePendingPayment(ctx context.Context, orderID uint64, payableAmount int64) (payment.PaymentOrder, error)
}

type CatalogService interface {
	ReserveCapacity(ctx context.Context, in catalog.ReserveInput) (catalog.Reservation, error)
}

type Service struct {
	repo       Repo
	billingSvc BillingService
	paymentSvc PaymentService
	catalogSvc CatalogService
}

func NewService(repo Repo, billingSvc BillingService, paymentSvc PaymentService, catalogSvc CatalogService) *Service {
	return &Service{
		repo:       repo,
		billingSvc: billingSvc,
		paymentSvc: paymentSvc,
		catalogSvc: catalogSvc,
	}
}

func (s *Service) CreateOrder(ctx context.Context, in CreateInput) (CreateResult, error) {
	if in.UserID == 0 || in.SKUID == 0 || in.RegionID == 0 || in.Cycle == "" {
		return CreateResult{}, errorsx.ErrBadRequest
	}

	price, err := s.billingSvc.Quote(ctx, in.SKUID, in.Cycle)
	if err != nil {
		return CreateResult{}, err
	}

	reservation, err := s.catalogSvc.ReserveCapacity(ctx, catalog.ReserveInput{
		UserID:   in.UserID,
		SKUID:    in.SKUID,
		RegionID: in.RegionID,
	})
	if err != nil {
		return CreateResult{}, err
	}

	orderRow, err := s.repo.CreateOrder(ctx, CreateOrderParams{
		UserID:         in.UserID,
		SKUID:          in.SKUID,
		RegionID:       in.RegionID,
		Cycle:          in.Cycle,
		OriginalAmount: price.OriginalAmount,
		DiscountAmount: price.DiscountAmount,
		PayableAmount:  price.PayableAmount,
	})
	if err != nil {
		return CreateResult{}, err
	}

	if err := s.repo.BindReservation(ctx, orderRow.ID, reservation.ID); err != nil {
		return CreateResult{}, err
	}
	orderRow.ReservationID = reservation.ID

	paymentRow, err := s.paymentSvc.CreatePendingPayment(ctx, orderRow.ID, price.PayableAmount)
	if err != nil {
		return CreateResult{}, err
	}

	return CreateResult{
		Order:        orderRow,
		PaymentOrder: paymentRow,
	}, nil
}
