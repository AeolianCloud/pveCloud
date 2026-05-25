package invoice

import "time"

type Application struct {
	ID                uint64     `gorm:"column:id;primaryKey"`
	InvoiceNo         string     `gorm:"column:invoice_no"`
	UserID            uint64     `gorm:"column:user_id"`
	ClientToken       string     `gorm:"column:client_token"`
	InvoiceType       string     `gorm:"column:invoice_type"`
	TitleType         string     `gorm:"column:title_type"`
	Title             string     `gorm:"column:title"`
	TaxNo             *string    `gorm:"column:tax_no"`
	Email             *string    `gorm:"column:email"`
	AmountCents       uint64     `gorm:"column:amount_cents"`
	Currency          string     `gorm:"column:currency"`
	Status            string     `gorm:"column:status"`
	Remark            *string    `gorm:"column:remark"`
	AdminNote         *string    `gorm:"column:admin_note"`
	RejectReason      *string    `gorm:"column:reject_reason"`
	CancelReason      *string    `gorm:"column:cancel_reason"`
	InvoiceCode       *string    `gorm:"column:invoice_code"`
	InvoiceNumber     *string    `gorm:"column:invoice_number"`
	InvoiceFileID     *uint64    `gorm:"column:invoice_file_id"`
	AcceptedByAdminID *uint64    `gorm:"column:accepted_by_admin_id"`
	RejectedByAdminID *uint64    `gorm:"column:rejected_by_admin_id"`
	IssuedByAdminID   *uint64    `gorm:"column:issued_by_admin_id"`
	AcceptedAt        *time.Time `gorm:"column:accepted_at"`
	RejectedAt        *time.Time `gorm:"column:rejected_at"`
	CancelledAt       *time.Time `gorm:"column:cancelled_at"`
	IssuedAt          *time.Time `gorm:"column:issued_at"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
}

func (Application) TableName() string { return "invoice_applications" }

type ApplicationOrder struct {
	ID               uint64     `gorm:"column:id;primaryKey"`
	InvoiceID        uint64     `gorm:"column:invoice_id"`
	InvoiceNo        string     `gorm:"column:invoice_no"`
	UserID           uint64     `gorm:"column:user_id"`
	OrderID          uint64     `gorm:"column:order_id"`
	OrderNo          string     `gorm:"column:order_no"`
	OrderType        string     `gorm:"column:order_type"`
	OrderAmountCents uint64     `gorm:"column:order_amount_cents"`
	Currency         string     `gorm:"column:currency"`
	PaymentStatus    string     `gorm:"column:payment_status"`
	PaidAt           *time.Time `gorm:"column:paid_at"`
	ProductName      *string    `gorm:"column:product_name"`
	PlanName         *string    `gorm:"column:plan_name"`
	StatusSnapshot   string     `gorm:"column:status_snapshot"`
	ActiveOrderID    *uint64    `gorm:"column:active_order_id;->"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (ApplicationOrder) TableName() string { return "invoice_application_orders" }

type ApplicationRow struct {
	Application
	Username    string
	UserEmail   string `gorm:"column:user_email"`
	DisplayName *string
	OrderCount  int
}

type EligibleOrderRow struct {
	ID                uint64
	OrderNo           string
	OrderType         string
	RelatedInstanceNo *string
	TotalAmountCents  uint64
	Currency          string
	PaymentStatus     string
	PaidAt            *time.Time
	ProductName       string
	PlanName          string
	InvoiceOccupied   bool
}
