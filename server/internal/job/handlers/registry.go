package handlers

import "github.com/AeolianCloud/pveCloud/server/internal/job"

type Registrar interface {
	RegisterHandlers(job.Registry)
}

type Func func(job.Registry)

func (f Func) RegisterHandlers(registry job.Registry) {
	if f == nil {
		return
	}
	f(registry)
}

func NewRegistry(registrars ...Registrar) job.Registry {
	registry := job.NewRegistry()
	registerBuiltins(registry)
	for _, registrar := range registrars {
		if registrar == nil {
			continue
		}
		registrar.RegisterHandlers(registry)
	}
	return registry
}

func registerBuiltins(registry job.Registry) {
	registerInstanceCreate(registry)
	registerInstanceRenew(registry)
	registerOrderExpire(registry)
	registerPaymentCheck(registry)
	registerInstanceStatusSync(registry)
}
