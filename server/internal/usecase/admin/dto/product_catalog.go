package dto

import "time"

type ProductListQuery struct {
	Page    int    `form:"page" validate:"omitempty,min=1"`
	PerPage int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Keyword string `form:"keyword" validate:"omitempty,max=96"`
	Type    string `form:"type" validate:"omitempty,oneof=server"`
	Status  string `form:"status" validate:"omitempty,oneof=draft active inactive"`
}

type ProductRequest struct {
	ProductNo   string  `json:"product_no" validate:"omitempty,max=64"`
	Type        string  `json:"type" validate:"omitempty,oneof=server"`
	Slug        string  `json:"slug" validate:"required,min=2,max=96"`
	Name        string  `json:"name" validate:"required,min=1,max=128"`
	Summary     *string `json:"summary" validate:"omitempty,max=255"`
	Description *string `json:"description" validate:"omitempty,max=5000"`
	Status      string  `json:"status" validate:"omitempty,oneof=draft active inactive"`
	Visible     bool    `json:"visible"`
	SortOrder   int     `json:"sort_order" validate:"omitempty,min=0,max=100000"`
}

type ProductStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft active inactive"`
}

type ProductItem struct {
	ID          uint64    `json:"id"`
	ProductNo   string    `json:"product_no"`
	Type        string    `json:"type"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Summary     *string   `json:"summary"`
	Description *string   `json:"description"`
	Status      string    `json:"status"`
	Visible     bool      `json:"visible"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductPlanListQuery struct {
	Page      int    `form:"page" validate:"omitempty,min=1"`
	PerPage   int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	ProductID uint64 `form:"product_id" validate:"omitempty,min=1"`
	Keyword   string `form:"keyword" validate:"omitempty,max=96"`
	Status    string `form:"status" validate:"omitempty,oneof=draft active inactive sold_out"`
}

type ProductPlanRequest struct {
	PlanNo         string  `json:"plan_no" validate:"omitempty,max=64"`
	ProductID      uint64  `json:"product_id" validate:"required,min=1"`
	Code           string  `json:"code" validate:"required,min=2,max=96"`
	Name           string  `json:"name" validate:"required,min=1,max=128"`
	Summary        *string `json:"summary" validate:"omitempty,max=255"`
	CPUCores       int     `json:"cpu_cores" validate:"required,min=1,max=512"`
	MemoryMB       int     `json:"memory_mb" validate:"required,min=128,max=1048576"`
	SystemDiskGB   int     `json:"system_disk_gb" validate:"required,min=1,max=1048576"`
	DataDiskGB     int     `json:"data_disk_gb" validate:"omitempty,min=0,max=1048576"`
	BandwidthMbps  int     `json:"bandwidth_mbps" validate:"required,min=1,max=100000"`
	TrafficGB      *int    `json:"traffic_gb" validate:"omitempty,min=0,max=10000000"`
	PublicIPCount  int     `json:"public_ip_count" validate:"omitempty,min=0,max=1024"`
	Virtualization string  `json:"virtualization" validate:"required,oneof=kvm"`
	Architecture   string  `json:"architecture" validate:"required,oneof=x86_64"`
	IsFeatured     bool    `json:"is_featured"`
	Status         string  `json:"status" validate:"required,oneof=draft active inactive sold_out"`
	Visible        bool    `json:"visible"`
	SortOrder      int     `json:"sort_order" validate:"omitempty,min=0,max=100000"`
}

type ProductPlanStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft active inactive sold_out"`
}

