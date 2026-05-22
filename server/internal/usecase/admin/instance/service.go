package instance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	domaininstance "github.com/AeolianCloud/pveCloud/server/internal/domain/instance"
	domainorder "github.com/AeolianCloud/pveCloud/server/internal/domain/order"
	"github.com/AeolianCloud/pveCloud/server/internal/integration/mcppve"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	mysqlinstance "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/instance"
	mysqlorder "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/order"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	adminaudit "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/audit"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	adminsupport "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/support"
)

const objectType = "instance"

var (
	ErrOperationPending      = errors.New("instance operation is still running")
	errExpiredReleaseSkipped = errors.New("expired release task skipped")
)

type AdminAuditService = adminaudit.AdminAuditService
type AdminAuditWriteInput = adminaudit.AdminAuditWriteInput

type Service struct {
	db        *gorm.DB
	orders    *mysqlorder.Repository
	instances *mysqlinstance.Repository
	mcp       *mcppve.Client
	lifecycle config.InstanceLifecycleConfig
	audit     *AdminAuditService
}

func NewService(db *gorm.DB, mcp *mcppve.Client, audit *AdminAuditService, lifecycle config.InstanceLifecycleConfig) *Service {
	if audit == nil {
		audit = adminaudit.NewAdminAuditService(db)
	}
	return &Service{db: db, orders: mysqlorder.NewRepository(db), instances: mysqlinstance.NewRepository(db), mcp: mcp, lifecycle: lifecycle, audit: audit}
}

func (s *Service) ListMappings(ctx context.Context, query admindto.InstanceMappingListQuery) (admindto.PageResponse[admindto.InstanceMappingItem], error) {
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	rows, total, err := s.instances.ListMappings(ctx, mysqlinstance.MappingFilters{Status: query.Status, PlanNo: query.PlanNo, RegionNo: query.RegionNo, TemplateNo: query.TemplateNo, NetworkTypeNo: query.NetworkTypeNo}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.InstanceMappingItem]{}, err
	}
	items := make([]admindto.InstanceMappingItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, mappingItem(row))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func (s *Service) CreateMapping(ctx context.Context, operatorID uint64, req admindto.InstanceMappingRequest) (admindto.InstanceMappingItem, error) {
	if req.VMIDStart > req.VMIDEnd || req.NextVMID < req.VMIDStart || req.NextVMID > req.VMIDEnd {
		return admindto.InstanceMappingItem{}, apperrors.ErrValidation.WithMessage("虚拟机编号范围不合法")
	}
	if err := validateCIPackages(req.CIPackages); err != nil {
		return admindto.InstanceMappingItem{}, err
	}
	mapping := mappingFromRequest(req)
	if mapping.MappingNo == "" {
		mapping.MappingNo = fmt.Sprintf("MAP-%d", time.Now().UnixNano())
	}
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		if err := s.instances.CreateMapping(ctx, tx, &mapping); err != nil {
			return err
		}
		return s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: "instance_mapping.create", ObjectType: "instance_mapping", ObjectID: mapping.MappingNo, AfterData: mappingAudit(mapping), Remark: "创建实例交付映射"})
	})
	if err != nil {
		return admindto.InstanceMappingItem{}, err
	}
	return mappingItem(mapping), nil
}

func (s *Service) UpdateMapping(ctx context.Context, operatorID uint64, id uint64, req admindto.InstanceMappingRequest) (admindto.InstanceMappingItem, error) {
	if req.VMIDStart > req.VMIDEnd || req.NextVMID < req.VMIDStart || req.NextVMID > req.VMIDEnd {
		return admindto.InstanceMappingItem{}, apperrors.ErrValidation.WithMessage("虚拟机编号范围不合法")
	}
	if err := validateCIPackages(req.CIPackages); err != nil {
		return admindto.InstanceMappingItem{}, err
	}
	var updated mysqlinstance.ProvisionMapping
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.instances.MappingByIDForUpdate(ctx, tx, id)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("交付映射不存在")
		}
		if err != nil {
			return err
		}
		next := mappingFromRequest(req)
		if strings.TrimSpace(next.MappingNo) == "" {
			next.MappingNo = current.MappingNo
		}
		if current.NextVMID > 0 && next.NextVMID < current.NextVMID {
			return apperrors.ErrConflict.WithMessage("下一个虚拟机编号不能回退")
		}
		if err := s.instances.UpdateMapping(ctx, tx, id, mappingUpdateMap(next)); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: "instance_mapping.update", ObjectType: "instance_mapping", ObjectID: current.MappingNo, BeforeData: mappingAudit(current), AfterData: mappingAudit(next), Remark: "更新实例交付映射"}); err != nil {
			return err
		}
		updated, err = s.instances.MappingByIDForUpdate(ctx, tx, id)
		return err
	})
	if err != nil {
		return admindto.InstanceMappingItem{}, err
	}
	return mappingItem(updated), nil
}

func (s *Service) Nodes(ctx context.Context) ([]admindto.MCPNode, error) {
	result, err := s.mcp.Nodes(ctx)
	if err != nil {
		return nil, externalError(err)
	}
	return mcpNodes(result), nil
}

func (s *Service) Node(ctx context.Context, node string) (admindto.MCPNode, error) {
	node = strings.TrimSpace(node)
	result, err := s.mcp.Node(ctx, node)
	if err != nil {
		return admindto.MCPNode{}, externalError(err)
	}
	return mcpNode(result, node), nil
}

func (s *Service) NodeVMs(ctx context.Context, node string) ([]admindto.MCPVM, error) {
	result, err := s.mcp.NodeVMs(ctx, strings.TrimSpace(node))
	if err != nil {
		return nil, externalError(err)
	}
	return mcpVMs(result), nil
}

func (s *Service) Storage(ctx context.Context) ([]admindto.MCPStorage, error) {
	result, err := s.mcp.Storage(ctx)
	if err != nil {
		return nil, externalError(err)
	}
	return mcpStorageList(result), nil
}

