package payment

import (
	"context"
	"fmt"
	"time"

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

type TxRepo interface {
	HasSuccessfulCallback(ctx context.Context, paymentOrderNo string) bool
	InsertCallbackLog(ctx context.Context, paymentOrderNo string, rawPayload []byte) error
	MarkSuccessAndMoveOrderPaid(ctx context.Context, paymentOrderNo string) (uint64, error)
	InsertPendingProvisionTask(ctx context.Context, orderID uint64) error
}

type Repo interface {
	CreatePendingPayment(ctx context.Context, orderID uint64, payableAmount int64) (PaymentOrder, error)
	WithTx(ctx context.Context, fn func(TxRepo) error) error
}

type Service struct {
	repo Repo
	now  func() time.Time
}

func NewService(repo Repo) *Service {
	return &Service{
		repo: repo,
		now:  time.Now,
	}
}

func (s *Service) CreatePendingPayment(ctx context.Context, orderID uint64, payableAmount int64) (PaymentOrder, error) {
	if orderID == 0 || payableAmount <= 0 {
		return PaymentOrder{}, errorsx.ErrBadRequest
	}

	if s.repo != nil {
		return s.repo.CreatePendingPayment(ctx, orderID, payableAmount)
	}

	return PaymentOrder{
		PaymentOrderNo: fmt.Sprintf("P%d", s.now().UnixNano()),
		OrderID:        orderID,
		PayStatus:      "pending",
		PayableAmount:  payableAmount,
	}, nil
}

func (s *Service) MarkPaymentSuccess(ctx context.Context, paymentOrderNo string, rawPayload []byte) error {
	if paymentOrderNo == "" {
		return errorsx.ErrBadRequest
	}
	if s.repo == nil {
		return errorsx.ErrInternal
	}

	return s.repo.WithTx(ctx, func(txRepo TxRepo) error {
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
