package job

import (
	"context"
	"time"

	"pvecloud/backend/internal/repository"
	"pvecloud/backend/internal/service"
)

// ExpireChecker 每日扫描实例到期状态并执行提醒/关机/销毁。
type ExpireChecker struct {
	instanceRepo    *repository.InstanceRepository
	instanceService *service.InstanceService
}

// NewExpireChecker 创建到期检查任务。
func NewExpireChecker(instanceRepo *repository.InstanceRepository, instanceService *service.InstanceService) *ExpireChecker {
	return &ExpireChecker{instanceRepo: instanceRepo, instanceService: instanceService}
}

// Start 启动每日任务。
func (j *ExpireChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			j.scan(ctx)
		}
	}
}

func (j *ExpireChecker) scan(ctx context.Context) {
	instances, err := j.instanceRepo.ListNeedExpireHandling(ctx)
	if err != nil {
		return
	}
	for _, item := range instances {
		_ = j.instanceService.HandleExpireForJob(ctx, item)
	}
}
