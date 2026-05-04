package models

import "time"

/**
 * AdminUser 映射 admin_users 管理员账号表。
 */
type AdminUser struct {
	ID           uint64     `gorm:"column:id;primaryKey"`
	Username     string     `gorm:"column:username"`
	Email        *string    `gorm:"column:email"`
	PasswordHash string     `gorm:"column:password_hash"`
	DisplayName  string     `gorm:"column:display_name"`
	Status       string     `gorm:"column:status"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
	LastLoginIP  *string    `gorm:"column:last_login_ip"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
}

/**
 * TableName 返回管理员账号表名。
 *
 * @return string 表名
 */
func (AdminUser) TableName() string {
	return "admin_users"
}

/**
 * AdminRole 映射 admin_roles 管理端角色表。
 */
type AdminRole struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	Code        string    `gorm:"column:code"`
	Name        string    `gorm:"column:name"`
	Description *string   `gorm:"column:description"`
	Status      string    `gorm:"column:status"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

/**
 * TableName 返回管理端角色表名。
 *
 * @return string 表名
 */
func (AdminRole) TableName() string {
	return "admin_roles"
}

/**
 * AdminPermission 映射 admin_permissions 管理端权限码表。
 */
type AdminPermission struct {
	ID            uint64    `gorm:"column:id;primaryKey"`
	Code          string    `gorm:"column:code"`
	Name          string    `gorm:"column:name"`
	Type          string    `gorm:"column:type"`
	ParentCode    *string   `gorm:"column:parent_code"`
	Path          *string   `gorm:"column:path"`
	Icon          *string   `gorm:"column:icon"`
	SortOrder     int       `gorm:"column:sort_order"`
	VisibleInMenu bool      `gorm:"column:visible_in_menu"`
	GroupName     string    `gorm:"column:group_name"`
	Description   *string   `gorm:"column:description"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

/**
 * TableName 返回管理端权限码表名。
 *
 * @return string 表名
 */
func (AdminPermission) TableName() string {
	return "admin_permissions"
}

/**
 * AdminSession 映射 admin_sessions 管理端登录会话表。
 */
