package task

import "time"

type Task struct {
	ID            uint64    `json:"id"`
	TaskNo        string    `json:"task_no"`
	TaskType      string    `json:"task_type"`
	BusinessType  string    `json:"business_type"`
	BusinessID    uint64    `json:"business_id"`
	Status        string    `json:"status"`
	Payload       []byte    `json:"payload"`
	NextRunAt     time.Time `json:"next_run_at"`
	RetryCount    int       `json:"retry_count"`
	MaxRetryCount int       `json:"max_retry_count"`
	LockedBy      string    `json:"locked_by"`
	LockedAt      time.Time `json:"locked_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateInput struct {
	TaskType     string `json:"task_type"`
	BusinessType string `json:"business_type"`
	BusinessID   uint64 `json:"business_id"`
	Payload      []byte `json:"payload"`
}

type CreateTaskParams struct {
	TaskType      string    `json:"task_type"`
	BusinessType  string    `json:"business_type"`
	BusinessID    uint64    `json:"business_id"`
	Status        string    `json:"status"`
	Payload       []byte    `json:"payload"`
	NextRunAt     time.Time `json:"next_run_at"`
	MaxRetryCount int       `json:"max_retry_count"`
}
