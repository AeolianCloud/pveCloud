package ticket

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct{ db *gorm.DB }

type ListFilters struct {
	UserID          uint64
	Status          string
	Category        string
	Priority        string
	TicketNo        string
	OrderNo         string
	InstanceNo      string
	UserKeyword     string
	DateFrom        string
	DateTo          string
	AssigneeAdminID uint64
	TagID           uint64
	SLAStatus       string
}

type TagListFilters struct {
	Keyword    string
	Visibility string
	Status     string
}

type AssigneeCandidateFilters struct {
	Keyword string
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) CreateTicket(ctx context.Context, db *gorm.DB, ticket *Ticket) error {
	return r.queryDB(db).WithContext(ctx).Create(ticket).Error
}

func (r *Repository) CreateMessage(ctx context.Context, db *gorm.DB, message *TicketMessage) error {
	return r.queryDB(db).WithContext(ctx).Create(message).Error
}

func (r *Repository) CreateAttachment(ctx context.Context, db *gorm.DB, attachment *TicketMessageAttachment) error {
	return r.queryDB(db).WithContext(ctx).Create(attachment).Error
}

func (r *Repository) CreateEvent(ctx context.Context, db *gorm.DB, event *TicketEvent) error {
	return r.queryDB(db).WithContext(ctx).Create(event).Error
}

func (r *Repository) CreateInternalNote(ctx context.Context, db *gorm.DB, note *TicketInternalNote) error {
	return r.queryDB(db).WithContext(ctx).Create(note).Error
}

func (r *Repository) CreateCollaborator(ctx context.Context, db *gorm.DB, collaborator *TicketCollaborator) error {
	return r.queryDB(db).WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(collaborator).Error
}

func (r *Repository) DeleteCollaborator(ctx context.Context, db *gorm.DB, ticketID uint64, adminID uint64) error {
	return r.queryDB(db).WithContext(ctx).Where("ticket_id = ? AND admin_id = ?", ticketID, adminID).Delete(&TicketCollaborator{}).Error
}

