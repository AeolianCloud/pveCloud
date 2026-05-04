package dto

import "time"

type RealNameConfig struct {
	Enabled             bool     `json:"enabled"`
	RequiredForOrder    bool     `json:"required_for_order"`
	ResubmitEnabled     bool     `json:"resubmit_enabled"`
	MaxSubmitAttempts   int      `json:"max_submit_attempts"`
	IDCardFrontRequired bool     `json:"id_card_front_required"`
	IDCardBackRequired  bool     `json:"id_card_back_required"`
	HoldCardRequired    bool     `json:"hold_card_required"`
	ImageMaxSizeMB      int      `json:"image_max_size_mb"`
	AllowedImageTypes   []string `json:"allowed_image_types"`
	ReviewNotice        string   `json:"review_notice"`
}

type RealNameApplicationSummary struct {
	ApplicationNo  string     `json:"application_no"`
	RealName       string     `json:"real_name"`
	IDType         string     `json:"id_type"`
	IDNumberMasked string     `json:"id_number_masked"`
	Status         string     `json:"status"`
	RejectReason   *string    `json:"reject_reason"`
	SubmitAttempt  uint       `json:"submit_attempt"`
	CreatedAt      time.Time  `json:"created_at"`
	ReviewedAt     *time.Time `json:"reviewed_at"`
}

type RealNameStatusResponse struct {
	Status      string                      `json:"status"`
	Application *RealNameApplicationSummary `json:"application"`
	Config      RealNameConfig              `json:"config"`
}

type RealNameSubmitRequest struct {
	RealName          string  `json:"real_name" validate:"required,min=2,max=64"`
	IDType            string  `json:"id_type" validate:"required,oneof=id_card"`
	IDNumber          string  `json:"id_number" validate:"required,min=6,max=32"`
	IDCardFrontFileID *uint64 `json:"id_card_front_file_id" validate:"omitempty,min=1"`
	IDCardBackFileID  *uint64 `json:"id_card_back_file_id" validate:"omitempty,min=1"`
	HoldCardFileID    *uint64 `json:"hold_card_file_id" validate:"omitempty,min=1"`
}

type RealNameFileUploadResponse struct {
	ID           uint64    `json:"id"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	Size         uint64    `json:"size"`
	CreatedAt    time.Time `json:"created_at"`
}
