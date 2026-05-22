package instance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"

	domaininstance "github.com/AeolianCloud/pveCloud/server/internal/domain/instance"
	domainorder "github.com/AeolianCloud/pveCloud/server/internal/domain/order"
	"github.com/AeolianCloud/pveCloud/server/internal/integration/mcppve"
	mysqlinstance "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/instance"
	mysqlorder "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/order"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
	weblogging "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/logging"
)

const (
	defaultPage    = 1
	defaultPerPage = 15
	maxPerPage     = 100
)

type Service struct {
	db        *gorm.DB
	instances *mysqlinstance.Repository
	orders    *mysqlorder.Repository
	logs      *weblogging.Recorder
	mcp       *mcppve.Client
}

func NewService(db *gorm.DB, mcp *mcppve.Client) *Service {
	return &Service{db: db, instances: mysqlinstance.NewRepository(db), orders: mysqlorder.NewRepository(db), logs: weblogging.NewRecorder(db), mcp: mcp}
}

func (s *Service) List(ctx context.Context, userID uint64, query webdto.InstanceListQuery) (webdto.PageResponse[webdto.InstanceItem], error) {
	if !domaininstance.IsKnownStatus(query.Status) {
		return webdto.PageResponse[webdto.InstanceItem]{}, apperrors.ErrValidation.WithMessage("实例状态不支持")
	}
	page, perPage := normalizePage(query.Page, query.PerPage)
	rows, total, err := s.instances.ListInstances(ctx, mysqlinstance.InstanceFilters{UserID: userID, Status: query.Status}, perPage, (page-1)*perPage)
	if err != nil {
		return webdto.PageResponse[webdto.InstanceItem]{}, err
	}
	items := make([]webdto.InstanceItem, 0, len(rows))
	for _, row := range rows {
		var latest *webdto.RenewalOrderSummary
		if renewal, err := s.orders.LatestRenewalByInstanceNo(ctx, row.InstanceNo); err == nil {
			latest = renewalSummary(renewal)
		}
		items = append(items, instanceItem(row.Instance, latest))
	}
	return pageResponse(items, total, page, perPage), nil
}

func (s *Service) Detail(ctx context.Context, userID uint64, instanceNo string) (webdto.InstanceDetail, error) {
	row, err := s.instances.UserInstance(ctx, userID, strings.TrimSpace(instanceNo))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return webdto.InstanceDetail{}, apperrors.ErrNotFound.WithMessage("实例不存在")
	}
	if err != nil {
		return webdto.InstanceDetail{}, err
	}
	ops, err := s.instances.Operations(ctx, row.ID, 10)
	if err != nil {
		return webdto.InstanceDetail{}, err
	}
	var latest *webdto.RenewalOrderSummary
	if renewal, err := s.orders.LatestRenewalByInstanceNo(ctx, row.InstanceNo); err == nil {
		latest = renewalSummary(renewal)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return webdto.InstanceDetail{}, err
	}
	return instanceDetail(row, ops, latest), nil
}

func (s *Service) CreateRenewalOrder(ctx context.Context, userID uint64, instanceNo string, req webdto.RenewalOrderCreateRequest) (webdto.OrderDetail, error) {
	clientToken := strings.TrimSpace(req.ClientToken)
	instanceNo = strings.TrimSpace(instanceNo)
	if existing, err := s.orders.FindByUserClientToken(ctx, userID, clientToken); err == nil {
		if existing.OrderType == domainorder.TypeRenewal && existing.RelatedInstanceNo != nil && strings.TrimSpace(*existing.RelatedInstanceNo) == instanceNo {
			return webOrderDetail(existing), nil
		}
		return webdto.OrderDetail{}, apperrors.ErrConflict.WithMessage("幂等键已被其它订单使用")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return webdto.OrderDetail{}, err
	}
	var created mysqlorder.Order
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.instances.InstanceForUpdate(ctx, tx, instanceNo)
		if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && current.UserID != userID) {
			return apperrors.ErrNotFound.WithMessage("实例不存在")
		}
		if err != nil {
			return err
		}
		if current.Status == domaininstance.StatusReleased || current.Status == domaininstance.StatusReleasing {
			return apperrors.ErrConflict.WithMessage("当前实例不能创建续费订单")
		}
		networkTypeNo := ""
		if current.NetworkTypeNo != nil {
			networkTypeNo = *current.NetworkTypeNo
		}
		selection, err := s.orders.CatalogSelection(ctx, current.PlanNo, strings.TrimSpace(req.BillingCycle), current.RegionNo, current.TemplateNo, networkTypeNo)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrValidation.WithMessage("当前套餐续费价格不可用")
		}
		if err != nil {
			return err
		}
		created = renewalOrderFromSelection(userID, current.InstanceNo, clientToken, selection)
		if err := s.orders.Create(ctx, tx, &created); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if existing, findErr := s.orders.FindByUserClientToken(ctx, userID, clientToken); findErr == nil {
			if existing.OrderType == domainorder.TypeRenewal && existing.RelatedInstanceNo != nil && strings.TrimSpace(*existing.RelatedInstanceNo) == instanceNo {
				return webOrderDetail(existing), nil
			}
			return webdto.OrderDetail{}, apperrors.ErrConflict.WithMessage("幂等键已被其它订单使用")
		}
		return webdto.OrderDetail{}, err
	}
	_ = s.logs.BusinessNoTx(ctx, weblogging.Snapshot(userID, "", ""), "order", "order.renewal.create", "order", created.OrderNo, "创建续费订单")
	order, err := s.orders.FindByOrderNo(ctx, created.OrderNo)
	if err != nil {
		return webdto.OrderDetail{}, err
	}
	return webOrderDetail(order), nil
}

