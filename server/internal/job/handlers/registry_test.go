package handlers

import (
	"context"
	"strings"
	"testing"

	"github.com/AeolianCloud/pveCloud/server/internal/job"
)

func TestNewRegistryRegistersBuiltInHandlers(t *testing.T) {
	registry := NewRegistry()

	taskTypes := []string{
		job.TaskTypeInstanceCreate,
		job.TaskTypeInstanceRenew,
		job.TaskTypeOrderExpire,
		job.TaskTypePaymentCheck,
		job.TaskTypeInstanceStatusSync,
	}
	for _, taskType := range taskTypes {
		if registry[taskType] == nil {
			t.Fatalf("expected built-in handler registered for %s", taskType)
		}
		result := registry[taskType](context.Background(), job.AsyncTask{TaskType: taskType})
		if result.Retryable {
			t.Fatalf("expected %s stub failure to be permanent", taskType)
		}
		if result.Err == nil {
			t.Fatalf("expected %s stub to return an error", taskType)
		}
		if !strings.Contains(result.Err.Error(), taskType) {
			t.Fatalf("expected %s stub error to mention task type, got %q", taskType, result.Err.Error())
		}
	}
}

func TestNewRegistryRegistersProvidedHandlers(t *testing.T) {
	registry := NewRegistry(Func(func(r job.Registry) {
		r.Register(job.TaskTypeInstanceStatusSync, func(ctx context.Context, task job.AsyncTask) job.HandlerResult {
			return job.Succeeded(nil)
		})
	}))

	if registry[job.TaskTypeInstanceStatusSync] == nil {
		t.Fatal("expected task handler registered from handlers boundary")
	}
	result := registry[job.TaskTypeInstanceStatusSync](context.Background(), job.AsyncTask{TaskType: job.TaskTypeInstanceStatusSync})
	if result.Err != nil {
		t.Fatalf("expected override handler success, got %v", result.Err)
	}
}
