package dto

import "time"

type RealNameApplicationListQuery struct {
	Page           int    `form:"page" validate:"omitempty,min=1"`
	PerPage        int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Keyword        string `form:"keyword" validate:"omitempty,max=96"`
	Status         string `form:"status" validate:"omitempty,oneof=pending approved rejected"`
	IDType         string `form:"id_type" validate:"omitempty,oneof=id_card"`
	Provider       string `form:"provider" validate:"omitempty,oneof=alipay wechat manual"`
	ProviderStatus string `form:"provider_status" validate:"omitempty,max=64"`
	DateFrom       string `form:"date_from" validate:"omitempty,max=32"`
	DateTo         string `form:"date_to" validate:"omitempty,max=32"`
}

type RealNameReviewRequest struct {
	Status string `json:"status" validate:"required,oneof=approved rejected"`
	Reason string `json:"reason" validate:"omitempty,max=500"`
}

type RealNameUserSummary struct {
	ID          uint64  `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
	Status      string  `json:"status"`
}

type RealNameApplicationItem struct {
	ID                    uint64              `json:"id"`
	ApplicationNo         string              `json:"application_no"`
	User                  RealNameUserSummary `json:"user"`
	RealName              string              `json:"real_name"`
	IDType                string              `json:"id_type"`
	IDNumberMasked        string              `json:"id_number_masked"`
	VerificationProvider  *string             `json:"verification_provider"`
	ProviderApplicationID *string             `json:"provider_application_id"`
	ProviderStatus        *string             `json:"provider_status"`
	ProviderResultCode    *string             `json:"provider_result_code"`
	ProviderResultMessage *string             `json:"provider_result_message"`
	ProviderTraceID       *string             `json:"provider_trace_id"`
	Status                string              `json:"status"`
	SubmitAttempt         uint                `json:"submit_attempt"`
	FailureReason         *string             `json:"failure_reason"`
	ProviderStartedAt     *time.Time          `json:"provider_started_at"`
	ProviderFinishedAt    *time.Time          `json:"provider_finished_at"`
	CreatedAt             time.Time           `json:"created_at"`
	UpdatedAt             time.Time           `json:"updated_at"`
}
