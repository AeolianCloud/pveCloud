package job

import (
	"context"
	"errors"
	"time"
)

const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
)

var (
	ErrNoTask              = errors.New("no pending async task")
	ErrUnregisteredHandler = errors.New("unregistered async task handler")
)

/**
 * Config 表示 Worker 调度所需配置。
 */
type Config struct {
	WorkerID     string
	PollInterval time.Duration
	LockTTL      time.Duration
	BatchSize    int
}

/**
 * Store 定义异步任务持久化操作。
 */
type Store interface {
	Claim(ctx context.Context, now time.Time) (AsyncTask, error)
	MarkSucceeded(ctx context.Context, task AsyncTask, result any, now time.Time) error
	MarkRetryableFailure(ctx context.Context, task AsyncTask, err error, nextRunAt time.Time, now time.Time) error
	MarkFailed(ctx context.Context, task AsyncTask, err error, now time.Time) error
}

/**
 * Handler 执行具体任务类型的业务逻辑。
 */
type Handler func(ctx context.Context, task AsyncTask) HandlerResult

/**
 * Registry 保存任务类型到 handler 的注册关系。
 */
type Registry map[string]Handler

/**
 * NewRegistry 创建空任务 handler 注册表。
 */
func NewRegistry() Registry {
	return Registry{}
}

/**
 * Register 注册任务 handler。
 */
func (r Registry) Register(taskType string, handler Handler) {
	if r == nil || handler == nil {
		return
	}
	r[taskType] = handler
}

/**
 * HandlerResult 表示任务 handler 的执行结果。
 */
type HandlerResult struct {
	Result    any
	Err       error
	Retryable bool
}

func Succeeded(result any) HandlerResult {
	return HandlerResult{Result: result}
}

func RetryableFailure(err error) HandlerResult {
	return HandlerResult{Err: err, Retryable: true}
}

func PermanentFailure(err error) HandlerResult {
	return HandlerResult{Err: err}
}
