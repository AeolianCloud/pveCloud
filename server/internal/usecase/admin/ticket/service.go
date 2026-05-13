package ticket

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	domainfile "github.com/AeolianCloud/pveCloud/server/internal/domain/file"
	domainticket "github.com/AeolianCloud/pveCloud/server/internal/domain/ticket"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	mysqlfile "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/file"
	mysqlticket "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/ticket"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	adminaudit "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/audit"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	adminsupport "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/support"
)

const (
	ticketObjectType       = "ticket"
	ticketReplyAction      = "ticket.reply"
	ticketCloseAction      = "ticket.close"
	maxAttachmentsPerReply = 5
	multipartOverheadBytes = int64(1 << 20)
)

type AdminAuditService = adminaudit.AdminAuditService
type AdminAuditWriteInput = adminaudit.AdminAuditWriteInput

type Service struct {
	db      *gorm.DB
	tickets *mysqlticket.Repository
	files   *mysqlfile.Repository
	audit   *AdminAuditService
	config  config.StorageConfig
}

func NewService(db *gorm.DB, audit *AdminAuditService, storage config.StorageConfig) *Service {
	if audit == nil {
		audit = adminaudit.NewAdminAuditService(db)
	}
	return &Service{db: db, tickets: mysqlticket.NewRepository(db), files: mysqlfile.NewRepository(db), audit: audit, config: storage}
}

func (s *Service) List(ctx context.Context, query admindto.TicketListQuery) (admindto.PageResponse[admindto.AdminTicketItem], error) {
	if !domainticket.IsKnownStatus(query.Status) || !domainticket.IsKnownCategoryOrEmpty(query.Category) || !domainticket.IsKnownPriority(query.Priority) {
		return admindto.PageResponse[admindto.AdminTicketItem]{}, apperrors.ErrValidation.WithMessage("工单筛选条件不支持")
	}
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	rows, total, err := s.tickets.List(ctx, mysqlticket.ListFilters{Status: query.Status, Category: query.Category, Priority: query.Priority, TicketNo: query.TicketNo, OrderNo: query.OrderNo, UserKeyword: query.UserKeyword, DateFrom: query.DateFrom, DateTo: query.DateTo}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.AdminTicketItem]{}, err
	}
	items := make([]admindto.AdminTicketItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, adminTicketItem(row))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func (s *Service) Detail(ctx context.Context, ticketNo string) (admindto.AdminTicketDetail, error) {
	row, err := s.tickets.Detail(ctx, strings.TrimSpace(ticketNo))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return admindto.AdminTicketDetail{}, apperrors.ErrNotFound.WithMessage("工单不存在")
	}
	if err != nil {
		return admindto.AdminTicketDetail{}, err
	}
	return s.detailFromRow(ctx, row)
}

func (s *Service) Reply(ctx context.Context, operatorID uint64, ticketNo string, req admindto.TicketMessageRequest, headers []*multipart.FileHeader) (admindto.AdminTicketDetail, error) {
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return admindto.AdminTicketDetail{}, apperrors.ErrValidation.WithMessage("回复内容不能为空")
	}
	uploads, err := s.prepareUploads(operatorID, headers)
	if err != nil {
		return admindto.AdminTicketDetail{}, err
	}
	defer cleanupUploadsOnError(&err, uploads)
	var savedTicketNo string
	err = mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.tickets.TicketForUpdate(ctx, tx, strings.TrimSpace(ticketNo))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("工单不存在")
		}
		if err != nil {
			return err
		}
		if !domainticket.CanReply(current.Status) {
			return apperrors.ErrConflict.WithMessage("当前工单已关闭，不能继续回复")
		}
		now := time.Now()
		message := mysqlticket.TicketMessage{TicketID: current.ID, SenderType: domainticket.SenderAdmin, SenderAdminID: &operatorID, Content: content}
		if err := s.tickets.CreateMessage(ctx, tx, &message); err != nil {
			return err
		}
		if err := s.persistUploads(ctx, tx, current, message, uploads); err != nil {
			return err
		}
		updates := map[string]any{"status": domainticket.StatusWaitingUser, "last_message_at": now, "last_admin_message_at": now}
		if err := s.tickets.UpdateTicket(ctx, tx, current.ID, updates); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: ticketReplyAction, ObjectType: ticketObjectType, ObjectID: current.TicketNo, BeforeData: auditSnapshot(current), AfterData: map[string]any{"message_id": message.ID, "status": domainticket.StatusWaitingUser, "attachment_count": len(uploads)}, Remark: "回复工单"}); err != nil {
			return err
		}
		savedTicketNo = current.TicketNo
		return nil
	})
	if err != nil {
		return admindto.AdminTicketDetail{}, err
	}
	return s.Detail(ctx, savedTicketNo)
}

