package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"pvecloud/backend/internal/model"
)

// TicketRepository 封装工单及回复的数据访问。
type TicketRepository struct {
	db *gorm.DB
}

// AdminTicketView 表示后台工单列表项，附带用户信息。
type AdminTicketView struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"user_id"`
	UserEmail  string    `json:"user_email"`
	InstanceID *uint     `json:"instance_id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Priority   string    `json:"priority"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// NewTicketRepository 创建工单仓储。
func NewTicketRepository(db *gorm.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

// CreateTicket 创建工单。
func (r *TicketRepository) CreateTicket(ctx context.Context, ticket *model.Ticket) error {
	return r.db.WithContext(ctx).Create(ticket).Error
}

// GetTicketByID 查询工单。
func (r *TicketRepository) GetTicketByID(ctx context.Context, id uint) (*model.Ticket, error) {
	var ticket model.Ticket
	err := r.db.WithContext(ctx).First(&ticket, id).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

// UpdateTicket 更新工单。
func (r *TicketRepository) UpdateTicket(ctx context.Context, ticket *model.Ticket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

// ListUserTickets 查询用户自己的工单。
func (r *TicketRepository) ListUserTickets(ctx context.Context, userID uint, status string) ([]model.Ticket, error) {
	var tickets []model.Ticket
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC").Find(&tickets).Error
	return tickets, err
}

// ListAdminTickets 查询后台工单，并附带提交用户信息。
func (r *TicketRepository) ListAdminTickets(ctx context.Context, status string) ([]AdminTicketView, error) {
	var tickets []AdminTicketView
	query := r.db.WithContext(ctx).
		Table("tickets t").
		Select("t.id, t.user_id, t.instance_id, t.title, t.content, t.priority, t.status, t.created_at, t.updated_at, u.email AS user_email").
		Joins("LEFT JOIN users u ON u.id = t.user_id")
	if status != "" {
		query = query.Where("t.status = ?", status)
	}
	err := query.Order("t.created_at DESC").Find(&tickets).Error
	return tickets, err
}

// CreateReply 创建工单回复。
func (r *TicketRepository) CreateReply(ctx context.Context, reply *model.TicketReply) error {
	return r.db.WithContext(ctx).Create(reply).Error
}

// ListReplies 查询工单回复。
func (r *TicketRepository) ListReplies(ctx context.Context, ticketID uint) ([]model.TicketReply, error) {
	var replies []model.TicketReply
	err := r.db.WithContext(ctx).Where("ticket_id = ?", ticketID).Order("created_at ASC").Find(&replies).Error
	return replies, err
}
