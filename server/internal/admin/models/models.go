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
	ID            uint64    `gorm:"column:id;primaryKey"`
	OriginalName  string    `gorm:"column:original_name"`
	StoredName    string    `gorm:"column:stored_name"`
	MimeType      string    `gorm:"column:mime_type"`
	Extension     string    `gorm:"column:extension"`
	Size          uint64    `gorm:"column:size"`
	StoragePath   string    `gorm:"column:storage_path"`
	StorageDriver string    `gorm:"column:storage_driver"`
	Checksum      string    `gorm:"column:checksum"`
	UploaderID    uint64    `gorm:"column:uploader_id"`
	Status        string    `gorm:"column:status"`
	CreatedAt     time.Time `gorm:"column:created_at"`
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
