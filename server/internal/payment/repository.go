package payment

import (
	"context"

	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
)

type Repository interface {
	CreatePendingPayment(ctx context.Context, q database.Querier, orderID uint64, payableAmount int64) (PaymentOrder, error)
	GetByPaymentOrderNo(ctx context.Context, paymentOrderNo string) (PaymentOrder, error)
}
