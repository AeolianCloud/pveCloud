package ticket

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	ticketObjectType        = "ticket"
	ticketReplyAction       = "ticket.reply"
	ticketCloseAction       = "ticket.close"
	ticketAssignAction      = "ticket.assign"
	ticketCollaborateAction = "ticket.collaborate"
	ticketNoteAction        = "ticket.note"
	ticketPriorityAction    = "ticket.priority_upgrade"
	ticketTagAction         = "ticket.tags_replace"
	ticketTagCreateAction   = "ticket.tag.create"
	ticketTagUpdateAction   = "ticket.tag.update"
	maxAttachmentsPerReply  = 5
	multipartOverheadBytes  = int64(1 << 20)
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
	rows, total, err := s.tickets.List(ctx, mysqlticket.ListFilters{Status: query.Status, Category: query.Category, Priority: query.Priority, TicketNo: query.TicketNo, OrderNo: query.OrderNo, InstanceNo: query.InstanceNo, UserKeyword: query.UserKeyword, DateFrom: query.DateFrom, DateTo: query.DateTo, AssigneeAdminID: query.AssigneeAdminID, TagID: query.TagID, SLAStatus: query.SLAStatus}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.AdminTicketItem]{}, err
	}
	tagsByTicket, err := s.tagsByRows(ctx, rows, false)
	if err != nil {
		return admindto.PageResponse[admindto.AdminTicketItem]{}, err
	}
	items := make([]admindto.AdminTicketItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, adminTicketItem(row, tagsByTicket[row.ID]))
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
		if current.FirstRespondedAt == nil {
			updates["first_responded_at"] = now
		}
		if err := s.tickets.UpdateTicket(ctx, tx, current.ID, updates); err != nil {
			return err
		}
		if err := s.recordEvent(ctx, tx, current.ID, domainticket.EventTypeAdminReply, &operatorID, nil, map[string]any{"message_id": message.ID, "attachment_count": len(uploads)}, "回复工单"); err != nil {
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
		updates := map[string]any{"status": domainticket.StatusClosed, "closed_by_type": sender, "closed_by_admin_id": operatorID, "closed_at": now, "resolved_at": now, "close_reason": textutil.NormalizeOptionalString(req.Reason)}
		if err := s.tickets.UpdateTicket(ctx, tx, current.ID, updates); err != nil {
			return err
		}
		if err := s.recordEvent(ctx, tx, current.ID, domainticket.EventTypeAdminClose, &operatorID, auditSnapshot(current), updates, optionalStringValue(req.Reason)); err != nil {
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

func (s *Service) AssigneeCandidates(ctx context.Context, query admindto.AssigneeCandidateQuery) (admindto.PageResponse[admindto.TicketAdminSummary], error) {
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	rows, total, err := s.tickets.AssigneeCandidates(ctx, mysqlticket.AssigneeCandidateFilters{Keyword: query.Keyword}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.TicketAdminSummary]{}, err
	}
	items := make([]admindto.TicketAdminSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, adminSummary(row))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func (s *Service) Assign(ctx context.Context, operatorID uint64, ticketNo string, req admindto.TicketAssignRequest) (admindto.AdminTicketDetail, error) {
	if ok, err := s.tickets.IsAssignableAdmin(ctx, req.AssigneeAdminID); err != nil {
		return admindto.AdminTicketDetail{}, err
	} else if !ok {
		return admindto.AdminTicketDetail{}, apperrors.ErrValidation.WithMessage("目标管理员不可指派")
	}
	var savedTicketNo string
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.lockOpenTicket(ctx, tx, ticketNo)
		if err != nil {
			return err
		}
		now := time.Now()
		eventType := domainticket.EventTypeAssign
		if current.AssigneeAdminID != nil {
			eventType = domainticket.EventTypeTransfer
		}
		updates := map[string]any{"assignee_admin_id": req.AssigneeAdminID, "assigned_by_admin_id": operatorID, "assigned_at": now}
		if err := s.tickets.UpdateTicket(ctx, tx, current.ID, updates); err != nil {
			return err
		}
		if err := s.recordEvent(ctx, tx, current.ID, eventType, &operatorID, auditSnapshot(current), updates, optionalStringValue(req.Reason)); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: ticketAssignAction, ObjectType: ticketObjectType, ObjectID: current.TicketNo, BeforeData: auditSnapshot(current), AfterData: updates, Remark: "指派工单"}); err != nil {
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

func (s *Service) AddCollaborator(ctx context.Context, operatorID uint64, ticketNo string, req admindto.TicketCollaboratorRequest) (admindto.AdminTicketDetail, error) {
	if ok, err := s.tickets.IsAssignableAdmin(ctx, req.AdminID); err != nil {
		return admindto.AdminTicketDetail{}, err
	} else if !ok {
		return admindto.AdminTicketDetail{}, apperrors.ErrValidation.WithMessage("目标管理员不可协作")
	}
	var savedTicketNo string
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.lockOpenTicket(ctx, tx, ticketNo)
		if err != nil {
			return err
		}
		row := mysqlticket.TicketCollaborator{TicketID: current.ID, AdminID: req.AdminID, CreatedByAdminID: &operatorID}
		if err := s.tickets.CreateCollaborator(ctx, tx, &row); err != nil {
			return err
		}
		after := map[string]any{"admin_id": req.AdminID}
		if err := s.recordEvent(ctx, tx, current.ID, domainticket.EventTypeCollaboratorAdd, &operatorID, nil, after, "添加协作者"); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: ticketCollaborateAction, ObjectType: ticketObjectType, ObjectID: current.TicketNo, AfterData: after, Remark: "添加工单协作者"}); err != nil {
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

func (s *Service) RemoveCollaborator(ctx context.Context, operatorID uint64, ticketNo string, adminID uint64) (admindto.AdminTicketDetail, error) {
	var savedTicketNo string
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.lockOpenTicket(ctx, tx, ticketNo)
		if err != nil {
			return err
		}
		if err := s.tickets.DeleteCollaborator(ctx, tx, current.ID, adminID); err != nil {
			return err
		}
		before := map[string]any{"admin_id": adminID}
		if err := s.recordEvent(ctx, tx, current.ID, domainticket.EventTypeCollaboratorRemove, &operatorID, before, nil, "移除协作者"); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: ticketCollaborateAction, ObjectType: ticketObjectType, ObjectID: current.TicketNo, BeforeData: before, Remark: "移除工单协作者"}); err != nil {
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

func (s *Service) AddInternalNote(ctx context.Context, operatorID uint64, ticketNo string, req admindto.TicketInternalNoteRequest) (admindto.AdminTicketDetail, error) {
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return admindto.AdminTicketDetail{}, apperrors.ErrValidation.WithMessage("内部备注不能为空")
	}
	var savedTicketNo string
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.ticketForUpdate(ctx, tx, ticketNo)
		if err != nil {
			return err
		}
		note := mysqlticket.TicketInternalNote{TicketID: current.ID, AdminID: operatorID, Content: content}
		if err := s.tickets.CreateInternalNote(ctx, tx, &note); err != nil {
			return err
		}
		after := map[string]any{"note_id": note.ID}
		if err := s.recordEvent(ctx, tx, current.ID, domainticket.EventTypeInternalNote, &operatorID, nil, after, "追加内部备注"); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: ticketNoteAction, ObjectType: ticketObjectType, ObjectID: current.TicketNo, AfterData: after, Remark: "追加工单内部备注"}); err != nil {
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

func (s *Service) UpgradePriority(ctx context.Context, operatorID uint64, ticketNo string, req admindto.TicketPriorityRequest) (admindto.AdminTicketDetail, error) {
	priority := strings.TrimSpace(req.Priority)
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return admindto.AdminTicketDetail{}, apperrors.ErrValidation.WithMessage("升级原因不能为空")
	}
	var savedTicketNo string
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.lockOpenTicket(ctx, tx, ticketNo)
		if err != nil {
			return err
		}
		if !domainticket.CanUpgradePriority(current.Priority, priority) {
			return apperrors.ErrConflict.WithMessage("只能升级到更高优先级")
		}
		firstDue, resolutionDue := slaDeadlines(current.CreatedAt, priority)
		updates := map[string]any{"priority": priority}
		if current.FirstRespondedAt == nil && deadlineEarlier(current.FirstResponseDueAt, firstDue) {
			updates["first_response_due_at"] = firstDue
		}
		if current.ResolvedAt == nil && deadlineEarlier(current.ResolutionDueAt, resolutionDue) {
			updates["resolution_due_at"] = resolutionDue
		}
		if err := s.tickets.UpdateTicket(ctx, tx, current.ID, updates); err != nil {
			return err
		}
		if err := s.recordEvent(ctx, tx, current.ID, domainticket.EventTypePriorityUpgrade, &operatorID, auditSnapshot(current), updates, reason); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: ticketPriorityAction, ObjectType: ticketObjectType, ObjectID: current.TicketNo, BeforeData: auditSnapshot(current), AfterData: updates, Remark: "升级工单优先级"}); err != nil {
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

func (s *Service) ReplaceTags(ctx context.Context, operatorID uint64, ticketNo string, req admindto.TicketTagsRequest) (admindto.AdminTicketDetail, error) {
	tagIDs := uniqueUint64s(req.TagIDs)
	if len(tagIDs) > 20 {
		return admindto.AdminTicketDetail{}, apperrors.ErrValidation.WithMessage("工单标签最多 20 个")
	}
	tags, err := s.tickets.ActiveTagsByIDs(ctx, tagIDs)
	if err != nil {
		return admindto.AdminTicketDetail{}, err
	}
	if len(tags) != len(tagIDs) {
		return admindto.AdminTicketDetail{}, apperrors.ErrValidation.WithMessage("存在不可用标签")
	}
	var savedTicketNo string
	err = mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.ticketForUpdate(ctx, tx, ticketNo)
		if err != nil {
			return err
		}
		if err := s.tickets.ReplaceTagBindings(ctx, tx, current.ID, tagIDs, operatorID); err != nil {
			return err
		}
		after := map[string]any{"tag_ids": tagIDs}
		if err := s.recordEvent(ctx, tx, current.ID, domainticket.EventTypeTagsReplace, &operatorID, nil, after, "更新工单标签"); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: ticketTagAction, ObjectType: ticketObjectType, ObjectID: current.TicketNo, AfterData: after, Remark: "更新工单标签"}); err != nil {
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

func (s *Service) Tags(ctx context.Context, query admindto.TicketTagListQuery) (admindto.PageResponse[admindto.TicketTagItem], error) {
	if !domainticket.IsKnownTagVisibilityOrEmpty(query.Visibility) || !domainticket.IsKnownTagStatusOrEmpty(query.Status) {
		return admindto.PageResponse[admindto.TicketTagItem]{}, apperrors.ErrValidation.WithMessage("标签筛选条件不支持")
	}
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	rows, total, err := s.tickets.Tags(ctx, mysqlticket.TagListFilters{Keyword: query.Keyword, Visibility: query.Visibility, Status: query.Status}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.TicketTagItem]{}, err
	}
	items := make([]admindto.TicketTagItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, tagItem(row))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func (s *Service) CreateTag(ctx context.Context, operatorID uint64, req admindto.TicketTagCreateRequest) (admindto.TicketTagItem, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return admindto.TicketTagItem{}, apperrors.ErrValidation.WithMessage("标签名称不能为空")
	}
	var created mysqlticket.TicketTag
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		if err := s.ensureTagNameUnique(ctx, tx, 0, name); err != nil {
			return err
		}
		created = mysqlticket.TicketTag{Name: name, Color: textutil.NormalizeOptionalString(req.Color), Visibility: req.Visibility, Status: req.Status, SortOrder: req.SortOrder, CreatedByAdminID: &operatorID, UpdatedByAdminID: &operatorID}
		if err := s.tickets.CreateTag(ctx, tx, &created); err != nil {
			return err
		}
		return s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: ticketTagCreateAction, ObjectType: "ticket_tag", ObjectID: textutil.Uint64String(created.ID), AfterData: tagAuditSnapshot(created), Remark: "创建工单标签"})
	})
	if err != nil {
		return admindto.TicketTagItem{}, err
	}
	return tagItem(created), nil
}

