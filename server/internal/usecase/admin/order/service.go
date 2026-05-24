package order

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

const objectType = "order"

type AdminAuditService = adminaudit.AdminAuditService
type AdminAuditWriteInput = adminaudit.AdminAuditWriteInput

type Service struct {
	db        *gorm.DB
	orders    *mysqlorder.Repository
	instances *mysqlinstance.Repository
	lifecycle config.InstanceLifecycleConfig
	audit     *AdminAuditService
}

func NewService(db *gorm.DB, audit *AdminAuditService, lifecycle config.InstanceLifecycleConfig) *Service {
	if audit == nil {
		audit = adminaudit.NewAdminAuditService(db)
	}
	return &Service{db: db, orders: mysqlorder.NewRepository(db), instances: mysqlinstance.NewRepository(db), lifecycle: lifecycle, audit: audit}
}

func (s *Service) List(ctx context.Context, query admindto.OrderListQuery) (admindto.PageResponse[admindto.AdminOrderItem], error) {
	if !domainorder.IsKnownStatus(query.Status) {
		return admindto.PageResponse[admindto.AdminOrderItem]{}, apperrors.ErrValidation.WithMessage("订单状态不支持")
	}
	if !domainorder.IsKnownType(query.OrderType) {
		return admindto.PageResponse[admindto.AdminOrderItem]{}, apperrors.ErrValidation.WithMessage("订单类型不支持")
	}
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	rows, total, err := s.orders.List(ctx, mysqlorder.ListFilters{OrderType: query.OrderType, Status: query.Status, OrderNo: query.OrderNo, UserKeyword: query.UserKeyword, DateFrom: query.DateFrom, DateTo: query.DateTo}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.AdminOrderItem]{}, err
	}
	items := make([]admindto.AdminOrderItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, adminOrderItem(row))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func (s *Service) Detail(ctx context.Context, orderNo string) (admindto.AdminOrderDetail, error) {
	row, err := s.orders.Detail(ctx, strings.TrimSpace(orderNo))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return admindto.AdminOrderDetail{}, apperrors.ErrNotFound.WithMessage("订单不存在")
	}
	if err != nil {
		return admindto.AdminOrderDetail{}, err
	}
	return adminOrderDetail(row), nil
}

func (s *Service) UpdateAdminNote(ctx context.Context, operatorID uint64, orderNo string, req admindto.OrderAdminNoteRequest) (admindto.AdminOrderDetail, error) {
	return s.update(ctx, operatorID, orderNo, func(current mysqlorder.Order) (map[string]any, string, error) {
		return map[string]any{"admin_note": textutil.NormalizeOptionalString(req.AdminNote)}, "order.admin_note.update", nil
	})
}

func (s *Service) Cancel(ctx context.Context, operatorID uint64, orderNo string, req admindto.OrderStatusRequest) (admindto.AdminOrderDetail, error) {
	return s.update(ctx, operatorID, orderNo, func(current mysqlorder.Order) (map[string]any, string, error) {
		if !domainorder.CanCancel(current.Status) {
			return nil, "", apperrors.ErrConflict.WithMessage("当前订单状态不能取消")
		}
		now := time.Now()
		return map[string]any{"status": domainorder.StatusCancelled, "cancel_reason": textutil.NormalizeOptionalString(req.Reason), "cancelled_at": now}, "order.cancel", nil
	})
}

func (s *Service) Close(ctx context.Context, operatorID uint64, orderNo string, req admindto.OrderStatusRequest) (admindto.AdminOrderDetail, error) {
	return s.update(ctx, operatorID, orderNo, func(current mysqlorder.Order) (map[string]any, string, error) {
		if !domainorder.CanClose(current.Status) {
			return nil, "", apperrors.ErrConflict.WithMessage("当前订单状态不能关闭")
		}
		now := time.Now()
		return map[string]any{"status": domainorder.StatusClosed, "closed_reason": textutil.NormalizeOptionalString(req.Reason), "closed_at": now}, "order.close", nil
	})
}

