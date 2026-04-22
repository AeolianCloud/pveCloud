package payment

import (
	"context"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
)

type PaymentOrder struct {
	ID             uint64     `json:"id"`
	PaymentOrderNo string     `json:"payment_order_no"`
	OrderID        uint64     `json:"order_id"`
	PayStatus      string     `json:"pay_status"`
	PayableAmount  int64      `json:"payable_amount"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
}

type CallbackTxRepository interface {
	HasSuccessfulCallback(ctx context.Context, paymentOrderNo string) bool
	InsertCallbackLog(ctx context.Context, paymentOrderNo string, rawPayload []byte) error
	MarkSuccessAndMoveOrderPaid(ctx context.Context, paymentOrderNo string) (uint64, error)
	InsertPendingProvisionTask(ctx context.Context, orderID uint64) error
}

type CallbackStore interface {
	WithTx(ctx context.Context, fn func(CallbackTxRepository) error) error
}

type Service struct {
	repo          Repository
	callbackStore CallbackStore
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func NewServiceWithCallbackStore(repo Repository, callbackStore CallbackStore) *Service {
	return &Service{
		repo:          repo,
		callbackStore: callbackStore,
	}
}

func (s *Service) GetPaymentOrder(ctx context.Context, paymentOrderNo string) (PaymentOrder, error) {
	if paymentOrderNo == "" {
		return PaymentOrder{}, errorsx.ErrBadRequest
	}
	if s.repo == nil {
		return PaymentOrder{}, errorsx.ErrInternal
	}
	return s.repo.GetByPaymentOrderNo(ctx, paymentOrderNo)
}

func (s *Service) CreatePendingPayment(ctx context.Context, q database.Querier, orderID uint64, payableAmount int64) (PaymentOrder, error) {
	if orderID == 0 || payableAmount <= 0 {
		return PaymentOrder{}, errorsx.ErrBadRequest
	}
	if s.repo == nil {
		return PaymentOrder{}, errorsx.ErrInternal
	}

	return s.repo.CreatePendingPayment(ctx, q, orderID, payableAmount)
}

func (s *Service) MarkPaymentSuccess(ctx context.Context, paymentOrderNo string, rawPayload []byte) error {
	if paymentOrderNo == "" {
		return errorsx.ErrBadRequest
	}
	if s.callbackStore == nil {
		return errorsx.ErrInternal
	}

	return s.callbackStore.WithTx(ctx, func(txRepo CallbackTxRepository) error {
		if txRepo.HasSuccessfulCallback(ctx, paymentOrderNo) {
			return nil
		}
		if err := txRepo.InsertCallbackLog(ctx, paymentOrderNo, rawPayload); err != nil {
			return err
		}
		orderID, err := txRepo.MarkSuccessAndMoveOrderPaid(ctx, paymentOrderNo)
		if err != nil {
			return err
		}
		return txRepo.InsertPendingProvisionTask(ctx, orderID)
	})
}
