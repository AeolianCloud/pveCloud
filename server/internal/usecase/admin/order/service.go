package order

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	domainorder "github.com/AeolianCloud/pveCloud/server/internal/domain/order"
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
	db     *gorm.DB
	orders *mysqlorder.Repository
	audit  *AdminAuditService
}

func NewService(db *gorm.DB, audit *AdminAuditService) *Service {
	if audit == nil {
		audit = adminaudit.NewAdminAuditService(db)
	}
	return &Service{db: db, orders: mysqlorder.NewRepository(db), audit: audit}
}

func (s *Service) List(ctx context.Context, query admindto.OrderListQuery) (admindto.PageResponse[admindto.AdminOrderItem], error) {
	if !domainorder.IsKnownStatus(query.Status) {
		return admindto.PageResponse[admindto.AdminOrderItem]{}, apperrors.ErrValidation.WithMessage("订单状态不支持")
	}
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	rows, total, err := s.orders.List(ctx, mysqlorder.ListFilters{Status: query.Status, OrderNo: query.OrderNo, UserKeyword: query.UserKeyword, DateFrom: query.DateFrom, DateTo: query.DateTo}, perPage, (page-1)*perPage)
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

func (s *Service) update(ctx context.Context, operatorID uint64, orderNo string, build func(mysqlorder.Order) (map[string]any, string, error)) (admindto.AdminOrderDetail, error) {
	var updated mysqlorder.OrderRow
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
		updated, err = s.orders.Detail(ctx, current.OrderNo)
		return err
	})
	if err != nil {
		return admindto.AdminOrderDetail{}, err
	}
	return adminOrderDetail(updated), nil
}

func adminOrderItem(row mysqlorder.OrderRow) admindto.AdminOrderItem {
	return admindto.AdminOrderItem{OrderNo: row.OrderNo, User: admindto.OrderUserSummary{ID: row.UserID, Username: row.Username, Email: row.Email, DisplayName: row.DisplayName}, Status: row.Status, ProductName: row.ProductName, PlanName: row.PlanName, BillingCycle: row.BillingCycle, TotalAmountCents: row.TotalAmountCents, Currency: row.Currency, AdminNote: row.AdminNote, CreatedAt: row.CreatedAt, CancelledAt: row.CancelledAt, ClosedAt: row.ClosedAt}
}

func adminOrderDetail(row mysqlorder.OrderRow) admindto.AdminOrderDetail {
	return admindto.AdminOrderDetail{AdminOrderItem: adminOrderItem(row), UserNote: row.UserNote, CancelReason: row.CancelReason, ClosedReason: row.ClosedReason, ProductNo: row.ProductNo, ProductType: row.ProductType, ProductSummary: row.ProductSummary, PlanNo: row.PlanNo, PlanCode: row.PlanCode, PlanSummary: row.PlanSummary, CPUCores: row.CPUCores, MemoryMB: row.MemoryMB, SystemDiskGB: row.SystemDiskGB, DataDiskGB: row.DataDiskGB, BandwidthMbps: row.BandwidthMbps, TrafficGB: row.TrafficGB, PublicIPCount: row.PublicIPCount, Virtualization: row.Virtualization, Architecture: row.Architecture, PriceCents: row.PriceCents, OriginalPriceCents: row.OriginalPriceCents, Quantity: row.Quantity, RegionNo: row.RegionNo, RegionCode: row.RegionCode, RegionName: row.RegionName, TemplateNo: row.TemplateNo, TemplateCode: row.TemplateCode, TemplateName: row.TemplateName, OSFamily: row.OSFamily, OSDistribution: row.OSDistribution, OSVersion: row.OSVersion, OSArchitecture: row.OSArchitecture}
}

func auditSnapshot(order mysqlorder.Order) map[string]any {
	return map[string]any{"order_no": order.OrderNo, "status": order.Status, "admin_note": order.AdminNote, "cancel_reason": order.CancelReason, "closed_reason": order.ClosedReason}
}
