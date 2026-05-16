package dto

type ClientErrorLogRequest struct {
	RequestID    string `json:"request_id" validate:"omitempty,max=64"`
	PagePath     string `json:"page_path" validate:"required,max=255"`
	ErrorType    string `json:"error_type" validate:"required,max=64"`
	Message      string `json:"message" validate:"required,max=500"`
	Stack        string `json:"stack" validate:"omitempty,max=5000"`
	APIPath      string `json:"api_path" validate:"omitempty,max=255"`
	HTTPStatus   *int   `json:"http_status" validate:"omitempty,min=100,max=599"`
	BusinessCode *int   `json:"business_code" validate:"omitempty"`
	Browser      string `json:"browser" validate:"omitempty,max=255"`
	OS           string `json:"os" validate:"omitempty,max=255"`
	AppVersion   string `json:"app_version" validate:"omitempty,max=64"`
}