func (s *Service) UpdateTag(ctx context.Context, operatorID uint64, id uint64, req admindto.TicketTagUpdateRequest) (admindto.TicketTagItem, error) {
	var updated mysqlticket.TicketTag
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.tickets.FindTagByID(ctx, tx, id)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("标签不存在")
		}
		if err != nil {
			return err
		}
		updates := map[string]any{"updated_by_admin_id": operatorID}
		if req.Name != nil {
			name := strings.TrimSpace(*req.Name)
			if name == "" {
				return apperrors.ErrValidation.WithMessage("标签名称不能为空")
			}
			if err := s.ensureTagNameUnique(ctx, tx, id, name); err != nil {
				return err
			}
			updates["name"] = name
		}
		if req.Color != nil {
			updates["color"] = textutil.NormalizeOptionalString(req.Color)
		}
		if req.Visibility != nil {
			updates["visibility"] = strings.TrimSpace(*req.Visibility)
		}
		if req.Status != nil {
			updates["status"] = strings.TrimSpace(*req.Status)
		}
		if req.SortOrder != nil {
			updates["sort_order"] = *req.SortOrder
		}
		if err := s.tickets.UpdateTag(ctx, tx, id, updates); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: ticketTagUpdateAction, ObjectType: "ticket_tag", ObjectID: textutil.Uint64String(id), BeforeData: tagAuditSnapshot(current), AfterData: updates, Remark: "更新工单标签"}); err != nil {
			return err
		}
		updated, err = s.tickets.FindTagByID(ctx, tx, id)
		return err
	})
	if err != nil {
		return admindto.TicketTagItem{}, err
	}
	return tagItem(updated), nil
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
	tagsByTicket, err := s.tickets.TagsByTicketIDs(ctx, []uint64{row.ID}, false)
	if err != nil {
		return admindto.AdminTicketDetail{}, err
	}
	collaborators, err := s.tickets.Collaborators(ctx, row.ID)
	if err != nil {
		return admindto.AdminTicketDetail{}, err
	}
	notes, err := s.tickets.InternalNotes(ctx, row.ID)
	if err != nil {
		return admindto.AdminTicketDetail{}, err
	}
	events, err := s.tickets.Events(ctx, row.ID)
	if err != nil {
		return admindto.AdminTicketDetail{}, err
	}
	result := admindto.AdminTicketDetail{AdminTicketItem: adminTicketItem(row, tagsByTicket[row.ID]), CloseReason: row.CloseReason, Collaborators: collaboratorItems(collaborators), InternalNotes: noteItems(notes), Events: eventItems(events)}
	for _, message := range messages {
		result.Messages = append(result.Messages, admindto.AdminTicketMessage{ID: message.ID, SenderType: message.SenderType, SenderName: senderName(message), Content: message.Content, Attachments: byMessage[message.ID], CreatedAt: message.CreatedAt})
	}
	return result, nil
}

