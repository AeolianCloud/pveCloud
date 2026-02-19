package repository

import (
	"context"

	"gorm.io/gorm"
	"pvecloud/backend/internal/model"
)

// TaskRepository 封装任务状态读写。
type TaskRepository struct {
	db *gorm.DB
}

// NewTaskRepository 创建任务仓储。
func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// Create 新建任务。
func (r *TaskRepository) Create(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// GetByID 查询任务。
func (r *TaskRepository) GetByID(ctx context.Context, id uint) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Update 更新任务状态。
func (r *TaskRepository) Update(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// ListPendingAndRunning 查询待同步的任务。
func (r *TaskRepository) ListPendingAndRunning(ctx context.Context) ([]model.Task, error) {
	var tasks []model.Task
	err := r.db.WithContext(ctx).Where("status IN ?", []string{"pending", "running"}).Order("id ASC").Find(&tasks).Error
	return tasks, err
}
