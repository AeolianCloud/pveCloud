package instance

import (
	"context"
	"fmt"

	"github.com/AeolianCloud/pveCloud/server/internal/catalog"
	"github.com/AeolianCloud/pveCloud/server/internal/resource"
)

type AuditService interface {
	Record(ctx context.Context, event string, businessID uint64) error
}

type NotificationService interface {
	SendProvisionSuccess(ctx context.Context, userID uint64, instanceNo string) error
	SendProvisionFailure(ctx context.Context, userID uint64, orderID uint64) error
}

type Service struct {
	repo            Repo
	vmClient        resource.VMClient
	auditSvc        AuditService
	notificationSvc NotificationService
}

func NewService(repo Repo, vmClient resource.VMClient, auditSvc AuditService, notificationSvc NotificationService) *Service {
	return &Service{
		repo:            repo,
		vmClient:        vmClient,
		auditSvc:        auditSvc,
		notificationSvc: notificationSvc,
	}
}

func (s *Service) HandleCreateInstanceTask(ctx context.Context, orderID uint64) (ProvisionResult, error) {
	if existing, found, err := s.repo.FindProvisionResultByOrder(ctx, orderID); err != nil {
		return ProvisionResult{}, err
	} else if found {
		return existing, nil
	}

	orderRow, reservation, err := s.repo.LoadPaidOrderForProvision(ctx, orderID)
	if err != nil {
		return ProvisionResult{}, err
	}

	vmResp, err := s.vmClient.CreateVM(ctx, buildCreateRequest(orderRow, reservation))
	if err != nil {
		_ = s.auditSvc.Record(ctx, "order.provision.failed", orderID)
		_ = s.notificationSvc.SendProvisionFailure(ctx, orderRow.UserID, orderID)
		return ProvisionResult{}, err
	}

	result, err := s.repo.CreateInstanceAndActivateOrder(ctx, orderRow, reservation, vmResp)
	if err != nil {
		return ProvisionResult{}, err
	}

	_ = s.auditSvc.Record(ctx, "order.provision.success", orderRow.ID)
	_ = s.notificationSvc.SendProvisionSuccess(ctx, orderRow.UserID, result.Instance.InstanceNo)

	return result, nil
}

func (s *Service) ListByUser(ctx context.Context, userID uint64) ([]Instance, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) ListAll(ctx context.Context) ([]Instance, error) {
	return s.repo.ListAll(ctx)
}

func buildCreateRequest(orderRow PaidOrder, reservation catalog.Reservation) resource.CreateVMRequest {
	return resource.CreateVMRequest{
		OrderID:  orderRow.ID,
		NodeID:   reservation.NodeID,
		UserID:   orderRow.UserID,
		Hostname: fmt.Sprintf("inst-%d", orderRow.ID),
	}
}