func (s *Service) Provision(ctx context.Context, operatorID uint64, orderNo string) (admindto.ProvisionResponse, error) {
	if !s.mcp.Enabled() {
		return admindto.ProvisionResponse{}, mcpUnavailableError()
	}
	var created mysqlinstance.Instance
	var op mysqlinstance.Operation
	var mapping mysqlinstance.ProvisionMapping
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		order, err := s.orders.OrderForUpdate(ctx, tx, strings.TrimSpace(orderNo))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("订单不存在")
		}
		if err != nil {
			return err
		}
		if !domainorder.CanProvision(order.Status) {
			return apperrors.ErrConflict.WithMessage("当前订单状态不能交付")
		}
		if existing, err := s.instances.InstanceByOrderID(ctx, order.ID); err == nil {
			return apperrors.ErrConflict.WithMessage("订单已存在实例：" + existing.InstanceNo)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		mapping, err = s.instances.MappingForProvision(ctx, tx, order.PlanNo, order.RegionNo, order.TemplateNo, order.NetworkTypeNo)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrValidation.WithMessage("缺少匹配的实例交付映射")
		}
		if err != nil {
			return err
		}
		if mapping.NextVMID > mapping.VMIDEnd {
			return apperrors.ErrConflict.WithMessage("交付映射虚拟机编号已耗尽")
		}
		vmid := mapping.NextVMID
		if err := s.instances.AdvanceMappingVMID(ctx, tx, mapping.ID, vmid+1); err != nil {
			return err
		}
		created = instanceFromOrder(order, mapping, vmid)
		if err := s.instances.CreateInstance(ctx, tx, &created); err != nil {
			return err
		}
		op = newOperation(created.ID, &order.ID, &operatorID, nil, domaininstance.OperationProvision)
		if err := s.instances.CreateOperation(ctx, tx, &op); err != nil {
			return err
		}
		if err := s.orders.Update(ctx, tx, order.ID, map[string]any{"status": domainorder.StatusProvisioning}); err != nil {
			return err
		}
		return s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: "instance.provision", ObjectType: objectType, ObjectID: created.InstanceNo, AfterData: instanceAudit(created), Remark: "触发实例交付"})
	})
	if err != nil {
		return admindto.ProvisionResponse{}, err
	}
	accepted, callErr := s.mcp.CreateVM(ctx, mapping.Node, createVMRequest(created, mapping))
	if callErr != nil {
		_ = s.markOperationFailed(context.Background(), created.ID, op.ID, callErr)
		return admindto.ProvisionResponse{}, externalError(callErr)
	}
	if err := s.instances.UpdateOperation(ctx, nil, op.ID, map[string]any{"external_operation_id": nullableString(accepted.OperationID), "operation_location": nullableString(accepted.OperationLocation), "resource_location": nullableString(accepted.Location)}); err != nil {
		return admindto.ProvisionResponse{}, err
	}
	if err := s.instances.UpdateInstance(ctx, nil, created.ID, map[string]any{"external_resource_location": nullableString(accepted.Location)}); err != nil {
		return admindto.ProvisionResponse{}, err
	}
	if err := s.enqueueOperationSync(ctx, nil, created.InstanceNo, op.OperationNo); err != nil {
		return admindto.ProvisionResponse{}, err
	}
	return s.provisionResponse(ctx, created.InstanceNo)
}

func (s *Service) List(ctx context.Context, query admindto.InstanceListQuery) (admindto.PageResponse[admindto.InstanceItem], error) {
	if !domaininstance.IsKnownStatus(query.Status) {
		return admindto.PageResponse[admindto.InstanceItem]{}, apperrors.ErrValidation.WithMessage("实例状态不支持")
	}
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	rows, total, err := s.instances.ListInstances(ctx, mysqlinstance.InstanceFilters{Status: query.Status, InstanceNo: query.InstanceNo, OrderNo: query.OrderNo, UserKeyword: query.UserKeyword, DateFrom: query.DateFrom, DateTo: query.DateTo}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.InstanceItem]{}, err
	}
	items := make([]admindto.InstanceItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, instanceItem(row))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func (s *Service) Detail(ctx context.Context, instanceNo string) (admindto.InstanceDetail, error) {
	return s.detail(ctx, strings.TrimSpace(instanceNo))
}

func (s *Service) Start(ctx context.Context, operatorID uint64, instanceNo string) (admindto.InstanceDetail, error) {
	return s.operate(ctx, instanceNo, &operatorID, nil, domaininstance.OperationStart)
}

func (s *Service) Stop(ctx context.Context, operatorID uint64, instanceNo string) (admindto.InstanceDetail, error) {
	return s.operate(ctx, instanceNo, &operatorID, nil, domaininstance.OperationStop)
}

func (s *Service) Release(ctx context.Context, operatorID uint64, instanceNo string) (admindto.InstanceDetail, error) {
	return s.operate(ctx, instanceNo, &operatorID, nil, domaininstance.OperationRelease)
}

func (s *Service) Sync(ctx context.Context, operatorID uint64, instanceNo string) (admindto.InstanceDetail, error) {
	return s.sync(ctx, strings.TrimSpace(instanceNo), &operatorID, true)
}

func (s *Service) SyncByWorker(ctx context.Context, instanceNo string) (admindto.InstanceDetail, error) {
	return s.sync(ctx, strings.TrimSpace(instanceNo), nil, false)
}

func (s *Service) ReleaseExpiredByWorker(ctx context.Context, instanceNo string, expectedExpiresAt time.Time) (admindto.InstanceDetail, error) {
	instanceNo = strings.TrimSpace(instanceNo)
	detail, err := s.operateWithGuardWithPendingError(ctx, instanceNo, nil, nil, domaininstance.OperationRelease, ErrOperationPending, func(current mysqlinstance.Instance) error {
		if current.Status == domaininstance.StatusReleased || current.Status == domaininstance.StatusReleasing {
			return errExpiredReleaseSkipped
		}
		if current.ExpiresAt == nil || current.ExpiresAt.After(time.Now()) {
			return errExpiredReleaseSkipped
		}
		if !current.ExpiresAt.Truncate(time.Millisecond).Equal(expectedExpiresAt.Truncate(time.Millisecond)) {
			return errExpiredReleaseSkipped
		}
		return nil
	})
	if errors.Is(err, errExpiredReleaseSkipped) {
		return s.detail(ctx, instanceNo)
	}
	return detail, err
}

