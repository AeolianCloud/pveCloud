package order

import (
	"context"

	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
)

type Repository interface {
	CreateOrder(ctx context.Context, q database.Querier, in CreateOrderParams) (Order, error)
	BindReservation(ctx context.Context, q database.Querier, orderID, reservationID uint64) error
	GetOrderByID(ctx context.Context, orderID uint64) (Order, error)
	ListOrdersByUser(ctx context.Context, userID uint64) ([]Order, error)
}
