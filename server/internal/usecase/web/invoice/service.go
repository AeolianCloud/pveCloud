package invoice

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	mysqlerr "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	domainfile "github.com/AeolianCloud/pveCloud/server/internal/domain/file"
	domaininvoice "github.com/AeolianCloud/pveCloud/server/internal/domain/invoice"
	domainorder "github.com/AeolianCloud/pveCloud/server/internal/domain/order"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	mysqlfile "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/file"
	mysqlinvoice "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/invoice"
	mysqlorder "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/order"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
)

const (
	defaultPage    = 1
	defaultPerPage = 15
	maxPerPage     = 100
)

type Service struct {
	db       *gorm.DB
	invoices *mysqlinvoice.Repository
	files    *mysqlfile.Repository
	storage  config.StorageConfig
}

func NewService(db *gorm.DB, storage config.StorageConfig) *Service {
	return &Service{db: db, invoices: mysqlinvoice.NewRepository(db), files: mysqlfile.NewRepository(db), storage: storage}
}

func (s *Service) EligibleOrders(ctx context.Context, userID uint64, query webdto.InvoiceEligibleOrderQuery) (webdto.PageResponse[webdto.InvoiceEligibleOrderItem], error) {
	page, perPage := normalizePage(query.Page, query.PerPage)
	rows, total, err := s.invoices.EligibleOrders(ctx, mysqlinvoice.EligibleOrderFilters{UserID: userID, Keyword: query.Keyword, DateFrom: query.DateFrom, DateTo: query.DateTo}, perPage, (page-1)*perPage)
	if err != nil {
		return webdto.PageResponse[webdto.InvoiceEligibleOrderItem]{}, err
	}
	items := make([]webdto.InvoiceEligibleOrderItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, webdto.InvoiceEligibleOrderItem{
			OrderNo:           row.OrderNo,
			OrderType:         row.OrderType,
			RelatedInstanceNo: row.RelatedInstanceNo,
			AmountCents:       row.TotalAmountCents,
			Currency:          row.Currency,
			PaymentStatus:     row.PaymentStatus,
			PaidAt:            row.PaidAt,
			ProductName:       row.ProductName,
			PlanName:          row.PlanName,
			InvoiceOccupied:   row.InvoiceOccupied,
		})
	}
	return pageResponse(items, total, page, perPage), nil
}

