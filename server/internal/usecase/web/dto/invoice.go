package dto

import "time"

type InvoiceEligibleOrderQuery struct {
	Page     int    `form:"page" validate:"omitempty,min=1"`
	PerPage  int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword" validate:"omitempty,max=128"`
	DateFrom string `form:"date_from" validate:"omitempty,max=32"`
	DateTo   string `form:"date_to" validate:"omitempty,max=32"`
}

type InvoiceListQuery struct {
	Page     int    `form:"page" validate:"omitempty,min=1"`
	PerPage  int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Status   string `form:"status" validate:"omitempty,oneof=pending processing issued rejected cancelled"`
	DateFrom string `form:"date_from" validate:"omitempty,max=32"`
	DateTo   string `form:"date_to" validate:"omitempty,max=32"`
}

type InvoiceCreateRequest struct {
	OrderNos    []string `json:"order_nos" validate:"required,min=1,dive,required,max=64"`
	TitleType   string   `json:"title_type" validate:"required,oneof=personal company"`
	Title       string   `json:"title" validate:"required,max=100"`
	TaxNo       *string  `json:"tax_no" validate:"omitempty,max=64"`
	Email       *string  `json:"email" validate:"omitempty,email,max=128"`
	Remark      *string  `json:"remark" validate:"omitempty,max=500"`
	ClientToken string   `json:"client_token" validate:"required,max=128"`
}

type InvoiceCancelRequest struct {
	Reason *string `json:"reason" validate:"omitempty,max=500"`
}

type InvoiceEligibleOrderItem struct {
	OrderNo           string     `json:"order_no"`
	OrderType         string     `json:"order_type"`
	RelatedInstanceNo *string    `json:"related_instance_no"`
	AmountCents       uint64     `json:"amount_cents"`
	Currency          string     `json:"currency"`
	PaymentStatus     string     `json:"payment_status"`
	PaidAt            *time.Time `json:"paid_at"`
	ProductName       string     `json:"product_name"`
	PlanName          string     `json:"plan_name"`
	InvoiceOccupied   bool       `json:"invoice_occupied"`
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
	InvoiceNo     string     `json:"invoice_no"`
	InvoiceType   string     `json:"invoice_type"`
	TitleType     string     `json:"title_type"`
	Title         string     `json:"title"`
	AmountCents   uint64     `json:"amount_cents"`
	Currency      string     `json:"currency"`
	Status        string     `json:"status"`
	OrderCount    int        `json:"order_count"`
	InvoiceNumber *string    `json:"invoice_number"`
	IssuedAt      *time.Time `json:"issued_at"`
	CreatedAt     time.Time  `json:"created_at"`
	CanCancel     bool       `json:"can_cancel"`
	CanDownload   bool       `json:"can_download"`
	DownloadURL   *string    `json:"download_url"`
}

type InvoiceDetail struct {
	InvoiceItem
	TaxNo        *string            `json:"tax_no"`
	Email        *string            `json:"email"`
	Remark       *string            `json:"remark"`
	RejectReason *string            `json:"reject_reason"`
	CancelReason *string            `json:"cancel_reason"`
	InvoiceCode  *string            `json:"invoice_code"`
	AcceptedAt   *time.Time         `json:"accepted_at"`
	RejectedAt   *time.Time         `json:"rejected_at"`
	CancelledAt  *time.Time         `json:"cancelled_at"`
	Orders       []InvoiceOrderItem `json:"orders"`
}