func (s *Service) UpdateExpiresAt(ctx context.Context, operatorID uint64, instanceNo string, req admindto.InstanceExpiresAtRequest) (admindto.InstanceDetail, error) {
	if !req.ExpiresAt.After(time.Now()) {
		return admindto.InstanceDetail{}, apperrors.ErrValidation.WithMessage("到期时间必须晚于当前时间")
	}
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.instances.InstanceForUpdate(ctx, tx, strings.TrimSpace(instanceNo))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("实例不存在")
		}
		if err != nil {
			return err
		}
		if current.Status == domaininstance.StatusReleased {
			return apperrors.ErrConflict.WithMessage("已释放实例不能调整到期时间")
		}
		expiresAt := normalizeDBTime(req.ExpiresAt)
		updates := map[string]any{"expires_at": expiresAt, "expire_notice_sent_at": nil}
		if s.lifecycle.AutoReleaseEnabled {
			updates["expire_release_scheduled_at"] = normalizeDBTime(expiresAt.Add(time.Duration(s.lifecycle.ExpireReleaseAfterSeconds) * time.Second))
		} else {
			updates["expire_release_scheduled_at"] = nil
		}
		if err := s.instances.UpdateInstance(ctx, tx, current.ID, updates); err != nil {
			return err
		}
		if err := s.enqueueLifecycleTasks(ctx, tx, current.InstanceNo, expiresAt); err != nil {
			return err
		}
		return s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: "instance.expires_at.update", ObjectType: objectType, ObjectID: current.InstanceNo, BeforeData: instanceAudit(current), AfterData: map[string]any{"expires_at": expiresAt, "remark": req.Remark}, Remark: "调整实例到期时间"})
	})
	if err != nil {
		return admindto.InstanceDetail{}, err
	}
	return s.detail(ctx, strings.TrimSpace(instanceNo))
}

func (s *Service) operate(ctx context.Context, instanceNo string, adminID *uint64, userID *uint64, action string) (admindto.InstanceDetail, error) {
	return s.operateWithGuard(ctx, instanceNo, adminID, userID, action, nil)
}

func (s *Service) operateWithGuard(ctx context.Context, instanceNo string, adminID *uint64, userID *uint64, action string, guard func(mysqlinstance.Instance) error) (admindto.InstanceDetail, error) {
	return s.operateWithGuardWithPendingError(ctx, instanceNo, adminID, userID, action, apperrors.ErrConflict.WithMessage("实例已有未完成操作"), guard)
}

func (s *Service) operateWithGuardWithPendingError(ctx context.Context, instanceNo string, adminID *uint64, userID *uint64, action string, pendingErr error, guard func(mysqlinstance.Instance) error) (admindto.InstanceDetail, error) {
	if !s.mcp.Enabled() {
		return admindto.InstanceDetail{}, mcpUnavailableError()
	}
	var row mysqlinstance.Instance
	var op mysqlinstance.Operation
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.instances.InstanceForUpdate(ctx, tx, strings.TrimSpace(instanceNo))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("实例不存在")
		}
		if err != nil {
			return err
		}
		if !canOperate(current.Status, action) {
			return apperrors.ErrConflict.WithMessage("当前实例状态不能执行该操作")
		}
		if guard != nil {
			if err := guard(current); err != nil {
				return err
			}
		}
		if err := s.ensureNoRunningOperation(ctx, tx, current.ID, pendingErr); err != nil {
			return err
		}
		row = current
		op = newOperation(current.ID, &current.OrderID, adminID, userID, action)
		if err := s.instances.CreateOperation(ctx, tx, &op); err != nil {
			return err
		}
		if action == domaininstance.OperationRelease {
			if err := s.instances.UpdateInstance(ctx, tx, current.ID, map[string]any{"status": domaininstance.StatusReleasing}); err != nil {
				return err
			}
		}
		return s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: adminID, Action: "instance." + action, ObjectType: objectType, ObjectID: current.InstanceNo, BeforeData: instanceAudit(current), AfterData: map[string]any{"action": action}, Remark: "触发实例操作"})
	})
	if err != nil {
		return admindto.InstanceDetail{}, err
	}
	accepted, callErr := s.callOperation(ctx, row, action)
	if callErr != nil {
		_ = s.markOperationFailed(context.Background(), row.ID, op.ID, callErr)
		return admindto.InstanceDetail{}, externalError(callErr)
	}
	if err := s.instances.UpdateOperation(ctx, nil, op.ID, map[string]any{"external_operation_id": nullableString(accepted.OperationID), "operation_location": nullableString(accepted.OperationLocation), "resource_location": nullableString(accepted.Location)}); err != nil {
		return admindto.InstanceDetail{}, err
	}
	if err := s.enqueueOperationSync(ctx, nil, row.InstanceNo, op.OperationNo); err != nil {
		return admindto.InstanceDetail{}, err
	}
	return s.detail(ctx, row.InstanceNo)
}

