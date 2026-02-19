package job

import (
	"context"
	"time"

	"pvecloud/backend/internal/pveclient"
	"pvecloud/backend/internal/repository"
)

// TaskSyncer 每 10 秒扫描 pending/running 任务，调用 PVE Client 获取状态并回写业务数据。
type TaskSyncer struct {
	taskRepo     *repository.TaskRepository
	orderRepo    *repository.OrderRepository
	instanceRepo *repository.InstanceRepository
	walletRepo   *repository.WalletRepository
	pve          pveclient.PVEClient
}

// NewTaskSyncer 创建任务同步器。
func NewTaskSyncer(
	taskRepo *repository.TaskRepository,
	orderRepo *repository.OrderRepository,
	instanceRepo *repository.InstanceRepository,
	walletRepo *repository.WalletRepository,
	pve pveclient.PVEClient,
) *TaskSyncer {
	return &TaskSyncer{taskRepo: taskRepo, orderRepo: orderRepo, instanceRepo: instanceRepo, walletRepo: walletRepo, pve: pve}
}

// Start 启动循环扫描。
func (j *TaskSyncer) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			j.syncOnce(ctx)
		}
	}
}

func (j *TaskSyncer) syncOnce(ctx context.Context) {
	tasks, err := j.taskRepo.ListPendingAndRunning(ctx)
	if err != nil {
		return
	}
	for _, task := range tasks {
		if task.PveTaskID == "" {
			continue
		}
		status, err := j.pve.GetTaskStatus(ctx, task.PveTaskID)
		if err != nil {
			continue
		}
		task.Status = status.Status
		task.Progress = status.Progress
		task.Message = status.Message
		_ = j.taskRepo.Update(ctx, &task)

		if task.Status == "success" {
			if task.OrderID != nil {
				if order, err := j.orderRepo.GetByID(ctx, *task.OrderID); err == nil {
					order.Status = "active"
					_ = j.orderRepo.Update(ctx, order)
				}
			}
			if task.InstanceID != nil {
				if inst, err := j.instanceRepo.GetByID(ctx, *task.InstanceID); err == nil {
					inst.Status = "active"
					_ = j.instanceRepo.Update(ctx, inst)
				}
			}
		}

		if task.Status == "failed" && task.OrderID != nil {
			if order, err := j.orderRepo.GetByID(ctx, *task.OrderID); err == nil {
				order.Status = "failed"
				_ = j.orderRepo.Update(ctx, order)
				_ = j.walletRepo.ChangeBalance(ctx, order.UserID, order.Amount, "refund", task.OrderID, "任务失败自动退款")
			}
		}
	}
}
