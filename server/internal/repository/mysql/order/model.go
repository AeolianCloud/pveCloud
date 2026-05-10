package order

import "time"

type Order struct {
	ID                 uint64     `gorm:"column:id;primaryKey"`
	OrderNo            string     `gorm:"column:order_no"`
	UserID             uint64     `gorm:"column:user_id"`
	ClientToken        string     `gorm:"column:client_token"`
	Status             string     `gorm:"column:status"`
	ProductNo          string     `gorm:"column:product_no"`
	ProductType        string     `gorm:"column:product_type"`
	ProductName        string     `gorm:"column:product_name"`
	ProductSummary     *string    `gorm:"column:product_summary"`
	PlanNo             string     `gorm:"column:plan_no"`
	PlanCode           string     `gorm:"column:plan_code"`
	PlanName           string     `gorm:"column:plan_name"`
	PlanSummary        *string    `gorm:"column:plan_summary"`
	CPUCores           int        `gorm:"column:cpu_cores"`
	MemoryMB           int        `gorm:"column:memory_mb"`
	SystemDiskGB       int        `gorm:"column:system_disk_gb"`
	DataDiskGB         int        `gorm:"column:data_disk_gb"`
	BandwidthMbps      int        `gorm:"column:bandwidth_mbps"`
	TrafficGB          *int       `gorm:"column:traffic_gb"`
	PublicIPCount      int        `gorm:"column:public_ip_count"`
	Virtualization     string     `gorm:"column:virtualization"`
	Architecture       string     `gorm:"column:architecture"`
	BillingCycle       string     `gorm:"column:billing_cycle"`
	PriceCents         uint64     `gorm:"column:price_cents"`
	OriginalPriceCents *uint64    `gorm:"column:original_price_cents"`
	Currency           string     `gorm:"column:currency"`
	Quantity           int        `gorm:"column:quantity"`
	TotalAmountCents   uint64     `gorm:"column:total_amount_cents"`
	RegionNo           string     `gorm:"column:region_no"`
	RegionCode         string     `gorm:"column:region_code"`
	RegionName         string     `gorm:"column:region_name"`
	TemplateNo         string     `gorm:"column:template_no"`
	TemplateCode       string     `gorm:"column:template_code"`
	TemplateName       string     `gorm:"column:template_name"`
	OSFamily           string     `gorm:"column:os_family"`
	OSDistribution     string     `gorm:"column:os_distribution"`
	OSVersion          string     `gorm:"column:os_version"`
	OSArchitecture     string     `gorm:"column:os_architecture"`
	UserNote           *string    `gorm:"column:user_note"`
	AdminNote          *string    `gorm:"column:admin_note"`
	CancelReason       *string    `gorm:"column:cancel_reason"`
	ClosedReason       *string    `gorm:"column:closed_reason"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at"`
	CancelledAt        *time.Time `gorm:"column:cancelled_at"`
	ClosedAt           *time.Time `gorm:"column:closed_at"`
}

func (Order) TableName() string { return "orders" }

type CatalogSelection struct {
	ProductNo          string
	ProductType        string
	ProductName        string
	ProductSummary     *string
	PlanNo             string
	PlanCode           string
	PlanName           string
	PlanSummary        *string
	CPUCores           int
	MemoryMB           int
	SystemDiskGB       int
	DataDiskGB         int
	BandwidthMbps      int
	TrafficGB          *int
	PublicIPCount      int
	Virtualization     string
	Architecture       string
	BillingCycle       string
	PriceCents         uint64
	OriginalPriceCents *uint64
	Currency           string
	RegionNo           string
	RegionCode         string
	RegionName         string
	TemplateNo         string
	TemplateCode       string
	TemplateName       string
	OSFamily           string
	OSDistribution     string
	OSVersion          string
	OSArchitecture     string
}

type UserSummary struct {
	ID          uint64
	Username    string
	Email       string
	DisplayName *string
}

type OrderRow struct {
	Order
	Username    string
	Email       string
	DisplayName *string
}