func (s *Service) ensureNoRunningOperation(ctx context.Context, tx *gorm.DB, instanceID uint64, pendingErr error) error {
	_, err := s.instances.LatestRunningOperationForUpdate(ctx, tx, instanceID, domaininstance.OperationSync)
	if err == nil {
		if pendingErr != nil {
			return pendingErr
		}
		return apperrors.ErrConflict.WithMessage("实例已有未完成操作")
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

func (s *Service) sync(ctx context.Context, instanceNo string, adminID *uint64, recordSyncOperation bool) (admindto.InstanceDetail, error) {
	if !s.mcp.Enabled() {
		return admindto.InstanceDetail{}, mcpUnavailableError()
	}
	var row mysqlinstance.Instance
	var latestOp mysqlinstance.Operation
	var syncOp mysqlinstance.Operation
	hasLatestOp := false
	latestOpSucceeded := false
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.instances.InstanceForUpdate(ctx, tx, instanceNo)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("实例不存在")
		}
		if err != nil {
			return err
		}
		if op, err := s.instances.LatestOperationExcluding(ctx, current.ID, domaininstance.OperationSync); err == nil {
			latestOp = op
			hasLatestOp = true
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		row = current
		if recordSyncOperation {
			syncOp = newOperation(current.ID, &current.OrderID, adminID, nil, domaininstance.OperationSync)
			if err := s.instances.CreateOperation(ctx, tx, &syncOp); err != nil {
				return err
			}
			return s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: adminID, Action: "instance.sync", ObjectType: objectType, ObjectID: current.InstanceNo, BeforeData: instanceAudit(current), Remark: "同步实例状态"})
		}
		return nil
	})
	if err != nil {
		return admindto.InstanceDetail{}, err
	}

	if hasLatestOp && latestOp.Status == domaininstance.OperationStatusRunning {
		if latestOp.ExternalOperationID == nil || strings.TrimSpace(*latestOp.ExternalOperationID) == "" {
			if recordSyncOperation {
				if err := s.markSyncSucceeded(ctx, syncOp.ID); err != nil {
					return admindto.InstanceDetail{}, err
				}
				return s.detail(ctx, row.InstanceNo)
			}
			return admindto.InstanceDetail{}, ErrOperationPending
		}
		result, callErr := s.mcp.Operation(ctx, strings.TrimSpace(*latestOp.ExternalOperationID))
		if callErr != nil {
			if recordSyncOperation {
				_ = s.markSyncFailed(context.Background(), syncOp.ID, callErr)
			}
			return admindto.InstanceDetail{}, externalError(callErr)
		}
		if isOperationFailed(result.Status) || result.Error != nil {
			if err := s.applyOperationFailure(ctx, row, latestOp, syncOp, recordSyncOperation, result); err != nil {
				return admindto.InstanceDetail{}, err
			}
			return s.detail(ctx, row.InstanceNo)
		}
		if isOperationSucceeded(result.Status) {
			now := time.Now()
			if err := s.instances.UpdateOperation(ctx, nil, latestOp.ID, map[string]any{"status": domaininstance.OperationStatusSucceeded, "resource_location": nullableString(result.ResourceLocation), "completed_at": now}); err != nil {
				return admindto.InstanceDetail{}, err
			}
			latestOpSucceeded = true
		} else {
			if recordSyncOperation {
				if err := s.markSyncSucceeded(ctx, syncOp.ID); err != nil {
					return admindto.InstanceDetail{}, err
				}
				return s.detail(ctx, row.InstanceNo)
			}
			return admindto.InstanceDetail{}, ErrOperationPending
		}
	} else if hasLatestOp && latestOp.Status == domaininstance.OperationStatusSucceeded {
		latestOpSucceeded = true
	}
	if row.Status == domaininstance.StatusReleasing {
		if !latestOpSucceeded || latestOp.Action != domaininstance.OperationRelease {
			if recordSyncOperation {
				if err := s.markSyncSucceeded(ctx, syncOp.ID); err != nil {
					return admindto.InstanceDetail{}, err
				}
				return s.detail(ctx, row.InstanceNo)
			}
			return admindto.InstanceDetail{}, ErrOperationPending
		}
		if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
			now := time.Now()
			if recordSyncOperation {
				if err := s.instances.UpdateOperation(ctx, tx, syncOp.ID, map[string]any{"status": domaininstance.OperationStatusSucceeded, "completed_at": now}); err != nil {
					return err
				}
			}
			return s.instances.UpdateInstance(ctx, tx, row.ID, map[string]any{"status": domaininstance.StatusReleased, "released_at": now})
		}); err != nil {
			return admindto.InstanceDetail{}, err
		}
		return s.detail(ctx, row.InstanceNo)
	}
	vm, callErr := s.mcp.VM(ctx, row.ExternalNode, row.ExternalVMID)
	if callErr != nil {
		if recordSyncOperation {
			_ = s.markSyncFailed(context.Background(), syncOp.ID, callErr)
		}
		if err := s.instances.UpdateInstance(ctx, nil, row.ID, map[string]any{"status": domaininstance.StatusError, "last_error_code": nullableString("mcp_query_failed"), "last_error_message": nullableString(externalStoredMessage(callErr))}); err != nil {
			return admindto.InstanceDetail{}, err
		}
		return admindto.InstanceDetail{}, externalError(callErr)
	}
	mappedStatus := domaininstance.MapVMStatus(vm.Status)
	if err := s.applyVMStatus(ctx, row, syncOp, recordSyncOperation, mappedStatus); err != nil {
		return admindto.InstanceDetail{}, err
	}
	return s.detail(ctx, row.InstanceNo)
}

func (s *Service) applyOperationFailure(ctx context.Context, row mysqlinstance.Instance, latestOp mysqlinstance.Operation, syncOp mysqlinstance.Operation, recordSyncOperation bool, result mcppve.Operation) error {
	message := "虚拟化操作失败"
	code := "mcp_operation_failed"
	if result.Error != nil {
		code = result.Error.Code
	}
	now := time.Now()
	return mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		if err := s.instances.UpdateOperation(ctx, tx, latestOp.ID, map[string]any{"status": domaininstance.OperationStatusFailed, "error_code": nullableString(code), "error_message": nullableString(message), "completed_at": now}); err != nil {
			return err
		}
		if recordSyncOperation {
			if err := s.instances.UpdateOperation(ctx, tx, syncOp.ID, map[string]any{"status": domaininstance.OperationStatusSucceeded, "completed_at": now}); err != nil {
				return err
			}
		}
		return s.instances.UpdateInstance(ctx, tx, row.ID, map[string]any{"status": domaininstance.StatusError, "last_error_code": nullableString(code), "last_error_message": nullableString(message)})
	})
}

