package instance_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/catalog"
	"github.com/AeolianCloud/pveCloud/server/internal/instance"
	"github.com/AeolianCloud/pveCloud/server/internal/resource"
)

type fakeInstanceRepo struct {
	existing     *instance.ProvisionResult
	findCalled   bool
	createCalled bool
	loadCalled   bool
}

func (f *fakeInstanceRepo) FindProvisionResultByOrder(ctx context.Context, orderID uint64) (instance.ProvisionResult, bool, error) {
	f.findCalled = true
	if f.existing == nil {
		return instance.ProvisionResult{}, false, nil
	}
	return *f.existing, true, nil
}

func (f *fakeInstanceRepo) LoadPaidOrderForProvision(ctx context.Context, orderID uint64) (instance.PaidOrder, catalog.Reservation, error) {
	f.loadCalled = true
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
	f.createCalled = true
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

type fakeVMClient struct {
	err error
}

func (f *fakeVMClient) CreateVM(ctx context.Context, req resource.CreateVMRequest) (resource.CreateVMResponse, error) {
	if f.err != nil {
		return resource.CreateVMResponse{}, f.err
	}
	return resource.CreateVMResponse{
		InstanceRef: "vm-9001",
		Status:      "running",
	}, nil
}

func (f *fakeVMClient) StartVM(ctx context.Context, instanceRef string) error  { return nil }
func (f *fakeVMClient) StopVM(ctx context.Context, instanceRef string) error   { return nil }
func (f *fakeVMClient) RebootVM(ctx context.Context, instanceRef string) error { return nil }
func (f *fakeVMClient) ReinstallVM(ctx context.Context, req resource.ReinstallVMRequest) error {
	return nil
}

type fakeAuditService struct {
	recorded bool
	events   []string
}

func (f *fakeAuditService) Record(ctx context.Context, event string, businessID uint64) error {
	f.recorded = true
	f.events = append(f.events, event)
	return nil
}

type fakeNotificationService struct {
	sent         bool
	failureSent  bool
	failureOrder uint64
}

func (f *fakeNotificationService) SendProvisionSuccess(ctx context.Context, userID uint64, instanceNo string) error {
	f.sent = true
	return nil
}

func (f *fakeNotificationService) SendProvisionFailure(ctx context.Context, userID uint64, orderID uint64) error {
	f.failureSent = true
	f.failureOrder = orderID
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

func TestHandleCreateInstanceTaskReturnsExistingProvisionResult(t *testing.T) {
	repo := &fakeInstanceRepo{
		existing: &instance.ProvisionResult{
			Instance: instance.Instance{ID: 9001, InstanceNo: "I9001", Status: "running"},
			Service:  instance.ServiceFact{InstanceID: 9001, BillingStatus: "active"},
		},
	}
	vmClient := &fakeVMClient{}
	svc := instance.NewService(repo, vmClient, &fakeAuditService{}, &fakeNotificationService{})

	result, err := svc.HandleCreateInstanceTask(context.Background(), 5001)
	if err != nil {
		t.Fatalf("handle create instance task: %v", err)
	}
	if result.Instance.InstanceNo != "I9001" {
		t.Fatalf("expected existing instance I9001, got %+v", result.Instance)
	}
	if repo.loadCalled || repo.createCalled {
		t.Fatalf("expected existing result to short-circuit provisioning")
	}
}

func TestHandleCreateInstanceTaskRecordsAuditOnProviderFailure(t *testing.T) {
	repo := &fakeInstanceRepo{}
	auditSvc := &fakeAuditService{}
	notificationSvc := &fakeNotificationService{}
	svc := instance.NewService(
		repo,
		&fakeVMClient{err: resource.Retryable(errors.New("temporary provider failure"), time.Minute)},
		auditSvc,
		notificationSvc,
	)

	_, err := svc.HandleCreateInstanceTask(context.Background(), 5001)
	if err == nil {
		t.Fatalf("expected provider error")
	}
	var providerErr *resource.ProviderError
	if !errors.As(err, &providerErr) || !providerErr.Retryable {
		t.Fatalf("expected retryable provider error, got %v", err)
	}
	if !auditSvc.recorded || len(auditSvc.events) == 0 || auditSvc.events[0] != "order.provision.failed" {
		t.Fatalf("expected failed provision audit event, got %+v", auditSvc.events)
	}
	if !notificationSvc.failureSent || notificationSvc.failureOrder != 5001 {
		t.Fatalf("expected failed provision notification, got %+v", notificationSvc)
	}
}