func (s *Service) Start(ctx context.Context, userID uint64, instanceNo string) (webdto.InstanceDetail, error) {
	return s.operate(ctx, userID, instanceNo, domaininstance.OperationStart)
}

func (s *Service) Stop(ctx context.Context, userID uint64, instanceNo string) (webdto.InstanceDetail, error) {
	return s.operate(ctx, userID, instanceNo, domaininstance.OperationStop)
}

func (s *Service) operate(ctx context.Context, userID uint64, instanceNo string, action string) (webdto.InstanceDetail, error) {
	if !s.mcp.Enabled() {
		return webdto.InstanceDetail{}, mcpUnavailableError()
	}
	var row mysqlinstance.Instance
	var op mysqlinstance.Operation
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.instances.InstanceForUpdate(ctx, tx, strings.TrimSpace(instanceNo))
		if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && current.UserID != userID) {
			return apperrors.ErrNotFound.WithMessage("实例不存在")
		}
		if err != nil {
			return err
		}
		if !canOperate(current.Status, action) {
			return apperrors.ErrConflict.WithMessage("当前实例状态不能执行该操作")
		}
		if err := s.ensureNoRunningOperation(ctx, tx, current.ID); err != nil {
			return err
		}
		row = current
		op = mysqlinstance.Operation{OperationNo: fmt.Sprintf("OP-%d", time.Now().UnixNano()), InstanceID: current.ID, OrderID: &current.OrderID, UserID: &userID, Action: action, Status: domaininstance.OperationStatusRunning}
		return s.instances.CreateOperation(ctx, tx, &op)
	})
	if err != nil {
		return webdto.InstanceDetail{}, err
	}
	accepted, callErr := s.callOperation(ctx, row, action)
	if callErr != nil {
		now := time.Now()
		message := externalStoredMessage(callErr)
		if len(message) > 500 {
			message = message[:500]
		}
		_ = s.instances.UpdateOperation(context.Background(), nil, op.ID, map[string]any{"status": domaininstance.OperationStatusFailed, "error_code": "mcp_call_failed", "error_message": message, "completed_at": now})
		return webdto.InstanceDetail{}, mcpUnavailableError()
	}
	if err := s.instances.UpdateOperation(ctx, nil, op.ID, map[string]any{"external_operation_id": nullableString(accepted.OperationID), "operation_location": nullableString(accepted.OperationLocation), "resource_location": nullableString(accepted.Location)}); err != nil {
		return webdto.InstanceDetail{}, err
	}
	if err := s.enqueueOperationSync(ctx, row.InstanceNo, op.OperationNo); err != nil {
		return webdto.InstanceDetail{}, err
	}
	return s.Detail(ctx, userID, row.InstanceNo)
}

