package dto

import "time"

type RealNameConfig struct {
	Enabled           bool     `json:"enabled"`
	RequiredForOrder  bool     `json:"required_for_order"`
	AllowedProviders  []string `json:"allowed_providers"`
	DefaultProvider   string   `json:"default_provider"`
	ResubmitEnabled   bool     `json:"resubmit_enabled"`
	MaxSubmitAttempts int      `json:"max_submit_attempts"`
	ReviewNotice      string   `json:"review_notice"`
}

type RealNameApplicationSummary struct {
	ApplicationNo        string     `json:"application_no"`
	RealName             string     `json:"real_name"`
	IDType               string     `json:"id_type"`
	IDNumberMasked       string     `json:"id_number_masked"`
	VerificationProvider *string    `json:"verification_provider"`
	ProviderStatus       *string    `json:"provider_status"`
	Status               string     `json:"status"`
	FailureReason        *string    `json:"failure_reason"`
	SubmitAttempt        uint       `json:"submit_attempt"`
	CreatedAt            time.Time  `json:"created_at"`
	VerifiedAt           *time.Time `json:"verified_at"`
}

type RealNameStatusResponse struct {
	Status      string                      `json:"status"`
	Application *RealNameApplicationSummary `json:"application"`
	Config      RealNameConfig              `json:"config"`
}

type RealNameSubmitRequest struct {
	RealName string `json:"real_name" validate:"required,min=2,max=64"`
	IDType   string `json:"id_type" validate:"required,oneof=id_card"`
	IDNumber string `json:"id_number" validate:"required,min=6,max=32"`
	Provider string `json:"provider" validate:"omitempty,oneof=alipay wechat manual"`
}

type RealNameProviderAction struct {
	Provider    string     `json:"provider"`
	ActionType  string     `json:"action_type"`
	RedirectURL string     `json:"redirect_url"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

type RealNameSubmitResponse struct {
	Application    RealNameApplicationSummary `json:"application"`
	ProviderAction RealNameProviderAction     `json:"provider_action"`
}

type RealNameSyncRequest struct {
	ApplicationNo string `json:"application_no" validate:"omitempty,max=64"`
}
