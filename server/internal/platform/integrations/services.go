package integrations

import (
	"github.com/AeolianCloud/pveCloud/server/internal/platform/integrations/notify"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/integrations/payment"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/integrations/pve"
)

type Services struct {
	PVE     pve.Adapter
	Payment payment.Adapter
	Notify  notify.Adapter
}

func NewServices() Services {
	return Services{
		PVE:     pve.NewStub(),
		Payment: payment.NewStub(),
		Notify:  notify.NewStub(),
	}
}
