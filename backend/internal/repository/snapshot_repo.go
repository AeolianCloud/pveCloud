package repository

import (
	"context"

	"gorm.io/gorm"
	"pvecloud/backend/internal/model"
)

// SnapshotRepository 封装快照记录读写。
type SnapshotRepository struct {
	db *gorm.DB
}

// NewSnapshotRepository 创建快照仓储。
func NewSnapshotRepository(db *gorm.DB) *SnapshotRepository {
	return &SnapshotRepository{db: db}
}

// Create 新建快照记录。
func (r *SnapshotRepository) Create(ctx context.Context, snapshot *model.InstanceSnapshot) error {
	return r.db.WithContext(ctx).Create(snapshot).Error
}

// ListByInstance 查询实例快照列表。
func (r *SnapshotRepository) ListByInstance(ctx context.Context, instanceID uint) ([]model.InstanceSnapshot, error) {
	var snapshots []model.InstanceSnapshot
	err := r.db.WithContext(ctx).Where("instance_id = ?", instanceID).Order("created_at DESC").Find(&snapshots).Error
	return snapshots, err
}

// DeleteByName 删除指定快照。
func (r *SnapshotRepository) DeleteByName(ctx context.Context, instanceID uint, name string) error {
	return r.db.WithContext(ctx).Where("instance_id = ? AND name = ?", instanceID, name).Delete(&model.InstanceSnapshot{}).Error
}
