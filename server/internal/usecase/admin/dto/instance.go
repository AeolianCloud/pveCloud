package dto

import "time"

type InstanceListQuery struct {
	Page        int    `form:"page" validate:"omitempty,min=1"`
	PerPage     int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Status      string `form:"status" validate:"omitempty,oneof=creating running stopped error releasing released"`
	InstanceNo  string `form:"instance_no" validate:"omitempty,max=64"`
	OrderNo     string `form:"order_no" validate:"omitempty,max=64"`
	UserKeyword string `form:"user_keyword" validate:"omitempty,max=128"`
	DateFrom    string `form:"date_from" validate:"omitempty,max=32"`
	DateTo      string `form:"date_to" validate:"omitempty,max=32"`
}

type InstanceMappingListQuery struct {
	Page          int    `form:"page" validate:"omitempty,min=1"`
	PerPage       int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Status        string `form:"status" validate:"omitempty,oneof=active inactive"`
	PlanNo        string `form:"plan_no" validate:"omitempty,max=64"`
	RegionNo      string `form:"region_no" validate:"omitempty,max=64"`
	TemplateNo    string `form:"template_no" validate:"omitempty,max=64"`
	NetworkTypeNo string `form:"network_type_no" validate:"omitempty,max=64"`
}

type InstanceMappingRequest struct {
	MappingNo       string  `json:"mapping_no" validate:"omitempty,max=64"`
	ProductNo       *string `json:"product_no" validate:"omitempty,max=64"`
	PlanNo          string  `json:"plan_no" validate:"required,max=64"`
	RegionNo        string  `json:"region_no" validate:"required,max=64"`
	TemplateNo      string  `json:"template_no" validate:"required,max=64"`
	NetworkTypeNo   string  `json:"network_type_no" validate:"omitempty,max=64"`
	Node            string  `json:"node" validate:"required,max=128"`
	Storage         string  `json:"storage" validate:"required,max=128"`
	DiskSource      string  `json:"disk_source" validate:"required,max=255"`
	DiskFormat      *string `json:"disk_format" validate:"omitempty,max=32"`
	DiskInterface   *string `json:"disk_interface" validate:"omitempty,max=32"`
	SnippetsStorage *string `json:"snippets_storage" validate:"omitempty,max=128"`
	CIUser          *string `json:"ci_user" validate:"omitempty,max=64"`
	SSHKeys         *string `json:"ssh_keys" validate:"omitempty,max=10000"`
	IPConfig0       *string `json:"ip_config0" validate:"omitempty,max=255"`
	Nameserver      *string `json:"nameserver" validate:"omitempty,max=128"`
	SearchDomain    *string `json:"search_domain" validate:"omitempty,max=128"`
	CIPackages      *string `json:"ci_packages" validate:"omitempty,max=2000"`
	AptMirror       *string `json:"apt_mirror" validate:"omitempty,max=255"`
	VMIDStart       uint    `json:"vmid_start" validate:"required,min=1"`
	VMIDEnd         uint    `json:"vmid_end" validate:"required,min=1"`
	NextVMID        uint    `json:"next_vmid" validate:"required,min=1"`
	Status          string  `json:"status" validate:"required,oneof=active inactive"`
	Remark          *string `json:"remark" validate:"omitempty,max=500"`
}

type InstanceMappingItem struct {
	ID              uint64    `json:"id"`
	MappingNo       string    `json:"mapping_no"`
	ProductNo       *string   `json:"product_no"`
	PlanNo          string    `json:"plan_no"`
	RegionNo        string    `json:"region_no"`
	TemplateNo      string    `json:"template_no"`
	NetworkTypeNo   string    `json:"network_type_no"`
	Node            string    `json:"node"`
	Storage         string    `json:"storage"`
	DiskSource      string    `json:"disk_source"`
	DiskFormat      *string   `json:"disk_format"`
	DiskInterface   *string   `json:"disk_interface"`
	SnippetsStorage *string   `json:"snippets_storage"`
	CIUser          *string   `json:"ci_user"`
	SSHKeys         *string   `json:"ssh_keys"`
	IPConfig0       *string   `json:"ip_config0"`
	Nameserver      *string   `json:"nameserver"`
	SearchDomain    *string   `json:"search_domain"`
	CIPackages      *string   `json:"ci_packages"`
	AptMirror       *string   `json:"apt_mirror"`
	VMIDStart       uint      `json:"vmid_start"`
	VMIDEnd         uint      `json:"vmid_end"`
	NextVMID        uint      `json:"next_vmid"`
	Status          string    `json:"status"`
	Remark          *string   `json:"remark"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type InstanceItem struct {
	InstanceNo      string           `json:"instance_no"`
	OrderNo         string           `json:"order_no"`
	User            OrderUserSummary `json:"user"`
	Status          string           `json:"status"`
	ProductName     string           `json:"product_name"`
	PlanName        string           `json:"plan_name"`
	RegionName      string           `json:"region_name"`
	NetworkTypeName *string          `json:"network_type_name"`
	TemplateName    string           `json:"template_name"`
	ExternalNode    string           `json:"external_node"`
	ExternalVMID    uint             `json:"external_vmid"`
	CreatedAt       time.Time        `json:"created_at"`
	ReleasedAt      *time.Time       `json:"released_at"`
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
	ExternalResourceLocation *string             `json:"external_resource_location"`
	LastErrorCode            *string             `json:"last_error_code"`
	LastErrorMessage         *string             `json:"last_error_message"`
	Operations               []InstanceOperation `json:"operations"`
}

type InstanceOperation struct {
	OperationNo         string     `json:"operation_no"`
	Action              string     `json:"action"`
	Status              string     `json:"status"`
	ExternalOperationID *string    `json:"external_operation_id"`
	OperationLocation   *string    `json:"operation_location"`
	ResourceLocation    *string    `json:"resource_location"`
	ErrorCode           *string    `json:"error_code"`
	ErrorMessage        *string    `json:"error_message"`
	CreatedAt           time.Time  `json:"created_at"`
	CompletedAt         *time.Time `json:"completed_at"`
}

type ProvisionResponse struct {
	Instance  InstanceDetail    `json:"instance"`
	Operation InstanceOperation `json:"operation"`
}

type MCPNode struct {
	Node   string `json:"node"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type MCPVM struct {
	VMID   uint   `json:"vmid"`
	Name   string `json:"name"`
	Status string `json:"status"`
	CPUs   int    `json:"cpus"`
	Mem    int64  `json:"mem"`
	MaxMem int64  `json:"maxmem"`
}

type MCPStorage struct {
	Storage string `json:"storage"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Status  string `json:"status"`
}
