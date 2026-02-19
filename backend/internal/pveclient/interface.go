package pveclient

import "context"

// CreateInstanceReq 表示创建实例时提交给 PVE 能力层的参数快照。
type CreateInstanceReq struct {
	Name          string `json:"name"`
	CPU           int    `json:"cpu"`
	MemoryMB      int    `json:"memory_mb"`
	DiskGB        int    `json:"disk_gb"`
	BandwidthMbps int    `json:"bandwidth_mbps"`
	Template      string `json:"template"`
	Password      string `json:"password"`
	RegionCode    string `json:"region_code"`
}

// TaskResult 描述 PVE 异步任务的提交结果。
type TaskResult struct {
	TaskID      string `json:"task_id"`
	PveTaskID   string `json:"pve_task_id"`
	InstanceID  string `json:"instance_id,omitempty"`
	Description string `json:"description"`
}

// InstanceStatus 描述实例当前运行态。
type InstanceStatus struct {
	InstanceID string `json:"instance_id"`
	Status     string `json:"status"`
	IP         string `json:"ip"`
	Node       string `json:"node"`
}

// Metrics 描述实例资源监控快照。
type Metrics struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	DiskUsage    float64 `json:"disk_usage"`
	NetworkInKB  int64   `json:"network_in_kb"`
	NetworkOutKB int64   `json:"network_out_kb"`
}

// ConsoleInfo 表示控制台直连所需信息。
type ConsoleInfo struct {
	URL    string `json:"url"`
	Ticket string `json:"ticket"`
}

// Snapshot 描述实例快照条目。
type Snapshot struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// TaskStatus 描述异步任务运行状态。
type TaskStatus struct {
	TaskID   string `json:"task_id"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
	Message  string `json:"message"`
}

// NodeStatus 描述 PVE 节点资源情况。
type NodeStatus struct {
	Node        string  `json:"node"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
}

// PVEClient 统一抽象所有 PVE 能力调用，业务层只依赖该接口。
type PVEClient interface {
	CreateInstance(ctx context.Context, req CreateInstanceReq) (*TaskResult, error)
	StartInstance(ctx context.Context, instanceID string) (*TaskResult, error)
	StopInstance(ctx context.Context, instanceID string) (*TaskResult, error)
	RebootInstance(ctx context.Context, instanceID string) (*TaskResult, error)
	DeleteInstance(ctx context.Context, instanceID string) (*TaskResult, error)
	GetInstanceStatus(ctx context.Context, instanceID string) (*InstanceStatus, error)
	GetInstanceMetrics(ctx context.Context, instanceID string) (*Metrics, error)
	GetConsoleToken(ctx context.Context, instanceID string) (*ConsoleInfo, error)
	CreateSnapshot(ctx context.Context, instanceID, name string) (*TaskResult, error)
	RestoreSnapshot(ctx context.Context, instanceID, snapshotName string) (*TaskResult, error)
	ListSnapshots(ctx context.Context, instanceID string) ([]*Snapshot, error)
	DeleteSnapshot(ctx context.Context, instanceID, snapshotName string) (*TaskResult, error)
	GetTaskStatus(ctx context.Context, taskID string) (*TaskStatus, error)
	GetNodeStatus(ctx context.Context, node string) (*NodeStatus, error)
}
