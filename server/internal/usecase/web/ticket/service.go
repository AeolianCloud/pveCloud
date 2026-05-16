package ticket

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
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
	mysqlorder "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/order"
	mysqlticket "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/ticket"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
	weblogging "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/logging"
)

const (
	defaultPage            = 1
	defaultPerPage         = 15
	maxPerPage             = 100
	maxAttachmentsPerReply = 5
	multipartOverheadBytes = int64(1 << 20)
)

type Service struct {
	db      *gorm.DB
	tickets *mysqlticket.Repository
	files   *mysqlfile.Repository
	orders  *mysqlorder.Repository
	config  config.StorageConfig
	logs    *weblogging.Recorder
}

func NewService(db *gorm.DB, storage config.StorageConfig) *Service {
	return &Service{
		db:      db,
		tickets: mysqlticket.NewRepository(db),
		files:   mysqlfile.NewRepository(db),
		orders:  mysqlorder.NewRepository(db),
		config:  storage,
		logs:    weblogging.NewRecorder(db),
	}
}

func (s *Service) Create(ctx context.Context, userID uint64, req webdto.TicketCreateRequest, headers []*multipart.FileHeader) (webdto.TicketDetail, error) {
	category := strings.TrimSpace(req.Category)
	priority := domainticket.NormalizePriority(strings.TrimSpace(req.Priority))
	if !domainticket.IsKnownCategory(category) || !domainticket.IsKnownPriority(priority) {
		return webdto.TicketDetail{}, apperrors.ErrValidation.WithMessage("工单分类或优先级不支持")
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return webdto.TicketDetail{}, apperrors.ErrValidation.WithMessage("工单内容不能为空")
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return webdto.TicketDetail{}, apperrors.ErrValidation.WithMessage("工单标题不能为空")
	}
	var orderID *uint64
	var orderNo *string
	if strings.TrimSpace(req.OrderNo) != "" {
		order, err := s.orders.UserOrder(ctx, userID, strings.TrimSpace(req.OrderNo))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return webdto.TicketDetail{}, apperrors.ErrValidation.WithMessage("关联订单不存在")
		}
		if err != nil {
			return webdto.TicketDetail{}, err
		}
		orderID = &order.ID
		orderNo = &order.OrderNo
	}
	uploads, err := s.prepareUploads(userID, headers)
	if err != nil {
		return webdto.TicketDetail{}, err
	}
	defer cleanupUploadsOnError(&err, uploads)

	now := time.Now()
	firstDue, resolutionDue := slaDeadlines(now, priority)
	ticket := mysqlticket.Ticket{TicketNo: fmt.Sprintf("TIC-%d", now.UnixNano()), UserID: userID, OrderID: orderID, OrderNo: orderNo, Category: category, Priority: priority, Title: title, Status: domainticket.StatusWaitingAdmin, LastMessageAt: now, LastUserMessageAt: &now, FirstResponseDueAt: &firstDue, ResolutionDueAt: &resolutionDue}
	err = mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		if err := s.tickets.CreateTicket(ctx, tx, &ticket); err != nil {
			return err
		}
		message := mysqlticket.TicketMessage{TicketID: ticket.ID, SenderType: domainticket.SenderUser, SenderUserID: &userID, Content: content}
		if err := s.tickets.CreateMessage(ctx, tx, &message); err != nil {
			return err
		}
		if err := s.persistUploads(ctx, tx, ticket, message, uploads); err != nil {
			return err
		}
		return s.logs.Business(ctx, tx, weblogging.Snapshot(userID, "", ""), "ticket", "ticket.create", "ticket", ticket.TicketNo, "工单创建")
	})
	if err != nil {
		return webdto.TicketDetail{}, err
	}
	return s.Detail(ctx, userID, ticket.TicketNo)
}

func (s *Service) List(ctx context.Context, userID uint64, query webdto.TicketListQuery) (webdto.PageResponse[webdto.TicketItem], error) {
	page, perPage := normalizePage(query.Page, query.PerPage)
	rows, total, err := s.tickets.List(ctx, mysqlticket.ListFilters{UserID: userID, Status: query.Status, Category: query.Category, Priority: query.Priority, OrderNo: query.OrderNo}, perPage, (page-1)*perPage)
	if err != nil {
		return webdto.PageResponse[webdto.TicketItem]{}, err
	}
	tagsByTicket, err := s.tagsByRows(ctx, rows)
	if err != nil {
		return webdto.PageResponse[webdto.TicketItem]{}, err
	}
	items := make([]webdto.TicketItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, webTicketItem(row.Ticket, tagsByTicket[row.ID]))
	}
	return pageResponse(items, total, page, perPage), nil
}

