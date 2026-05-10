package catalog

import "time"

/**
 * Product 映射 products 产品目录表。
 */
type Product struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	ProductNo   string    `gorm:"column:product_no"`
	Type        string    `gorm:"column:type"`
	Slug        string    `gorm:"column:slug"`
	Name        string    `gorm:"column:name"`
	Summary     *string   `gorm:"column:summary"`
	Description *string   `gorm:"column:description"`
	Status      string    `gorm:"column:status"`
	Visible     bool      `gorm:"column:visible"`
	SortOrder   int       `gorm:"column:sort_order"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (Product) TableName() string {
	return "products"
}

/**
 * ProductPlan 映射 product_plans 服务器套餐表。
 */
type ProductPlan struct {
	ID             uint64    `gorm:"column:id;primaryKey"`
	PlanNo         string    `gorm:"column:plan_no"`
	ProductID      uint64    `gorm:"column:product_id"`
	Code           string    `gorm:"column:code"`
	Name           string    `gorm:"column:name"`
	Summary        *string   `gorm:"column:summary"`
	CPUCores       int       `gorm:"column:cpu_cores"`
	MemoryMB       int       `gorm:"column:memory_mb"`
	SystemDiskGB   int       `gorm:"column:system_disk_gb"`
	DataDiskGB     int       `gorm:"column:data_disk_gb"`
	BandwidthMbps  int       `gorm:"column:bandwidth_mbps"`
	TrafficGB      *int      `gorm:"column:traffic_gb"`
	PublicIPCount  int       `gorm:"column:public_ip_count"`
	Virtualization string    `gorm:"column:virtualization"`
	Architecture   string    `gorm:"column:architecture"`
	IsFeatured     bool      `gorm:"column:is_featured"`
	Status         string    `gorm:"column:status"`
	Visible        bool      `gorm:"column:visible"`
	SortOrder      int       `gorm:"column:sort_order"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (ProductPlan) TableName() string {
	return "product_plans"
}

/**
 * PlanPrice 映射 plan_prices 套餐周期价格表。
 */
type PlanPrice struct {
	ID                 uint64    `gorm:"column:id;primaryKey"`
	PlanID             uint64    `gorm:"column:plan_id"`
	BillingCycle       string    `gorm:"column:billing_cycle"`
	PriceCents         uint64    `gorm:"column:price_cents"`
	OriginalPriceCents *uint64   `gorm:"column:original_price_cents"`
	Currency           string    `gorm:"column:currency"`
	Status             string    `gorm:"column:status"`
	SortOrder          int       `gorm:"column:sort_order"`
	CreatedAt          time.Time `gorm:"column:created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at"`
}

func (PlanPrice) TableName() string {
	return "plan_prices"
}

/**
 * SalesRegion 映射 sales_regions 销售地域表。
 */
type SalesRegion struct {
	ID        uint64    `gorm:"column:id;primaryKey"`
	RegionNo  string    `gorm:"column:region_no"`
	Code      string    `gorm:"column:code"`
	Name      string    `gorm:"column:name"`
	Country   *string   `gorm:"column:country"`
	City      *string   `gorm:"column:city"`
	Summary   *string   `gorm:"column:summary"`
	Status    string    `gorm:"column:status"`
	Visible   bool      `gorm:"column:visible"`
	SortOrder int       `gorm:"column:sort_order"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (SalesRegion) TableName() string {
	return "sales_regions"
}

/**
 * ServerOSTemplate 映射 server_os_templates 服务器系统模板表。
 */
type ServerOSTemplate struct {
	ID           uint64    `gorm:"column:id;primaryKey"`
	TemplateNo   string    `gorm:"column:template_no"`
	Code         string    `gorm:"column:code"`
	Name         string    `gorm:"column:name"`
	OSFamily     string    `gorm:"column:os_family"`
	Distribution string    `gorm:"column:distribution"`
	Version      string    `gorm:"column:version"`
	Architecture string    `gorm:"column:architecture"`
	Summary      *string   `gorm:"column:summary"`
	Status       string    `gorm:"column:status"`
	Visible      bool      `gorm:"column:visible"`
	SortOrder    int       `gorm:"column:sort_order"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (ServerOSTemplate) TableName() string {
	return "server_os_templates"
}

/**
 * NetworkType 映射 network_types 网络类型表。
 */
type NetworkType struct {
	ID            uint64    `gorm:"column:id;primaryKey"`
	NetworkTypeNo string    `gorm:"column:network_type_no"`
	Code          string    `gorm:"column:code"`
	Name          string    `gorm:"column:name"`
	Summary       *string   `gorm:"column:summary"`
	Status        string    `gorm:"column:status"`
	Visible       bool      `gorm:"column:visible"`
	SortOrder     int       `gorm:"column:sort_order"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (NetworkType) TableName() string {
	return "network_types"
}

/**
 * PlanRegion 映射 plan_regions 套餐销售地域关联表。
 */
type PlanRegion struct {
	PlanID    uint64    `gorm:"column:plan_id;primaryKey"`
	RegionID  uint64    `gorm:"column:region_id;primaryKey"`
	Status    string    `gorm:"column:status"`
	SortOrder int       `gorm:"column:sort_order"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (PlanRegion) TableName() string {
	return "plan_regions"
}

/**
 * PlanOSTemplate 映射 plan_os_templates 套餐服务器系统模板关联表。
 */
type PlanOSTemplate struct {
	PlanID     uint64    `gorm:"column:plan_id;primaryKey"`
	TemplateID uint64    `gorm:"column:template_id;primaryKey"`
	Status     string    `gorm:"column:status"`
	SortOrder  int       `gorm:"column:sort_order"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (PlanOSTemplate) TableName() string {
	return "plan_os_templates"
}

/**
 * PlanNetworkType 映射 plan_network_types 套餐网络类型关联表。
 */
type PlanNetworkType struct {
	PlanID        uint64    `gorm:"column:plan_id;primaryKey"`
	NetworkTypeID uint64    `gorm:"column:network_type_id;primaryKey"`
	Status        string    `gorm:"column:status"`
	SortOrder     int       `gorm:"column:sort_order"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (PlanNetworkType) TableName() string {
	return "plan_network_types"
}
