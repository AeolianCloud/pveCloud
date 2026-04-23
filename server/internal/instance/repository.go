package instance

import (
	"context"

	"github.com/AeolianCloud/pveCloud/server/internal/catalog"
	"github.com/AeolianCloud/pveCloud/server/internal/resource"
)

type Repo interface {
	FindProvisionResultByOrder(ctx context.Context, orderID uint64) (ProvisionResult, bool, error)
	LoadPaidOrderForProvision(ctx context.Context, orderID uint64) (PaidOrder, catalog.Reservation, error)
	CreateInstanceAndActivateOrder(ctx context.Context, orderRow PaidOrder, reservation catalog.Reservation, vmResp resource.CreateVMResponse) (ProvisionResult, error)
	ListByUser(ctx context.Context, userID uint64) ([]Instance, error)
	ListAll(ctx context.Context) ([]Instance, error)
}