func (s *Service) ticketForUpdate(ctx context.Context, tx *gorm.DB, ticketNo string) (mysqlticket.Ticket, error) {
	current, err := s.tickets.TicketForUpdate(ctx, tx, strings.TrimSpace(ticketNo))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return mysqlticket.Ticket{}, apperrors.ErrNotFound.WithMessage("工单不存在")
	}
	return current, err
}

func (s *Service) lockOpenTicket(ctx context.Context, tx *gorm.DB, ticketNo string) (mysqlticket.Ticket, error) {
	current, err := s.ticketForUpdate(ctx, tx, ticketNo)
	if err != nil {
		return mysqlticket.Ticket{}, err
	}
	if current.Status == domainticket.StatusClosed {
		return mysqlticket.Ticket{}, apperrors.ErrConflict.WithMessage("当前工单已关闭")
	}
	return current, nil
}

func (s *Service) recordEvent(ctx context.Context, tx *gorm.DB, ticketID uint64, eventType string, actorAdminID *uint64, before any, after any, remark string) error {
	beforeJSON := jsonString(before)
	afterJSON := jsonString(after)
	event := mysqlticket.TicketEvent{TicketID: ticketID, EventType: eventType, ActorAdminID: actorAdminID, BeforeData: beforeJSON, AfterData: afterJSON, Remark: textutil.NormalizeOptionalString(&remark)}
	return s.tickets.CreateEvent(ctx, tx, &event)
}

