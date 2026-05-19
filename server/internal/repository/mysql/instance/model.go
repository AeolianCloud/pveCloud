package instance

import "time"

type ProvisionMapping struct {
	ID              uint64    `gorm:"column:id;primaryKey"`
	MappingNo       string    `gorm:"column:mapping_no"`
	ProductNo       *string   `gorm:"column:product_no"`
	PlanNo          string    `gorm:"column:plan_no"`
	RegionNo        string    `gorm:"column:region_no"`
	TemplateNo      string    `gorm:"column:template_no"`
	NetworkTypeNo   string    `gorm:"column:network_type_no"`
	Node            string    `gorm:"column:node"`
	Storage         string    `gorm:"column:storage"`
	DiskSource      string    `gorm:"column:disk_source"`
	DiskFormat      *string   `gorm:"column:disk_format"`
	DiskInterface   *string   `gorm:"column:disk_interface"`
	SnippetsStorage *string   `gorm:"column:snippets_storage"`
	CIUser          *string   `gorm:"column:ci_user"`
	SSHKeys         *string   `gorm:"column:ssh_keys"`
	IPConfig0       *string   `gorm:"column:ip_config0"`
	Nameserver      *string   `gorm:"column:nameserver"`
	SearchDomain    *string   `gorm:"column:search_domain"`
	CIPackages      *string   `gorm:"column:ci_packages"`
	AptMirror       *string   `gorm:"column:apt_mirror"`
	VMIDStart       uint      `gorm:"column:vmid_start"`
	VMIDEnd         uint      `gorm:"column:vmid_end"`
	NextVMID        uint      `gorm:"column:next_vmid"`
	Status          string    `gorm:"column:status"`
	Remark          *string   `gorm:"column:remark"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

func (ProvisionMapping) TableName() string { return "instance_provision_mappings" }

type Instance struct {
	ID                       uint64     `gorm:"column:id;primaryKey"`
	InstanceNo               string     `gorm:"column:instance_no"`
	UserID                   uint64     `gorm:"column:user_id"`
	OrderID                  uint64     `gorm:"column:order_id"`
	OrderNo                  string     `gorm:"column:order_no"`
	Status                   string     `gorm:"column:status"`
	ProductNo                string     `gorm:"column:product_no"`
	ProductName              string     `gorm:"column:product_name"`
	PlanNo                   string     `gorm:"column:plan_no"`
	PlanName                 string     `gorm:"column:plan_name"`
	CPUCores                 int        `gorm:"column:cpu_cores"`
	MemoryMB                 int        `gorm:"column:memory_mb"`
	SystemDiskGB             int        `gorm:"column:system_disk_gb"`
	DataDiskGB               int        `gorm:"column:data_disk_gb"`
	BandwidthMbps            int        `gorm:"column:bandwidth_mbps"`
	RegionNo                 string     `gorm:"column:region_no"`
	RegionName               string     `gorm:"column:region_name"`
	NetworkTypeNo            *string    `gorm:"column:network_type_no"`
	NetworkTypeName          *string    `gorm:"column:network_type_name"`
	TemplateNo               string     `gorm:"column:template_no"`
	TemplateName             string     `gorm:"column:template_name"`
	OSFamily                 string     `gorm:"column:os_family"`
	OSDistribution           string     `gorm:"column:os_distribution"`
	OSVersion                string     `gorm:"column:os_version"`
	ExternalNode             string     `gorm:"column:external_node"`
	ExternalVMID             uint       `gorm:"column:external_vmid"`
	ExternalResourceLocation *string    `gorm:"column:external_resource_location"`
	LastErrorCode            *string    `gorm:"column:last_error_code"`
	LastErrorMessage         *string    `gorm:"column:last_error_message"`
	ServiceStartedAt         *time.Time `gorm:"column:service_started_at"`
	ExpiresAt                *time.Time `gorm:"column:expires_at"`
	ExpireNoticeSentAt       *time.Time `gorm:"column:expire_notice_sent_at"`
	ExpireReleaseScheduledAt *time.Time `gorm:"column:expire_release_scheduled_at"`
	ExpireReleasedAt         *time.Time `gorm:"column:expire_released_at"`
	CreatedAt                time.Time  `gorm:"column:created_at"`
	UpdatedAt                time.Time  `gorm:"column:updated_at"`
	ReleasedAt               *time.Time `gorm:"column:released_at"`
}

func (Instance) TableName() string { return "instances" }

type Operation struct {
	ID                  uint64     `gorm:"column:id;primaryKey"`
	OperationNo         string     `gorm:"column:operation_no"`
	InstanceID          uint64     `gorm:"column:instance_id"`
	OrderID             *uint64    `gorm:"column:order_id"`
	AdminID             *uint64    `gorm:"column:admin_id"`
	UserID              *uint64    `gorm:"column:user_id"`
	Action              string     `gorm:"column:action"`
	Status              string     `gorm:"column:status"`
	ExternalOperationID *string    `gorm:"column:external_operation_id"`
	OperationLocation   *string    `gorm:"column:operation_location"`
	ResourceLocation    *string    `gorm:"column:resource_location"`
	ErrorCode           *string    `gorm:"column:error_code"`
	ErrorMessage        *string    `gorm:"column:error_message"`
	CreatedAt           time.Time  `gorm:"column:created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at"`
	CompletedAt         *time.Time `gorm:"column:completed_at"`
}

func (Operation) TableName() string { return "instance_operations" }

type Task struct {
	ID               uint64     `gorm:"column:id;primaryKey"`
	TaskNo           string     `gorm:"column:task_no"`
	TaskType         string     `gorm:"column:task_type"`
	IdempotencyKey   *string    `gorm:"column:idempotency_key"`
	Status           string     `gorm:"column:status"`
	ObjectType       *string    `gorm:"column:object_type"`
	ObjectNo         *string    `gorm:"column:object_no"`
	Payload          *string    `gorm:"column:payload"`
	Result           *string    `gorm:"column:result"`
	Attempts         int        `gorm:"column:attempts"`
	MaxAttempts      int        `gorm:"column:max_attempts"`
	ScheduledAt      time.Time  `gorm:"column:scheduled_at"`
	LockedBy         *string    `gorm:"column:locked_by"`
	LockedUntil      *time.Time `gorm:"column:locked_until"`
	LastErrorCode    *string    `gorm:"column:last_error_code"`
	LastErrorMessage *string    `gorm:"column:last_error_message"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
	CompletedAt      *time.Time `gorm:"column:completed_at"`
}

func (Task) TableName() string { return "async_tasks" }

type Notification struct {
	ID                uint64     `gorm:"column:id;primaryKey"`
	NotificationNo    string     `gorm:"column:notification_no"`
	UserID            uint64     `gorm:"column:user_id"`
	Channel           string     `gorm:"column:channel"`
	Scene             string     `gorm:"column:scene"`
	Target            string     `gorm:"column:target"`
	Status            string     `gorm:"column:status"`
	Subject           *string    `gorm:"column:subject"`
	ContentSummary    *string    `gorm:"column:content_summary"`
	RelatedObjectType *string    `gorm:"column:related_object_type"`
	RelatedObjectNo   *string    `gorm:"column:related_object_no"`
	TaskNo            *string    `gorm:"column:task_no"`
	ErrorCode         *string    `gorm:"column:error_code"`
	ErrorMessage      *string    `gorm:"column:error_message"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
	SentAt            *time.Time `gorm:"column:sent_at"`
}

func (Notification) TableName() string { return "notifications" }

type InstanceRow struct {
	Instance
	Username    string
	Email       string
	DisplayName *string
}
