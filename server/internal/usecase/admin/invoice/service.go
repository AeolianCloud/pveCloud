package invoice

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	domainfile "github.com/AeolianCloud/pveCloud/server/internal/domain/file"
	domaininvoice "github.com/AeolianCloud/pveCloud/server/internal/domain/invoice"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	mysqlfile "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/file"
	mysqlinvoice "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/invoice"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	adminaudit "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/audit"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	adminsupport "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/support"
)

type AdminAuditService = adminaudit.AdminAuditService
type AdminAuditWriteInput = adminaudit.AdminAuditWriteInput

type Service struct {
	db       *gorm.DB
	invoices *mysqlinvoice.Repository
	files    *mysqlfile.Repository
	audit    *AdminAuditService
	storage  config.StorageConfig
}

func NewService(db *gorm.DB, audit *AdminAuditService, storage config.StorageConfig) *Service {
	if audit == nil {
		audit = adminaudit.NewAdminAuditService(db)
	}
	return &Service{db: db, invoices: mysqlinvoice.NewRepository(db), files: mysqlfile.NewRepository(db), audit: audit, storage: storage}
}

func (s *Service) List(ctx context.Context, query admindto.InvoiceListQuery) (admindto.PageResponse[admindto.InvoiceItem], error) {
	if !domaininvoice.IsKnownStatus(query.Status) {
		return admindto.PageResponse[admindto.InvoiceItem]{}, apperrors.ErrValidation.WithMessage("发票状态不支持")
	}
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	rows, total, err := s.invoices.AdminList(ctx, mysqlinvoice.ListFilters{Status: query.Status, InvoiceNo: query.InvoiceNo, OrderNo: query.OrderNo, UserKeyword: query.UserKeyword, TitleKeyword: query.TitleKeyword, DateFrom: query.DateFrom, DateTo: query.DateTo}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.InvoiceItem]{}, err
	}
	items := make([]admindto.InvoiceItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, adminInvoiceItem(row))
	}
	return pageResponse(items, total, page, perPage), nil
}

func (s *Service) Detail(ctx context.Context, invoiceNo string) (admindto.InvoiceDetail, error) {
	row, err := s.invoices.Detail(ctx, strings.TrimSpace(invoiceNo))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return admindto.InvoiceDetail{}, apperrors.ErrNotFound.WithMessage("发票申请不存在")
	}
	if err != nil {
		return admindto.InvoiceDetail{}, err
	}
	orders, err := s.invoices.Orders(ctx, row.ID)
	if err != nil {
		return admindto.InvoiceDetail{}, err
	}
	var fileSummary *admindto.InvoiceFileSummary
	if row.InvoiceFileID != nil {
		summary, err := s.fileSummary(ctx, row.Application)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return admindto.InvoiceDetail{}, err
		}
		fileSummary = summary
	}
	return adminInvoiceDetail(row, orders, fileSummary), nil
}

func (s *Service) Accept(ctx context.Context, operatorID uint64, invoiceNo string) (admindto.InvoiceDetail, error) {
	return s.statusUpdate(ctx, operatorID, invoiceNo, "invoice.accept", func(current mysqlinvoice.Application) (map[string]any, string, error) {
		if !domaininvoice.CanAccept(current.Status) {
			return nil, "", apperrors.ErrConflict.WithMessage("当前发票状态不能受理")
		}
		now := time.Now().Truncate(time.Millisecond)
		return map[string]any{"status": domaininvoice.StatusProcessing, "accepted_by_admin_id": operatorID, "accepted_at": now}, domaininvoice.StatusProcessing, nil
	})
}

func (s *Service) Reject(ctx context.Context, operatorID uint64, invoiceNo string, req admindto.InvoiceRejectRequest) (admindto.InvoiceDetail, error) {
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return admindto.InvoiceDetail{}, apperrors.ErrValidation.WithMessage("驳回原因不能为空")
	}
	return s.statusUpdate(ctx, operatorID, invoiceNo, "invoice.reject", func(current mysqlinvoice.Application) (map[string]any, string, error) {
		if !domaininvoice.CanReject(current.Status) {
			return nil, "", apperrors.ErrConflict.WithMessage("当前发票状态不能驳回")
		}
		now := time.Now().Truncate(time.Millisecond)
		return map[string]any{"status": domaininvoice.StatusRejected, "reject_reason": reason, "rejected_by_admin_id": operatorID, "rejected_at": now}, domaininvoice.StatusRejected, nil
	})
}