func (s *Service) Create(ctx context.Context, userID uint64, req webdto.InvoiceCreateRequest) (webdto.InvoiceDetail, error) {
	titleType := strings.TrimSpace(req.TitleType)
	title := strings.TrimSpace(req.Title)
	taxNo := textutil.NormalizeOptionalString(req.TaxNo)
	if !domaininvoice.IsKnownTitleType(titleType) {
		return webdto.InvoiceDetail{}, apperrors.ErrValidation.WithMessage("发票抬头类型不支持")
	}
	if title == "" {
		return webdto.InvoiceDetail{}, apperrors.ErrValidation.WithMessage("发票抬头不能为空")
	}
	if titleType == domaininvoice.TitleTypeCompany && taxNo == nil {
		return webdto.InvoiceDetail{}, apperrors.ErrValidation.WithMessage("企业抬头必须填写税号")
	}
	clientToken := strings.TrimSpace(req.ClientToken)
	if clientToken == "" {
		return webdto.InvoiceDetail{}, apperrors.ErrValidation.WithMessage("缺少发票申请幂等键")
	}
	if existing, err := s.invoices.FindByUserClientToken(ctx, userID, clientToken); err == nil {
		return s.Detail(ctx, userID, existing.InvoiceNo)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return webdto.InvoiceDetail{}, err
	}

	orderNos := normalizeOrderNos(req.OrderNos)
	if len(orderNos) == 0 {
		return webdto.InvoiceDetail{}, apperrors.ErrValidation.WithMessage("请选择要开票的订单")
	}
	var createdNo string
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		if existing, err := s.invoices.FindByUserClientTokenInTx(ctx, tx, userID, clientToken); err == nil {
			createdNo = existing.InvoiceNo
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		orders, err := s.invoices.LockUserOrdersByNos(ctx, tx, userID, orderNos)
		if err != nil {
			return err
		}
		if len(orders) != len(orderNos) {
			return apperrors.ErrNotFound.WithMessage("订单不存在或不可开票")
		}
		if err := validateInvoiceOrders(ctx, s.invoices, tx, orders); err != nil {
			return err
		}
		amount := uint64(0)
		for _, order := range orders {
			amount += order.TotalAmountCents
		}

		now := time.Now()
		app := mysqlinvoice.Application{
			InvoiceNo:   fmt.Sprintf("INV-%d", now.UnixNano()),
			UserID:      userID,
			ClientToken: clientToken,
			InvoiceType: domaininvoice.TypeElectronicNormal,
			TitleType:   titleType,
			Title:       title,
			TaxNo:       taxNo,
			Email:       textutil.NormalizeOptionalString(req.Email),
			AmountCents: amount,
			Currency:    "CNY",
			Status:      domaininvoice.StatusPending,
			Remark:      textutil.NormalizeOptionalString(req.Remark),
		}
		if err := s.invoices.CreateApplication(ctx, tx, &app); err != nil {
			return err
		}
		for _, order := range orders {
			productName := textutil.StringPtr(order.ProductName)
			planName := textutil.StringPtr(order.PlanName)
			item := mysqlinvoice.ApplicationOrder{
				InvoiceID:        app.ID,
				InvoiceNo:        app.InvoiceNo,
				UserID:           userID,
				OrderID:          order.ID,
				OrderNo:          order.OrderNo,
				OrderType:        normalizedOrderType(order),
				OrderAmountCents: order.TotalAmountCents,
				Currency:         order.Currency,
				PaymentStatus:    order.PaymentStatus,
				PaidAt:           order.PaidAt,
				ProductName:      productName,
				PlanName:         planName,
				StatusSnapshot:   domaininvoice.StatusPending,
			}
			if err := s.invoices.CreateApplicationOrder(ctx, tx, &item); err != nil {
				return err
			}
		}
		createdNo = app.InvoiceNo
		return nil
	})
	if err != nil {
		if existing, findErr := s.invoices.FindByUserClientToken(ctx, userID, clientToken); findErr == nil {
			return s.Detail(ctx, userID, existing.InvoiceNo)
		}
		if isDuplicate(err) {
			return webdto.InvoiceDetail{}, apperrors.ErrConflict.WithMessage("订单已被有效发票申请占用")
		}
		return webdto.InvoiceDetail{}, err
	}
	return s.Detail(ctx, userID, createdNo)
}

func (s *Service) List(ctx context.Context, userID uint64, query webdto.InvoiceListQuery) (webdto.PageResponse[webdto.InvoiceItem], error) {
	if !domaininvoice.IsKnownStatus(query.Status) {
		return webdto.PageResponse[webdto.InvoiceItem]{}, apperrors.ErrValidation.WithMessage("发票状态不支持")
	}
	page, perPage := normalizePage(query.Page, query.PerPage)
	rows, total, err := s.invoices.UserList(ctx, mysqlinvoice.ListFilters{UserID: userID, Status: query.Status, DateFrom: query.DateFrom, DateTo: query.DateTo}, perPage, (page-1)*perPage)
	if err != nil {
		return webdto.PageResponse[webdto.InvoiceItem]{}, err
	}
	items := make([]webdto.InvoiceItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, webInvoiceItem(row.Application, row.OrderCount))
	}
	return pageResponse(items, total, page, perPage), nil
}

func (s *Service) Detail(ctx context.Context, userID uint64, invoiceNo string) (webdto.InvoiceDetail, error) {
	row, err := s.invoices.UserDetail(ctx, userID, strings.TrimSpace(invoiceNo))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return webdto.InvoiceDetail{}, apperrors.ErrNotFound.WithMessage("发票申请不存在")
	}
	if err != nil {
		return webdto.InvoiceDetail{}, err
	}
	orders, err := s.invoices.Orders(ctx, row.ID)
	if err != nil {
		return webdto.InvoiceDetail{}, err
	}
	return webInvoiceDetail(row.Application, row.OrderCount, orders), nil
}

