package dto

import "time"

type RealNameApplicationListQuery struct {
	Page     int    `form:"page" validate:"omitempty,min=1"`
	PerPage  int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword" validate:"omitempty,max=96"`
	Status   string `form:"status" validate:"omitempty,oneof=pending approved rejected"`
	IDType   string `form:"id_type" validate:"omitempty,oneof=id_card"`
	DateFrom string `form:"date_from" validate:"omitempty,max=32"`
	DateTo   string `form:"date_to" validate:"omitempty,max=32"`
}

type RealNameReviewRequest struct {
	Status       string  `json:"status" validate:"required,oneof=approved rejected"`
	RejectReason *string `json:"reject_reason" validate:"omitempty,max=500"`
}

type RealNameUserSummary struct {
	ID          uint64  `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
	Status      string  `json:"status"`
}

type RealNameFileSummary struct {
	ID           uint64    `json:"id"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	Size         uint64    `json:"size"`
	CreatedAt    time.Time `json:"created_at"`
}

type RealNameApplicationItem struct {
	ID              uint64               `json:"id"`
	ApplicationNo   string               `json:"application_no"`
	User            RealNameUserSummary  `json:"user"`
	RealName        string               `json:"real_name"`
	IDType          string               `json:"id_type"`
	IDNumberMasked  string               `json:"id_number_masked"`
	Status          string               `json:"status"`
	SubmitAttempt   uint                 `json:"submit_attempt"`
	ReviewAdmin     *RealNameUserSummary `json:"review_admin"`
	ReviewedAt      *time.Time           `json:"reviewed_at"`
	RejectReason    *string              `json:"reject_reason"`
	IDCardFrontFile *RealNameFileSummary `json:"id_card_front_file,omitempty"`
	IDCardBackFile  *RealNameFileSummary `json:"id_card_back_file,omitempty"`
	HoldCardFile    *RealNameFileSummary `json:"hold_card_file,omitempty"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}
