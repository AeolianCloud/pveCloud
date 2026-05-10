package order

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"

	domainorder "github.com/AeolianCloud/pveCloud/server/internal/domain/order"
	mysqlorder "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/order"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
	webrealname "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/realname"
)

const (
	defaultPage    = 1
	defaultPerPage = 15
	maxPerPage     = 100
)

type Service struct {
	db       *gorm.DB
	orders   *mysqlorder.Repository
	realName *webrealname.RealNameService
}

func NewService(db *gorm.DB, realName *webrealname.RealNameService) *Service {
	return &Service{db: db, orders: mysqlorder.NewRepository(db), realName: realName}
}

func (s *Service) Create(ctx context.Context, userID uint64, req webdto.OrderCreateRequest) (webdto.OrderDetail, error) {
	if req.Quantity == 0 {
		req.Quantity = 1
	}
	if req.Quantity != 1 {
		return webdto.OrderDetail{}, apperrors.ErrValidation.WithMessage("订单数量当前仅支持 1")
	}
	clientToken := strings.TrimSpace(req.ClientToken)
	if existing, err := s.orders.FindByUserClientToken(ctx, userID, clientToken); err == nil {
		return webOrderDetail(existing), nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return webdto.OrderDetail{}, err
	}
	if s.realName != nil {
		if err := s.realName.RequireApprovedForOrder(ctx, userID); err != nil {
			return webdto.OrderDetail{}, err
		}
	}
	selection, err := s.orders.CatalogSelection(ctx, strings.TrimSpace(req.PlanNo), strings.TrimSpace(req.BillingCycle), strings.TrimSpace(req.RegionNo), strings.TrimSpace(req.TemplateNo))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return webdto.OrderDetail{}, apperrors.ErrValidation.WithMessage("套餐、价格、地域或系统模板不可购买")
	}
	if err != nil {
		return webdto.OrderDetail{}, err
	}
	order := orderFromSelection(userID, clientToken, req, selection)
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error { return s.orders.Create(ctx, tx, &order) }); err != nil {
		if existing, findErr := s.orders.FindByUserClientToken(ctx, userID, clientToken); findErr == nil {
			return webOrderDetail(existing), nil
		}
		return webdto.OrderDetail{}, err
	}
	return webOrderDetail(order), nil
}

func (s *Service) List(ctx context.Context, userID uint64, query webdto.OrderListQuery) (webdto.PageResponse[webdto.OrderItem], error) {
	page, perPage := normalizePage(query.Page, query.PerPage)
	rows, total, err := s.orders.List(ctx, mysqlorder.ListFilters{UserID: userID, Status: query.Status}, perPage, (page-1)*perPage)
	if err != nil {
		return webdto.PageResponse[webdto.OrderItem]{}, err
	}
	items := make([]webdto.OrderItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, webOrderItem(row.Order))
	}
	return pageResponse(items, total, page, perPage), nil
}

func (s *Service) Detail(ctx context.Context, userID uint64, orderNo string) (webdto.OrderDetail, error) {
	order, err := s.orders.UserOrder(ctx, userID, strings.TrimSpace(orderNo))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return webdto.OrderDetail{}, apperrors.ErrNotFound.WithMessage("订单不存在")
	}
	if err != nil {
		return webdto.OrderDetail{}, err
	}
	return webOrderDetail(order), nil
}

func (s *Service) Cancel(ctx context.Context, userID uint64, orderNo string, req webdto.OrderCancelRequest) (webdto.OrderDetail, error) {
	var updated mysqlorder.Order
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.orders.OrderForUpdate(ctx, tx, strings.TrimSpace(orderNo))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("订单不存在")
		}
		if err != nil {
			return err
		}
		if current.UserID != userID {
			return apperrors.ErrNotFound.WithMessage("订单不存在")
		}
		if !domainorder.CanCancel(current.Status) {
			return apperrors.ErrConflict.WithMessage("当前订单状态不能取消")
		}
		now := time.Now()
		updates := map[string]any{"status": domainorder.StatusCancelled, "cancel_reason": textutil.NormalizeOptionalString(req.Reason), "cancelled_at": now}
		if err := s.orders.Update(ctx, tx, current.ID, updates); err != nil {
			return err
		}
		updated, err = s.orders.OrderForUpdate(ctx, tx, current.OrderNo)
		return err
	})
	if err != nil {
		return webdto.OrderDetail{}, err
	}
	return webOrderDetail(updated), nil
}

