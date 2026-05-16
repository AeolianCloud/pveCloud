package logs

import "time"

type UserSecurityLog struct {
	ID            uint64     `gorm:"column:id;primaryKey"`
	UserID        *uint64    `gorm:"column:user_id"`
	Username      *string    `gorm:"column:username"`
	Email         *string    `gorm:"column:email"`
	SessionID     *string    `gorm:"column:session_id"`
	RequestID     *string    `gorm:"column:request_id"`
	RequestMethod *string    `gorm:"column:request_method"`
	RequestPath   *string    `gorm:"column:request_path"`
	Action        string     `gorm:"column:action"`
	Result        string     `gorm:"column:result"`
	IP            *string    `gorm:"column:ip"`
	UserAgent     *string    `gorm:"column:user_agent"`
	Remark        *string    `gorm:"column:remark"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
}

func (UserSecurityLog) TableName() string { return "user_security_logs" }

type UserSecurityLogRow struct {
	UserSecurityLog
	UserUsername    *string `gorm:"column:user_username"`
	UserEmail       *string `gorm:"column:user_email"`
	UserDisplayName *string `gorm:"column:user_display_name"`
}

type UserBusinessLog struct {
	ID            uint64     `gorm:"column:id;primaryKey"`
	UserID        uint64     `gorm:"column:user_id"`
	Username      *string    `gorm:"column:username"`
	Email         *string    `gorm:"column:email"`
	RequestID     *string    `gorm:"column:request_id"`
	RequestMethod *string    `gorm:"column:request_method"`
	RequestPath   *string    `gorm:"column:request_path"`
	Module        string     `gorm:"column:module"`
	Action        string     `gorm:"column:action"`
	ObjectType    string     `gorm:"column:object_type"`
	ObjectID      *string    `gorm:"column:object_id"`
	Summary       *string    `gorm:"column:summary"`
	IP            *string    `gorm:"column:ip"`
	UserAgent     *string    `gorm:"column:user_agent"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
}

func (UserBusinessLog) TableName() string { return "user_business_logs" }

type UserBusinessLogRow struct {
	UserBusinessLog
	UserUsername    *string `gorm:"column:user_username"`
	UserEmail       *string `gorm:"column:user_email"`
	UserDisplayName *string `gorm:"column:user_display_name"`
}

type FrontendErrorLog struct {
	ID            uint64     `gorm:"column:id;primaryKey"`
	SourceApp     string     `gorm:"column:source_app"`
	UserID        *uint64    `gorm:"column:user_id"`
	AdminID       *uint64    `gorm:"column:admin_id"`
	RequestID     *string    `gorm:"column:request_id"`
	PagePath      string     `gorm:"column:page_path"`
	ErrorType     string     `gorm:"column:error_type"`
	Message       string     `gorm:"column:message"`
	Stack         *string    `gorm:"column:stack"`
	APIPath       *string    `gorm:"column:api_path"`
	HTTPStatus    *int       `gorm:"column:http_status"`
	BusinessCode  *int       `gorm:"column:business_code"`
	Browser       *string    `gorm:"column:browser"`
	OS            *string    `gorm:"column:os"`
	AppVersion    *string    `gorm:"column:app_version"`
	IP            *string    `gorm:"column:ip"`
	UserAgent     *string    `gorm:"column:user_agent"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
}

func (FrontendErrorLog) TableName() string { return "frontend_error_logs" }

type BackendRuntimeLog struct {
	ID            uint64     `gorm:"column:id;primaryKey"`
	Level         string     `gorm:"column:level"`
	Category      string     `gorm:"column:category"`
	RequestID     *string    `gorm:"column:request_id"`
	RequestMethod *string    `gorm:"column:request_method"`
	RequestPath   *string    `gorm:"column:request_path"`
	Status        *int       `gorm:"column:status"`
	LatencyMS     *int64     `gorm:"column:latency_ms"`
	ClientIP      *string    `gorm:"column:client_ip"`
	Message       string     `gorm:"column:message"`
	Detail        *string    `gorm:"column:detail"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
}

func (BackendRuntimeLog) TableName() string { return "backend_runtime_logs" }

type ExportRecord struct {
	ID        uint64    `gorm:"column:id;primaryKey"`
	AdminID   uint64    `gorm:"column:admin_id"`
	LogType   string    `gorm:"column:log_type"`
	Filters   *string   `gorm:"column:filters"`
	RowCount  int       `gorm:"column:row_count"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (ExportRecord) TableName() string { return "log_export_records" }
