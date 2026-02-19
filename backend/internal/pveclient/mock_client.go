package pveclient

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MockPVEClient 提供可预测的模拟实现，便于前后端在无真实 PVE 环境下联调。
type MockPVEClient struct {
	taskCalls sync.Map
}

// NewMockClient 创建 MockPVEClient 实例。
func NewMockClient() *MockPVEClient {
	return &MockPVEClient{}
}

// CreateInstance 模拟提交实例创建任务。
func (m *MockPVEClient) CreateInstance(_ context.Context, req CreateInstanceReq) (*TaskResult, error) {
	id := fmt.Sprintf("mock-inst-%d", time.Now().UnixNano())
	tid := fmt.Sprintf("mock-task-%d", time.Now().UnixNano())
	return &TaskResult{TaskID: tid, PveTaskID: tid, InstanceID: id, Description: "mock create instance queued"}, nil
}

func (m *MockPVEClient) StartInstance(_ context.Context, instanceID string) (*TaskResult, error) {
	return &TaskResult{TaskID: "mock-start-" + instanceID, PveTaskID: "mock-start-" + instanceID, Description: "mock start queued"}, nil
}

func (m *MockPVEClient) StopInstance(_ context.Context, instanceID string) (*TaskResult, error) {
	return &TaskResult{TaskID: "mock-stop-" + instanceID, PveTaskID: "mock-stop-" + instanceID, Description: "mock stop queued"}, nil
}

func (m *MockPVEClient) RebootInstance(_ context.Context, instanceID string) (*TaskResult, error) {
	return &TaskResult{TaskID: "mock-reboot-" + instanceID, PveTaskID: "mock-reboot-" + instanceID, Description: "mock reboot queued"}, nil
}

func (m *MockPVEClient) DeleteInstance(_ context.Context, instanceID string) (*TaskResult, error) {
	return &TaskResult{TaskID: "mock-delete-" + instanceID, PveTaskID: "mock-delete-" + instanceID, Description: "mock delete queued"}, nil
}

func (m *MockPVEClient) GetInstanceStatus(_ context.Context, instanceID string) (*InstanceStatus, error) {
	return &InstanceStatus{InstanceID: instanceID, Status: "running", IP: "10.0.0.8", Node: "node-a"}, nil
}

func (m *MockPVEClient) GetInstanceMetrics(_ context.Context, _ string) (*Metrics, error) {
	return &Metrics{CPUUsage: 0.21, MemoryUsage: 0.45, DiskUsage: 0.60, NetworkInKB: 1200, NetworkOutKB: 800}, nil
}

func (m *MockPVEClient) GetConsoleToken(_ context.Context, instanceID string) (*ConsoleInfo, error) {
	return &ConsoleInfo{URL: "https://novnc.example/mock/" + instanceID, Ticket: "mock-ticket-123"}, nil
}

func (m *MockPVEClient) CreateSnapshot(_ context.Context, instanceID, name string) (*TaskResult, error) {
	id := "mock-snap-create-" + instanceID + "-" + name
	return &TaskResult{TaskID: id, PveTaskID: id, Description: "mock snapshot create queued"}, nil
}

func (m *MockPVEClient) RestoreSnapshot(_ context.Context, instanceID, snapshotName string) (*TaskResult, error) {
	id := "mock-snap-restore-" + instanceID + "-" + snapshotName
	return &TaskResult{TaskID: id, PveTaskID: id, Description: "mock snapshot restore queued"}, nil
}

func (m *MockPVEClient) ListSnapshots(_ context.Context, _ string) ([]*Snapshot, error) {
	return []*Snapshot{{Name: "snap-001", Status: "available", CreatedAt: time.Now().Add(-2 * time.Hour).Format(time.RFC3339)}}, nil
}

func (m *MockPVEClient) DeleteSnapshot(_ context.Context, instanceID, snapshotName string) (*TaskResult, error) {
	id := "mock-snap-delete-" + instanceID + "-" + snapshotName
	return &TaskResult{TaskID: id, PveTaskID: id, Description: "mock snapshot delete queued"}, nil
}

// GetTaskStatus 首次调用返回 running，后续调用返回 success，用于模拟异步任务推进。
func (m *MockPVEClient) GetTaskStatus(_ context.Context, taskID string) (*TaskStatus, error) {
	v, ok := m.taskCalls.Load(taskID)
	if !ok {
		m.taskCalls.Store(taskID, 1)
		return &TaskStatus{TaskID: taskID, Status: "running", Progress: 50, Message: "task is running"}, nil
	}
	m.taskCalls.Store(taskID, v.(int)+1)
	return &TaskStatus{TaskID: taskID, Status: "success", Progress: 100, Message: "task completed"}, nil
}

func (m *MockPVEClient) GetNodeStatus(_ context.Context, node string) (*NodeStatus, error) {
	return &NodeStatus{Node: node, CPUUsage: 0.33, MemoryUsage: 0.47, DiskUsage: 0.58}, nil
}
