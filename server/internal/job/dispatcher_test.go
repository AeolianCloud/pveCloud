package job

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"
)

func TestDispatcherMarksSucceededTask(t *testing.T) {
	now := time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC)
	store := newFakeStore([]AsyncTask{{ID: 1, TaskType: "ok", MaxRetries: 3}})
	dispatcher := newTestDispatcher(store, map[string]Handler{
		"ok": func(ctx context.Context, task AsyncTask) HandlerResult {
			return Succeeded(map[string]any{"done": true})
		},
	}, now)

	if err := dispatcher.RunOnce(context.Background()); err != nil {
		t.Fatalf("expected run once success, got %v", err)
	}
	if store.succeeded != 1 || store.failed != 0 || store.retryable != 0 {
		t.Fatalf("expected one succeeded task, got succeeded=%d failed=%d retry=%d", store.succeeded, store.failed, store.retryable)
	}
}

func TestDispatcherMarksRetryableFailure(t *testing.T) {
	now := time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC)
	store := newFakeStore([]AsyncTask{{ID: 1, TaskType: "retry", RetryCount: 1, MaxRetries: 3}})
	dispatcher := newTestDispatcher(store, map[string]Handler{
		"retry": func(ctx context.Context, task AsyncTask) HandlerResult {
			return RetryableFailure(errors.New("temporary outage"))
		},
	}, now)

	if err := dispatcher.RunOnce(context.Background()); err != nil {
		t.Fatalf("expected run once success, got %v", err)
	}
	if store.retryable != 1 || store.failed != 0 {
		t.Fatalf("expected retryable failure, got failed=%d retry=%d", store.failed, store.retryable)
	}
	expectedNextRun := now.Add(4 * time.Second)
	if !store.nextRunAt.Equal(expectedNextRun) {
		t.Fatalf("expected next run %s, got %s", expectedNextRun, store.nextRunAt)
	}
}

func TestDispatcherMarksFailedWhenRetriesExhausted(t *testing.T) {
	now := time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC)
	store := newFakeStore([]AsyncTask{{ID: 1, TaskType: "retry", RetryCount: 2, MaxRetries: 3}})
	dispatcher := newTestDispatcher(store, map[string]Handler{
		"retry": func(ctx context.Context, task AsyncTask) HandlerResult {
			return RetryableFailure(errors.New("still failing"))
		},
	}, now)

	if err := dispatcher.RunOnce(context.Background()); err != nil {
		t.Fatalf("expected run once success, got %v", err)
	}
	if store.failed != 1 || store.retryable != 0 {
		t.Fatalf("expected exhausted task failed, got failed=%d retry=%d", store.failed, store.retryable)
	}
}

func TestDispatcherFailsUnregisteredTaskType(t *testing.T) {
	now := time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC)
	store := newFakeStore([]AsyncTask{{ID: 1, TaskType: "missing", MaxRetries: 3}})
	dispatcher := newTestDispatcher(store, nil, now)

	if err := dispatcher.RunOnce(context.Background()); err != nil {
		t.Fatalf("expected run once success, got %v", err)
	}
	if store.failed != 1 {
		t.Fatalf("expected unregistered task failed, got %d", store.failed)
	}
	if !errors.Is(store.lastErr, ErrUnregisteredHandler) {
		t.Fatalf("expected unregistered handler error, got %v", store.lastErr)
	}
}

func TestRegistryRegisterAddsHandler(t *testing.T) {
	registry := NewRegistry()
	registry.Register("registered", func(ctx context.Context, task AsyncTask) HandlerResult {
		return Succeeded(nil)
	})
	if registry["registered"] == nil {
		t.Fatal("expected handler registered")
	}
}

func newTestDispatcher(store *fakeStore, handlers map[string]Handler, now time.Time) *Dispatcher {
	dispatcher := NewDispatcher(Config{
		WorkerID:     "worker-test",
		PollInterval: 2 * time.Second,
		LockTTL:      time.Minute,
		BatchSize:    10,
	}, store, handlers, slog.New(slog.NewTextHandler(testDiscard{}, nil)))
	dispatcher.now = func() time.Time { return now }
	return dispatcher
}

type fakeStore struct {
	queue     []AsyncTask
	succeeded int
	retryable int
	failed    int
	nextRunAt time.Time
	lastErr   error
}

func newFakeStore(tasks []AsyncTask) *fakeStore {
	return &fakeStore{queue: tasks}
}

func (s *fakeStore) Claim(ctx context.Context, now time.Time) (AsyncTask, error) {
	if len(s.queue) == 0 {
		return AsyncTask{}, ErrNoTask
	}
	task := s.queue[0]
	s.queue = s.queue[1:]
	return task, nil
}

func (s *fakeStore) MarkSucceeded(ctx context.Context, task AsyncTask, result any, now time.Time) error {
	s.succeeded++
	return nil
}

func (s *fakeStore) MarkRetryableFailure(ctx context.Context, task AsyncTask, err error, nextRunAt time.Time, now time.Time) error {
	s.retryable++
	s.nextRunAt = nextRunAt
	s.lastErr = err
	return nil
}

func (s *fakeStore) MarkFailed(ctx context.Context, task AsyncTask, err error, now time.Time) error {
	s.failed++
	s.lastErr = err
	return nil
}

type testDiscard struct{}

func (testDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}
