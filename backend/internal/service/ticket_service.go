package service

import (
	"context"
	"errors"

	"pvecloud/backend/internal/model"
	"pvecloud/backend/internal/repository"
)

// TicketService 负责工单创建、回复、状态流转。
type TicketService struct {
	ticketRepo *repository.TicketRepository
}

// NewTicketService 创建工单服务。
func NewTicketService(ticketRepo *repository.TicketRepository) *TicketService {
	return &TicketService{ticketRepo: ticketRepo}
}

// CreateTicket 创建用户工单。
func (s *TicketService) CreateTicket(ctx context.Context, userID uint, title, content, priority string, instanceID *uint) (*model.Ticket, error) {
	ticket := &model.Ticket{UserID: userID, InstanceID: instanceID, Title: title, Content: content, Priority: priority, Status: "open"}
	if ticket.Priority == "" {
		ticket.Priority = "medium"
	}
	if err := s.ticketRepo.CreateTicket(ctx, ticket); err != nil {
		return nil, err
	}
	return ticket, nil
}

// AddReply 新增工单回复，closed 状态不允许继续回复。
func (s *TicketService) AddReply(ctx context.Context, userID uint, ticketID uint, isAdmin bool, content string) error {
	ticket, err := s.ticketRepo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return err
	}
	if ticket.Status == "closed" {
		return errors.New("工单已关闭，无法继续回复")
	}
	if !isAdmin && ticket.UserID != userID {
		return WrapForbidden("无权限回复该工单")
	}
	if isAdmin && ticket.Status == "open" {
		ticket.Status = "processing"
		_ = s.ticketRepo.UpdateTicket(ctx, ticket)
	}
	reply := &model.TicketReply{TicketID: ticketID, UserID: userID, IsAdmin: isAdmin, Content: content}
	return s.ticketRepo.CreateReply(ctx, reply)
}

// ChangeStatus 修改工单状态，仅管理员可调用。
func (s *TicketService) ChangeStatus(ctx context.Context, ticketID uint, status string) error {
	ticket, err := s.ticketRepo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return err
	}
	ticket.Status = status
	return s.ticketRepo.UpdateTicket(ctx, ticket)
}

// ListUserTickets 查询当前用户工单。
func (s *TicketService) ListUserTickets(ctx context.Context, userID uint, status string) ([]model.Ticket, error) {
	return s.ticketRepo.ListUserTickets(ctx, userID, status)
}

// ListAdminTickets 查询后台工单。
func (s *TicketService) ListAdminTickets(ctx context.Context, status string) ([]repository.AdminTicketView, error) {
	return s.ticketRepo.ListAdminTickets(ctx, status)
}

// ListReplies 查询工单回复，普通用户仅可查看自己的工单回复。
func (s *TicketService) ListReplies(ctx context.Context, userID uint, ticketID uint, isAdmin bool) ([]model.TicketReply, error) {
	ticket, err := s.ticketRepo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if !isAdmin && ticket.UserID != userID {
		return nil, WrapForbidden("无权限查看该工单回复")
	}
	return s.ticketRepo.ListReplies(ctx, ticketID)
}