func (s *Service) Close(ctx context.Context, operatorID uint64, ticketNo string, req admindto.TicketCloseRequest) (admindto.AdminTicketDetail, error) {
	var savedTicketNo string
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.tickets.TicketForUpdate(ctx, tx, strings.TrimSpace(ticketNo))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("工单不存在")
		}
		if err != nil {
			return err
		}
		if !domainticket.CanClose(current.Status) {
			return apperrors.ErrConflict.WithMessage("当前工单已关闭")
		}
		now := time.Now()
		sender := domainticket.SenderAdmin
		updates := map[string]any{"status": domainticket.StatusClosed, "closed_by_type": sender, "closed_by_admin_id": operatorID, "closed_at": now, "close_reason": textutil.NormalizeOptionalString(req.Reason)}
		if err := s.tickets.UpdateTicket(ctx, tx, current.ID, updates); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: ticketCloseAction, ObjectType: ticketObjectType, ObjectID: current.TicketNo, BeforeData: auditSnapshot(current), AfterData: updates, Remark: "关闭工单"}); err != nil {
			return err
		}
		savedTicketNo = current.TicketNo
		return nil
	})
	if err != nil {
		return admindto.AdminTicketDetail{}, err
	}
	return s.Detail(ctx, savedTicketNo)
}

func (s *Service) DownloadPath(ctx context.Context, ticketNo string, fileID uint64) (string, string, string, error) {
	row, err := s.tickets.Detail(ctx, strings.TrimSpace(ticketNo))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", "", "", apperrors.ErrNotFound.WithMessage("附件不存在")
	}
	if err != nil {
		return "", "", "", err
	}
	ok, err := s.tickets.AttachmentBelongsToTicket(ctx, row.ID, fileID)
	if err != nil {
		return "", "", "", err
	}
	if !ok {
		return "", "", "", apperrors.ErrNotFound.WithMessage("附件不存在")
	}
	return s.attachmentPath(ctx, fileID)
}

func (s *Service) MaxUploadRequestBytes() int64 {
	if s.config.MaxSize <= 0 {
		return multipartOverheadBytes
	}
	return s.config.MaxSize*maxAttachmentsPerReply + multipartOverheadBytes
}

func (s *Service) detailFromRow(ctx context.Context, row mysqlticket.TicketRow) (admindto.AdminTicketDetail, error) {
	messages, err := s.tickets.Messages(ctx, row.ID)
	if err != nil {
		return admindto.AdminTicketDetail{}, err
	}
	attachments, err := s.tickets.MessageAttachments(ctx, row.ID)
	if err != nil {
		return admindto.AdminTicketDetail{}, err
	}
	byMessage := make(map[uint64][]admindto.AdminTicketAttachment)
	for _, item := range attachments {
		byMessage[item.MessageID] = append(byMessage[item.MessageID], admindto.AdminTicketAttachment{FileID: item.FileID, OriginalName: item.OriginalName, MimeType: item.MimeType, Extension: item.Extension, Size: item.Size, DownloadURL: fmt.Sprintf("/admin-api/tickets/%s/attachments/%d/download", row.TicketNo, item.FileID)})
	}
	result := admindto.AdminTicketDetail{AdminTicketItem: adminTicketItem(row), CloseReason: row.CloseReason}
	for _, message := range messages {
		result.Messages = append(result.Messages, admindto.AdminTicketMessage{ID: message.ID, SenderType: message.SenderType, SenderName: senderName(message), Content: message.Content, Attachments: byMessage[message.ID], CreatedAt: message.CreatedAt})
	}
	return result, nil
}

