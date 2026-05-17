package instance

import (
	"context"
	"strings"

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
