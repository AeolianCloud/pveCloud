package task_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/task"
)

type fakeTaskRepo struct {
	tasks map[string]task.Task
}

func (f *fakeTaskRepo) FindByBusinessKey(ctx context.Context, taskType, businessType string, businessID uint64) (task.Task, bool, error) {
	if f.tasks == nil {
		f.tasks = make(map[string]task.Task)
	}
	key := taskType + ":" + businessType + ":" + strconv.FormatUint(businessID, 10)
	value, ok := f.tasks[key]
	return value, ok, nil
}

func (f *fakeTaskRepo) CreateTask(ctx context.Context, in task.CreateTaskParams) (task.Task, error) {
	if f.tasks == nil {
		f.tasks = make(map[string]task.Task)
	}
	row := task.Task{
		ID:           8001,
		TaskNo:       "T8001",
		TaskType:     in.TaskType,
		BusinessType: in.BusinessType,
		BusinessID:   in.BusinessID,
		Status:       in.Status,
		Payload:      in.Payload,
		NextRunAt:    in.NextRunAt,
		MaxRetryCount: in.MaxRetryCount,
	}
	key := in.TaskType + ":" + in.BusinessType + ":" + strconv.FormatUint(in.BusinessID, 10)
	f.tasks[key] = row
	return row, nil
}

func TestCreateUniqueTaskForBusinessKey(t *testing.T) {
	svc := task.NewService(&fakeTaskRepo{})
	first, err := svc.CreateTask(context.Background(), task.CreateInput{
		TaskType:     "create_instance",
		BusinessType: "order",
		BusinessID:   5001,
	})
	if err != nil {
		t.Fatalf("create first task: %v", err)
	}

	second, err := svc.CreateTask(context.Background(), task.CreateInput{
		TaskType:     "create_instance",
		BusinessType: "order",
		BusinessID:   5001,
	})
	if err != nil {
		t.Fatalf("create second task: %v", err)
	}
	if first.TaskNo != second.TaskNo {
		t.Fatalf("expected same task no, got %s and %s", first.TaskNo, second.TaskNo)
	}
	if second.Status != "pending" {
		t.Fatalf("expected pending status, got %s", second.Status)
	}
	if second.NextRunAt.Before(time.Now().Add(-time.Minute)) {
		t.Fatalf("expected next run time to be current or future")
	}
}
