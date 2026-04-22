package instance_test

import (
	"context"
	"testing"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/catalog"
	"github.com/AeolianCloud/pveCloud/server/internal/instance"
	"github.com/AeolianCloud/pveCloud/server/internal/resource"
)

type fakeInstanceRepo struct{}

func (f *fakeInstanceRepo) LoadPaidOrderForProvision(ctx context.Context, orderID uint64) (instance.PaidOrder, catalog.Reservation, error) {
	return instance.PaidOrder{
			ID:            orderID,
			UserID:        1001,
			OrderNo:       "O5001",
			SKUID:         2001,
			RegionID:      3001,
			Cycle:         "month",
			PayableAmount: 10000,
		}, catalog.Reservation{
			ID:        6001,
			UserID:    1001,
			SKUID:     2001,
			RegionID:  3001,
			NodeID:    4001,
			Status:    "consumed",
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}, nil
}

func (f *fakeInstanceRepo) CreateInstanceAndActivateOrder(ctx context.Context, orderRow instance.PaidOrder, reservation catalog.Reservation, vmResp resource.CreateVMResponse) (instance.ProvisionResult, error) {
	return instance.ProvisionResult{
		Instance: instance.Instance{
			ID:          9001,
			InstanceNo:  "I9001",
			UserID:      orderRow.UserID,
			OrderID:     orderRow.ID,
			NodeID:      reservation.NodeID,
			Status:      "running",
			InstanceRef: vmResp.InstanceRef,
		},
		Service: instance.ServiceFact{
			InstanceID:           9001,
			CurrentPeriodStartAt: time.Now(),
			CurrentPeriodEndAt:   time.Now().Add(30 * 24 * time.Hour),
			BillingStatus:        "active",
		},
	}, nil
}

func (f *fakeInstanceRepo) ListByUser(ctx context.Context, userID uint64) ([]instance.Instance, error) {
	return nil, nil
}

func (f *fakeInstanceRepo) ListAll(ctx context.Context) ([]instance.Instance, error) {
	return nil, nil
}

type fakeVMClient struct{}

func (f *fakeVMClient) CreateVM(ctx context.Context, req resource.CreateVMRequest) (resource.CreateVMResponse, error) {
	return resource.CreateVMResponse{
		InstanceRef: "vm-9001",
		Status:      "running",
	}, nil
}

func (f *fakeVMClient) StartVM(ctx context.Context, instanceRef string) error { return nil }
func (f *fakeVMClient) StopVM(ctx context.Context, instanceRef string) error { return nil }
func (f *fakeVMClient) RebootVM(ctx context.Context, instanceRef string) error { return nil }
func (f *fakeVMClient) ReinstallVM(ctx context.Context, req resource.ReinstallVMRequest) error {
	return nil
}

type fakeAuditService struct {
	recorded bool
}

func (f *fakeAuditService) Record(ctx context.Context, event string, businessID uint64) error {
	f.recorded = true
	return nil
}

type fakeNotificationService struct {
	sent bool
}

func (f *fakeNotificationService) SendProvisionSuccess(ctx context.Context, userID uint64, instanceNo string) error {
	f.sent = true
	return nil
}

func TestProvisionFromPaidOrderCreatesInstanceAndServiceFact(t *testing.T) {
	svc := instance.NewService(&fakeInstanceRepo{}, &fakeVMClient{}, &fakeAuditService{}, &fakeNotificationService{})
	result, err := svc.HandleCreateInstanceTask(context.Background(), 5001)
	if err != nil {
		t.Fatalf("handle create instance task: %v", err)
	}
	if result.Instance.Status != "running" {
		t.Fatalf("expected instance status running, got %s", result.Instance.Status)
	}
	if result.Service.CurrentPeriodEndAt.IsZero() {
		t.Fatalf("expected current period end at to be set")
	}
}
