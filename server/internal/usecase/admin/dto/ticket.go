package dto

import "time"

type TicketListQuery struct {
	Page            int    `form:"page" validate:"omitempty,min=1"`
	PerPage         int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Status          string `form:"status" validate:"omitempty,oneof=waiting_admin waiting_user closed"`
	Category        string `form:"category" validate:"omitempty,oneof=account order product technical billing other"`
	Priority        string `form:"priority" validate:"omitempty,oneof=low normal high urgent"`
	TicketNo        string `form:"ticket_no" validate:"omitempty,max=64"`
	OrderNo         string `form:"order_no" validate:"omitempty,max=64"`
	UserKeyword     string `form:"user_keyword" validate:"omitempty,max=128"`
	DateFrom        string `form:"date_from" validate:"omitempty,max=32"`
	DateTo          string `form:"date_to" validate:"omitempty,max=32"`
	AssigneeAdminID uint64 `form:"assignee_admin_id" validate:"omitempty,min=1"`
	TagID           uint64 `form:"tag_id" validate:"omitempty,min=1"`
	SLAStatus       string `form:"sla_status" validate:"omitempty,oneof=normal first_response_overdue resolution_overdue"`
}

type TicketMessageRequest struct {
	Content string `validate:"required,max=5000"`
}

type TicketCloseRequest struct {
	Reason *string `json:"reason" validate:"omitempty,max=500"`
}

type TicketAssignRequest struct {
	AssigneeAdminID uint64  `json:"assignee_admin_id" validate:"required,min=1"`
	Reason          *string `json:"reason" validate:"omitempty,max=500"`
}

type TicketCollaboratorRequest struct {
	AdminID uint64 `json:"admin_id" validate:"required,min=1"`
}

type TicketInternalNoteRequest struct {
	Content string `json:"content" validate:"required,max=5000"`
}

type TicketPriorityRequest struct {
	Priority string `json:"priority" validate:"required,oneof=low normal high urgent"`
	Reason   string `json:"reason" validate:"required,max=500"`
}

type TicketTagsRequest struct {
	TagIDs []uint64 `json:"tag_ids" validate:"required,max=20,dive,min=1"`
}

type TicketTagListQuery struct {
	Page       int    `form:"page" validate:"omitempty,min=1"`
	PerPage    int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Keyword    string `form:"keyword" validate:"omitempty,max=64"`
	Visibility string `form:"visibility" validate:"omitempty,oneof=public internal"`
	Status     string `form:"status" validate:"omitempty,oneof=active disabled"`
}

type TicketTagCreateRequest struct {
	Name       string  `json:"name" validate:"required,max=40"`
	Color      *string `json:"color" validate:"omitempty,max=32"`
	Visibility string  `json:"visibility" validate:"required,oneof=public internal"`
	Status     string  `json:"status" validate:"required,oneof=active disabled"`
	SortOrder  int     `json:"sort_order" validate:"omitempty,min=0,max=9999"`
}

type TicketTagUpdateRequest struct {
	Name       *string `json:"name" validate:"omitempty,max=40"`
	Color      *string `json:"color" validate:"omitempty,max=32"`
	Visibility *string `json:"visibility" validate:"omitempty,oneof=public internal"`
	Status     *string `json:"status" validate:"omitempty,oneof=active disabled"`
	SortOrder  *int    `json:"sort_order" validate:"omitempty,min=0,max=9999"`
}

type AssigneeCandidateQuery struct {
	Page    int    `form:"page" validate:"omitempty,min=1"`
	PerPage int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Keyword string `form:"keyword" validate:"omitempty,max=96"`
}

type TicketUserSummary struct {
	ID          uint64  `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
}

type TicketAdminSummary struct {
	ID          uint64  `json:"id"`
	Username    string  `json:"username"`
	Email       *string `json:"email"`
	DisplayName string  `json:"display_name"`
	Status      string  `json:"status,omitempty"`
}

type TicketTagItem struct {
	ID         uint64    `json:"id"`
	Name       string    `json:"name"`
	Color      *string   `json:"color"`
	Visibility string    `json:"visibility"`
	Status     string    `json:"status"`
	SortOrder  int       `json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type TicketSLAInfo struct {
	FirstResponseDueAt *time.Time `json:"first_response_due_at"`
	FirstRespondedAt   *time.Time `json:"first_responded_at"`
	ResolutionDueAt    *time.Time `json:"resolution_due_at"`
	ResolvedAt         *time.Time `json:"resolved_at"`
	Status             string     `json:"status"`
}

type AdminTicketItem struct {
	TicketNo      string              `json:"ticket_no"`
	User          TicketUserSummary   `json:"user"`
	Title         string              `json:"title"`
	Category      string              `json:"category"`
	Priority      string              `json:"priority"`
	Status        string              `json:"status"`
	Assignee      *TicketAdminSummary `json:"assignee"`
	Tags          []TicketTagItem     `json:"tags"`
	SLA           TicketSLAInfo       `json:"sla"`
	OrderNo       *string             `json:"order_no"`
	LastMessageAt time.Time           `json:"last_message_at"`
	CreatedAt     time.Time           `json:"created_at"`
	ClosedAt      *time.Time          `json:"closed_at"`
}

type AdminTicketDetail struct {
	AdminTicketItem
	CloseReason   *string              `json:"close_reason"`
	Messages      []AdminTicketMessage `json:"messages"`
	Collaborators []TicketAdminSummary `json:"collaborators"`
	InternalNotes []TicketInternalNote `json:"internal_notes"`
	Events        []TicketEvent        `json:"events"`
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

type TicketInternalNote struct {
	ID        uint64             `json:"id"`
	Admin     TicketAdminSummary `json:"admin"`
	Content   string             `json:"content"`
	CreatedAt time.Time          `json:"created_at"`
}

type TicketEvent struct {
	ID         uint64              `json:"id"`
	EventType  string              `json:"event_type"`
	Actor      *TicketActorSummary `json:"actor"`
	BeforeData *string             `json:"before_data"`
	AfterData  *string             `json:"after_data"`
	Remark     *string             `json:"remark"`
	CreatedAt  time.Time           `json:"created_at"`
}

type TicketActorSummary struct {
	Type        string  `json:"type"`
	ID          uint64  `json:"id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
}
