package resource

import "context"

type CreateVMRequest struct {
	OrderID  uint64 `json:"order_id"`
	NodeID   uint64 `json:"node_id"`
	UserID   uint64 `json:"user_id"`
	Hostname string `json:"hostname"`
}

type CreateVMResponse struct {
	InstanceRef string `json:"instance_ref"`
	Status      string `json:"status"`
}

type ReinstallVMRequest struct {
	InstanceRef string `json:"instance_ref"`
	ImageID     uint64 `json:"image_id"`
}

type VMClient interface {
	CreateVM(ctx context.Context, req CreateVMRequest) (CreateVMResponse, error)
	StartVM(ctx context.Context, instanceRef string) error
	StopVM(ctx context.Context, instanceRef string) error
	RebootVM(ctx context.Context, instanceRef string) error
	ReinstallVM(ctx context.Context, req ReinstallVMRequest) error
}