func (r *Repository) ReplaceTagBindings(ctx context.Context, db *gorm.DB, ticketID uint64, tagIDs []uint64, adminID uint64) error {
	target := r.queryDB(db).WithContext(ctx)
	if err := target.Where("ticket_id = ?", ticketID).Delete(&TicketTagBinding{}).Error; err != nil {
		return err
	}
	for _, tagID := range uniqueUint64s(tagIDs) {
		binding := TicketTagBinding{TicketID: ticketID, TagID: tagID, CreatedByAdminID: &adminID}
		if err := target.Create(&binding).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) CreateTag(ctx context.Context, db *gorm.DB, tag *TicketTag) error {
	return r.queryDB(db).WithContext(ctx).Create(tag).Error
}

func (r *Repository) UpdateTag(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Model(&TicketTag{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) UpdateTicket(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Model(&Ticket{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) TicketForUpdate(ctx context.Context, db *gorm.DB, ticketNo string) (Ticket, error) {
	var ticket Ticket
	err := r.queryDB(db).WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("ticket_no = ?", ticketNo).First(&ticket).Error
	return ticket, err
}

func (r *Repository) UserTicket(ctx context.Context, userID uint64, ticketNo string) (TicketRow, error) {
	var row TicketRow
	err := r.baseDetailQuery(ctx).Where("tickets.user_id = ? AND tickets.ticket_no = ?", userID, ticketNo).Take(&row).Error
	return row, err
}

func (r *Repository) Detail(ctx context.Context, ticketNo string) (TicketRow, error) {
	var row TicketRow
	err := r.baseDetailQuery(ctx).Where("tickets.ticket_no = ?", ticketNo).Take(&row).Error
	return row, err
}

func (r *Repository) FindTagByID(ctx context.Context, db *gorm.DB, id uint64) (TicketTag, error) {
	var tag TicketTag
	err := r.queryDB(db).WithContext(ctx).Where("id = ?", id).First(&tag).Error
	return tag, err
}

func (r *Repository) ActiveTagsByIDs(ctx context.Context, ids []uint64) ([]TicketTag, error) {
	var tags []TicketTag
	if len(ids) == 0 {
		return tags, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ? AND status = ?", uniqueUint64s(ids), "active").Find(&tags).Error
	return tags, err
}

func (r *Repository) CountTagsByName(ctx context.Context, db *gorm.DB, excludeID uint64, name string) (int64, error) {
	var count int64
	query := r.queryDB(db).WithContext(ctx).Model(&TicketTag{}).Where("name = ?", strings.TrimSpace(name))
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *Repository) List(ctx context.Context, filters ListFilters, limit, offset int) ([]TicketRow, int64, error) {
	query := r.applyFilters(r.baseDetailQuery(ctx), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []TicketRow
	if err := query.Order("tickets.last_message_at DESC, tickets.id DESC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repository) Tags(ctx context.Context, filters TagListFilters, limit, offset int) ([]TicketTag, int64, error) {
	query := r.applyTagFilters(r.db.WithContext(ctx).Model(&TicketTag{}), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var tags []TicketTag
	if err := query.Order("sort_order ASC, id ASC").Limit(limit).Offset(offset).Find(&tags).Error; err != nil {
		return nil, 0, err
	}
	return tags, total, nil
}

func (r *Repository) AssigneeCandidates(ctx context.Context, filters AssigneeCandidateFilters, limit, offset int) ([]AdminSummary, int64, error) {
	base := r.assigneeCandidateQuery(ctx, filters)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []AdminSummary
	err := r.assigneeCandidateQuery(ctx, filters).
		Order("admin_users.id ASC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) IsAssignableAdmin(ctx context.Context, adminID uint64) (bool, error) {
	var count int64
	err := r.assigneeCandidateQuery(ctx, AssigneeCandidateFilters{}).Where("admin_users.id = ?", adminID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) Messages(ctx context.Context, ticketID uint64) ([]MessageRow, error) {
	var rows []MessageRow
	err := r.db.WithContext(ctx).Table("ticket_messages").
		Select(`ticket_messages.*,
			users.username AS username, users.email AS user_email, users.display_name AS user_display_name,
			admin_users.username AS admin_username, admin_users.display_name AS admin_display_name`).
		Joins("LEFT JOIN users ON users.id = ticket_messages.sender_user_id").
		Joins("LEFT JOIN admin_users ON admin_users.id = ticket_messages.sender_admin_id").
		Where("ticket_messages.ticket_id = ?", ticketID).
		Order("ticket_messages.created_at ASC, ticket_messages.id ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *Repository) MessageAttachments(ctx context.Context, ticketID uint64) ([]MessageAttachmentRow, error) {
	var rows []MessageAttachmentRow
	err := r.db.WithContext(ctx).Table("ticket_message_attachments").
		Select("ticket_message_attachments.*, file_attachments.original_name, file_attachments.mime_type, file_attachments.extension, file_attachments.size").
		Joins("JOIN file_attachments ON file_attachments.id = ticket_message_attachments.file_id").
		Where("ticket_message_attachments.ticket_id = ? AND file_attachments.status = ?", ticketID, "active").
		Order("ticket_message_attachments.message_id ASC, ticket_message_attachments.sort_order ASC, ticket_message_attachments.id ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *Repository) TagsByTicketIDs(ctx context.Context, ticketIDs []uint64, publicOnly bool) (map[uint64][]TicketTag, error) {
	result := make(map[uint64][]TicketTag)
	if len(ticketIDs) == 0 {
		return result, nil
	}
	var rows []TicketTagRow
	query := r.db.WithContext(ctx).Table("ticket_tag_bindings").
		Select("ticket_tags.*, ticket_tag_bindings.ticket_id").
		Joins("JOIN ticket_tags ON ticket_tags.id = ticket_tag_bindings.tag_id").
		Where("ticket_tag_bindings.ticket_id IN ?", uniqueUint64s(ticketIDs)).
		Order("ticket_tags.sort_order ASC, ticket_tags.id ASC")
	if publicOnly {
		query = query.Where("ticket_tags.visibility = ?", "public")
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.TicketID == nil {
			continue
		}
		result[*row.TicketID] = append(result[*row.TicketID], row.TicketTag)
	}
	return result, nil
}

func (r *Repository) InternalNotes(ctx context.Context, ticketID uint64) ([]TicketNoteRow, error) {
	var rows []TicketNoteRow
	err := r.db.WithContext(ctx).Table("ticket_internal_notes").
		Select("ticket_internal_notes.*, admin_users.username AS admin_username, admin_users.email AS admin_email, admin_users.display_name AS admin_display_name").
		Joins("JOIN admin_users ON admin_users.id = ticket_internal_notes.admin_id").
		Where("ticket_internal_notes.ticket_id = ?", ticketID).
		Order("ticket_internal_notes.created_at ASC, ticket_internal_notes.id ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *Repository) Collaborators(ctx context.Context, ticketID uint64) ([]TicketCollaboratorRow, error) {
	var rows []TicketCollaboratorRow
	err := r.db.WithContext(ctx).Table("ticket_collaborators").
		Select("ticket_collaborators.*, admin_users.username AS admin_username, admin_users.email AS admin_email, admin_users.display_name AS admin_display_name, admin_users.status AS admin_status").
		Joins("JOIN admin_users ON admin_users.id = ticket_collaborators.admin_id").
		Where("ticket_collaborators.ticket_id = ?", ticketID).
		Order("ticket_collaborators.created_at ASC, ticket_collaborators.id ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *Repository) Events(ctx context.Context, ticketID uint64) ([]TicketEventRow, error) {
	var rows []TicketEventRow
	err := r.db.WithContext(ctx).Table("ticket_events").
		Select(`ticket_events.*,
			admin_users.username AS actor_admin_username, admin_users.display_name AS actor_admin_display_name,
			users.username AS actor_user_username, users.display_name AS actor_user_display_name`).
		Joins("LEFT JOIN admin_users ON admin_users.id = ticket_events.actor_admin_id").
		Joins("LEFT JOIN users ON users.id = ticket_events.actor_user_id").
		Where("ticket_events.ticket_id = ?", ticketID).
		Order("ticket_events.created_at ASC, ticket_events.id ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *Repository) AttachmentBelongsToTicket(ctx context.Context, ticketID uint64, fileID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&TicketMessageAttachment{}).Where("ticket_id = ? AND file_id = ?", ticketID, fileID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) baseDetailQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table("tickets").
		Select(`tickets.*, users.username, users.email, users.display_name,
			assignee.username AS assignee_username,
			assignee.email AS assignee_email,
			assignee.display_name AS assignee_display_name`).
		Joins("JOIN users ON users.id = tickets.user_id").
		Joins("LEFT JOIN admin_users AS assignee ON assignee.id = tickets.assignee_admin_id")
}

func (r *Repository) applyFilters(db *gorm.DB, filters ListFilters) *gorm.DB {
	if filters.UserID > 0 {
		db = db.Where("tickets.user_id = ?", filters.UserID)
	}
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("tickets.status = ?", strings.TrimSpace(filters.Status))
	}
	if strings.TrimSpace(filters.Category) != "" {
		db = db.Where("tickets.category = ?", strings.TrimSpace(filters.Category))
	}
	if strings.TrimSpace(filters.Priority) != "" {
		db = db.Where("tickets.priority = ?", strings.TrimSpace(filters.Priority))
	}
	if strings.TrimSpace(filters.TicketNo) != "" {
		db = db.Where("tickets.ticket_no = ?", strings.TrimSpace(filters.TicketNo))
	}
	if strings.TrimSpace(filters.OrderNo) != "" {
		db = db.Where("tickets.order_no = ?", strings.TrimSpace(filters.OrderNo))
	}
	if strings.TrimSpace(filters.InstanceNo) != "" {
		db = db.Where("tickets.instance_no = ?", strings.TrimSpace(filters.InstanceNo))
	}
	if keyword := strings.TrimSpace(filters.UserKeyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("users.username LIKE ? OR users.email LIKE ? OR users.display_name LIKE ?", like, like, like)
	}
	if strings.TrimSpace(filters.DateFrom) != "" {
		db = db.Where("tickets.created_at >= ?", strings.TrimSpace(filters.DateFrom))
	}
	if strings.TrimSpace(filters.DateTo) != "" {
		db = db.Where("tickets.created_at <= ?", strings.TrimSpace(filters.DateTo))
	}
	if filters.AssigneeAdminID > 0 {
		db = db.Where("tickets.assignee_admin_id = ?", filters.AssigneeAdminID)
	}
	if filters.TagID > 0 {
		db = db.Joins("JOIN ticket_tag_bindings ON ticket_tag_bindings.ticket_id = tickets.id AND ticket_tag_bindings.tag_id = ?", filters.TagID)
	}
	switch strings.TrimSpace(filters.SLAStatus) {
	case "first_response_overdue":
		db = db.Where("tickets.first_responded_at IS NULL AND tickets.first_response_due_at IS NOT NULL AND tickets.first_response_due_at < ?", nowExpr())
	case "resolution_overdue":
		db = db.Where("tickets.resolved_at IS NULL AND tickets.resolution_due_at IS NOT NULL AND tickets.resolution_due_at < ?", nowExpr())
	case "normal":
		db = db.Where(`NOT (
			(tickets.first_responded_at IS NULL AND tickets.first_response_due_at IS NOT NULL AND tickets.first_response_due_at < ?)
			OR (tickets.resolved_at IS NULL AND tickets.resolution_due_at IS NOT NULL AND tickets.resolution_due_at < ?)
		)`, nowExpr(), nowExpr())
	}
	return db
}

func (r *Repository) applyTagFilters(db *gorm.DB, filters TagListFilters) *gorm.DB {
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}
	if strings.TrimSpace(filters.Visibility) != "" {
		db = db.Where("visibility = ?", strings.TrimSpace(filters.Visibility))
	}
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("status = ?", strings.TrimSpace(filters.Status))
	}
	return db
}

func (r *Repository) assigneeCandidateQuery(ctx context.Context, filters AssigneeCandidateFilters) *gorm.DB {
	query := r.db.WithContext(ctx).Table("admin_users").
		Select("DISTINCT admin_users.id, admin_users.username, admin_users.email, admin_users.display_name, admin_users.status").
		Joins("JOIN admin_user_roles ON admin_user_roles.admin_id = admin_users.id").
		Joins("JOIN admin_role_permissions ON admin_role_permissions.role_id = admin_user_roles.role_id").
		Joins("JOIN admin_permissions ON admin_permissions.id = admin_role_permissions.permission_id").
		Where("admin_users.deleted_at IS NULL").
		Where("admin_users.status = ?", "active").
		Where("admin_permissions.code IN ?", []string{"page.tickets", "ticket:reply", "ticket:*"})
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("admin_users.username LIKE ? OR admin_users.email LIKE ? OR admin_users.display_name LIKE ?", like, like, like)
	}
	return query
}

func (r *Repository) queryDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return r.db
}

func nowExpr() time.Time {
	return time.Now()
}

func uniqueUint64s(values []uint64) []uint64 {
	if len(values) == 0 {
		return nil
	}
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
