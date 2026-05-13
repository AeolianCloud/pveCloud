package ticket

import "time"

type Ticket struct {
	ID                 uint64     `gorm:"column:id;primaryKey"`
	TicketNo           string     `gorm:"column:ticket_no"`
	UserID             uint64     `gorm:"column:user_id"`
	OrderID            *uint64    `gorm:"column:order_id"`
	OrderNo            *string    `gorm:"column:order_no"`
	Category           string     `gorm:"column:category"`
	Priority           string     `gorm:"column:priority"`
	Title              string     `gorm:"column:title"`
	Status             string     `gorm:"column:status"`
	LastMessageAt      time.Time  `gorm:"column:last_message_at"`
	LastUserMessageAt  *time.Time `gorm:"column:last_user_message_at"`
	LastAdminMessageAt *time.Time `gorm:"column:last_admin_message_at"`
	ClosedByType       *string    `gorm:"column:closed_by_type"`
	ClosedByUserID     *uint64    `gorm:"column:closed_by_user_id"`
	ClosedByAdminID    *uint64    `gorm:"column:closed_by_admin_id"`
	ClosedAt           *time.Time `gorm:"column:closed_at"`
	CloseReason        *string    `gorm:"column:close_reason"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at"`
}

func (Ticket) TableName() string { return "tickets" }

type TicketMessage struct {
	ID            uint64    `gorm:"column:id;primaryKey"`
	TicketID      uint64    `gorm:"column:ticket_id"`
	SenderType    string    `gorm:"column:sender_type"`
	SenderUserID  *uint64   `gorm:"column:sender_user_id"`
	SenderAdminID *uint64   `gorm:"column:sender_admin_id"`
	Content       string    `gorm:"column:content"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (TicketMessage) TableName() string { return "ticket_messages" }

type TicketMessageAttachment struct {
	ID        uint64    `gorm:"column:id;primaryKey"`
	TicketID  uint64    `gorm:"column:ticket_id"`
	MessageID uint64    `gorm:"column:message_id"`
	FileID    uint64    `gorm:"column:file_id"`
	SortOrder int       `gorm:"column:sort_order"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (TicketMessageAttachment) TableName() string { return "ticket_message_attachments" }

type UserSummary struct {
	ID          uint64
	Username    string
	Email       string
	DisplayName *string
}

type TicketRow struct {
	Ticket
	Username    string
	Email       string
	DisplayName *string
}

type MessageRow struct {
	TicketMessage
	Username         *string
	UserEmail        *string
	UserDisplayName  *string
	AdminUsername    *string
	AdminDisplayName *string
}

type MessageAttachmentRow struct {
	TicketMessageAttachment
	OriginalName string
	MimeType     string
	Extension    string
	Size         uint64
	CreatedAt    time.Time
}