func (s *Service) Issue(ctx context.Context, operatorID uint64, invoiceNo string, req admindto.InvoiceIssueRequest) (admindto.InvoiceDetail, error) {
	if strings.TrimSpace(req.InvoiceNumber) == "" {
		return admindto.InvoiceDetail{}, apperrors.ErrValidation.WithMessage("发票号码不能为空")
	}
	if req.IssuedAt.IsZero() {
		return admindto.InvoiceDetail{}, apperrors.ErrValidation.WithMessage("开票时间不能为空")
	}
	var savedNo string
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.invoices.ApplicationForUpdate(ctx, tx, strings.TrimSpace(invoiceNo))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("发票申请不存在")
		}
		if err != nil {
			return err
		}
		if !domaininvoice.CanIssue(current.Status) {
			return apperrors.ErrConflict.WithMessage("当前发票状态不能开票")
		}
		attachment, err := s.files.FindAttachmentByIDForUpdate(ctx, tx, req.FileID)
		if errors.Is(err, gorm.ErrRecordNotFound) || attachment.Status != "active" {
			return apperrors.ErrNotFound.WithMessage("发票PDF文件不存在")
		}
		if err != nil {
			return err
		}
		if attachment.MimeType != "application/pdf" {
			return apperrors.ErrValidation.WithMessage("发票文件必须是 PDF")
		}
		fileID := attachment.ID
		updates := map[string]any{
			"status":             domaininvoice.StatusIssued,
			"invoice_code":       textutil.NormalizeOptionalString(req.InvoiceCode),
			"invoice_number":     strings.TrimSpace(req.InvoiceNumber),
			"invoice_file_id":    fileID,
			"issued_by_admin_id": operatorID,
			"issued_at":          req.IssuedAt.Truncate(time.Millisecond),
		}
		if err := s.invoices.UpdateApplication(ctx, tx, current.ID, updates); err != nil {
			return err
		}
		// `issued` 仍占用订单可开票资格，快照同事务更新可以修复 processing->issued 的投影漂移。
		if err := s.invoices.UpdateOrderStatusSnapshot(ctx, tx, current.ID, domaininvoice.StatusIssued); err != nil {
			return err
		}
		refName := current.InvoiceNo
		refPath := fmt.Sprintf("/invoices/%s", current.InvoiceNo)
		ref := mysqlfile.FileAttachmentReference{FileID: fileID, RefType: domaininvoice.FileRefType, RefID: textutil.Uint64String(current.ID), RefName: &refName, RefPath: &refPath}
		if err := s.files.CreateReference(ctx, tx, &ref); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: "invoice.issue", ObjectType: "invoice", ObjectID: current.InvoiceNo, BeforeData: auditSnapshot(current), AfterData: map[string]any{"status": domaininvoice.StatusIssued, "invoice_number": req.InvoiceNumber, "file_id": fileID, "order_nos": orderNoSummary(ctx, s.invoices, current.ID)}, Remark: "登记发票"}); err != nil {
			return err
		}
		savedNo = current.InvoiceNo
		return nil
	})
	if err != nil {
		return admindto.InvoiceDetail{}, err
	}
	return s.Detail(ctx, savedNo)
}

func (s *Service) UpdateAdminNote(ctx context.Context, operatorID uint64, invoiceNo string, req admindto.InvoiceAdminNoteRequest) (admindto.InvoiceDetail, error) {
	var savedNo string
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.invoices.ApplicationForUpdate(ctx, tx, strings.TrimSpace(invoiceNo))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("发票申请不存在")
		}
		if err != nil {
			return err
		}
		updates := map[string]any{"admin_note": textutil.NormalizeOptionalString(req.AdminNote)}
		if err := s.invoices.UpdateApplication(ctx, tx, current.ID, updates); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: "invoice.admin_note.update", ObjectType: "invoice", ObjectID: current.InvoiceNo, BeforeData: map[string]any{"admin_note": current.AdminNote}, AfterData: updates, Remark: "更新发票后台备注"}); err != nil {
			return err
		}
		savedNo = current.InvoiceNo
		return nil
	})
	if err != nil {
		return admindto.InvoiceDetail{}, err
	}
	return s.Detail(ctx, savedNo)
}

func (s *Service) DownloadPath(ctx context.Context, invoiceNo string) (string, string, string, error) {
	row, err := s.invoices.Detail(ctx, strings.TrimSpace(invoiceNo))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", "", "", apperrors.ErrNotFound.WithMessage("发票文件不存在")
	}
	if err != nil {
		return "", "", "", err
	}
	return s.invoiceFilePath(ctx, row.Application)
}

