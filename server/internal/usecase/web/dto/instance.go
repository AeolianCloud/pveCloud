package dto

import "time"

type InstanceListQuery struct {
	Page    int    `form:"page" validate:"omitempty,min=1"`
	PerPage int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Status  string `form:"status" validate:"omitempty,oneof=creating running stopped error releasing released"`
}

type InstanceItem struct {
	InstanceNo              string               `json:"instance_no"`
	OrderNo                 string               `json:"order_no"`
	Status                  string               `json:"status"`
	ProductName             string               `json:"product_name"`
	PlanName                string               `json:"plan_name"`
	RegionName              string               `json:"region_name"`
	NetworkTypeName         *string              `json:"network_type_name"`
	TemplateName            string               `json:"template_name"`
	ServiceStartedAt        *time.Time           `json:"service_started_at"`
	ExpiresAt               *time.Time           `json:"expires_at"`
	ExpireStatus            string               `json:"expire_status"`
	ReleaseCountdownSeconds *int64               `json:"release_countdown_seconds"`
	LatestRenewalOrder      *RenewalOrderSummary `json:"latest_renewal_order"`
	CreatedAt               time.Time            `json:"created_at"`
	ReleasedAt              *time.Time           `json:"released_at"`
}

type InstanceDetail struct {
	InstanceItem
	ProductNo                string              `json:"product_no"`
	PlanNo                   string              `json:"plan_no"`
	CPUCores                 int                 `json:"cpu_cores"`
	MemoryMB                 int                 `json:"memory_mb"`
	SystemDiskGB             int                 `json:"system_disk_gb"`
	DataDiskGB               int                 `json:"data_disk_gb"`
	BandwidthMbps            int                 `json:"bandwidth_mbps"`
	RegionNo                 string              `json:"region_no"`
	NetworkTypeNo            *string             `json:"network_type_no"`
	TemplateNo               string              `json:"template_no"`
	OSFamily                 string              `json:"os_family"`
	OSDistribution           string              `json:"os_distribution"`
	OSVersion                string              `json:"os_version"`
	ExpireNoticeSentAt       *time.Time          `json:"expire_notice_sent_at"`
	ExpireReleaseScheduledAt *time.Time          `json:"expire_release_scheduled_at"`
	ExpireReleasedAt         *time.Time          `json:"expire_released_at"`
	RenewalAvailable         bool                `json:"renewal_available"`
	Operations               []InstanceOperation `json:"operations"`
}

type RenewalOrderSummary struct {
	OrderNo          string     `json:"order_no"`
	Status           string     `json:"status"`
	PaymentStatus    string     `json:"payment_status"`
	BillingCycle     string     `json:"billing_cycle"`
	TotalAmountCents uint64     `json:"total_amount_cents"`
	Currency         string     `json:"currency"`
	PaidAt           *time.Time `json:"paid_at"`
	CreatedAt        time.Time  `json:"created_at"`
}

type InstanceOperation struct {
	OperationNo string     `json:"operation_no"`
	Action      string     `json:"action"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
}