func (s *Service) ensureTagNameUnique(ctx context.Context, tx *gorm.DB, excludeID uint64, name string) error {
	count, err := s.tickets.CountTagsByName(ctx, tx, excludeID, name)
	if err != nil {
		return err
	}
	if count > 0 {
		return apperrors.ErrConflict.WithMessage("标签名称已存在")
	}
	return nil
}

func (s *Service) tagsByRows(ctx context.Context, rows []mysqlticket.TicketRow, publicOnly bool) (map[uint64][]mysqlticket.TicketTag, error) {
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return s.tickets.TagsByTicketIDs(ctx, ids, publicOnly)
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

func adminTicketItem(row mysqlticket.TicketRow, tags []mysqlticket.TicketTag) admindto.AdminTicketItem {
	return admindto.AdminTicketItem{TicketNo: row.TicketNo, User: admindto.TicketUserSummary{ID: row.UserID, Username: row.Username, Email: row.Email, DisplayName: row.DisplayName}, Title: row.Title, Category: row.Category, Priority: row.Priority, Status: row.Status, Assignee: assigneeSummary(row), Tags: tagItems(tags), SLA: slaInfo(row.Ticket), OrderNo: row.OrderNo, InstanceNo: row.InstanceNo, LastMessageAt: row.LastMessageAt, CreatedAt: row.CreatedAt, ClosedAt: row.ClosedAt}
}

func assigneeSummary(row mysqlticket.TicketRow) *admindto.TicketAdminSummary {
	if row.AssigneeAdminID == nil {
		return nil
	}
	return &admindto.TicketAdminSummary{ID: *row.AssigneeAdminID, Username: stringValue(row.AssigneeUsername), Email: row.AssigneeEmail, DisplayName: stringValue(row.AssigneeDisplayName)}
}

func adminSummary(row mysqlticket.AdminSummary) admindto.TicketAdminSummary {
	return admindto.TicketAdminSummary{ID: row.ID, Username: row.Username, Email: row.Email, DisplayName: row.DisplayName, Status: row.Status}
}

func tagItem(tag mysqlticket.TicketTag) admindto.TicketTagItem {
	return admindto.TicketTagItem{ID: tag.ID, Name: tag.Name, Color: tag.Color, Visibility: tag.Visibility, Status: tag.Status, SortOrder: tag.SortOrder, CreatedAt: tag.CreatedAt, UpdatedAt: tag.UpdatedAt}
}

func tagItems(tags []mysqlticket.TicketTag) []admindto.TicketTagItem {
	items := make([]admindto.TicketTagItem, 0, len(tags))
	for _, tag := range tags {
		items = append(items, tagItem(tag))
	}
	return items
}

func collaboratorItems(rows []mysqlticket.TicketCollaboratorRow) []admindto.TicketAdminSummary {
	items := make([]admindto.TicketAdminSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, admindto.TicketAdminSummary{ID: row.AdminID, Username: row.AdminUsername, Email: row.AdminEmail, DisplayName: row.AdminDisplayName, Status: row.AdminStatus})
	}
	return items
}

