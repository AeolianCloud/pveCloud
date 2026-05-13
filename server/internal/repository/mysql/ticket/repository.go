package ticket

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct{ db *gorm.DB }

type ListFilters struct {
	UserID      uint64
	Status      string
	Category    string
	Priority    string
	TicketNo    string
	OrderNo     string
	UserKeyword string
	DateFrom    string
	DateTo      string
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

func (r *Repository) AttachmentBelongsToTicket(ctx context.Context, ticketID uint64, fileID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&TicketMessageAttachment{}).Where("ticket_id = ? AND file_id = ?", ticketID, fileID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) baseDetailQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table("tickets").
		Select("tickets.*, users.username, users.email, users.display_name").
		Joins("JOIN users ON users.id = tickets.user_id")
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
	return db
}

func (r *Repository) queryDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return r.db
}