func (s *Service) persistUploads(ctx context.Context, tx *gorm.DB, ticket mysqlticket.Ticket, message mysqlticket.TicketMessage, uploads []storedUpload) error {
	for index := range uploads {
		upload := &uploads[index]
		if err := s.files.CreateAttachment(ctx, tx, &upload.Attachment); err != nil {
			return err
		}
		link := mysqlticket.TicketMessageAttachment{TicketID: ticket.ID, MessageID: message.ID, FileID: upload.Attachment.ID, SortOrder: index}
		if err := s.tickets.CreateAttachment(ctx, tx, &link); err != nil {
			return err
		}
		refName := ticket.TicketNo
		refPath := fmt.Sprintf("/tickets/%s", ticket.TicketNo)
		ref := mysqlfile.FileAttachmentReference{FileID: upload.Attachment.ID, RefType: "ticket_message", RefID: textutil.Uint64String(message.ID), RefName: &refName, RefPath: &refPath}
		if err := s.files.CreateReference(ctx, tx, &ref); err != nil {
			return err
		}
	}
	return nil
}

type storedUpload struct {
	Attachment   mysqlfile.FileAttachment
	AbsolutePath string
}

func (s *Service) prepareUploads(operatorID uint64, headers []*multipart.FileHeader) ([]storedUpload, error) {
	if len(headers) > maxAttachmentsPerReply {
		return nil, apperrors.ErrValidation.WithMessage("单条消息最多上传 5 个附件")
	}
	uploads := make([]storedUpload, 0, len(headers))
	for _, header := range headers {
		if header == nil {
			continue
		}
		upload, err := s.prepareUpload(operatorID, header)
		if err != nil {
			cleanupStoredUploads(uploads)
			return nil, err
		}
		uploads = append(uploads, upload)
	}
	return uploads, nil
}

func (s *Service) prepareUpload(operatorID uint64, header *multipart.FileHeader) (storedUpload, error) {
	originalName := sanitizeOriginalName(header.Filename)
	if originalName == "" {
		return storedUpload{}, apperrors.ErrValidation.WithMessage("文件名不能为空")
	}
	if header.Size > s.config.MaxSize {
		return storedUpload{}, apperrors.ErrValidation.WithMessage(fmt.Sprintf("文件大小超过限制，最大允许 %d 字节", s.config.MaxSize))
	}
	file, err := header.Open()
	if err != nil {
		return storedUpload{}, apperrors.ErrInternal.WithMessage("读取文件失败")
	}
	defer file.Close()

	sniff := make([]byte, 512)
	sniffSize, err := io.ReadFull(file, sniff)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return storedUpload{}, apperrors.ErrInternal.WithMessage("读取文件失败")
	}
	if sniffSize == 0 {
		return storedUpload{}, apperrors.ErrValidation.WithMessage("文件内容不能为空")
	}
	sniff = sniff[:sniffSize]
	if int64(sniffSize) > s.config.MaxSize {
		return storedUpload{}, apperrors.ErrValidation.WithMessage(fmt.Sprintf("文件大小超过限制，最大允许 %d 字节", s.config.MaxSize))
	}

	mimeType := http.DetectContentType(sniff)
	if err := validateUpload(originalName, header, mimeType, s.config.AllowedTypes); err != nil {
		return storedUpload{}, err
	}
	uuid, err := generateUUID()
	if err != nil {
		return storedUpload{}, apperrors.ErrInternal.WithMessage("生成文件名失败")
	}
	ext := strings.ToLower(filepath.Ext(originalName))
	if ext != "" {
		ext = ext[1:]
	}
	storedName := uuid + "." + ext
	now := time.Now()
	storagePath := filepath.Join(fmt.Sprintf("%04d", now.Year()), fmt.Sprintf("%02d", now.Month()), fmt.Sprintf("%02d", now.Day()), storedName)
	absolutePath, err := s.absoluteStoragePath(storagePath)
	if err != nil {
		return storedUpload{}, err
	}
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0755); err != nil {
		return storedUpload{}, apperrors.ErrInternal.WithMessage("创建存储目录失败")
	}
	out, err := os.OpenFile(absolutePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return storedUpload{}, apperrors.ErrInternal.WithMessage("保存文件失败")
	}
	hash := sha256.New()
	writer := io.MultiWriter(out, hash)
	writtenBytes, err := writer.Write(sniff)
	written := int64(writtenBytes)
	if err == nil {
		var copied int64
		copied, err = io.Copy(writer, io.LimitReader(file, s.config.MaxSize-written+1))
		written += copied
	}
	closeErr := out.Close()
	if err != nil || closeErr != nil {
		_ = os.Remove(absolutePath)
		return storedUpload{}, apperrors.ErrInternal.WithMessage("保存文件失败")
	}
	if written > s.config.MaxSize {
		_ = os.Remove(absolutePath)
		return storedUpload{}, apperrors.ErrValidation.WithMessage(fmt.Sprintf("文件大小超过限制，最大允许 %d 字节", s.config.MaxSize))
	}

	return storedUpload{
		AbsolutePath: absolutePath,
		Attachment: mysqlfile.FileAttachment{
			OriginalName:  originalName,
			StoredName:    storedName,
			MimeType:      mimeType,
			Extension:     ext,
			Size:          uint64(written),
			StoragePath:   storagePath,
			StorageDriver: "local",
			Checksum:      hex.EncodeToString(hash.Sum(nil)),
			UploaderID:    &operatorID,
			Status:        "active",
		},
	}, nil
}

