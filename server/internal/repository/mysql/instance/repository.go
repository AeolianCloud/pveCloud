package instance

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct{ db *gorm.DB }

type InstanceFilters struct {
	UserID      uint64
	Status      string
	InstanceNo  string
	OrderNo     string
	UserKeyword string
	DateFrom    string
	DateTo      string
}

type MappingFilters struct {
	Status        string
	PlanNo        string
	RegionNo      string
	TemplateNo    string
	NetworkTypeNo string
}

type TaskFilters struct {
	TaskType   string
	Status     string
	ObjectType string
	ObjectNo   string
	DateFrom   string
	DateTo     string
}

type NotificationFilters struct {
	UserID   uint64
	Scene    string
	Status   string
	Target   string
	DateFrom string
	DateTo   string
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) CreateMapping(ctx context.Context, db *gorm.DB, mapping *ProvisionMapping) error {
	return r.queryDB(db).WithContext(ctx).Create(mapping).Error
}

func (r *Repository) UpdateMapping(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Model(&ProvisionMapping{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) MappingByID(ctx context.Context, id uint64) (ProvisionMapping, error) {
	var mapping ProvisionMapping
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&mapping).Error
	return mapping, err
}

func (r *Repository) ListMappings(ctx context.Context, filters MappingFilters, limit, offset int) ([]ProvisionMapping, int64, error) {
	query := r.applyMappingFilters(r.db.WithContext(ctx).Model(&ProvisionMapping{}), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []ProvisionMapping
	if err := query.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repository) MappingForProvision(ctx context.Context, db *gorm.DB, planNo, regionNo, templateNo, networkTypeNo string) (ProvisionMapping, error) {
	var mapping ProvisionMapping
	err := r.queryDB(db).WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("plan_no = ? AND region_no = ? AND template_no = ? AND status = ?", planNo, regionNo, templateNo, "active").
		Where("network_type_no = ? OR network_type_no = ''", strings.TrimSpace(networkTypeNo)).
		Order("network_type_no DESC, id ASC").
		First(&mapping).Error
	return mapping, err
}

func (r *Repository) AdvanceMappingVMID(ctx context.Context, db *gorm.DB, id uint64, nextVMID uint) error {
	return r.queryDB(db).WithContext(ctx).Model(&ProvisionMapping{}).Where("id = ?", id).Update("next_vmid", nextVMID).Error
}

func (r *Repository) CreateInstance(ctx context.Context, db *gorm.DB, instance *Instance) error {
	return r.queryDB(db).WithContext(ctx).Create(instance).Error
}

func (r *Repository) UpdateInstance(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Model(&Instance{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) InstanceForUpdate(ctx context.Context, db *gorm.DB, instanceNo string) (Instance, error) {
	var row Instance
	err := r.queryDB(db).WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("instance_no = ?", instanceNo).First(&row).Error
	return row, err
}

func (r *Repository) InstanceByOrderID(ctx context.Context, orderID uint64) (Instance, error) {
	var row Instance
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&row).Error
	return row, err
}

func (r *Repository) UserInstance(ctx context.Context, userID uint64, instanceNo string) (Instance, error) {
	var row Instance
	err := r.db.WithContext(ctx).Where("user_id = ? AND instance_no = ?", userID, instanceNo).First(&row).Error
	return row, err
}

func (r *Repository) ListInstances(ctx context.Context, filters InstanceFilters, limit, offset int) ([]InstanceRow, int64, error) {
	query := r.applyInstanceFilters(r.db.WithContext(ctx).Table("instances").Joins("JOIN users ON users.id = instances.user_id"), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []InstanceRow
	if err := query.Select("instances.*, users.username, users.email, users.display_name").Order("instances.created_at DESC, instances.id DESC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repository) Detail(ctx context.Context, instanceNo string) (InstanceRow, error) {
	var row InstanceRow
	err := r.db.WithContext(ctx).Table("instances").Select("instances.*, users.username, users.email, users.display_name").Joins("JOIN users ON users.id = instances.user_id").Where("instances.instance_no = ?", instanceNo).Take(&row).Error
	return row, err
}

func (r *Repository) CreateOperation(ctx context.Context, db *gorm.DB, op *Operation) error {
	return r.queryDB(db).WithContext(ctx).Create(op).Error
}

func (r *Repository) UpdateOperation(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Model(&Operation{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) LatestOperation(ctx context.Context, instanceID uint64) (Operation, error) {
	var op Operation
	err := r.db.WithContext(ctx).Where("instance_id = ?", instanceID).Order("created_at DESC, id DESC").First(&op).Error
	return op, err
}

func (r *Repository) LatestOperationExcluding(ctx context.Context, instanceID uint64, excludedAction string) (Operation, error) {
	var op Operation
	err := r.db.WithContext(ctx).Where("instance_id = ? AND action <> ?", instanceID, excludedAction).Order("created_at DESC, id DESC").First(&op).Error
	return op, err
}

func (r *Repository) Operations(ctx context.Context, instanceID uint64, limit int) ([]Operation, error) {
	var rows []Operation
	err := r.db.WithContext(ctx).Where("instance_id = ?", instanceID).Order("created_at DESC, id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *Repository) CreateTask(ctx context.Context, db *gorm.DB, task *Task) error {
	return r.queryDB(db).WithContext(ctx).Create(task).Error
}

func (r *Repository) CreateTaskIgnoreDuplicate(ctx context.Context, db *gorm.DB, task *Task) error {
	return r.queryDB(db).WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(task).Error
}

func (r *Repository) UpdateTask(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Model(&Task{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) TaskByNo(ctx context.Context, taskNo string) (Task, error) {
	var task Task
	err := r.db.WithContext(ctx).Where("task_no = ?", taskNo).First(&task).Error
	return task, err
}

func (r *Repository) TaskForUpdate(ctx context.Context, db *gorm.DB, taskNo string) (Task, error) {
	var task Task
	err := r.queryDB(db).WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("task_no = ?", taskNo).First(&task).Error
	return task, err
}

func (r *Repository) ListTasks(ctx context.Context, filters TaskFilters, limit, offset int) ([]Task, int64, error) {
	query := r.applyTaskFilters(r.db.WithContext(ctx).Model(&Task{}), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []Task
	if err := query.Order("scheduled_at DESC, id DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repository) ClaimTasks(ctx context.Context, db *gorm.DB, workerID string, limit int, lockUntil time.Time) ([]Task, error) {
	if limit <= 0 {
		return nil, nil
	}
	now := time.Now()
	var rows []Task
	query := r.queryDB(db).WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("scheduled_at <= ?", now).
		Where("(status = ? OR (status = ? AND locked_until IS NOT NULL AND locked_until < ?))", "pending", "running", now).
		Order("scheduled_at ASC, id ASC").
		Limit(limit)
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	for i := range rows {
		if err := r.UpdateTask(ctx, db, rows[i].ID, map[string]any{"status": "running", "locked_by": workerID, "locked_until": lockUntil, "attempts": rows[i].Attempts + 1}); err != nil {
			return nil, err
		}
		rows[i].Status = "running"
		rows[i].LockedBy = &workerID
		rows[i].LockedUntil = &lockUntil
		rows[i].Attempts++
	}
	return rows, nil
}

func (r *Repository) CreateNotification(ctx context.Context, db *gorm.DB, notification *Notification) error {
	return r.queryDB(db).WithContext(ctx).Create(notification).Error
}

func (r *Repository) CreateNotificationIgnoreDuplicate(ctx context.Context, db *gorm.DB, notification *Notification) error {
	return r.queryDB(db).WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(notification).Error
}

func (r *Repository) UpdateNotification(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Model(&Notification{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) NotificationByNo(ctx context.Context, notificationNo string) (Notification, error) {
	var notification Notification
	err := r.db.WithContext(ctx).Where("notification_no = ?", notificationNo).First(&notification).Error
	return notification, err
}

func (r *Repository) queryDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return r.db
}

func (r *Repository) applyMappingFilters(db *gorm.DB, filters MappingFilters) *gorm.DB {
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("status = ?", strings.TrimSpace(filters.Status))
	}
	if strings.TrimSpace(filters.PlanNo) != "" {
		db = db.Where("plan_no = ?", strings.TrimSpace(filters.PlanNo))
	}
	if strings.TrimSpace(filters.RegionNo) != "" {
		db = db.Where("region_no = ?", strings.TrimSpace(filters.RegionNo))
	}
	if strings.TrimSpace(filters.TemplateNo) != "" {
		db = db.Where("template_no = ?", strings.TrimSpace(filters.TemplateNo))
	}
	if strings.TrimSpace(filters.NetworkTypeNo) != "" {
		db = db.Where("network_type_no = ?", strings.TrimSpace(filters.NetworkTypeNo))
	}
	return db
}

func (r *Repository) applyInstanceFilters(db *gorm.DB, filters InstanceFilters) *gorm.DB {
	if filters.UserID > 0 {
		db = db.Where("instances.user_id = ?", filters.UserID)
	}
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("instances.status = ?", strings.TrimSpace(filters.Status))
	}
	if strings.TrimSpace(filters.InstanceNo) != "" {
		db = db.Where("instances.instance_no = ?", strings.TrimSpace(filters.InstanceNo))
	}
	if strings.TrimSpace(filters.OrderNo) != "" {
		db = db.Where("instances.order_no = ?", strings.TrimSpace(filters.OrderNo))
	}
	if keyword := strings.TrimSpace(filters.UserKeyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("users.username LIKE ? OR users.email LIKE ? OR users.display_name LIKE ?", like, like, like)
	}
	if strings.TrimSpace(filters.DateFrom) != "" {
		db = db.Where("instances.created_at >= ?", strings.TrimSpace(filters.DateFrom))
	}
	if strings.TrimSpace(filters.DateTo) != "" {
		db = db.Where("instances.created_at <= ?", strings.TrimSpace(filters.DateTo))
	}
	return db
}

func (r *Repository) applyTaskFilters(db *gorm.DB, filters TaskFilters) *gorm.DB {
	if strings.TrimSpace(filters.TaskType) != "" {
		db = db.Where("task_type = ?", strings.TrimSpace(filters.TaskType))
	}
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("status = ?", strings.TrimSpace(filters.Status))
	}
	if strings.TrimSpace(filters.ObjectType) != "" {
		db = db.Where("object_type = ?", strings.TrimSpace(filters.ObjectType))
	}
	if strings.TrimSpace(filters.ObjectNo) != "" {
		db = db.Where("object_no = ?", strings.TrimSpace(filters.ObjectNo))
	}
	if strings.TrimSpace(filters.DateFrom) != "" {
		db = db.Where("created_at >= ?", strings.TrimSpace(filters.DateFrom))
	}
	if strings.TrimSpace(filters.DateTo) != "" {
		db = db.Where("created_at <= ?", strings.TrimSpace(filters.DateTo))
	}
	return db
}