func noteItems(rows []mysqlticket.TicketNoteRow) []admindto.TicketInternalNote {
	items := make([]admindto.TicketInternalNote, 0, len(rows))
	for _, row := range rows {
		items = append(items, admindto.TicketInternalNote{ID: row.ID, Admin: admindto.TicketAdminSummary{ID: row.AdminID, Username: row.AdminUsername, Email: row.AdminEmail, DisplayName: row.AdminDisplayName}, Content: row.Content, CreatedAt: row.CreatedAt})
	}
	return items
}

func eventItems(rows []mysqlticket.TicketEventRow) []admindto.TicketEvent {
	items := make([]admindto.TicketEvent, 0, len(rows))
	for _, row := range rows {
		items = append(items, admindto.TicketEvent{ID: row.ID, EventType: row.EventType, Actor: actorSummary(row), BeforeData: row.BeforeData, AfterData: row.AfterData, Remark: row.Remark, CreatedAt: row.CreatedAt})
	}
	return items
}

func actorSummary(row mysqlticket.TicketEventRow) *admindto.TicketActorSummary {
	if row.ActorAdminID != nil {
		display := row.ActorAdminDisplayName
		return &admindto.TicketActorSummary{Type: "admin", ID: *row.ActorAdminID, Username: stringValue(row.ActorAdminUsername), DisplayName: display}
	}
	if row.ActorUserID != nil {
		display := row.ActorUserDisplayName
		return &admindto.TicketActorSummary{Type: "user", ID: *row.ActorUserID, Username: stringValue(row.ActorUserUsername), DisplayName: display}
	}
	return nil
}