func (s *Service) Detail(ctx context.Context, userID uint64, ticketNo string) (webdto.TicketDetail, error) {
	row, err := s.tickets.UserTicket(ctx, userID, strings.TrimSpace(ticketNo))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return webdto.TicketDetail{}, apperrors.ErrNotFound.WithMessage("工单不存在")
	}
	if err != nil {
		return webdto.TicketDetail{}, err
	}
	return s.detailFromRow(ctx, row)
}

func (s *Service) Reply(ctx context.Context, userID uint64, ticketNo string, req webdto.TicketMessageRequest, headers []*multipart.FileHeader) (webdto.TicketDetail, error) {
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return webdto.TicketDetail{}, apperrors.ErrValidation.WithMessage("回复内容不能为空")
	}
	uploads, err := s.prepareUploads(userID, headers)
	if err != nil {
		return webdto.TicketDetail{}, err
	}
	defer cleanupUploadsOnError(&err, uploads)
	var savedTicketNo string
	err = mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.tickets.TicketForUpdate(ctx, tx, strings.TrimSpace(ticketNo))
		if errors.Is(err, gorm.ErrRecordNotFound) || current.UserID != userID {
			return apperrors.ErrNotFound.WithMessage("工单不存在")
		}
		if err != nil {
			return err
		}
		if !domainticket.CanReply(current.Status) {
			return apperrors.ErrConflict.WithMessage("当前工单已关闭，不能继续回复")
		}
		now := time.Now()
		message := mysqlticket.TicketMessage{TicketID: current.ID, SenderType: domainticket.SenderUser, SenderUserID: &userID, Content: content}
		if err := s.tickets.CreateMessage(ctx, tx, &message); err != nil {
			return err
		}
		if err := s.persistUploads(ctx, tx, current, message, uploads); err != nil {
			return err
		}
		if err := s.tickets.UpdateTicket(ctx, tx, current.ID, map[string]any{"status": domainticket.StatusWaitingAdmin, "last_message_at": now, "last_user_message_at": now}); err != nil {
			return err
		}
		savedTicketNo = current.TicketNo
		return s.logs.Business(ctx, tx, weblogging.Snapshot(userID, "", ""), "ticket", "ticket.reply", "ticket", current.TicketNo, "工单回复")
	})
	if err != nil {
		return webdto.TicketDetail{}, err
	}
	return s.Detail(ctx, userID, savedTicketNo)
}

func (s *Service) Close(ctx context.Context, userID uint64, ticketNo string, req webdto.TicketCloseRequest) (webdto.TicketDetail, error) {
	var savedTicketNo string
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.tickets.TicketForUpdate(ctx, tx, strings.TrimSpace(ticketNo))
		if errors.Is(err, gorm.ErrRecordNotFound) || current.UserID != userID {
			return apperrors.ErrNotFound.WithMessage("工单不存在")
		}
		if err != nil {
			return err
		}
		if !domainticket.CanClose(current.Status) {
			return apperrors.ErrConflict.WithMessage("当前工单已关闭")
		}
		now := time.Now()
		sender := domainticket.SenderUser
		if err := s.tickets.UpdateTicket(ctx, tx, current.ID, map[string]any{"status": domainticket.StatusClosed, "closed_by_type": sender, "closed_by_user_id": userID, "closed_at": now, "resolved_at": now, "close_reason": textutil.NormalizeOptionalString(req.Reason)}); err != nil {
			return err
		}
		savedTicketNo = current.TicketNo
		return s.logs.Business(ctx, tx, weblogging.Snapshot(userID, "", ""), "ticket", "ticket.close", "ticket", current.TicketNo, "工单关闭")
	})
	if err != nil {
		return webdto.TicketDetail{}, err
	}
	return s.Detail(ctx, userID, savedTicketNo)
}

func (s *Service) DownloadPath(ctx context.Context, userID uint64, ticketNo string, fileID uint64) (string, string, string, error) {
	row, err := s.tickets.UserTicket(ctx, userID, strings.TrimSpace(ticketNo))
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
	maxSize := s.config.MaxSize
	if maxSize <= 0 {
		return multipartOverheadBytes
	}
	return maxSize*maxAttachmentsPerReply + multipartOverheadBytes
}

