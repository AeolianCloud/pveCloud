package dto

import "time"

type PageResponse[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PerPage  int   `json:"per_page"`
	LastPage int   `json:"last_page"`
}

type OrderCreateRequest struct {
	PlanNo        string  `json:"plan_no" validate:"required,max=64"`
	BillingCycle  string  `json:"billing_cycle" validate:"required,oneof=monthly quarterly semi_yearly yearly"`
	RegionNo      string  `json:"region_no" validate:"required,max=64"`
	TemplateNo    string  `json:"template_no" validate:"required,max=64"`
	NetworkTypeNo string  `json:"network_type_no" validate:"required,max=64"`
	Quantity      int     `json:"quantity" validate:"omitempty,min=1,max=1"`
	ClientToken   string  `json:"client_token" validate:"required,max=128"`
	UserNote      *string `json:"user_note" validate:"omitempty,max=500"`
}

type OrderListQuery struct {
	Page    int    `form:"page" validate:"omitempty,min=1"`
	PerPage int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Status  string `form:"status" validate:"omitempty,oneof=pending provisioning fulfilled cancelled closed"`
}

type OrderItem struct {
	OrderNo          string     `json:"order_no"`
	Status           string     `json:"status"`
	ProductName      string     `json:"product_name"`
	PlanName         string     `json:"plan_name"`
	BillingCycle     string     `json:"billing_cycle"`
	NetworkTypeName  string     `json:"network_type_name"`
	TotalAmountCents uint64     `json:"total_amount_cents"`
	Currency         string     `json:"currency"`
	CreatedAt        time.Time  `json:"created_at"`
	CancelledAt      *time.Time `json:"cancelled_at"`
	ClosedAt         *time.Time `json:"closed_at"`
}

type OrderDetail struct {
	OrderItem
	UserNote           *string `json:"user_note"`
	ProductNo          string  `json:"product_no"`
	ProductType        string  `json:"product_type"`
	ProductSummary     *string `json:"product_summary"`
	PlanNo             string  `json:"plan_no"`
	PlanCode           string  `json:"plan_code"`
	PlanSummary        *string `json:"plan_summary"`
	CPUCores           int     `json:"cpu_cores"`
	MemoryMB           int     `json:"memory_mb"`
	SystemDiskGB       int     `json:"system_disk_gb"`
	DataDiskGB         int     `json:"data_disk_gb"`
	BandwidthMbps      int     `json:"bandwidth_mbps"`
	TrafficGB          *int    `json:"traffic_gb"`
	PublicIPCount      int     `json:"public_ip_count"`
	Virtualization     string  `json:"virtualization"`
	Architecture       string  `json:"architecture"`
	PriceCents         uint64  `json:"price_cents"`
	OriginalPriceCents *uint64 `json:"original_price_cents"`
	Quantity           int     `json:"quantity"`
	RegionNo           string  `json:"region_no"`
	RegionCode         string  `json:"region_code"`
	RegionName         string  `json:"region_name"`
	NetworkTypeNo      string  `json:"network_type_no"`
	NetworkTypeCode    string  `json:"network_type_code"`
	NetworkTypeName    string  `json:"network_type_name"`
	TemplateNo         string  `json:"template_no"`
	TemplateCode       string  `json:"template_code"`
	TemplateName       string  `json:"template_name"`
	OSFamily           string  `json:"os_family"`
	OSDistribution     string  `json:"os_distribution"`
	OSVersion          string  `json:"os_version"`
	OSArchitecture     string  `json:"os_architecture"`
}

type OrderCancelRequest struct {
	Reason *string `json:"reason" validate:"omitempty,max=500"`
}