func slaInfo(ticket mysqlticket.Ticket) admindto.TicketSLAInfo {
	return admindto.TicketSLAInfo{FirstResponseDueAt: ticket.FirstResponseDueAt, FirstRespondedAt: ticket.FirstRespondedAt, ResolutionDueAt: ticket.ResolutionDueAt, ResolvedAt: ticket.ResolvedAt, Status: slaStatus(ticket)}
}

func slaStatus(ticket mysqlticket.Ticket) string {
	now := time.Now()
	if ticket.ResolvedAt == nil && ticket.ResolutionDueAt != nil && ticket.ResolutionDueAt.Before(now) {
		return domainticket.SLAStatusResolutionOverdue
	}
	if ticket.FirstRespondedAt == nil && ticket.FirstResponseDueAt != nil && ticket.FirstResponseDueAt.Before(now) {
		return domainticket.SLAStatusFirstResponseOverdue
	}
	return domainticket.SLAStatusNormal
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

func deadlineEarlier(current *time.Time, next time.Time) bool {
	return current == nil || next.Before(*current)
}

func tagAuditSnapshot(tag mysqlticket.TicketTag) map[string]any {
	return map[string]any{"id": tag.ID, "name": tag.Name, "color": tag.Color, "visibility": tag.Visibility, "status": tag.Status, "sort_order": tag.SortOrder}
}

func jsonString(value any) *string {
	if value == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	result := string(data)
	return &result
}

func optionalStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func uniqueUint64s(values []uint64) []uint64 {
	seen := make(map[uint64]struct{}, len(values))
	result := make([]uint64, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
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
	return map[string]any{"ticket_no": ticket.TicketNo, "status": ticket.Status, "priority": ticket.Priority, "category": ticket.Category, "assignee_admin_id": ticket.AssigneeAdminID, "close_reason": ticket.CloseReason}
}
