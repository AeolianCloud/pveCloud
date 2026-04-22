package order

import (
	"context"
	"database/sql"

	"github.com/AeolianCloud/pveCloud/server/internal/billing"
	"github.com/AeolianCloud/pveCloud/server/internal/catalog"
	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/payment"
)

type BillingService interface {
	Quote(ctx context.Context, skuID uint64, cycle string) (BillingQuote, error)
	CreateRecord(ctx context.Context, q database.Querier, in billing.CreateRecordInput) (billing.Record, error)
}

type PaymentService interface {
	CreatePendingPayment(ctx context.Context, q database.Querier, orderID uint64, payableAmount int64) (payment.PaymentOrder, error)
}

type CatalogService interface {
	ReserveCapacityWithQuerier(ctx context.Context, q database.Querier, in catalog.ReserveInput) (catalog.Reservation, error)
}

type Service struct {
	db         *sql.DB
	repo       Repository
	billingSvc BillingService
	paymentSvc PaymentService
	catalogSvc CatalogService
}

func NewService(db *sql.DB, repo Repository, billingSvc BillingService, paymentSvc PaymentService, catalogSvc CatalogService) *Service {
	return &Service{
		db:         db,
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
	if s.db == nil {
		return CreateResult{}, errorsx.ErrInternal
	}

	price, err := s.billingSvc.Quote(ctx, in.SKUID, in.Cycle)
	if err != nil {
		return CreateResult{}, err
	}

	var result CreateResult

	err = database.WithTx(ctx, s.db, func(tx *sql.Tx) error {
		reservation, err := s.catalogSvc.ReserveCapacityWithQuerier(ctx, tx, catalog.ReserveInput{
			UserID:   in.UserID,
			SKUID:    in.SKUID,
			RegionID: in.RegionID,
		})
		if err != nil {
			return err
		}

		orderRow, err := s.repo.CreateOrder(ctx, tx, CreateOrderParams{
			UserID:         in.UserID,
			SKUID:          in.SKUID,
			RegionID:       in.RegionID,
			Cycle:          in.Cycle,
			OriginalAmount: price.OriginalAmount,
			DiscountAmount: price.DiscountAmount,
			PayableAmount:  price.PayableAmount,
		})
		if err != nil {
			return err
		}

		if _, err := s.billingSvc.CreateRecord(ctx, tx, billing.CreateRecordInput{
			OrderID:        orderRow.ID,
			BillingType:    "create",
			Cycle:          in.Cycle,
			OriginalAmount: price.OriginalAmount,
			DiscountAmount: price.DiscountAmount,
			PayableAmount:  price.PayableAmount,
		}); err != nil {
			return err
		}

		if err := s.repo.BindReservation(ctx, tx, orderRow.ID, reservation.ID); err != nil {
			return err
		}
		orderRow.ReservationID = reservation.ID

		paymentRow, err := s.paymentSvc.CreatePendingPayment(ctx, tx, orderRow.ID, price.PayableAmount)
		if err != nil {
			return err
		}

		result = CreateResult{
			Order:        orderRow,
			PaymentOrder: paymentRow,
		}
		return nil
	})
	if err != nil {
		return CreateResult{}, err
	}

	return result, nil
}
