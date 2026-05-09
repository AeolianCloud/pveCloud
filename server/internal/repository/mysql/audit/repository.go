package audit

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

type LogFilters struct {
	AdminID    uint64
	Action     string
	ObjectType string
	ObjectID   string
	DateFrom   *time.Time
	DateTo     *time.Time
	DateToOpen bool
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, db *gorm.DB, log AdminAuditLog) error {
	return r.queryDB(db).WithContext(ctx).Create(&log).Error
}

func (r *Repository) Logs(ctx context.Context, filters LogFilters, limit int, offset int) ([]LogRow, int64, error) {
	query := r.applyFilters(r.db.WithContext(ctx).Table("admin_audit_logs"), filters)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []LogRow
	if err := query.
		Select(`admin_audit_logs.*,
			COALESCE(admin_audit_logs.admin_username, admin_users.username) AS actor_username,
			COALESCE(admin_audit_logs.admin_display_name, admin_users.display_name) AS actor_display_name,
			admin_users.email AS admin_email`).
		Joins("LEFT JOIN admin_users ON admin_users.id = admin_audit_logs.admin_id").
		Order("admin_audit_logs.id DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *Repository) applyFilters(db *gorm.DB, filters LogFilters) *gorm.DB {
	if filters.AdminID > 0 {
		db = db.Where("admin_id = ?", filters.AdminID)
	}
	if filters.Action != "" {
		db = db.Where("action = ?", filters.Action)
	}
	if filters.ObjectType != "" {
		db = db.Where("object_type = ?", filters.ObjectType)
	}
	if filters.ObjectID != "" {
		db = db.Where("object_id = ?", filters.ObjectID)
	}
	if filters.DateFrom != nil {
		db = db.Where("created_at >= ?", *filters.DateFrom)
	}
	if filters.DateTo != nil {
		if filters.DateToOpen {
			db = db.Where("created_at < ?", *filters.DateTo)
		} else {
			db = db.Where("created_at <= ?", *filters.DateTo)
		}
	}
	return db
}

func (r *Repository) queryDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return r.db
}