type AdminSession struct {
	ID           uint64     `gorm:"column:id;primaryKey"`
	SessionID    string     `gorm:"column:session_id"`
	AdminID      uint64     `gorm:"column:admin_id"`
	Status       string     `gorm:"column:status"`
	IssuedAt     time.Time  `gorm:"column:issued_at"`
	ExpiresAt    time.Time  `gorm:"column:expires_at"`
	LastSeenAt   *time.Time `gorm:"column:last_seen_at"`
	LastSeenIP   *string    `gorm:"column:last_seen_ip"`
	UserAgent    *string    `gorm:"column:user_agent"`
	RevokedAt    *time.Time `gorm:"column:revoked_at"`
	RevokeReason *string    `gorm:"column:revoke_reason"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

/**
 * TableName 返回管理端登录会话表名。
 *
 * @return string 表名
 */
func (AdminSession) TableName() string {
	return "admin_sessions"
}

/**
 * SystemConfig 映射 system_configs 系统配置表。
 */
type SystemConfig struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	ConfigKey   string    `gorm:"column:config_key"`
	ConfigValue *string   `gorm:"column:config_value"`
	ValueType   string    `gorm:"column:value_type"`
	GroupName   string    `gorm:"column:group_name"`
	IsSecret    bool      `gorm:"column:is_secret"`
	Description *string   `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

/**
 * TableName 返回系统配置表名。
 *
 * @return string 表名
 */
func (SystemConfig) TableName() string {
	return "system_configs"
}

/**
 * User 映射 users 用户端账号表。
 */
type User struct {
	ID           uint64    `gorm:"column:id;primaryKey"`
	Username     string    `gorm:"column:username"`
	Email        string    `gorm:"column:email"`
	PasswordHash string    `gorm:"column:password_hash"`
	DisplayName  *string   `gorm:"column:display_name"`
	Status       string    `gorm:"column:status"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

/**
 * TableName 返回用户端账号表名。
 */
func (User) TableName() string {
	return "users"
}

/**
 * UserSession 映射 user_sessions 用户端登录会话表。
 */
type UserSession struct {
	ID           uint64     `gorm:"column:id;primaryKey"`
	UserID       uint64     `gorm:"column:user_id"`
	SessionID    string     `gorm:"column:session_id"`
	Status       string     `gorm:"column:status"`
	IssuedAt     time.Time  `gorm:"column:issued_at"`
	ExpiresAt    time.Time  `gorm:"column:expires_at"`
	RevokedAt    *time.Time `gorm:"column:revoked_at"`
	RevokeReason *string    `gorm:"column:revoke_reason"`
	LastSeenAt   *time.Time `gorm:"column:last_seen_at"`
	LastSeenIP   *string    `gorm:"column:last_seen_ip"`
	UserAgent    *string    `gorm:"column:user_agent"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

/**
 * TableName 返回用户端登录会话表名。
 */
func (UserSession) TableName() string {
	return "user_sessions"
}

/**
 * UserPasswordResetToken 映射 user_password_reset_tokens 用户端密码重置 Token 表。
 */
type UserPasswordResetToken struct {
	ID          uint64     `gorm:"column:id;primaryKey"`
	UserID      uint64     `gorm:"column:user_id"`
	TokenHash   string     `gorm:"column:token_hash"`
	Status      string     `gorm:"column:status"`
	ExpiresAt   time.Time  `gorm:"column:expires_at"`
	UsedAt      *time.Time `gorm:"column:used_at"`
	RequestedIP *string    `gorm:"column:requested_ip"`
	UserAgent   *string    `gorm:"column:user_agent"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
}

/**
 * TableName 返回用户端密码重置 Token 表名。
 */
func (UserPasswordResetToken) TableName() string {
	return "user_password_reset_tokens"
}

/**
 * AdminAuditLog 映射 admin_audit_logs 后台操作审计表。
 */
type AdminAuditLog struct {
	ID               uint64    `gorm:"column:id;primaryKey"`
	AdminID          *uint64   `gorm:"column:admin_id"`
	AdminUsername    *string   `gorm:"column:admin_username"`
	AdminDisplayName *string   `gorm:"column:admin_display_name"`
	SessionID        *string   `gorm:"column:session_id"`
	RequestID        *string   `gorm:"column:request_id"`
	RequestMethod    *string   `gorm:"column:request_method"`
	RequestPath      *string   `gorm:"column:request_path"`
	Action           string    `gorm:"column:action"`
	ObjectType       string    `gorm:"column:object_type"`
	ObjectID         *string   `gorm:"column:object_id"`
	BeforeData       *string   `gorm:"column:before_data"`
	AfterData        *string   `gorm:"column:after_data"`
	IP               *string   `gorm:"column:ip"`
	UserAgent        *string   `gorm:"column:user_agent"`
	Remark           *string   `gorm:"column:remark"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

/**
 * TableName 返回后台操作审计表名。
 */
func (AdminAuditLog) TableName() string {
	return "admin_audit_logs"
}

/**
 * FileAttachment 映射 file_attachments 文件附件表。
 */
type FileAttachment struct {
	ID             uint64    `gorm:"column:id;primaryKey"`
	OriginalName   string    `gorm:"column:original_name"`
	StoredName     string    `gorm:"column:stored_name"`
	MimeType       string    `gorm:"column:mime_type"`
	Extension      string    `gorm:"column:extension"`
	Size           uint64    `gorm:"column:size"`
	StoragePath    string    `gorm:"column:storage_path"`
	StorageDriver  string    `gorm:"column:storage_driver"`
	Checksum       string    `gorm:"column:checksum"`
	UploaderID     uint64    `gorm:"column:uploader_id"`
	UploaderUserID *uint64   `gorm:"column:uploader_user_id"`
	Status         string    `gorm:"column:status"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

/**
 * TableName 返回文件附件表名。
 */
func (FileAttachment) TableName() string {
	return "file_attachments"
}

/**
 * FileAttachmentReference 映射 file_attachment_references 文件引用表。
 */
type FileAttachmentReference struct {
	ID        uint64    `gorm:"column:id;primaryKey"`
	FileID    uint64    `gorm:"column:file_id"`
	RefType   string    `gorm:"column:ref_type"`
	RefID     string    `gorm:"column:ref_id"`
	RefName   *string   `gorm:"column:ref_name"`
	RefPath   *string   `gorm:"column:ref_path"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

/**
 * TableName 返回文件引用表名。
 */
func (FileAttachmentReference) TableName() string {
	return "file_attachment_references"
}

/**
 * UserRealNameApplication 映射用户实名申请表。
 */
type UserRealNameApplication struct {
	ID                uint64     `gorm:"column:id;primaryKey"`
	ApplicationNo     string     `gorm:"column:application_no"`
	UserID            uint64     `gorm:"column:user_id"`
	RealName          string     `gorm:"column:real_name"`
	IDType            string     `gorm:"column:id_type"`
	IDNumberDigest    string     `gorm:"column:id_number_digest"`
	IDNumberMasked    string     `gorm:"column:id_number_masked"`
	IDCardFrontFileID *uint64    `gorm:"column:id_card_front_file_id"`
	IDCardBackFileID  *uint64    `gorm:"column:id_card_back_file_id"`
	HoldCardFileID    *uint64    `gorm:"column:hold_card_file_id"`
	Status            string     `gorm:"column:status"`
	ReviewAdminID     *uint64    `gorm:"column:review_admin_id"`
	ReviewedAt        *time.Time `gorm:"column:reviewed_at"`
	RejectReason      *string    `gorm:"column:reject_reason"`
	SubmitAttempt     uint       `gorm:"column:submit_attempt"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
}

func (UserRealNameApplication) TableName() string {
	return "user_real_name_applications"
}

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
