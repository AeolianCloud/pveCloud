package job

import "time"

/**
 * AsyncTask 映射 async_tasks 异步任务表。
 */
type AsyncTask struct {
	ID             uint64     `gorm:"column:id;primaryKey"`
	TaskNo         string     `gorm:"column:task_no"`
	TaskType       string     `gorm:"column:task_type"`
	IdempotencyKey string     `gorm:"column:idempotency_key"`
	BizType        string     `gorm:"column:biz_type"`
	BizID          uint64     `gorm:"column:biz_id"`
	Status         string     `gorm:"column:status"`
	Priority       int        `gorm:"column:priority"`
	RetryCount     uint       `gorm:"column:retry_count"`
	MaxRetries     uint       `gorm:"column:max_retries"`
	RunAt          time.Time  `gorm:"column:run_at"`
	LockedBy       *string    `gorm:"column:locked_by"`
	LockedUntil    *time.Time `gorm:"column:locked_until"`
	LastError      *string    `gorm:"column:last_error"`
	Payload        *string    `gorm:"column:payload"`
	Result         *string    `gorm:"column:result"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
	FinishedAt     *time.Time `gorm:"column:finished_at"`
}

/**
 * TableName 返回异步任务表名。
 *
 * @return string 表名
 */
func (AsyncTask) TableName() string {
	return "async_tasks"
}