func (s *Service) Cancel(ctx context.Context, userID uint64, invoiceNo string, req webdto.InvoiceCancelRequest) (webdto.InvoiceDetail, error) {
	var savedNo string
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.invoices.ApplicationForUpdate(ctx, tx, strings.TrimSpace(invoiceNo))
		if errors.Is(err, gorm.ErrRecordNotFound) || current.UserID != userID {
			return apperrors.ErrNotFound.WithMessage("发票申请不存在")
		}
		if err != nil {
			return err
		}
		if !domaininvoice.CanCancel(current.Status) {
			return apperrors.ErrConflict.WithMessage("当前发票状态不能取消")
		}
		now := time.Now().Truncate(time.Millisecond)
		updates := map[string]any{"status": domaininvoice.StatusCancelled, "cancel_reason": textutil.NormalizeOptionalString(req.Reason), "cancelled_at": now}
		if err := s.invoices.UpdateApplication(ctx, tx, current.ID, updates); err != nil {
			return err
		}
		// 订单占用依赖明细表状态快照的生成列，主状态变更必须同事务同步快照。
		if err := s.invoices.UpdateOrderStatusSnapshot(ctx, tx, current.ID, domaininvoice.StatusCancelled); err != nil {
			return err
		}
		savedNo = current.InvoiceNo
		return nil
	})
	if err != nil {
		return webdto.InvoiceDetail{}, err
	}
	return s.Detail(ctx, userID, savedNo)
}

func (s *Service) DownloadPath(ctx context.Context, userID uint64, invoiceNo string) (string, string, string, error) {
	row, err := s.invoices.UserDetail(ctx, userID, strings.TrimSpace(invoiceNo))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", "", "", apperrors.ErrNotFound.WithMessage("发票文件不存在")
	}
	if err != nil {
		return "", "", "", err
	}
	return s.invoiceFilePath(ctx, row.Application)
}

func validateInvoiceOrders(ctx context.Context, invoices *mysqlinvoice.Repository, tx *gorm.DB, orders []mysqlorder.Order) error {
	orderIDs := make([]uint64, 0, len(orders))
	for _, order := range orders {
		if order.Currency != "CNY" {
			return apperrors.ErrConflict.WithMessage("订单币种不可开票")
		}
		if order.PaymentStatus != domainorder.PaymentStatusPaid && order.PaymentStatus != domainorder.PaymentStatusManualConfirmed {
			return apperrors.ErrConflict.WithMessage("仅已支付订单可申请发票")
		}
		if order.Status == domainorder.StatusCancelled || order.Status == domainorder.StatusClosed {
			return apperrors.ErrConflict.WithMessage("当前订单状态不可开票")
		}
		orderIDs = append(orderIDs, order.ID)
	}
	if count, err := invoices.RefundBlockingOrderCount(ctx, tx, orderIDs); err != nil {
		return err
	} else if count > 0 {
		return apperrors.ErrConflict.WithMessage("订单已存在退款或退款处理中")
	}
	if count, err := invoices.ActiveOrderCount(ctx, tx, orderIDs); err != nil {
		return err
	} else if count > 0 {
		return apperrors.ErrConflict.WithMessage("订单已被有效发票申请占用")
	}
	return nil
}