func orderFromSelection(userID uint64, clientToken string, req webdto.OrderCreateRequest, selection mysqlorder.CatalogSelection) mysqlorder.Order {
	return mysqlorder.Order{OrderNo: fmt.Sprintf("ORD-%d", time.Now().UnixNano()), UserID: userID, ClientToken: clientToken, Status: domainorder.StatusPending, ProductNo: selection.ProductNo, ProductType: selection.ProductType, ProductName: selection.ProductName, ProductSummary: selection.ProductSummary, PlanNo: selection.PlanNo, PlanCode: selection.PlanCode, PlanName: selection.PlanName, PlanSummary: selection.PlanSummary, CPUCores: selection.CPUCores, MemoryMB: selection.MemoryMB, SystemDiskGB: selection.SystemDiskGB, DataDiskGB: selection.DataDiskGB, BandwidthMbps: selection.BandwidthMbps, TrafficGB: selection.TrafficGB, PublicIPCount: selection.PublicIPCount, Virtualization: selection.Virtualization, Architecture: selection.Architecture, BillingCycle: selection.BillingCycle, PriceCents: selection.PriceCents, OriginalPriceCents: selection.OriginalPriceCents, Currency: selection.Currency, Quantity: 1, TotalAmountCents: selection.PriceCents, RegionNo: selection.RegionNo, RegionCode: selection.RegionCode, RegionName: selection.RegionName, TemplateNo: selection.TemplateNo, TemplateCode: selection.TemplateCode, TemplateName: selection.TemplateName, OSFamily: selection.OSFamily, OSDistribution: selection.OSDistribution, OSVersion: selection.OSVersion, OSArchitecture: selection.OSArchitecture, UserNote: textutil.NormalizeOptionalString(req.UserNote)}
}

func webOrderItem(order mysqlorder.Order) webdto.OrderItem {
	return webdto.OrderItem{OrderNo: order.OrderNo, Status: order.Status, ProductName: order.ProductName, PlanName: order.PlanName, BillingCycle: order.BillingCycle, TotalAmountCents: order.TotalAmountCents, Currency: order.Currency, CreatedAt: order.CreatedAt, CancelledAt: order.CancelledAt, ClosedAt: order.ClosedAt}
}

func webOrderDetail(order mysqlorder.Order) webdto.OrderDetail {
	return webdto.OrderDetail{OrderItem: webOrderItem(order), UserNote: order.UserNote, ProductNo: order.ProductNo, ProductType: order.ProductType, ProductSummary: order.ProductSummary, PlanNo: order.PlanNo, PlanCode: order.PlanCode, PlanSummary: order.PlanSummary, CPUCores: order.CPUCores, MemoryMB: order.MemoryMB, SystemDiskGB: order.SystemDiskGB, DataDiskGB: order.DataDiskGB, BandwidthMbps: order.BandwidthMbps, TrafficGB: order.TrafficGB, PublicIPCount: order.PublicIPCount, Virtualization: order.Virtualization, Architecture: order.Architecture, PriceCents: order.PriceCents, OriginalPriceCents: order.OriginalPriceCents, Quantity: order.Quantity, RegionNo: order.RegionNo, RegionCode: order.RegionCode, RegionName: order.RegionName, TemplateNo: order.TemplateNo, TemplateCode: order.TemplateCode, TemplateName: order.TemplateName, OSFamily: order.OSFamily, OSDistribution: order.OSDistribution, OSVersion: order.OSVersion, OSArchitecture: order.OSArchitecture}
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
