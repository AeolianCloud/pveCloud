package instance

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"

	domaininstance "github.com/AeolianCloud/pveCloud/server/internal/domain/instance"
	"github.com/AeolianCloud/pveCloud/server/internal/integration/mcppve"
	mysqlinstance "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/instance"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
)

const (
	defaultPage    = 1
	defaultPerPage = 15
	maxPerPage     = 100
)

type Service struct {
	db        *gorm.DB
	instances *mysqlinstance.Repository
	mcp       *mcppve.Client
}

func NewService(db *gorm.DB, mcp *mcppve.Client) *Service {
	return &Service{db: db, instances: mysqlinstance.NewRepository(db), mcp: mcp}
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
		items = append(items, instanceItem(row.Instance))
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
	return instanceDetail(row, ops), nil
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
	_ = s.instances.UpdateOperation(ctx, nil, op.ID, map[string]any{"external_operation_id": nullableString(accepted.OperationID), "operation_location": nullableString(accepted.OperationLocation), "resource_location": nullableString(accepted.Location)})
	return s.Detail(ctx, userID, row.InstanceNo)
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

func instanceItem(row mysqlinstance.Instance) webdto.InstanceItem {
	return webdto.InstanceItem{InstanceNo: row.InstanceNo, OrderNo: row.OrderNo, Status: row.Status, ProductName: row.ProductName, PlanName: row.PlanName, RegionName: row.RegionName, NetworkTypeName: row.NetworkTypeName, TemplateName: row.TemplateName, CreatedAt: row.CreatedAt, ReleasedAt: row.ReleasedAt}
}

func instanceDetail(row mysqlinstance.Instance, ops []mysqlinstance.Operation) webdto.InstanceDetail {
	items := make([]webdto.InstanceOperation, 0, len(ops))
	for _, op := range ops {
		items = append(items, webdto.InstanceOperation{OperationNo: op.OperationNo, Action: op.Action, Status: op.Status, CreatedAt: op.CreatedAt, CompletedAt: op.CompletedAt})
	}
	return webdto.InstanceDetail{InstanceItem: instanceItem(row), ProductNo: row.ProductNo, PlanNo: row.PlanNo, CPUCores: row.CPUCores, MemoryMB: row.MemoryMB, SystemDiskGB: row.SystemDiskGB, DataDiskGB: row.DataDiskGB, BandwidthMbps: row.BandwidthMbps, RegionNo: row.RegionNo, NetworkTypeNo: row.NetworkTypeNo, TemplateNo: row.TemplateNo, OSFamily: row.OSFamily, OSDistribution: row.OSDistribution, OSVersion: row.OSVersion, Operations: items}
}

func nullableString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
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
