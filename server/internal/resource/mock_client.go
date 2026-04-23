package resource

import (
	"context"
	"fmt"
)

type MockClient struct{}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (c *MockClient) CreateVM(ctx context.Context, req CreateVMRequest) (CreateVMResponse, error) {
	return CreateVMResponse{
		InstanceRef: fmt.Sprintf("mock-vm-%d", req.OrderID),
		Status:      "running",
	}, nil
}

func (c *MockClient) StartVM(ctx context.Context, instanceRef string) error {
	return nil
}

func (c *MockClient) StopVM(ctx context.Context, instanceRef string) error {
	return nil
}

func (c *MockClient) RebootVM(ctx context.Context, instanceRef string) error {
	return nil
}

func (c *MockClient) ReinstallVM(ctx context.Context, req ReinstallVMRequest) error {
	return nil
}
