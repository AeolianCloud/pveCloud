package task

import (
	"context"
	"testing"
)

func TestDispatchingExecutorRoutesCreateInstance(t *testing.T) {
	var calledOrderID uint64
	executor := NewDispatchingExecutor(func(ctx context.Context, orderID uint64) error {
		calledOrderID = orderID
		return nil
	})

	err := executor.Execute(context.Background(), Task{
		TaskType:     "create_instance",
		BusinessType: "order",
		BusinessID:   5001,
	})
	if err != nil {
		t.Fatalf("execute task: %v", err)
	}
	if calledOrderID != 5001 {
		t.Fatalf("expected order id 5001, got %d", calledOrderID)
	}
}
