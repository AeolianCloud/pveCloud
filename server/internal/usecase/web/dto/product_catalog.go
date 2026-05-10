package dto

type ServerCatalogResponse struct {
	Products []ServerCatalogProduct `json:"products"`
}

type ServerCatalogProduct struct {
	ProductNo   string              `json:"product_no"`
	Slug        string              `json:"slug"`
	Name        string              `json:"name"`
	Summary     *string             `json:"summary"`
	Description *string             `json:"description"`
	Plans       []ServerCatalogPlan `json:"plans"`
}

type ServerCatalogPlan struct {
	PlanNo         string                     `json:"plan_no"`
	Code           string                     `json:"code"`
	Name           string                     `json:"name"`
	Summary        *string                    `json:"summary"`
	CPUCores       int                        `json:"cpu_cores"`
	MemoryMB       int                        `json:"memory_mb"`
	SystemDiskGB   int                        `json:"system_disk_gb"`
	DataDiskGB     int                        `json:"data_disk_gb"`
	BandwidthMbps  int                        `json:"bandwidth_mbps"`
	TrafficGB      *int                       `json:"traffic_gb"`
	PublicIPCount  int                        `json:"public_ip_count"`
	Virtualization string                     `json:"virtualization"`
	Architecture   string                     `json:"architecture"`
	IsFeatured     bool                       `json:"is_featured"`
	Status         string                     `json:"status"`
	Prices         []ServerCatalogPlanPrice   `json:"prices"`
	Regions        []ServerCatalogRegion      `json:"regions"`
	OSTemplates    []ServerCatalogOSTemplate  `json:"os_templates"`
	NetworkTypes   []ServerCatalogNetworkType `json:"network_types"`
}

type ServerCatalogPlanPrice struct {
	BillingCycle       string  `json:"billing_cycle"`
	PriceCents         uint64  `json:"price_cents"`
	OriginalPriceCents *uint64 `json:"original_price_cents"`
	Currency           string  `json:"currency"`
}

type ServerCatalogRegion struct {
	RegionNo string  `json:"region_no"`
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	Country  *string `json:"country"`
	City     *string `json:"city"`
	Summary  *string `json:"summary"`
}

type ServerCatalogOSTemplate struct {
	TemplateNo   string  `json:"template_no"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	OSFamily     string  `json:"os_family"`
	Distribution string  `json:"distribution"`
	Version      string  `json:"version"`
	Architecture string  `json:"architecture"`
	Summary      *string `json:"summary"`
}

type ServerCatalogNetworkType struct {
	NetworkTypeNo string  `json:"network_type_no"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Summary       *string `json:"summary"`
}
