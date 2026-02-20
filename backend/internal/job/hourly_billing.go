package job

import (
	"context"
	"log"
	"time"

	"pvecloud/backend/internal/pveclient"
	"pvecloud/backend/internal/repository"
	"pvecloud/backend/internal/service"
)

// HourlyBilling 每小时扣除按小时计费实例费用，余额不足则关机。
type HourlyBilling struct {
	instanceRepo *repository.InstanceRepository
	billing      *service.BillingService
	pve          pveclient.PVEClient
}

// NewHourlyBilling 创建小时计费任务。
func NewHourlyBilling(instanceRepo *repository.InstanceRepository, billing *service.BillingService, pve pveclient.PVEClient) *HourlyBilling {
	return &HourlyBilling{instanceRepo: instanceRepo, billing: billing, pve: pve}
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
		// 当前以固定 1 元/小时扣费，后续可替换为按实例套餐动态计算。
		hourlyPrice := 1.0
		if err := j.billing.Consume(ctx, inst.UserID, hourlyPrice, &inst.OrderID, "按小时计费扣费"); err != nil {
			log.Printf("hourly billing insufficient balance, user_id=%d instance_id=%d", inst.UserID, inst.ID)
			if inst.PVEInstanceID != "" {
				if _, stopErr := j.pve.StopInstance(ctx, inst.PVEInstanceID); stopErr != nil {
					log.Printf("hourly billing stop instance failed, instance_id=%d err=%v", inst.ID, stopErr)
				}
			}
			inst.Status = "suspended"
			_ = j.instanceRepo.Update(ctx, &inst)
			log.Printf("notify user_id=%d balance insufficient, instance_id=%d suspended", inst.UserID, inst.ID)
		} else if inst.Status == "suspended" {
			if inst.PVEInstanceID != "" {
				if _, startErr := j.pve.StartInstance(ctx, inst.PVEInstanceID); startErr != nil {
					log.Printf("hourly billing resume instance failed, instance_id=%d err=%v", inst.ID, startErr)
					continue
				}
			}
			inst.Status = "running"
			_ = j.instanceRepo.Update(ctx, &inst)
			log.Printf("notify user_id=%d balance recovered, instance_id=%d resumed", inst.UserID, inst.ID)
		}
	}
}
