package dto

import "time"

type TicketListQuery struct {
	Page        int    `form:"page" validate:"omitempty,min=1"`
	PerPage     int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Status      string `form:"status" validate:"omitempty,oneof=waiting_admin waiting_user closed"`
	Category    string `form:"category" validate:"omitempty,oneof=account order product technical billing other"`
	Priority    string `form:"priority" validate:"omitempty,oneof=low normal high urgent"`
	TicketNo    string `form:"ticket_no" validate:"omitempty,max=64"`
	OrderNo     string `form:"order_no" validate:"omitempty,max=64"`
	UserKeyword string `form:"user_keyword" validate:"omitempty,max=128"`
	DateFrom    string `form:"date_from" validate:"omitempty,max=32"`
	DateTo      string `form:"date_to" validate:"omitempty,max=32"`
}

type TicketMessageRequest struct {
	Content string `validate:"required,max=5000"`
}

type TicketCloseRequest struct {
	Reason *string `json:"reason" validate:"omitempty,max=500"`
}

type TicketUserSummary struct {
	ID          uint64  `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
}

type AdminTicketItem struct {
	TicketNo      string            `json:"ticket_no"`
	User          TicketUserSummary `json:"user"`
	Title         string            `json:"title"`
	Category      string            `json:"category"`
	Priority      string            `json:"priority"`
	Status        string            `json:"status"`
	OrderNo       *string           `json:"order_no"`
	LastMessageAt time.Time         `json:"last_message_at"`
	CreatedAt     time.Time         `json:"created_at"`
	ClosedAt      *time.Time        `json:"closed_at"`
}

type AdminTicketDetail struct {
	AdminTicketItem
	CloseReason *string              `json:"close_reason"`
	Messages    []AdminTicketMessage `json:"messages"`
}

type AdminTicketMessage struct {
	ID          uint64                  `json:"id"`
	SenderType  string                  `json:"sender_type"`
	SenderName  string                  `json:"sender_name"`
	Content     string                  `json:"content"`
	Attachments []AdminTicketAttachment `json:"attachments"`
	CreatedAt   time.Time               `json:"created_at"`
}

type AdminTicketAttachment struct {
	FileID       uint64 `json:"file_id"`
	OriginalName string `json:"original_name"`
	MimeType     string `json:"mime_type"`
	Extension    string `json:"extension"`
	Size         uint64 `json:"size"`
	DownloadURL  string `json:"download_url"`
}
