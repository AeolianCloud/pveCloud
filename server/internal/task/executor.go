package task

import (
	"context"
	"fmt"
	"time"
)

type Executor interface {
	Execute(ctx context.Context, row Task) error
}

type CreateInstanceFunc func(ctx context.Context, orderID uint64) error

type DispatchingExecutor struct {
	createInstance CreateInstanceFunc
}

func NewDispatchingExecutor(createInstance CreateInstanceFunc) *DispatchingExecutor {
	return &DispatchingExecutor{createInstance: createInstance}
}

func (e *DispatchingExecutor) Execute(ctx context.Context, row Task) error {
	switch row.TaskType {
	case "create_instance":
		if e.createInstance == nil {
			return fmt.Errorf("create_instance handler is not configured")
		}
		return e.createInstance(ctx, row.BusinessID)
	default:
		return fmt.Errorf("unsupported task type: %s", row.TaskType)
	}
}

type RetryableError struct {
	Err   error
	Delay time.Duration
}

func (e *RetryableError) Error() string {
	if e == nil || e.Err == nil {
		return "retryable task error"
	}
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func Retryable(err error, delay time.Duration) error {
	if err == nil {
		return nil
	}
	if delay <= 0 {
		delay = time.Minute
	}
	return &RetryableError{
		Err:   err,
		Delay: delay,
	}
}
