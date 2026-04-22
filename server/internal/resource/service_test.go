package resource_test

import (
	"context"
	"testing"

	"github.com/AeolianCloud/pveCloud/server/internal/resource"
)

type fakeClient struct {
	lastCreate resource.CreateVMRequest
}

func (f *fakeClient) CreateVM(ctx context.Context, req resource.CreateVMRequest) (resource.CreateVMResponse, error) {
	f.lastCreate = req
	return resource.CreateVMResponse{InstanceRef: "vm-1", Status: "running"}, nil
}

func (f *fakeClient) StartVM(ctx context.Context, instanceRef string) error { return nil }
func (f *fakeClient) StopVM(ctx context.Context, instanceRef string) error { return nil }
func (f *fakeClient) RebootVM(ctx context.Context, instanceRef string) error { return nil }
func (f *fakeClient) ReinstallVM(ctx context.Context, req resource.ReinstallVMRequest) error {
	return nil
}

func TestServiceCreateVMDelegatesToClient(t *testing.T) {
	client := &fakeClient{}
	svc := resource.NewService(client)

	resp, err := svc.CreateVM(context.Background(), resource.CreateVMRequest{
		OrderID:  5001,
		NodeID:   4001,
		UserID:   1001,
		Hostname: "vm-1",
	})
	if err != nil {
		t.Fatalf("create vm: %v", err)
	}
	if resp.InstanceRef != "vm-1" {
		t.Fatalf("expected instance ref vm-1, got %s", resp.InstanceRef)
	}
	if client.lastCreate.Hostname != "vm-1" {
		t.Fatalf("expected client create request to be captured")
	}
}
