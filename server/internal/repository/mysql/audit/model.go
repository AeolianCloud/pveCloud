package audit

import "time"

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
 * TableName 返回后台操作审计表。
 */
func (AdminAuditLog) TableName() string {
	return "admin_audit_logs"
}

type LogRow struct {
	AdminAuditLog
	ActorUsername    *string `gorm:"column:actor_username"`
	ActorDisplayName *string `gorm:"column:actor_display_name"`
	AdminEmail       *string `gorm:"column:admin_email"`
}
