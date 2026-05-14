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
	AssigneeAdminID    *uint64    `gorm:"column:assignee_admin_id"`
	AssignedByAdminID  *uint64    `gorm:"column:assigned_by_admin_id"`
	AssignedAt         *time.Time `gorm:"column:assigned_at"`
	LastMessageAt      time.Time  `gorm:"column:last_message_at"`
	LastUserMessageAt  *time.Time `gorm:"column:last_user_message_at"`
	LastAdminMessageAt *time.Time `gorm:"column:last_admin_message_at"`
	FirstResponseDueAt *time.Time `gorm:"column:first_response_due_at"`
	FirstRespondedAt   *time.Time `gorm:"column:first_responded_at"`
	ResolutionDueAt    *time.Time `gorm:"column:resolution_due_at"`
	ResolvedAt         *time.Time `gorm:"column:resolved_at"`
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

type TicketTag struct {
	ID               uint64    `gorm:"column:id;primaryKey"`
	Name             string    `gorm:"column:name"`
	Color            *string   `gorm:"column:color"`
	Visibility       string    `gorm:"column:visibility"`
	Status           string    `gorm:"column:status"`
	SortOrder        int       `gorm:"column:sort_order"`
	CreatedByAdminID *uint64   `gorm:"column:created_by_admin_id"`
	UpdatedByAdminID *uint64   `gorm:"column:updated_by_admin_id"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (TicketTag) TableName() string { return "ticket_tags" }

type TicketTagBinding struct {
	ID               uint64    `gorm:"column:id;primaryKey"`
	TicketID         uint64    `gorm:"column:ticket_id"`
	TagID            uint64    `gorm:"column:tag_id"`
	CreatedByAdminID *uint64   `gorm:"column:created_by_admin_id"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

func (TicketTagBinding) TableName() string { return "ticket_tag_bindings" }

type TicketInternalNote struct {
	ID        uint64    `gorm:"column:id;primaryKey"`
	TicketID  uint64    `gorm:"column:ticket_id"`
	AdminID   uint64    `gorm:"column:admin_id"`
	Content   string    `gorm:"column:content"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (TicketInternalNote) TableName() string { return "ticket_internal_notes" }

type TicketCollaborator struct {
	ID               uint64    `gorm:"column:id;primaryKey"`
	TicketID         uint64    `gorm:"column:ticket_id"`
	AdminID          uint64    `gorm:"column:admin_id"`
	CreatedByAdminID *uint64   `gorm:"column:created_by_admin_id"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

func (TicketCollaborator) TableName() string { return "ticket_collaborators" }

type TicketEvent struct {
	ID           uint64    `gorm:"column:id;primaryKey"`
	TicketID     uint64    `gorm:"column:ticket_id"`
	EventType    string    `gorm:"column:event_type"`
	ActorAdminID *uint64   `gorm:"column:actor_admin_id"`
	ActorUserID  *uint64   `gorm:"column:actor_user_id"`
	BeforeData   *string   `gorm:"column:before_data"`
	AfterData    *string   `gorm:"column:after_data"`
	Remark       *string   `gorm:"column:remark"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (TicketEvent) TableName() string { return "ticket_events" }

type UserSummary struct {
	ID          uint64
	Username    string
	Email       string
	DisplayName *string
}

type AdminSummary struct {
	ID          uint64
	Username    string
	Email       *string
	DisplayName string
	Status      string
}

type TicketRow struct {
	Ticket
	Username            string
	Email               string
	DisplayName         *string
	AssigneeUsername    *string
	AssigneeEmail       *string
	AssigneeDisplayName *string
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

type TicketTagRow struct {
	TicketTag
	TicketID *uint64
}

type TicketNoteRow struct {
	TicketInternalNote
	AdminUsername    string
	AdminEmail       *string
	AdminDisplayName string
}

type TicketCollaboratorRow struct {
	TicketCollaborator
	AdminUsername    string
	AdminEmail       *string
	AdminDisplayName string
	AdminStatus      string
}

type TicketEventRow struct {
	TicketEvent
	ActorAdminUsername    *string
	ActorAdminDisplayName *string
	ActorUserUsername     *string
	ActorUserDisplayName  *string
}
