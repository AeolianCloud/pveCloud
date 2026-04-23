package resource

import (
	"context"
	"time"
)

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

type ProviderError struct {
	Err       error
	Retryable bool
	Delay     time.Duration
}

func (e *ProviderError) Error() string {
	if e == nil || e.Err == nil {
		return "provider error"
	}
	return e.Err.Error()
}

func (e *ProviderError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func Retryable(err error, delay time.Duration) error {
	if err == nil {
		return nil
	}
	if delay <= 0 {
		delay = time.Minute
	}
	return &ProviderError{
		Err:       err,
		Retryable: true,
		Delay:     delay,
	}
}

func Terminal(err error) error {
	if err == nil {
		return nil
	}
	return &ProviderError{Err: err}
}