func (s *Service) invoiceFilePath(ctx context.Context, app mysqlinvoice.Application) (string, string, string, error) {
	if app.Status != domaininvoice.StatusIssued || app.InvoiceFileID == nil {
		return "", "", "", apperrors.ErrConflict.WithMessage("当前发票状态不可下载")
	}
	ok, err := s.invoices.FileReferenceExists(ctx, *app.InvoiceFileID, app.ID)
	if err != nil {
		return "", "", "", err
	}
	if !ok {
		return "", "", "", apperrors.ErrNotFound.WithMessage("发票文件不存在")
	}
	attachment, err := s.files.FindAttachmentByID(ctx, *app.InvoiceFileID)
	if errors.Is(err, gorm.ErrRecordNotFound) || attachment.Status != "active" {
		return "", "", "", apperrors.ErrNotFound.WithMessage("发票文件不存在")
	}
	if err != nil {
		return "", "", "", err
	}
	if attachment.MimeType != "application/pdf" {
		return "", "", "", apperrors.ErrConflict.WithMessage("发票文件类型不正确")
	}
	absolutePath, err := safeStoragePath(s.storage.LocalPath, attachment.StoragePath)
	if err != nil {
		return "", "", "", err
	}
	if _, err := os.Stat(absolutePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", "", "", apperrors.ErrNotFound.WithMessage("发票文件不存在")
		}
		return "", "", "", err
	}
	return absolutePath, attachment.MimeType, attachment.OriginalName, nil
}

func webInvoiceItem(app mysqlinvoice.Application, orderCount int) webdto.InvoiceItem {
	item := webdto.InvoiceItem{
		InvoiceNo:     app.InvoiceNo,
		InvoiceType:   app.InvoiceType,
		TitleType:     app.TitleType,
		Title:         app.Title,
		AmountCents:   app.AmountCents,
		Currency:      app.Currency,
		Status:        app.Status,
		OrderCount:    orderCount,
		InvoiceNumber: app.InvoiceNumber,
		IssuedAt:      app.IssuedAt,
		CreatedAt:     app.CreatedAt,
		CanCancel:     domaininvoice.CanCancel(app.Status),
		CanDownload:   app.Status == domaininvoice.StatusIssued && app.InvoiceFileID != nil,
	}
	if item.CanDownload {
		downloadURL := fmt.Sprintf("/api/invoices/%s/download", app.InvoiceNo)
		item.DownloadURL = &downloadURL
	}
	return item
}

func webInvoiceDetail(app mysqlinvoice.Application, orderCount int, orders []mysqlinvoice.ApplicationOrder) webdto.InvoiceDetail {
	detail := webdto.InvoiceDetail{InvoiceItem: webInvoiceItem(app, orderCount), TaxNo: app.TaxNo, Email: app.Email, Remark: app.Remark, RejectReason: app.RejectReason, CancelReason: app.CancelReason, InvoiceCode: app.InvoiceCode, AcceptedAt: app.AcceptedAt, RejectedAt: app.RejectedAt, CancelledAt: app.CancelledAt}
	detail.Orders = make([]webdto.InvoiceOrderItem, 0, len(orders))
	for _, order := range orders {
		detail.Orders = append(detail.Orders, webdto.InvoiceOrderItem{OrderNo: order.OrderNo, OrderType: order.OrderType, OrderAmountCents: order.OrderAmountCents, Currency: order.Currency, PaymentStatus: order.PaymentStatus, PaidAt: order.PaidAt, ProductName: order.ProductName, PlanName: order.PlanName})
	}
	return detail
}

func normalizeOrderNos(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func normalizedOrderType(order mysqlorder.Order) string {
	if strings.TrimSpace(order.OrderType) == "" {
		return domainorder.TypePurchase
	}
	return order.OrderType
}

func safeStoragePath(rootPath string, storagePath string) (string, error) {
	cleanPath := filepath.Clean(strings.TrimSpace(storagePath))
	if !domainfile.IsSafeRelativeStoragePath(cleanPath) {
		return "", apperrors.ErrNotFound.WithMessage("发票文件不存在")
	}
	root, err := filepath.Abs(rootPath)
	if err != nil {
		return "", err
	}
	target, err := filepath.Abs(filepath.Join(root, cleanPath))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", apperrors.ErrNotFound.WithMessage("发票文件不存在")
	}
	return target, nil
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

func isDuplicate(err error) bool {
	var mysqlErr *mysqlerr.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
