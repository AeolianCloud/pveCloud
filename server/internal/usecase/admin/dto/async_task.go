package dto

import "time"

type AsyncTaskListQuery struct {
	Page       int    `form:"page" validate:"omitempty,min=1"`
	PerPage    int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	TaskType   string `form:"task_type" validate:"omitempty,max=64"`
	Status     string `form:"status" validate:"omitempty,oneof=pending running succeeded failed cancelled"`
	ObjectType string `form:"object_type" validate:"omitempty,max=64"`
	ObjectNo   string `form:"object_no" validate:"omitempty,max=64"`
	DateFrom   string `form:"date_from" validate:"omitempty,max=32"`
	DateTo     string `form:"date_to" validate:"omitempty,max=32"`
}

type AsyncTaskItem struct {
	TaskNo           string     `json:"task_no"`
	TaskType         string     `json:"task_type"`
	Status           string     `json:"status"`
	ObjectType       *string    `json:"object_type"`
	ObjectNo         *string    `json:"object_no"`
	ScheduledAt      time.Time  `json:"scheduled_at"`
	Attempts         int        `json:"attempts"`
	MaxAttempts      int        `json:"max_attempts"`
	LastErrorCode    *string    `json:"last_error_code"`
	LastErrorMessage *string    `json:"last_error_message"`
	LockedBy         *string    `json:"locked_by"`
	LockedUntil      *time.Time `json:"locked_until"`
	CreatedAt        time.Time  `json:"created_at"`
	CompletedAt      *time.Time `json:"completed_at"`
}

type AsyncTaskRetryRequest struct {
	Remark *string `json:"remark" validate:"omitempty,max=500"`
}
