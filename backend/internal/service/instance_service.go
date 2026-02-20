package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pvecloud/backend/internal/model"
	"pvecloud/backend/internal/pveclient"
	"pvecloud/backend/internal/repository"
)

// InstanceService 封装实例操作、快照管理、控制台访问等逻辑。
type InstanceService struct {
	instanceRepo *repository.InstanceRepository
	snapshotRepo *repository.SnapshotRepository
	pve          pveclient.PVEClient
}

// NewInstanceService 创建实例服务。
func NewInstanceService(instanceRepo *repository.InstanceRepository, snapshotRepo *repository.SnapshotRepository, pve pveclient.PVEClient) *InstanceService {
	return &InstanceService{instanceRepo: instanceRepo, snapshotRepo: snapshotRepo, pve: pve}
}

// ListUserInstances 查询用户实例列表。
func (s *InstanceService) ListUserInstances(ctx context.Context, userID uint) ([]model.Instance, error) {
	return s.instanceRepo.ListByUser(ctx, userID)
}

// GetUserInstance 查询单实例并验证归属。
func (s *InstanceService) GetUserInstance(ctx context.Context, userID uint, instanceID uint) (*model.Instance, error) {
	inst, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	if inst.UserID != userID {
		return nil, WrapForbidden("无权限访问该实例")
	}
	status, err := s.pve.GetInstanceStatus(ctx, inst.PVEInstanceID)
	if err == nil {
		inst.Status = status.Status
		inst.IP = status.IP
		_ = s.instanceRepo.Update(ctx, inst)
	}
	return inst, nil
}

// Operate 执行开关机重启操作。
func (s *InstanceService) Operate(ctx context.Context, userID uint, instanceID uint, action string) (*pveclient.TaskResult, error) {
	inst, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	if inst.UserID != userID {
		return nil, WrapForbidden("无权限操作该实例")
	}

	var result *pveclient.TaskResult
	switch action {
	case "start":
		result, err = s.pve.StartInstance(ctx, inst.PVEInstanceID)
		inst.Status = "starting"
	case "stop":
		result, err = s.pve.StopInstance(ctx, inst.PVEInstanceID)
		inst.Status = "stopping"
	case "reboot":
		result, err = s.pve.RebootInstance(ctx, inst.PVEInstanceID)
		inst.Status = "rebooting"
	default:
		return nil, errors.New("不支持的操作")
	}
	if err != nil {
		return nil, err
	}
	_ = s.instanceRepo.Update(ctx, inst)
	return result, nil
}

// GetConsole 获取控制台 token/url。
func (s *InstanceService) GetConsole(ctx context.Context, userID uint, instanceID uint) (*pveclient.ConsoleInfo, error) {
	inst, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	if inst.UserID != userID {
		return nil, WrapForbidden("无权限访问该实例")
	}
	return s.pve.GetConsoleToken(ctx, inst.PVEInstanceID)
}

// ListSnapshots 查询实例快照。
func (s *InstanceService) ListSnapshots(ctx context.Context, userID uint, instanceID uint) ([]model.InstanceSnapshot, error) {
	inst, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	if inst.UserID != userID {
		return nil, WrapForbidden("无权限访问该实例")
	}
	return s.snapshotRepo.ListByInstance(ctx, instanceID)
}

// CreateSnapshot 创建快照并写本地记录。
func (s *InstanceService) CreateSnapshot(ctx context.Context, userID uint, instanceID uint, name string) (*pveclient.TaskResult, error) {
	inst, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	if inst.UserID != userID {
		return nil, WrapForbidden("无权限访问该实例")
	}
	result, err := s.pve.CreateSnapshot(ctx, inst.PVEInstanceID, name)
	if err != nil {
		return nil, err
	}
	snapshot := &model.InstanceSnapshot{InstanceID: instanceID, Name: name, Status: "creating"}
	_ = s.snapshotRepo.Create(ctx, snapshot)
	return result, nil
}

// RestoreSnapshot 恢复快照。
func (s *InstanceService) RestoreSnapshot(ctx context.Context, userID uint, instanceID uint, name string) (*pveclient.TaskResult, error) {
	inst, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	if inst.UserID != userID {
		return nil, WrapForbidden("无权限访问该实例")
	}
	return s.pve.RestoreSnapshot(ctx, inst.PVEInstanceID, name)
}

// DeleteSnapshot 删除快照。
func (s *InstanceService) DeleteSnapshot(ctx context.Context, userID uint, instanceID uint, name string) (*pveclient.TaskResult, error) {
	inst, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	if inst.UserID != userID {
		return nil, WrapForbidden("无权限访问该实例")
	}
	result, err := s.pve.DeleteSnapshot(ctx, inst.PVEInstanceID, name)
	if err != nil {
		return nil, err
	}
	_ = s.snapshotRepo.DeleteByName(ctx, instanceID, name)
	return result, nil
}

// SyncStatusForJob 给定时任务调用，刷新实例状态。
func (s *InstanceService) SyncStatusForJob(ctx context.Context, instance model.Instance) error {
	status, err := s.pve.GetInstanceStatus(ctx, instance.PVEInstanceID)
	if err != nil {
		return err
	}
	instance.Status = status.Status
	instance.IP = status.IP
	return s.instanceRepo.Update(ctx, &instance)
}

// HandleExpireForJob 到期处理：提醒、关机、超期销毁。
func (s *InstanceService) HandleExpireForJob(ctx context.Context, instance model.Instance) error {
	if instance.ExpireAt == nil {
		return nil
	}
	now := time.Now()
	d := instance.ExpireAt.Sub(now)
	if d <= 72*time.Hour && d > 24*time.Hour {
		fmt.Printf("notify user %d: instance %d will expire in 3 days\n", instance.UserID, instance.ID)
	}
	if d <= 24*time.Hour && d > 0 {
		fmt.Printf("notify user %d: instance %d will expire in 1 day\n", instance.UserID, instance.ID)
	}
	if d <= 0 && d > -72*time.Hour {
		_, _ = s.pve.StopInstance(ctx, instance.PVEInstanceID)
		instance.Status = "expired"
		return s.instanceRepo.Update(ctx, &instance)
	}
	if d <= -72*time.Hour {
		_, _ = s.pve.DeleteInstance(ctx, instance.PVEInstanceID)
		now := time.Now()
		instance.DeletedAt = &now
		instance.Status = "deleted"
		return s.instanceRepo.Update(ctx, &instance)
	}
	return nil
}
