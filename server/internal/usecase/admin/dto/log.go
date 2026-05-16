package dto

import "time"

type UserSecurityLogQuery struct {
	Page      int    `form:"page" validate:"omitempty,min=1"`
	PerPage   int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	UserID    uint64 `form:"user_id" validate:"omitempty,min=1"`
	Username  string `form:"username" validate:"omitempty,max=96"`
	Action    string `form:"action" validate:"omitempty,max=96"`
	Result    string `form:"result" validate:"omitempty,oneof=success failed limited"`
	RequestID string `form:"request_id" validate:"omitempty,max=64"`
	IP        string `form:"ip" validate:"omitempty,max=64"`
	DateFrom  string `form:"date_from" validate:"omitempty,max=32"`
	DateTo    string `form:"date_to" validate:"omitempty,max=32"`
}

type UserBusinessLogQuery struct {
	Page       int    `form:"page" validate:"omitempty,min=1"`
	PerPage    int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	UserID     uint64 `form:"user_id" validate:"omitempty,min=1"`
	Module     string `form:"module" validate:"omitempty,max=64"`
	Action     string `form:"action" validate:"omitempty,max=96"`
	ObjectType string `form:"object_type" validate:"omitempty,max=64"`
	ObjectID   string `form:"object_id" validate:"omitempty,max=128"`
	RequestID  string `form:"request_id" validate:"omitempty,max=64"`
	DateFrom   string `form:"date_from" validate:"omitempty,max=32"`
	DateTo     string `form:"date_to" validate:"omitempty,max=32"`
}

type FrontendErrorLogQuery struct {
	Page      int    `form:"page" validate:"omitempty,min=1"`
	PerPage   int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	SourceApp string `form:"source_app" validate:"omitempty,oneof=admin web"`
	PagePath  string `form:"page_path" validate:"omitempty,max=255"`
	ErrorType string `form:"error_type" validate:"omitempty,max=64"`
	APIPath   string `form:"api_path" validate:"omitempty,max=255"`
	HTTPStatus int   `form:"http_status" validate:"omitempty,min=100,max=599"`
	RequestID string `form:"request_id" validate:"omitempty,max=64"`
	DateFrom  string `form:"date_from" validate:"omitempty,max=32"`
	DateTo    string `form:"date_to" validate:"omitempty,max=32"`
}

type BackendRuntimeLogQuery struct {
	Page       int    `form:"page" validate:"omitempty,min=1"`
	PerPage    int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Level      string `form:"level" validate:"omitempty,max=16"`
	Category   string `form:"category" validate:"omitempty,max=32"`
	Status     int    `form:"status" validate:"omitempty,min=100,max=599"`
	RequestID  string `form:"request_id" validate:"omitempty,max=64"`
	RequestPath string `form:"request_path" validate:"omitempty,max=255"`
	DateFrom   string `form:"date_from" validate:"omitempty,max=32"`
	DateTo     string `form:"date_to" validate:"omitempty,max=32"`
}

type LogUserSummary struct {
	ID          uint64  `json:"id"`
	Username    string  `json:"username"`
	Email       *string `json:"email"`
	DisplayName *string `json:"display_name"`
}

type UserSecurityLogItem struct {
	ID            uint64          `json:"id"`
	User          *LogUserSummary `json:"user"`
	SessionID     *string         `json:"session_id"`
	RequestID     *string         `json:"request_id"`
	RequestMethod *string         `json:"request_method"`
	RequestPath   *string         `json:"request_path"`
	Action        string          `json:"action"`
	Result        string          `json:"result"`
	IP            *string         `json:"ip"`
	UserAgent     *string         `json:"user_agent"`
	Remark        *string         `json:"remark"`
	CreatedAt     time.Time       `json:"created_at"`
}

type UserBusinessLogItem struct {
	ID            uint64          `json:"id"`
	User          LogUserSummary  `json:"user"`
	RequestID     *string         `json:"request_id"`
	RequestMethod *string         `json:"request_method"`
	RequestPath   *string         `json:"request_path"`
	Module        string          `json:"module"`
	Action        string          `json:"action"`
	ObjectType    string          `json:"object_type"`
	ObjectID      *string         `json:"object_id"`
	Summary       *string         `json:"summary"`
	IP            *string         `json:"ip"`
	UserAgent     *string         `json:"user_agent"`
	CreatedAt     time.Time       `json:"created_at"`
}

type LogExportRecordRequest struct {
	LogType string `json:"log_type" validate:"required,max=64"`
}

type FrontendErrorLogItem struct {
	ID           uint64         `json:"id"`
	SourceApp    string         `json:"source_app"`
	UserID       *uint64        `json:"user_id"`
	AdminID      *uint64        `json:"admin_id"`
	RequestID    *string        `json:"request_id"`
	PagePath     string         `json:"page_path"`
	ErrorType    string         `json:"error_type"`
	Message      string         `json:"message"`
	Stack        *string        `json:"stack"`
	APIPath      *string        `json:"api_path"`
	HTTPStatus   *int           `json:"http_status"`
	BusinessCode *int           `json:"business_code"`
	Browser      *string        `json:"browser"`
	OS           *string        `json:"os"`
	AppVersion   *string        `json:"app_version"`
	IP           *string        `json:"ip"`
	UserAgent    *string        `json:"user_agent"`
	CreatedAt    time.Time      `json:"created_at"`
}

type BackendRuntimeLogItem struct {
	ID            uint64   `json:"id"`
	Level         string   `json:"level"`
	Category      string   `json:"category"`
	RequestID     *string  `json:"request_id"`
	RequestMethod *string  `json:"request_method"`
	RequestPath   *string  `json:"request_path"`
	Status        *int     `json:"status"`
	LatencyMS     *int64   `json:"latency_ms"`
	ClientIP      *string  `json:"client_ip"`
	Message       string   `json:"message"`
	Detail        *string  `json:"detail"`
	CreatedAt     time.Time `json:"created_at"`
}

type LogExportRecordItem struct {
	ID        uint64    `json:"id"`
	AdminID   uint64    `json:"admin_id"`
	LogType   string    `json:"log_type"`
	RowCount  int       `json:"row_count"`
	CreatedAt time.Time `json:"created_at"`
}