type ProductPlanItem struct {
	ID             uint64    `json:"id"`
	PlanNo         string    `json:"plan_no"`
	ProductID      uint64    `json:"product_id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	Summary        *string   `json:"summary"`
	CPUCores       int       `json:"cpu_cores"`
	MemoryMB       int       `json:"memory_mb"`
	SystemDiskGB   int       `json:"system_disk_gb"`
	DataDiskGB     int       `json:"data_disk_gb"`
	BandwidthMbps  int       `json:"bandwidth_mbps"`
	TrafficGB      *int      `json:"traffic_gb"`
	PublicIPCount  int       `json:"public_ip_count"`
	Virtualization string    `json:"virtualization"`
	Architecture   string    `json:"architecture"`
	IsFeatured     bool      `json:"is_featured"`
	Status         string    `json:"status"`
	Visible        bool      `json:"visible"`
	SortOrder      int       `json:"sort_order"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PlanPriceRequest struct {
	BillingCycle       string  `json:"billing_cycle" validate:"required,oneof=monthly quarterly semi_yearly yearly"`
	PriceCents         uint64  `json:"price_cents" validate:"required,min=1"`
	OriginalPriceCents *uint64 `json:"original_price_cents" validate:"omitempty,min=0"`
	Currency           string  `json:"currency" validate:"required,oneof=CNY"`
	Status             string  `json:"status" validate:"required,oneof=active inactive"`
	SortOrder          int     `json:"sort_order" validate:"omitempty,min=0,max=100000"`
}

type PlanPriceListRequest struct {
	Prices []PlanPriceRequest `json:"prices" validate:"required,dive"`
}

type PlanPriceItem struct {
	ID                 uint64    `json:"id"`
	PlanID             uint64    `json:"plan_id"`
	BillingCycle       string    `json:"billing_cycle"`
	PriceCents         uint64    `json:"price_cents"`
	OriginalPriceCents *uint64   `json:"original_price_cents"`
	Currency           string    `json:"currency"`
	Status             string    `json:"status"`
	SortOrder          int       `json:"sort_order"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type SalesRegionListQuery struct {
	Keyword string `form:"keyword" validate:"omitempty,max=96"`
	Status  string `form:"status" validate:"omitempty,oneof=active inactive"`
}

type SalesRegionRequest struct {
	RegionNo  string  `json:"region_no" validate:"omitempty,max=64"`
	Code      string  `json:"code" validate:"required,min=2,max=64"`
	Name      string  `json:"name" validate:"required,min=1,max=128"`
	Country   *string `json:"country" validate:"omitempty,max=64"`
	City      *string `json:"city" validate:"omitempty,max=64"`
	Summary   *string `json:"summary" validate:"omitempty,max=255"`
	Status    string  `json:"status" validate:"required,oneof=active inactive"`
	Visible   bool    `json:"visible"`
	SortOrder int     `json:"sort_order" validate:"omitempty,min=0,max=100000"`
}

type SalesRegionItem struct {
	ID        uint64    `json:"id"`
	RegionNo  string    `json:"region_no"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Country   *string   `json:"country"`
	City      *string   `json:"city"`
	Summary   *string   `json:"summary"`
	Status    string    `json:"status"`
	Visible   bool      `json:"visible"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ServerOSTemplateListQuery struct {
	Keyword string `form:"keyword" validate:"omitempty,max=96"`
	Status  string `form:"status" validate:"omitempty,oneof=active inactive"`
}

type ServerOSTemplateRequest struct {
	TemplateNo   string  `json:"template_no" validate:"omitempty,max=64"`
	Code         string  `json:"code" validate:"required,min=2,max=96"`
	Name         string  `json:"name" validate:"required,min=1,max=128"`
	OSFamily     string  `json:"os_family" validate:"required,oneof=linux windows bsd"`
	Distribution string  `json:"distribution" validate:"required,min=1,max=64"`
	Version      string  `json:"version" validate:"required,min=1,max=64"`
	Architecture string  `json:"architecture" validate:"required,oneof=x86_64"`
	Summary      *string `json:"summary" validate:"omitempty,max=255"`
	Status       string  `json:"status" validate:"required,oneof=active inactive"`
	Visible      bool    `json:"visible"`
	SortOrder    int     `json:"sort_order" validate:"omitempty,min=0,max=100000"`
}

type ServerOSTemplateItem struct {
	ID           uint64    `json:"id"`
	TemplateNo   string    `json:"template_no"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	OSFamily     string    `json:"os_family"`
	Distribution string    `json:"distribution"`
	Version      string    `json:"version"`
	Architecture string    `json:"architecture"`
	Summary      *string   `json:"summary"`
	Status       string    `json:"status"`
	Visible      bool      `json:"visible"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type NetworkTypeListQuery struct {
	Keyword string `form:"keyword" validate:"omitempty,max=96"`
	Status  string `form:"status" validate:"omitempty,oneof=active inactive"`
}

type NetworkTypeRequest struct {
	NetworkTypeNo string  `json:"network_type_no" validate:"omitempty,max=64"`
	Code          string  `json:"code" validate:"required,min=2,max=64"`
	Name          string  `json:"name" validate:"required,min=1,max=128"`
	Summary       *string `json:"summary" validate:"omitempty,max=255"`
	Status        string  `json:"status" validate:"required,oneof=active inactive"`
	Visible       bool    `json:"visible"`
	SortOrder     int     `json:"sort_order" validate:"omitempty,min=0,max=100000"`
}

type NetworkTypeItem struct {
	ID            uint64    `json:"id"`
	NetworkTypeNo string    `json:"network_type_no"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Summary       *string   `json:"summary"`
	Status        string    `json:"status"`
	Visible       bool      `json:"visible"`
	SortOrder     int       `json:"sort_order"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PlanRelationRequest struct {
	IDs []uint64 `json:"ids" validate:"required,dive,min=1"`
}

type PlanRelationResponse struct {
	PlanID     uint64   `json:"plan_id"`
	RelatedIDs []uint64 `json:"related_ids"`
}