func (s *Service) statusUpdate(ctx context.Context, operatorID uint64, invoiceNo string, action string, build func(mysqlinvoice.Application) (map[string]any, string, error)) (admindto.InvoiceDetail, error) {
	var savedNo string
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.invoices.ApplicationForUpdate(ctx, tx, strings.TrimSpace(invoiceNo))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("发票申请不存在")
		}
		if err != nil {
			return err
		}
		updates, nextStatus, err := build(current)
		if err != nil {
			return err
		}
		if err := s.invoices.UpdateApplication(ctx, tx, current.ID, updates); err != nil {
			return err
		}
		// 有效订单占用完全由明细状态快照驱动，所有状态流转必须与主表同事务落库。
		if err := s.invoices.UpdateOrderStatusSnapshot(ctx, tx, current.ID, nextStatus); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: action, ObjectType: "invoice", ObjectID: current.InvoiceNo, BeforeData: auditSnapshot(current), AfterData: map[string]any{"status": nextStatus, "order_nos": orderNoSummary(ctx, s.invoices, current.ID)}, Remark: "处理发票申请"}); err != nil {
			return err
		}
		savedNo = current.InvoiceNo
		return nil
	})
	if err != nil {
		return admindto.InvoiceDetail{}, err
	}
	return s.Detail(ctx, savedNo)
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

func (s *Service) fileSummary(ctx context.Context, app mysqlinvoice.Application) (*admindto.InvoiceFileSummary, error) {
	if app.InvoiceFileID == nil {
		return nil, nil
	}
	attachment, err := s.files.FindAttachmentByID(ctx, *app.InvoiceFileID)
	if err != nil {
		return nil, err
	}
	if attachment.Status != "active" {
		return nil, nil
	}
	return &admindto.InvoiceFileSummary{ID: attachment.ID, OriginalName: attachment.OriginalName, MimeType: attachment.MimeType, Size: attachment.Size, DownloadURL: fmt.Sprintf("/admin-api/invoices/%s/download", app.InvoiceNo)}, nil
}

func adminInvoiceItem(row mysqlinvoice.ApplicationRow) admindto.InvoiceItem {
	return admindto.InvoiceItem{
		InvoiceNo:     row.InvoiceNo,
		InvoiceType:   row.InvoiceType,
		User:          admindto.OrderUserSummary{ID: row.UserID, Username: row.Username, Email: row.UserEmail, DisplayName: row.DisplayName},
		TitleType:     row.TitleType,
		Title:         row.Title,
		AmountCents:   row.AmountCents,
		Currency:      row.Currency,
		Status:        row.Status,
		OrderCount:    row.OrderCount,
		InvoiceNumber: row.InvoiceNumber,
		CreatedAt:     row.CreatedAt,
		AcceptedAt:    row.AcceptedAt,
		IssuedAt:      row.IssuedAt,
	}
}

func adminInvoiceDetail(row mysqlinvoice.ApplicationRow, orders []mysqlinvoice.ApplicationOrder, file *admindto.InvoiceFileSummary) admindto.InvoiceDetail {
	detail := admindto.InvoiceDetail{InvoiceItem: adminInvoiceItem(row), TaxNo: row.TaxNo, Email: row.Email, Remark: row.Remark, AdminNote: row.AdminNote, RejectReason: row.RejectReason, CancelReason: row.CancelReason, InvoiceCode: row.InvoiceCode, RejectedAt: row.RejectedAt, CancelledAt: row.CancelledAt, File: file}
	detail.Orders = make([]admindto.InvoiceOrderItem, 0, len(orders))
	for _, order := range orders {
		detail.Orders = append(detail.Orders, admindto.InvoiceOrderItem{OrderNo: order.OrderNo, OrderType: order.OrderType, OrderAmountCents: order.OrderAmountCents, Currency: order.Currency, PaymentStatus: order.PaymentStatus, PaidAt: order.PaidAt, ProductName: order.ProductName, PlanName: order.PlanName})
	}
	return detail
}

func auditSnapshot(app mysqlinvoice.Application) map[string]any {
	return map[string]any{"invoice_no": app.InvoiceNo, "status": app.Status, "title_type": app.TitleType, "title": app.Title, "amount_cents": app.AmountCents, "currency": app.Currency, "invoice_number": app.InvoiceNumber}
}

func orderNoSummary(ctx context.Context, invoices *mysqlinvoice.Repository, invoiceID uint64) []string {
	rows, err := invoices.Orders(ctx, invoiceID)
	if err != nil {
		return nil
	}
	orderNos := make([]string, 0, len(rows))
	for _, row := range rows {
		orderNos = append(orderNos, row.OrderNo)
	}
	return orderNos
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

func pageResponse[T any](items []T, total int64, page, perPage int) admindto.PageResponse[T] {
	lastPage := 0
	if total > 0 {
		lastPage = int(math.Ceil(float64(total) / float64(perPage)))
	}
	return admindto.PageResponse[T]{List: items, Total: total, Page: page, PerPage: perPage, LastPage: lastPage}
}
