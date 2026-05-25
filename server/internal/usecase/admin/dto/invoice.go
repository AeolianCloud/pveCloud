package dto

import "time"

type InvoiceListQuery struct {
	Page         int    `form:"page" validate:"omitempty,min=1"`
	PerPage      int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Status       string `form:"status" validate:"omitempty,oneof=pending processing issued rejected cancelled"`
	InvoiceNo    string `form:"invoice_no" validate:"omitempty,max=64"`
	OrderNo      string `form:"order_no" validate:"omitempty,max=64"`
	UserKeyword  string `form:"user_keyword" validate:"omitempty,max=128"`
	TitleKeyword string `form:"title_keyword" validate:"omitempty,max=100"`
	DateFrom     string `form:"date_from" validate:"omitempty,max=32"`
	DateTo       string `form:"date_to" validate:"omitempty,max=32"`
}

type InvoiceRejectRequest struct {
	Reason string `json:"reason" validate:"required,max=500"`
}

type InvoiceIssueRequest struct {
	InvoiceCode   *string   `json:"invoice_code" validate:"omitempty,max=64"`
	InvoiceNumber string    `json:"invoice_number" validate:"required,max=128"`
	IssuedAt      time.Time `json:"issued_at" validate:"required"`
	FileID        uint64    `json:"file_id" validate:"required,min=1"`
}

type InvoiceAdminNoteRequest struct {
	AdminNote *string `json:"admin_note" validate:"omitempty,max=1000"`
}

type InvoiceFileSummary struct {
	ID           uint64 `json:"id"`
	OriginalName string `json:"original_name"`
	MimeType     string `json:"mime_type"`
	Size         uint64 `json:"size"`
	DownloadURL  string `json:"download_url"`
}

type InvoiceOrderItem struct {
	OrderNo          string     `json:"order_no"`
	OrderType        string     `json:"order_type"`
	OrderAmountCents uint64     `json:"order_amount_cents"`
	Currency         string     `json:"currency"`
	PaymentStatus    string     `json:"payment_status"`
	PaidAt           *time.Time `json:"paid_at"`
	ProductName      *string    `json:"product_name"`
	PlanName         *string    `json:"plan_name"`
}

type InvoiceItem struct {
	InvoiceNo     string           `json:"invoice_no"`
	InvoiceType   string           `json:"invoice_type"`
	User          OrderUserSummary `json:"user"`
	TitleType     string           `json:"title_type"`
	Title         string           `json:"title"`
	AmountCents   uint64           `json:"amount_cents"`
	Currency      string           `json:"currency"`
	Status        string           `json:"status"`
	OrderCount    int              `json:"order_count"`
	InvoiceNumber *string          `json:"invoice_number"`
	CreatedAt     time.Time        `json:"created_at"`
	AcceptedAt    *time.Time       `json:"accepted_at"`
	IssuedAt      *time.Time       `json:"issued_at"`
}

type InvoiceDetail struct {
	InvoiceItem
	TaxNo        *string             `json:"tax_no"`
	Email        *string             `json:"email"`
	Remark       *string             `json:"remark"`
	AdminNote    *string             `json:"admin_note"`
	RejectReason *string             `json:"reject_reason"`
	CancelReason *string             `json:"cancel_reason"`
	InvoiceCode  *string             `json:"invoice_code"`
	RejectedAt   *time.Time          `json:"rejected_at"`
	CancelledAt  *time.Time          `json:"cancelled_at"`
	Orders       []InvoiceOrderItem  `json:"orders"`
	File         *InvoiceFileSummary `json:"file"`
}