func (s *Service) ensureNoRunningOperation(ctx context.Context, tx *gorm.DB, instanceID uint64) error {
	_, err := s.instances.LatestRunningOperationForUpdate(ctx, tx, instanceID, domaininstance.OperationSync)
	if err == nil {
		return apperrors.ErrConflict.WithMessage("实例已有未完成操作")
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

func (s *Service) callOperation(ctx context.Context, row mysqlinstance.Instance, action string) (mcppve.AsyncAccepted, error) {
	switch action {
	case domaininstance.OperationStart:
		return s.mcp.StartVM(ctx, row.ExternalNode, row.ExternalVMID)
	case domaininstance.OperationStop:
		return s.mcp.StopVM(ctx, row.ExternalNode, row.ExternalVMID)
	default:
		return mcppve.AsyncAccepted{}, apperrors.ErrValidation.WithMessage("实例操作不支持")
	}
}

func canOperate(status string, action string) bool {
	switch action {
	case domaininstance.OperationStart:
		return domaininstance.CanStart(status)
	case domaininstance.OperationStop:
		return domaininstance.CanStop(status)
	default:
		return false
	}
}

func instanceItem(row mysqlinstance.Instance, latest *webdto.RenewalOrderSummary) webdto.InstanceItem {
	countdown := releaseCountdown(row)
	return webdto.InstanceItem{InstanceNo: row.InstanceNo, OrderNo: row.OrderNo, Status: row.Status, ProductName: row.ProductName, PlanName: row.PlanName, RegionName: row.RegionName, NetworkTypeName: row.NetworkTypeName, TemplateName: row.TemplateName, ServiceStartedAt: row.ServiceStartedAt, ExpiresAt: row.ExpiresAt, ExpireStatus: expireStatus(row), ReleaseCountdownSeconds: countdown, LatestRenewalOrder: latest, CreatedAt: row.CreatedAt, ReleasedAt: row.ReleasedAt}
}

func instanceDetail(row mysqlinstance.Instance, ops []mysqlinstance.Operation, latest *webdto.RenewalOrderSummary) webdto.InstanceDetail {
	items := make([]webdto.InstanceOperation, 0, len(ops))
	for _, op := range ops {
		items = append(items, webdto.InstanceOperation{OperationNo: op.OperationNo, Action: op.Action, Status: op.Status, CreatedAt: op.CreatedAt, CompletedAt: op.CompletedAt})
	}
	return webdto.InstanceDetail{InstanceItem: instanceItem(row, latest), ProductNo: row.ProductNo, PlanNo: row.PlanNo, CPUCores: row.CPUCores, MemoryMB: row.MemoryMB, SystemDiskGB: row.SystemDiskGB, DataDiskGB: row.DataDiskGB, BandwidthMbps: row.BandwidthMbps, RegionNo: row.RegionNo, NetworkTypeNo: row.NetworkTypeNo, TemplateNo: row.TemplateNo, OSFamily: row.OSFamily, OSDistribution: row.OSDistribution, OSVersion: row.OSVersion, ExpireNoticeSentAt: row.ExpireNoticeSentAt, ExpireReleaseScheduledAt: row.ExpireReleaseScheduledAt, ExpireReleasedAt: row.ExpireReleasedAt, RenewalAvailable: row.Status != domaininstance.StatusReleased && row.Status != domaininstance.StatusReleasing, Operations: items}
}

func renewalOrderFromSelection(userID uint64, instanceNo string, clientToken string, selection mysqlorder.CatalogSelection) mysqlorder.Order {
	relatedInstanceNo := instanceNo
	return mysqlorder.Order{OrderNo: fmt.Sprintf("ORD-%d", time.Now().UnixNano()), UserID: userID, ClientToken: clientToken, Status: domainorder.StatusPending, OrderType: domainorder.TypeRenewal, RelatedInstanceNo: &relatedInstanceNo, PaymentStatus: domainorder.PaymentStatusUnpaid, ProductNo: selection.ProductNo, ProductType: selection.ProductType, ProductName: selection.ProductName, ProductSummary: selection.ProductSummary, PlanNo: selection.PlanNo, PlanCode: selection.PlanCode, PlanName: selection.PlanName, PlanSummary: selection.PlanSummary, CPUCores: selection.CPUCores, MemoryMB: selection.MemoryMB, SystemDiskGB: selection.SystemDiskGB, DataDiskGB: selection.DataDiskGB, BandwidthMbps: selection.BandwidthMbps, TrafficGB: selection.TrafficGB, PublicIPCount: selection.PublicIPCount, Virtualization: selection.Virtualization, Architecture: selection.Architecture, BillingCycle: selection.BillingCycle, PriceCents: selection.PriceCents, OriginalPriceCents: selection.OriginalPriceCents, Currency: selection.Currency, Quantity: 1, TotalAmountCents: selection.PriceCents, RegionNo: selection.RegionNo, RegionCode: selection.RegionCode, RegionName: selection.RegionName, NetworkTypeNo: selection.NetworkTypeNo, NetworkTypeCode: selection.NetworkTypeCode, NetworkTypeName: selection.NetworkTypeName, TemplateNo: selection.TemplateNo, TemplateCode: selection.TemplateCode, TemplateName: selection.TemplateName, OSFamily: selection.OSFamily, OSDistribution: selection.OSDistribution, OSVersion: selection.OSVersion, OSArchitecture: selection.OSArchitecture}
}

func renewalSummary(order mysqlorder.Order) *webdto.RenewalOrderSummary {
	return &webdto.RenewalOrderSummary{OrderNo: order.OrderNo, Status: order.Status, PaymentStatus: order.PaymentStatus, BillingCycle: order.BillingCycle, TotalAmountCents: order.TotalAmountCents, Currency: order.Currency, PaidAt: order.PaidAt, CreatedAt: order.CreatedAt}
}

func (s *Service) enqueueOperationSync(ctx context.Context, instanceNo string, operationNo string) error {
	payload := map[string]string{"instance_no": instanceNo}
	data, _ := json.Marshal(payload)
	idempotencyKey := "operation_sync:" + strings.TrimSpace(operationNo)
	objectType := "instance"
	objectNo := strings.TrimSpace(instanceNo)
	task := mysqlinstance.Task{TaskNo: fmt.Sprintf("TASK-%d", time.Now().UnixNano()), TaskType: domaininstance.TaskTypeOperationSync, IdempotencyKey: &idempotencyKey, Status: domaininstance.TaskStatusPending, ObjectType: &objectType, ObjectNo: &objectNo, Payload: stringPtr(string(data)), MaxAttempts: 20, ScheduledAt: time.Now().Truncate(time.Millisecond)}
	return s.instances.CreateTaskIgnoreDuplicate(ctx, nil, &task)
}

func webOrderItem(order mysqlorder.Order) webdto.OrderItem {
	orderType := order.OrderType
	if orderType == "" {
		orderType = domainorder.TypePurchase
	}
	paymentStatus := order.PaymentStatus
	if paymentStatus == "" {
		paymentStatus = domainorder.PaymentStatusUnpaid
	}
	return webdto.OrderItem{OrderNo: order.OrderNo, OrderType: orderType, PaymentStatus: paymentStatus, Status: order.Status, RelatedInstanceNo: order.RelatedInstanceNo, ProductName: order.ProductName, PlanName: order.PlanName, BillingCycle: order.BillingCycle, NetworkTypeName: order.NetworkTypeName, TotalAmountCents: order.TotalAmountCents, Currency: order.Currency, CreatedAt: order.CreatedAt, PaidAt: order.PaidAt, CancelledAt: order.CancelledAt, ClosedAt: order.ClosedAt}
}

func webOrderDetail(order mysqlorder.Order) webdto.OrderDetail {
	return webdto.OrderDetail{OrderItem: webOrderItem(order), UserNote: order.UserNote, ProductNo: order.ProductNo, ProductType: order.ProductType, ProductSummary: order.ProductSummary, PlanNo: order.PlanNo, PlanCode: order.PlanCode, PlanSummary: order.PlanSummary, CPUCores: order.CPUCores, MemoryMB: order.MemoryMB, SystemDiskGB: order.SystemDiskGB, DataDiskGB: order.DataDiskGB, BandwidthMbps: order.BandwidthMbps, TrafficGB: order.TrafficGB, PublicIPCount: order.PublicIPCount, Virtualization: order.Virtualization, Architecture: order.Architecture, PriceCents: order.PriceCents, OriginalPriceCents: order.OriginalPriceCents, Quantity: order.Quantity, RegionNo: order.RegionNo, RegionCode: order.RegionCode, RegionName: order.RegionName, NetworkTypeNo: order.NetworkTypeNo, NetworkTypeCode: order.NetworkTypeCode, NetworkTypeName: order.NetworkTypeName, TemplateNo: order.TemplateNo, TemplateCode: order.TemplateCode, TemplateName: order.TemplateName, OSFamily: order.OSFamily, OSDistribution: order.OSDistribution, OSVersion: order.OSVersion, OSArchitecture: order.OSArchitecture}
}

func expireStatus(row mysqlinstance.Instance) string {
	if row.Status == domaininstance.StatusReleased || row.ExpireReleasedAt != nil {
		return "released"
	}
	if row.ExpiresAt == nil {
		return "unknown"
	}
	if row.ExpiresAt.After(time.Now()) {
		return "active"
	}
	return "expired"
}

func releaseCountdown(row mysqlinstance.Instance) *int64 {
	if row.ExpireReleaseScheduledAt == nil {
		return nil
	}
	seconds := int64(time.Until(*row.ExpireReleaseScheduledAt).Seconds())
	if seconds < 0 {
		seconds = 0
	}
	return &seconds
}

func nullableString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func stringPtr(value string) *string {
	return &value
}

func normalizePage(page, perPage int) (int, int) {
	if page < 1 {
		page = defaultPage
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}
	return page, perPage
}

func pageResponse[T any](items []T, total int64, page, perPage int) webdto.PageResponse[T] {
	lastPage := 0
	if total > 0 {
		lastPage = int(math.Ceil(float64(total) / float64(perPage)))
	}
	return webdto.PageResponse[T]{List: items, Total: total, Page: page, PerPage: perPage, LastPage: lastPage}
}

func mcpUnavailableError() error {
	return apperrors.ErrExternalUnavailable.WithMessage("虚拟化管理接口暂不可用")
}

func externalStoredMessage(err error) string {
	if err == nil {
		return ""
	}
	return "虚拟化管理接口调用失败"
}