func (s *Service) applyVMStatus(ctx context.Context, row mysqlinstance.Instance, syncOp mysqlinstance.Operation, recordSyncOperation bool, mappedStatus string) error {
	now := normalizeDBTime(time.Now())
	return mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		if recordSyncOperation {
			if err := s.instances.UpdateOperation(ctx, tx, syncOp.ID, map[string]any{"status": domaininstance.OperationStatusSucceeded, "completed_at": now}); err != nil {
				return err
			}
		}
		instanceUpdates := map[string]any{"status": mappedStatus, "last_error_code": nil, "last_error_message": nil}
		if (mappedStatus == domaininstance.StatusRunning || mappedStatus == domaininstance.StatusStopped) && row.Status == domaininstance.StatusCreating {
			order, err := s.orders.FindByOrderNo(ctx, row.OrderNo)
			if err != nil {
				return err
			}
			months, ok := domainorder.BillingCycleMonths(order.BillingCycle)
			if !ok {
				return apperrors.ErrValidation.WithMessage("订单周期不支持")
			}
			expiresAt := normalizeDBTime(now.AddDate(0, months, 0))
			instanceUpdates["service_started_at"] = now
			instanceUpdates["expires_at"] = expiresAt
			if s.lifecycle.AutoReleaseEnabled {
				instanceUpdates["expire_release_scheduled_at"] = normalizeDBTime(expiresAt.Add(time.Duration(s.lifecycle.ExpireReleaseAfterSeconds) * time.Second))
			}
			if err := s.enqueueLifecycleTasks(ctx, tx, row.InstanceNo, expiresAt); err != nil {
				return err
			}
			if err := s.orders.Update(ctx, tx, row.OrderID, map[string]any{"status": domainorder.StatusFulfilled}); err != nil {
				return err
			}
		}
		return s.instances.UpdateInstance(ctx, tx, row.ID, instanceUpdates)
	})
}

func (s *Service) markSyncSucceeded(ctx context.Context, operationID uint64) error {
	now := time.Now()
	return s.instances.UpdateOperation(ctx, nil, operationID, map[string]any{"status": domaininstance.OperationStatusSucceeded, "completed_at": now})
}

func (s *Service) callOperation(ctx context.Context, row mysqlinstance.Instance, action string) (mcppve.AsyncAccepted, error) {
	switch action {
	case domaininstance.OperationStart:
		return s.mcp.StartVM(ctx, row.ExternalNode, row.ExternalVMID)
	case domaininstance.OperationStop:
		return s.mcp.StopVM(ctx, row.ExternalNode, row.ExternalVMID)
	case domaininstance.OperationRelease:
		return s.mcp.DeleteVM(ctx, row.ExternalNode, row.ExternalVMID)
	default:
		return mcppve.AsyncAccepted{}, apperrors.ErrValidation.WithMessage("实例操作不支持")
	}
}

func (s *Service) markOperationFailed(ctx context.Context, instanceID uint64, operationID uint64, err error) error {
	now := time.Now()
	message := externalStoredMessage(err)
	if len(message) > 500 {
		message = message[:500]
	}
	if updateErr := s.instances.UpdateOperation(ctx, nil, operationID, map[string]any{"status": domaininstance.OperationStatusFailed, "error_code": nullableString("mcp_call_failed"), "error_message": nullableString(message), "completed_at": now}); updateErr != nil {
		return updateErr
	}
	return s.instances.UpdateInstance(ctx, nil, instanceID, map[string]any{"status": domaininstance.StatusError, "last_error_code": nullableString("mcp_call_failed"), "last_error_message": nullableString(message)})
}

func (s *Service) markSyncFailed(ctx context.Context, operationID uint64, err error) error {
	now := time.Now()
	message := externalStoredMessage(err)
	if len(message) > 500 {
		message = message[:500]
	}
	return s.instances.UpdateOperation(ctx, nil, operationID, map[string]any{"status": domaininstance.OperationStatusFailed, "error_code": nullableString("mcp_sync_failed"), "error_message": nullableString(message), "completed_at": now})
}

func (s *Service) provisionResponse(ctx context.Context, instanceNo string) (admindto.ProvisionResponse, error) {
	detail, err := s.detail(ctx, instanceNo)
	if err != nil {
		return admindto.ProvisionResponse{}, err
	}
	if len(detail.Operations) == 0 {
		return admindto.ProvisionResponse{Instance: detail}, nil
	}
	return admindto.ProvisionResponse{Instance: detail, Operation: detail.Operations[0]}, nil
}

func (s *Service) detail(ctx context.Context, instanceNo string) (admindto.InstanceDetail, error) {
	row, err := s.instances.Detail(ctx, instanceNo)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return admindto.InstanceDetail{}, apperrors.ErrNotFound.WithMessage("实例不存在")
	}
	if err != nil {
		return admindto.InstanceDetail{}, err
	}
	ops, err := s.instances.Operations(ctx, row.ID, 20)
	if err != nil {
		return admindto.InstanceDetail{}, err
	}
	var latest *admindto.RenewalOrderSummary
	if renewal, err := s.orders.LatestRenewalByInstanceNo(ctx, row.InstanceNo); err == nil {
		latest = renewalSummary(renewal)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return admindto.InstanceDetail{}, err
	}
	return instanceDetail(row, ops, latest), nil
}