func (s *Service) attachmentPath(ctx context.Context, fileID uint64) (string, string, string, error) {
	attachment, err := s.files.FindAttachmentByID(ctx, fileID)
	if errors.Is(err, gorm.ErrRecordNotFound) || attachment.Status != "active" {
		return "", "", "", apperrors.ErrNotFound.WithMessage("附件不存在")
	}
	if err != nil {
		return "", "", "", err
	}
	absolutePath, err := s.absoluteStoragePath(attachment.StoragePath)
	if err != nil {
		return "", "", "", err
	}
	if _, err := os.Stat(absolutePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", "", "", apperrors.ErrNotFound.WithMessage("附件不存在")
		}
		return "", "", "", err
	}
	return absolutePath, attachment.MimeType, attachment.OriginalName, nil
}

func (s *Service) absoluteStoragePath(storagePath string) (string, error) {
	cleanPath := filepath.Clean(strings.TrimSpace(storagePath))
	if !domainfile.IsSafeRelativeStoragePath(cleanPath) {
		return "", apperrors.ErrNotFound.WithMessage("附件不存在")
	}
	root, err := filepath.Abs(s.config.LocalPath)
	if err != nil {
		return "", err
	}
	target, err := filepath.Abs(filepath.Join(root, cleanPath))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", apperrors.ErrNotFound.WithMessage("附件不存在")
	}
	return target, nil
}

func cleanupUploadsOnError(err *error, uploads []storedUpload) {
	if err == nil || *err == nil {
		return
	}
	cleanupStoredUploads(uploads)
}

func cleanupStoredUploads(uploads []storedUpload) {
	for _, upload := range uploads {
		if upload.AbsolutePath != "" {
			_ = os.Remove(upload.AbsolutePath)
		}
	}
}

func validateUpload(originalName string, header *multipart.FileHeader, detectedMimeType string, allowedTypes []string) error {
	if err := domainfile.ValidateUpload(originalName, header.Header.Get("Content-Type"), detectedMimeType, allowedTypes); err != nil {
		return apperrors.ErrValidation.WithMessage(err.Error())
	}
	return nil
}

func sanitizeOriginalName(name string) string {
	name = strings.TrimSpace(filepath.Base(name))
	name = strings.ReplaceAll(name, "\x00", "")
	return name
}

func generateUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

func adminTicketItem(row mysqlticket.TicketRow) admindto.AdminTicketItem {
	return admindto.AdminTicketItem{TicketNo: row.TicketNo, User: admindto.TicketUserSummary{ID: row.UserID, Username: row.Username, Email: row.Email, DisplayName: row.DisplayName}, Title: row.Title, Category: row.Category, Priority: row.Priority, Status: row.Status, OrderNo: row.OrderNo, LastMessageAt: row.LastMessageAt, CreatedAt: row.CreatedAt, ClosedAt: row.ClosedAt}
}

func senderName(message mysqlticket.MessageRow) string {
	if message.SenderType == domainticket.SenderAdmin {
		if message.AdminDisplayName != nil && strings.TrimSpace(*message.AdminDisplayName) != "" {
			return *message.AdminDisplayName
		}
		if message.AdminUsername != nil {
			return *message.AdminUsername
		}
		return "客服"
	}
	if message.UserDisplayName != nil && strings.TrimSpace(*message.UserDisplayName) != "" {
		return *message.UserDisplayName
	}
	if message.Username != nil {
		return *message.Username
	}
	return "用户"
}

func auditSnapshot(ticket mysqlticket.Ticket) map[string]any {
	return map[string]any{"ticket_no": ticket.TicketNo, "status": ticket.Status, "priority": ticket.Priority, "category": ticket.Category, "close_reason": ticket.CloseReason}
}
