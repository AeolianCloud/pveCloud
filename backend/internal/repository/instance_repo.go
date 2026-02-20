package repository

import (
	"context"

	"gorm.io/gorm"
	"pvecloud/backend/internal/model"
)

// InstanceRepository 封装实例相关数据库操作。
type InstanceRepository struct {
	db *gorm.DB
}

// NewInstanceRepository 创建实例仓储。
func NewInstanceRepository(db *gorm.DB) *InstanceRepository {
	return &InstanceRepository{db: db}
}

// Create 新建实例记录。
func (r *InstanceRepository) Create(ctx context.Context, inst *model.Instance) error {
	return r.db.WithContext(ctx).Create(inst).Error
}

// GetByID 查询实例。
func (r *InstanceRepository) GetByID(ctx context.Context, id uint) (*model.Instance, error) {
	var inst model.Instance
	err := r.db.WithContext(ctx).First(&inst, id).Error
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

// ListByUser 查询用户实例。
func (r *InstanceRepository) ListByUser(ctx context.Context, userID uint) ([]model.Instance, error) {
	var list []model.Instance
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("id DESC").Find(&list).Error
	return list, err
}

// Update 更新实例。
func (r *InstanceRepository) Update(ctx context.Context, inst *model.Instance) error {
	return r.db.WithContext(ctx).Save(inst).Error
}

// ListNeedExpireHandling 查询需要到期处理的实例。
func (r *InstanceRepository) ListNeedExpireHandling(ctx context.Context) ([]model.Instance, error) {
	var list []model.Instance
	err := r.db.WithContext(ctx).Where("expire_at IS NOT NULL").Find(&list).Error
	return list, err
}

// ListHourlyBillingTargets 查询按小时计费实例。
func (r *InstanceRepository) ListHourlyBillingTargets(ctx context.Context) ([]model.Instance, error) {
	var list []model.Instance
	err := r.db.WithContext(ctx).Where("status IN ?", []string{"running", "suspended"}).Find(&list).Error
	return list, err
}

// ListForStatusSync 查询需要进行状态同步的实例（排除已删除实例）。
func (r *InstanceRepository) ListForStatusSync(ctx context.Context) ([]model.Instance, error) {
	var list []model.Instance
	err := r.db.WithContext(ctx).Where("status <> ?", "deleted").Find(&list).Error
	return list, err
}