func (s *Service) detailFromRow(ctx context.Context, row mysqlticket.TicketRow) (webdto.TicketDetail, error) {
	messages, err := s.tickets.Messages(ctx, row.ID)
	if err != nil {
		return webdto.TicketDetail{}, err
	}
	attachments, err := s.tickets.MessageAttachments(ctx, row.ID)
	if err != nil {
		return webdto.TicketDetail{}, err
	}
	byMessage := make(map[uint64][]webdto.TicketAttachment)
	for _, item := range attachments {
		byMessage[item.MessageID] = append(byMessage[item.MessageID], webdto.TicketAttachment{FileID: item.FileID, OriginalName: item.OriginalName, MimeType: item.MimeType, Extension: item.Extension, Size: item.Size, DownloadURL: fmt.Sprintf("/api/tickets/%s/attachments/%d/download", row.TicketNo, item.FileID)})
	}
	tagsByTicket, err := s.tickets.TagsByTicketIDs(ctx, []uint64{row.ID}, true)
	if err != nil {
		return webdto.TicketDetail{}, err
	}
	result := webdto.TicketDetail{TicketItem: webTicketItem(row.Ticket, tagsByTicket[row.ID]), CloseReason: row.CloseReason}
	for _, message := range messages {
		result.Messages = append(result.Messages, webdto.TicketMessage{ID: message.ID, SenderType: message.SenderType, SenderName: senderName(message), Content: message.Content, Attachments: byMessage[message.ID], CreatedAt: message.CreatedAt})
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
		refPath := fmt.Sprintf("/user/tickets/%s", ticket.TicketNo)
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

func (s *Service) prepareUploads(userID uint64, headers []*multipart.FileHeader) ([]storedUpload, error) {
	if len(headers) > maxAttachmentsPerReply {
		return nil, apperrors.ErrValidation.WithMessage("单条消息最多上传 5 个附件")
	}
	uploads := make([]storedUpload, 0, len(headers))
	for _, header := range headers {
		if header == nil {
			continue
		}
		upload, err := s.prepareUpload(userID, header)
		if err != nil {
			cleanupStoredUploads(uploads)
			return nil, err
		}
		uploads = append(uploads, upload)
	}
	return uploads, nil
}

func (s *Service) prepareUpload(userID uint64, header *multipart.FileHeader) (storedUpload, error) {
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

	mimeType := detectMimeType(sniff)
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
	storagePath := filepath.Join(
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()),
		storedName,
	)
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
			OriginalName:   originalName,
			StoredName:     storedName,
			MimeType:       mimeType,
			Extension:      ext,
			Size:           uint64(written),
			StoragePath:    storagePath,
			StorageDriver:  "local",
			Checksum:       hex.EncodeToString(hash.Sum(nil)),
			UploaderUserID: &userID,
			Status:         "active",
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

func validateUpload(originalName string, header *multipart.FileHeader, detectedMimeType string, allowedTypes []string) error {
	contentType := header.Header.Get("Content-Type")
	if err := domainfile.ValidateUpload(originalName, contentType, detectedMimeType, allowedTypes); err != nil {
		return apperrors.ErrValidation.WithMessage(err.Error())
	}
	return nil
}

func detectMimeType(sniff []byte) string {
	return http.DetectContentType(sniff)
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

func webTicketItem(ticket mysqlticket.Ticket, tags []mysqlticket.TicketTag) webdto.TicketItem {
	return webdto.TicketItem{TicketNo: ticket.TicketNo, Title: ticket.Title, Category: ticket.Category, Priority: ticket.Priority, Status: ticket.Status, Tags: webTagItems(tags), OrderNo: ticket.OrderNo, LastMessageAt: ticket.LastMessageAt, CreatedAt: ticket.CreatedAt, ClosedAt: ticket.ClosedAt}
}

func webTagItems(tags []mysqlticket.TicketTag) []webdto.TicketTagItem {
	items := make([]webdto.TicketTagItem, 0, len(tags))
	for _, tag := range tags {
		items = append(items, webdto.TicketTagItem{ID: tag.ID, Name: tag.Name, Color: tag.Color, Visibility: tag.Visibility})
	}
	return items
}

func (s *Service) tagsByRows(ctx context.Context, rows []mysqlticket.TicketRow) (map[uint64][]mysqlticket.TicketTag, error) {
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return s.tickets.TagsByTicketIDs(ctx, ids, true)
}

func slaDeadlines(createdAt time.Time, priority string) (time.Time, time.Time) {
	switch priority {
	case domainticket.PriorityLow:
		return createdAt.Add(48 * time.Hour), createdAt.Add(7 * 24 * time.Hour)
	case domainticket.PriorityHigh:
		return createdAt.Add(8 * time.Hour), createdAt.Add(3 * 24 * time.Hour)
	case domainticket.PriorityUrgent:
		return createdAt.Add(2 * time.Hour), createdAt.Add(24 * time.Hour)
	default:
		return createdAt.Add(24 * time.Hour), createdAt.Add(5 * 24 * time.Hour)
	}
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
