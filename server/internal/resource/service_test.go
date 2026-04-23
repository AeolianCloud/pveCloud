package resource_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/resource"
)

type fakeClient struct {
	lastCreate resource.CreateVMRequest
}

func (f *fakeClient) CreateVM(ctx context.Context, req resource.CreateVMRequest) (resource.CreateVMResponse, error) {
	f.lastCreate = req
	return resource.CreateVMResponse{InstanceRef: "vm-1", Status: "running"}, nil
}

func (f *fakeClient) StartVM(ctx context.Context, instanceRef string) error  { return nil }
func (f *fakeClient) StopVM(ctx context.Context, instanceRef string) error   { return nil }
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

func TestMockClientCreateVMReturnsStableResponse(t *testing.T) {
	client := resource.NewMockClient()

	resp, err := client.CreateVM(context.Background(), resource.CreateVMRequest{
		OrderID:  5001,
		NodeID:   4001,
		UserID:   1001,
		Hostname: "inst-5001",
	})
	if err != nil {
		t.Fatalf("create vm: %v", err)
	}
	if resp.InstanceRef != "mock-vm-5001" {
		t.Fatalf("expected mock-vm-5001, got %s", resp.InstanceRef)
	}
	if resp.Status != "running" {
		t.Fatalf("expected running status, got %s", resp.Status)
	}
}

func TestRetryableWrapsProviderError(t *testing.T) {
	err := resource.Retryable(errors.New("temporary"), 2*time.Minute)

	var providerErr *resource.ProviderError
	if !errors.As(err, &providerErr) {
		t.Fatalf("expected provider error wrapper")
	}
	if !providerErr.Retryable || providerErr.Delay != 2*time.Minute {
		t.Fatalf("unexpected provider error: %+v", providerErr)
	}
}