func mappingFromRequest(req admindto.InstanceMappingRequest) mysqlinstance.ProvisionMapping {
	return mysqlinstance.ProvisionMapping{MappingNo: strings.TrimSpace(req.MappingNo), ProductNo: normalizeOptional(req.ProductNo), PlanNo: strings.TrimSpace(req.PlanNo), RegionNo: strings.TrimSpace(req.RegionNo), TemplateNo: strings.TrimSpace(req.TemplateNo), NetworkTypeNo: strings.TrimSpace(req.NetworkTypeNo), Node: strings.TrimSpace(req.Node), Storage: strings.TrimSpace(req.Storage), DiskSource: strings.TrimSpace(req.DiskSource), DiskFormat: normalizeOptional(req.DiskFormat), DiskInterface: normalizeOptional(req.DiskInterface), SnippetsStorage: normalizeOptional(req.SnippetsStorage), CIUser: normalizeOptional(req.CIUser), SSHKeys: normalizeOptional(req.SSHKeys), IPConfig0: normalizeOptional(req.IPConfig0), Nameserver: normalizeOptional(req.Nameserver), SearchDomain: normalizeOptional(req.SearchDomain), CIPackages: normalizeOptional(req.CIPackages), AptMirror: normalizeOptional(req.AptMirror), VMIDStart: req.VMIDStart, VMIDEnd: req.VMIDEnd, NextVMID: req.NextVMID, Status: strings.TrimSpace(req.Status), Remark: normalizeOptional(req.Remark)}
}

func mappingUpdateMap(mapping mysqlinstance.ProvisionMapping) map[string]any {
	return map[string]any{"mapping_no": mapping.MappingNo, "product_no": mapping.ProductNo, "plan_no": mapping.PlanNo, "region_no": mapping.RegionNo, "template_no": mapping.TemplateNo, "network_type_no": mapping.NetworkTypeNo, "node": mapping.Node, "storage": mapping.Storage, "disk_source": mapping.DiskSource, "disk_format": mapping.DiskFormat, "disk_interface": mapping.DiskInterface, "snippets_storage": mapping.SnippetsStorage, "ci_user": mapping.CIUser, "ssh_keys": mapping.SSHKeys, "ip_config0": mapping.IPConfig0, "nameserver": mapping.Nameserver, "search_domain": mapping.SearchDomain, "ci_packages": mapping.CIPackages, "apt_mirror": mapping.AptMirror, "vmid_start": mapping.VMIDStart, "vmid_end": mapping.VMIDEnd, "next_vmid": mapping.NextVMID, "status": mapping.Status, "remark": mapping.Remark}
}

func instanceFromOrder(order mysqlorder.Order, mapping mysqlinstance.ProvisionMapping, vmid uint) mysqlinstance.Instance {
	return mysqlinstance.Instance{InstanceNo: fmt.Sprintf("INS-%d", time.Now().UnixNano()), UserID: order.UserID, OrderID: order.ID, OrderNo: order.OrderNo, Status: domaininstance.StatusCreating, ProductNo: order.ProductNo, ProductName: order.ProductName, PlanNo: order.PlanNo, PlanName: order.PlanName, CPUCores: order.CPUCores, MemoryMB: order.MemoryMB, SystemDiskGB: order.SystemDiskGB, DataDiskGB: order.DataDiskGB, BandwidthMbps: order.BandwidthMbps, RegionNo: order.RegionNo, RegionName: order.RegionName, NetworkTypeNo: nullableString(order.NetworkTypeNo), NetworkTypeName: nullableString(order.NetworkTypeName), TemplateNo: order.TemplateNo, TemplateName: order.TemplateName, OSFamily: order.OSFamily, OSDistribution: order.OSDistribution, OSVersion: order.OSVersion, ExternalNode: mapping.Node, ExternalVMID: vmid}
}

func createVMRequest(instance mysqlinstance.Instance, mapping mysqlinstance.ProvisionMapping) mcppve.CreateVMRequest {
	req := mcppve.CreateVMRequest{VMID: instance.ExternalVMID, Name: instance.InstanceNo, Cores: instance.CPUCores, Memory: instance.MemoryMB, Storage: mapping.Storage, DiskSource: mapping.DiskSource}
	req.DiskFormat = value(mapping.DiskFormat)
	req.DiskInterface = value(mapping.DiskInterface)
	req.CIUser = value(mapping.CIUser)
	req.SSHKeys = value(mapping.SSHKeys)
	req.IPConfig0 = value(mapping.IPConfig0)
	req.Nameserver = value(mapping.Nameserver)
	req.SearchDomain = value(mapping.SearchDomain)
	req.SnippetsStorage = value(mapping.SnippetsStorage)
	req.AptMirror = value(mapping.AptMirror)
	if mapping.CIPackages != nil {
		_ = json.Unmarshal([]byte(*mapping.CIPackages), &req.CIPackages)
	}
	return req
}

func newOperation(instanceID uint64, orderID *uint64, adminID *uint64, userID *uint64, action string) mysqlinstance.Operation {
	return mysqlinstance.Operation{OperationNo: fmt.Sprintf("OP-%d", time.Now().UnixNano()), InstanceID: instanceID, OrderID: orderID, AdminID: adminID, UserID: userID, Action: action, Status: domaininstance.OperationStatusRunning}
}

func (s *Service) enqueueOperationSync(ctx context.Context, tx *gorm.DB, instanceNo string, operationNo string) error {
	payload := map[string]string{"instance_no": instanceNo}
	data, _ := json.Marshal(payload)
	idempotencyKey := "operation_sync:" + strings.TrimSpace(operationNo)
	objectType := "instance"
	objectNo := strings.TrimSpace(instanceNo)
	task := mysqlinstance.Task{TaskNo: fmt.Sprintf("TASK-%d", time.Now().UnixNano()), TaskType: domaininstance.TaskTypeOperationSync, IdempotencyKey: &idempotencyKey, Status: domaininstance.TaskStatusPending, ObjectType: &objectType, ObjectNo: &objectNo, Payload: stringPtr(string(data)), MaxAttempts: 20, ScheduledAt: normalizeDBTime(time.Now())}
	return s.instances.CreateTaskIgnoreDuplicate(ctx, tx, &task)
}