func (s *Service) ConfirmRenewal(ctx context.Context, operatorID uint64, orderNo string, req admindto.OrderRenewalConfirmRequest) (admindto.AdminOrderDetail, error) {
	var updatedOrderNo string
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		order, err := s.orders.OrderForUpdate(ctx, tx, strings.TrimSpace(orderNo))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("订单不存在")
		}
		if err != nil {
			return err
		}
		if !domainorder.CanConfirmRenewal(order.Status, order.OrderType) {
			return apperrors.ErrConflict.WithMessage("当前订单不是可确认的续费订单")
		}
		if order.RelatedInstanceNo == nil || strings.TrimSpace(*order.RelatedInstanceNo) == "" {
			return apperrors.ErrConflict.WithMessage("续费订单未关联实例")
		}
		instance, err := s.instances.InstanceForUpdate(ctx, tx, *order.RelatedInstanceNo)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("关联实例不存在")
		}
		if err != nil {
			return err
		}
		if instance.UserID != order.UserID {
			return apperrors.ErrConflict.WithMessage("续费订单与实例归属不一致")
		}
		if instance.Status == domaininstance.StatusReleased || instance.Status == domaininstance.StatusReleasing {
			return apperrors.ErrConflict.WithMessage("已释放实例不能续费")
		}
		months, ok := domainorder.BillingCycleMonths(order.BillingCycle)
		if !ok {
			return apperrors.ErrValidation.WithMessage("续费周期不支持")
		}
		now := time.Now()
		nextExpiresAt := renewalExpiresAt(now, instance.ExpiresAt, months)
		instanceUpdates := renewalInstanceUpdates(s.lifecycle, nextExpiresAt)
		if err := s.instances.UpdateInstance(ctx, tx, instance.ID, instanceUpdates); err != nil {
			return err
		}
		if err := s.enqueueLifecycleTasks(ctx, tx, instance.InstanceNo, nextExpiresAt); err != nil {
			return err
		}
		orderUpdates := map[string]any{"status": domainorder.StatusFulfilled, "payment_status": domainorder.PaymentStatusManualConfirmed, "paid_at": now}
		if err := s.orders.Update(ctx, tx, order.ID, orderUpdates); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: "order.renewal.confirm", ObjectType: objectType, ObjectID: order.OrderNo, BeforeData: auditSnapshot(order), AfterData: map[string]any{"expires_at": nextExpiresAt, "payment_status": domainorder.PaymentStatusManualConfirmed, "remark": req.Remark}, Remark: firstNonEmptyValue(req.Remark, "人工确认续费订单")}); err != nil {
			return err
		}
		updatedOrderNo = order.OrderNo
		return nil
	})
	if err != nil {
		return admindto.AdminOrderDetail{}, err
	}
	updated, err := s.orders.Detail(ctx, updatedOrderNo)
	if err != nil {
		return admindto.AdminOrderDetail{}, err
	}
	return adminOrderDetail(updated), nil
}

func (s *Service) update(ctx context.Context, operatorID uint64, orderNo string, build func(mysqlorder.Order) (map[string]any, string, error)) (admindto.AdminOrderDetail, error) {
	var updatedOrderNo string
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.orders.OrderForUpdate(ctx, tx, strings.TrimSpace(orderNo))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("订单不存在")
		}
		if err != nil {
			return err
		}
		updates, action, err := build(current)
		if err != nil {
			return err
		}
		if err := s.orders.Update(ctx, tx, current.ID, updates); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: action, ObjectType: objectType, ObjectID: current.OrderNo, BeforeData: auditSnapshot(current), AfterData: updates, Remark: "处理订单"}); err != nil {
			return err
		}
		updatedOrderNo = current.OrderNo
		return nil
	})
	if err != nil {
		return admindto.AdminOrderDetail{}, err
	}
	updated, err := s.orders.Detail(ctx, updatedOrderNo)
	if err != nil {
		return admindto.AdminOrderDetail{}, err
	}
	return adminOrderDetail(updated), nil
}

func adminOrderItem(row mysqlorder.OrderRow) admindto.AdminOrderItem {
	orderType := row.OrderType
	if orderType == "" {
		orderType = domainorder.TypePurchase
	}
	paymentStatus := row.PaymentStatus
	if paymentStatus == "" {
		paymentStatus = domainorder.PaymentStatusUnpaid
	}
	return admindto.AdminOrderItem{OrderNo: row.OrderNo, OrderType: orderType, PaymentStatus: paymentStatus, User: admindto.OrderUserSummary{ID: row.UserID, Username: row.Username, Email: row.Email, DisplayName: row.DisplayName}, Status: row.Status, RelatedInstanceNo: row.RelatedInstanceNo, ProductName: row.ProductName, PlanName: row.PlanName, BillingCycle: row.BillingCycle, NetworkTypeName: row.NetworkTypeName, TotalAmountCents: row.TotalAmountCents, Currency: row.Currency, AdminNote: row.AdminNote, CreatedAt: row.CreatedAt, PaidAt: row.PaidAt, CancelledAt: row.CancelledAt, ClosedAt: row.ClosedAt}
}

