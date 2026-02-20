package job

import (
	"context"
	"log"
	"time"

	"pvecloud/backend/internal/repository"
	"pvecloud/backend/internal/service"
)

// InstanceStatusSyncer 每分钟同步实例状态，保持本地实例状态与 PVE 一致。
type InstanceStatusSyncer struct {
	instanceRepo    *repository.InstanceRepository
	instanceService *service.InstanceService
}

// NewInstanceStatusSyncer 创建实例状态同步任务。
func NewInstanceStatusSyncer(instanceRepo *repository.InstanceRepository, instanceService *service.InstanceService) *InstanceStatusSyncer {
	return &InstanceStatusSyncer{instanceRepo: instanceRepo, instanceService: instanceService}
}

// Start 启动每分钟实例状态同步任务。
func (j *InstanceStatusSyncer) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			j.sync(ctx)
		}
	}
}

func (j *InstanceStatusSyncer) sync(ctx context.Context) {
	instances, err := j.instanceRepo.ListForStatusSync(ctx)
	if err != nil {
		log.Printf("instance status sync list failed: %v", err)
		return
	}
	for _, item := range instances {
		if err := j.instanceService.SyncStatusForJob(ctx, item); err != nil {
			// PVE 不可达时保留上次状态，只记录日志。
			log.Printf("instance status sync failed, instance_id=%d pve_instance_id=%s err=%v", item.ID, item.PVEInstanceID, err)
		}
	}
}