func (s *Service) enqueueLifecycleTasks(ctx context.Context, tx *gorm.DB, instanceNo string, expiresAt time.Time) error {
	expiresAt = normalizeDBTime(expiresAt)
	payload := map[string]string{"instance_no": instanceNo, "expires_at": expiresAt.Format(time.RFC3339Nano)}
	data, _ := json.Marshal(payload)
	objectType := "instance"
	objectNo := strings.TrimSpace(instanceNo)
	noticeAt := normalizeDBTime(expiresAt.Add(-time.Duration(s.lifecycle.ExpireNoticeBeforeSeconds) * time.Second))
	if noticeAt.Before(time.Now()) {
		noticeAt = normalizeDBTime(time.Now())
	}
	noticeKey := "expiry_notice:" + objectNo + ":" + expiresAt.Format(time.RFC3339Nano)
	noticeTask := mysqlinstance.Task{TaskNo: fmt.Sprintf("TASK-%d", time.Now().UnixNano()), TaskType: domaininstance.TaskTypeExpiryNotice, IdempotencyKey: &noticeKey, Status: domaininstance.TaskStatusPending, ObjectType: &objectType, ObjectNo: &objectNo, Payload: stringPtr(string(data)), MaxAttempts: 10, ScheduledAt: noticeAt}
	if err := s.instances.CreateTaskIgnoreDuplicate(ctx, tx, &noticeTask); err != nil {
		return err
	}
	if !s.lifecycle.AutoReleaseEnabled {
		return nil
	}
	releaseKey := "expiry_release:" + objectNo + ":" + expiresAt.Format(time.RFC3339Nano)
	releaseTask := mysqlinstance.Task{TaskNo: fmt.Sprintf("TASK-%d", time.Now().UnixNano()+1), TaskType: domaininstance.TaskTypeExpiryRelease, IdempotencyKey: &releaseKey, Status: domaininstance.TaskStatusPending, ObjectType: &objectType, ObjectNo: &objectNo, Payload: stringPtr(string(data)), MaxAttempts: 10, ScheduledAt: normalizeDBTime(expiresAt.Add(time.Duration(s.lifecycle.ExpireReleaseAfterSeconds) * time.Second))}
	return s.instances.CreateTaskIgnoreDuplicate(ctx, tx, &releaseTask)
}

func canOperate(status string, action string) bool {
	switch action {
	case domaininstance.OperationStart:
		return domaininstance.CanStart(status)
	case domaininstance.OperationStop:
		return domaininstance.CanStop(status)
	case domaininstance.OperationRelease:
		return domaininstance.CanRelease(status)
	default:
		return false
	}
}

func isOperationSucceeded(status string) bool {
	normalized := strings.ToLower(strings.TrimSpace(status))
	return normalized == "succeeded" || normalized == "success" || normalized == "done"
}

func isOperationFailed(status string) bool {
	normalized := strings.ToLower(strings.TrimSpace(status))
	return normalized == "failed" || normalized == "canceled" || normalized == "cancelled"
}

func externalError(err error) error {
	if err == nil {
		return nil
	}
	return mcpUnavailableError()
}

func mcpNodes(value any) []admindto.MCPNode {
	items := anySlice(value)
	nodes := make([]admindto.MCPNode, 0, len(items))
	for _, item := range items {
		nodes = append(nodes, mcpNode(item, ""))
	}
	return nodes
}

func mcpNode(value any, fallback string) admindto.MCPNode {
	raw := anyMap(value)
	node := stringField(raw, "node")
	name := stringField(raw, "name")
	if node == "" {
		node = fallback
	}
	if node == "" {
		node = name
	}
	if name == "" {
		name = node
	}
	return admindto.MCPNode{Node: node, Name: name, Status: stringField(raw, "status")}
}

func mcpVMs(value any) []admindto.MCPVM {
	items := anySlice(value)
	vms := make([]admindto.MCPVM, 0, len(items))
	for _, item := range items {
		raw := anyMap(item)
		vms = append(vms, admindto.MCPVM{VMID: uintField(raw, "vmid"), Name: stringField(raw, "name"), Status: stringField(raw, "status"), CPUs: intField(raw, "cpus"), Mem: int64Field(raw, "mem"), MaxMem: int64Field(raw, "maxmem")})
	}
	return vms
}

func mcpStorageList(value any) []admindto.MCPStorage {
	items := anySlice(value)
	storage := make([]admindto.MCPStorage, 0, len(items))
	for _, item := range items {
		raw := anyMap(item)
		storageName := stringField(raw, "storage")
		name := stringField(raw, "name")
		if storageName == "" {
			storageName = name
		}
		if name == "" {
			name = storageName
		}
		storage = append(storage, admindto.MCPStorage{Storage: storageName, Name: name, Type: stringField(raw, "type"), Status: stringField(raw, "status")})
	}
	return storage
}

func anySlice(value any) []any {
	switch typed := value.(type) {
	case []any:
		return typed
	case []map[string]any:
		items := make([]any, 0, len(typed))
		for _, item := range typed {
			items = append(items, item)
		}
		return items
	default:
		return nil
	}
}

func anyMap(value any) map[string]any {
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return nil
}

func stringField(raw map[string]any, key string) string {
	if raw == nil {
		return ""
	}
	switch value := raw[key].(type) {
	case string:
		return strings.TrimSpace(value)
	case json.Number:
		return value.String()
	default:
		return ""
	}
}

func uintField(raw map[string]any, key string) uint {
	if raw == nil {
		return 0
	}
	switch value := raw[key].(type) {
	case float64:
		return uint(value)
	case json.Number:
		parsed, _ := value.Int64()
		if parsed > 0 {
			return uint(parsed)
		}
	case int:
		if value > 0 {
			return uint(value)
		}
	}
	return 0
}

func intField(raw map[string]any, key string) int {
	if raw == nil {
		return 0
	}
	switch value := raw[key].(type) {
	case float64:
		return int(value)
	case json.Number:
		parsed, _ := value.Int64()
		return int(parsed)
	case int:
		return value
	}
	return 0
}