func adminOrderDetail(row mysqlorder.OrderRow) admindto.AdminOrderDetail {
	return admindto.AdminOrderDetail{AdminOrderItem: adminOrderItem(row), UserNote: row.UserNote, CancelReason: row.CancelReason, ClosedReason: row.ClosedReason, ProductNo: row.ProductNo, ProductType: row.ProductType, ProductSummary: row.ProductSummary, PlanNo: row.PlanNo, PlanCode: row.PlanCode, PlanSummary: row.PlanSummary, CPUCores: row.CPUCores, MemoryMB: row.MemoryMB, SystemDiskGB: row.SystemDiskGB, DataDiskGB: row.DataDiskGB, BandwidthMbps: row.BandwidthMbps, TrafficGB: row.TrafficGB, PublicIPCount: row.PublicIPCount, Virtualization: row.Virtualization, Architecture: row.Architecture, PriceCents: row.PriceCents, OriginalPriceCents: row.OriginalPriceCents, Quantity: row.Quantity, RegionNo: row.RegionNo, RegionCode: row.RegionCode, RegionName: row.RegionName, NetworkTypeNo: row.NetworkTypeNo, NetworkTypeCode: row.NetworkTypeCode, NetworkTypeName: row.NetworkTypeName, TemplateNo: row.TemplateNo, TemplateCode: row.TemplateCode, TemplateName: row.TemplateName, OSFamily: row.OSFamily, OSDistribution: row.OSDistribution, OSVersion: row.OSVersion, OSArchitecture: row.OSArchitecture}
}

func auditSnapshot(order mysqlorder.Order) map[string]any {
	return map[string]any{"order_no": order.OrderNo, "status": order.Status, "order_type": order.OrderType, "payment_status": order.PaymentStatus, "related_instance_no": order.RelatedInstanceNo, "admin_note": order.AdminNote, "cancel_reason": order.CancelReason, "closed_reason": order.ClosedReason}
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
	noticeTask := mysqlinstance.Task{TaskNo: fmt.Sprintf("TASK-%d", time.Now().UnixNano()), TaskType: "instance_expiry_notice", IdempotencyKey: &noticeKey, Status: "pending", ObjectType: &objectType, ObjectNo: &objectNo, Payload: stringPtr(string(data)), MaxAttempts: 10, ScheduledAt: noticeAt}
	if err := s.instances.CreateTaskIgnoreDuplicate(ctx, tx, &noticeTask); err != nil {
		return err
	}
	if !s.lifecycle.AutoReleaseEnabled {
		return nil
	}
	releaseKey := "expiry_release:" + objectNo + ":" + expiresAt.Format(time.RFC3339Nano)
	releaseTask := mysqlinstance.Task{TaskNo: fmt.Sprintf("TASK-%d", time.Now().UnixNano()+1), TaskType: "instance_expiry_release", IdempotencyKey: &releaseKey, Status: "pending", ObjectType: &objectType, ObjectNo: &objectNo, Payload: stringPtr(string(data)), MaxAttempts: 10, ScheduledAt: normalizeDBTime(expiresAt.Add(time.Duration(s.lifecycle.ExpireReleaseAfterSeconds) * time.Second))}
	return s.instances.CreateTaskIgnoreDuplicate(ctx, tx, &releaseTask)
}

func stringPtr(value string) *string {
	return &value
}

func renewalExpiresAt(now time.Time, currentExpiresAt *time.Time, months int) time.Time {
	base := now
	// 未到期实例必须从原到期时间顺延，避免用户提前续费时损失剩余服务期。
	if currentExpiresAt != nil && currentExpiresAt.After(now) {
		base = *currentExpiresAt
	}
	return normalizeDBTime(base.AddDate(0, months, 0))
}

func renewalInstanceUpdates(lifecycle config.InstanceLifecycleConfig, nextExpiresAt time.Time) map[string]any {
	updates := map[string]any{"expires_at": nextExpiresAt, "expire_notice_sent_at": nil}
	if lifecycle.AutoReleaseEnabled {
		updates["expire_release_scheduled_at"] = normalizeDBTime(nextExpiresAt.Add(time.Duration(lifecycle.ExpireReleaseAfterSeconds) * time.Second))
	} else {
		updates["expire_release_scheduled_at"] = nil
	}
	return updates
}

func firstNonEmptyValue(value *string, fallback string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return fallback
	}
	return strings.TrimSpace(*value)
}

func normalizeDBTime(value time.Time) time.Time {
	return value.Truncate(time.Millisecond)
}
