package dto

import "time"

type OrderListQuery struct {
	Page        int    `form:"page" validate:"omitempty,min=1"`
	PerPage     int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Status      string `form:"status" validate:"omitempty,oneof=pending provisioning fulfilled cancelled closed"`
	OrderNo     string `form:"order_no" validate:"omitempty,max=64"`
	UserKeyword string `form:"user_keyword" validate:"omitempty,max=128"`
	DateFrom    string `form:"date_from" validate:"omitempty,max=32"`
	DateTo      string `form:"date_to" validate:"omitempty,max=32"`
}

type OrderUserSummary struct {
	ID          uint64  `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
}

type AdminOrderItem struct {
	OrderNo          string           `json:"order_no"`
	User             OrderUserSummary `json:"user"`
	Status           string           `json:"status"`
	ProductName      string           `json:"product_name"`
	PlanName         string           `json:"plan_name"`
	BillingCycle     string           `json:"billing_cycle"`
	NetworkTypeName  string           `json:"network_type_name"`
	TotalAmountCents uint64           `json:"total_amount_cents"`
	Currency         string           `json:"currency"`
	AdminNote        *string          `json:"admin_note"`
	CreatedAt        time.Time        `json:"created_at"`
	CancelledAt      *time.Time       `json:"cancelled_at"`
	ClosedAt         *time.Time       `json:"closed_at"`
}

type AdminOrderDetail struct {
	AdminOrderItem
	UserNote           *string `json:"user_note"`
	CancelReason       *string `json:"cancel_reason"`
	ClosedReason       *string `json:"closed_reason"`
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

type OrderAdminNoteRequest struct {
	AdminNote *string `json:"admin_note" validate:"omitempty,max=1000"`
}

type OrderStatusRequest struct {
	Reason *string `json:"reason" validate:"omitempty,max=500"`
}