func int64Field(raw map[string]any, key string) int64 {
	if raw == nil {
		return 0
	}
	switch value := raw[key].(type) {
	case float64:
		return int64(value)
	case json.Number:
		parsed, _ := value.Int64()
		return parsed
	case int64:
		return value
	case int:
		return int64(value)
	}
	return 0
}

func mcpUnavailableError() error {
	return apperrors.ErrExternalUnavailable.WithMessage("虚拟化管理接口暂不可用")
}

func mappingItem(row mysqlinstance.ProvisionMapping) admindto.InstanceMappingItem {
	return admindto.InstanceMappingItem{ID: row.ID, MappingNo: row.MappingNo, ProductNo: row.ProductNo, PlanNo: row.PlanNo, RegionNo: row.RegionNo, TemplateNo: row.TemplateNo, NetworkTypeNo: row.NetworkTypeNo, Node: row.Node, Storage: row.Storage, DiskSource: row.DiskSource, DiskFormat: row.DiskFormat, DiskInterface: row.DiskInterface, SnippetsStorage: row.SnippetsStorage, CIUser: row.CIUser, SSHKeys: row.SSHKeys, IPConfig0: row.IPConfig0, Nameserver: row.Nameserver, SearchDomain: row.SearchDomain, CIPackages: row.CIPackages, AptMirror: row.AptMirror, VMIDStart: row.VMIDStart, VMIDEnd: row.VMIDEnd, NextVMID: row.NextVMID, Status: row.Status, Remark: row.Remark, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func instanceItem(row mysqlinstance.InstanceRow) admindto.InstanceItem {
	return admindto.InstanceItem{InstanceNo: row.InstanceNo, OrderNo: row.OrderNo, User: admindto.OrderUserSummary{ID: row.UserID, Username: row.Username, Email: row.Email, DisplayName: row.DisplayName}, Status: row.Status, ProductName: row.ProductName, PlanName: row.PlanName, RegionName: row.RegionName, NetworkTypeName: row.NetworkTypeName, TemplateName: row.TemplateName, ExternalNode: row.ExternalNode, ExternalVMID: row.ExternalVMID, ServiceStartedAt: row.ServiceStartedAt, ExpiresAt: row.ExpiresAt, ExpireNoticeSentAt: row.ExpireNoticeSentAt, ExpireReleaseScheduledAt: row.ExpireReleaseScheduledAt, ExpireReleasedAt: row.ExpireReleasedAt, CreatedAt: row.CreatedAt, ReleasedAt: row.ReleasedAt}
}

func instanceDetail(row mysqlinstance.InstanceRow, ops []mysqlinstance.Operation, latest *admindto.RenewalOrderSummary) admindto.InstanceDetail {
	items := make([]admindto.InstanceOperation, 0, len(ops))
	for _, op := range ops {
		items = append(items, operationItem(op))
	}
	return admindto.InstanceDetail{InstanceItem: instanceItem(row), ProductNo: row.ProductNo, PlanNo: row.PlanNo, CPUCores: row.CPUCores, MemoryMB: row.MemoryMB, SystemDiskGB: row.SystemDiskGB, DataDiskGB: row.DataDiskGB, BandwidthMbps: row.BandwidthMbps, RegionNo: row.RegionNo, NetworkTypeNo: row.NetworkTypeNo, TemplateNo: row.TemplateNo, OSFamily: row.OSFamily, OSDistribution: row.OSDistribution, OSVersion: row.OSVersion, ExternalResourceLocation: row.ExternalResourceLocation, LastErrorCode: row.LastErrorCode, LastErrorMessage: row.LastErrorMessage, RenewalAvailable: row.Status != domaininstance.StatusReleased && row.Status != domaininstance.StatusReleasing, LatestRenewalOrder: latest, Operations: items}
}

func renewalSummary(order mysqlorder.Order) *admindto.RenewalOrderSummary {
	return &admindto.RenewalOrderSummary{OrderNo: order.OrderNo, Status: order.Status, PaymentStatus: order.PaymentStatus, BillingCycle: order.BillingCycle, TotalAmountCents: order.TotalAmountCents, Currency: order.Currency, PaidAt: order.PaidAt, CreatedAt: order.CreatedAt}
}

func operationItem(op mysqlinstance.Operation) admindto.InstanceOperation {
	return admindto.InstanceOperation{OperationNo: op.OperationNo, Action: op.Action, Status: op.Status, ExternalOperationID: op.ExternalOperationID, OperationLocation: op.OperationLocation, ResourceLocation: op.ResourceLocation, ErrorCode: op.ErrorCode, ErrorMessage: op.ErrorMessage, CreatedAt: op.CreatedAt, CompletedAt: op.CompletedAt}
}

func mappingAudit(mapping mysqlinstance.ProvisionMapping) map[string]any {
	return map[string]any{"mapping_no": mapping.MappingNo, "plan_no": mapping.PlanNo, "region_no": mapping.RegionNo, "template_no": mapping.TemplateNo, "network_type_no": mapping.NetworkTypeNo, "node": mapping.Node, "storage": mapping.Storage, "disk_source": mapping.DiskSource, "status": mapping.Status}
}

func instanceAudit(row mysqlinstance.Instance) map[string]any {
	return map[string]any{"instance_no": row.InstanceNo, "order_no": row.OrderNo, "status": row.Status, "node": row.ExternalNode, "vmid": row.ExternalVMID, "expires_at": row.ExpiresAt}
}

func normalizeOptional(value *string) *string {
	if value == nil {
		return nil
	}
	return textutil.NormalizeOptionalString(value)
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

func value(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return strings.TrimSpace(*ptr)
}

func normalizeDBTime(value time.Time) time.Time {
	return value.Truncate(time.Millisecond)
}

func validateCIPackages(value *string) error {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	var packages []string
	if err := json.Unmarshal([]byte(*value), &packages); err != nil || packages == nil {
		return apperrors.ErrValidation.WithMessage("初始化软件包配置必须是字符串数组")
	}
	return nil
}

func externalStoredMessage(err error) string {
	if err == nil {
		return ""
	}
	return "虚拟化管理接口调用失败"
}
