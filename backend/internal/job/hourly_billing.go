package job

import (
	"context"
	"time"

	"pvecloud/backend/internal/repository"
	"pvecloud/backend/internal/service"
)

// HourlyBilling 每小时扣除按小时计费实例费用，余额不足则关机。
type HourlyBilling struct {
	instanceRepo *repository.InstanceRepository
	billing      *service.BillingService
}

// NewHourlyBilling 创建小时计费任务。
func NewHourlyBilling(instanceRepo *repository.InstanceRepository, billing *service.BillingService) *HourlyBilling {
	return &HourlyBilling{instanceRepo: instanceRepo, billing: billing}
}

// Start 启动每小时扫描任务。
func (j *HourlyBilling) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			j.bill(ctx)
		}
	}
}

func (j *HourlyBilling) bill(ctx context.Context) {
	instances, err := j.instanceRepo.ListHourlyBillingTargets(ctx)
	if err != nil {
		return
	}
	for _, inst := range instances {
		// TODO: 真实实现应按实例绑定产品价格计算小时费率。
		hourlyPrice := 1.0
		if err := j.billing.Consume(ctx, inst.UserID, hourlyPrice, &inst.OrderID, "按小时计费扣费"); err != nil {
			inst.Status = "suspended"
			_ = j.instanceRepo.Update(ctx, &inst)
		} else if inst.Status == "suspended" {
			inst.Status = "running"
			_ = j.instanceRepo.Update(ctx, &inst)
		}
	}
}
