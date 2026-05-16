package logs

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

type UserSecurityFilters struct {
	UserID     uint64
	Username   string
	Action     string
	Result     string
	RequestID  string
	IP         string
	DateFrom   *time.Time
	DateTo     *time.Time
}

type UserBusinessFilters struct {
	UserID     uint64
	Module     string
	Action     string
	ObjectType string
	ObjectID   string
	RequestID  string
	DateFrom   *time.Time
	DateTo     *time.Time
}

type FrontendErrorFilters struct {
	SourceApp    string
	PagePath     string
	ErrorType    string
	APIPath      string
	HTTPStatus   int
	RequestID    string
	DateFrom     *time.Time
	DateTo       *time.Time
}

type BackendRuntimeFilters struct {
	Level       string
	Category    string
	Status      int
	RequestID   string
	RequestPath string
	DateFrom    *time.Time
	DateTo      *time.Time
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUserSecurityLog(ctx context.Context, db *gorm.DB, log *UserSecurityLog) error {
	return r.queryDB(db).WithContext(ctx).Create(log).Error
}

func (r *Repository) CreateUserBusinessLog(ctx context.Context, db *gorm.DB, log *UserBusinessLog) error {
	return r.queryDB(db).WithContext(ctx).Create(log).Error
}

func (r *Repository) CreateFrontendErrorLog(ctx context.Context, db *gorm.DB, log *FrontendErrorLog) error {
	return r.queryDB(db).WithContext(ctx).Create(log).Error
}

func (r *Repository) CreateBackendRuntimeLog(ctx context.Context, db *gorm.DB, log *BackendRuntimeLog) error {
	return r.queryDB(db).WithContext(ctx).Create(log).Error
}

func (r *Repository) CreateExportRecord(ctx context.Context, db *gorm.DB, record *ExportRecord) error {
	return r.queryDB(db).WithContext(ctx).Create(record).Error
}

func (r *Repository) UserSecurityLogs(ctx context.Context, filters UserSecurityFilters, limit, offset int) ([]UserSecurityLogRow, int64, error) {
	query := r.applyUserSecurityFilters(r.db.WithContext(ctx).Table("user_security_logs"), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []UserSecurityLogRow
	if err := query.Select("user_security_logs.*, users.username AS user_username, users.email AS user_email, users.display_name AS user_display_name").
		Joins("LEFT JOIN users ON users.id = user_security_logs.user_id").
		Order("user_security_logs.id DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repository) UserBusinessLogs(ctx context.Context, filters UserBusinessFilters, limit, offset int) ([]UserBusinessLogRow, int64, error) {
	query := r.applyUserBusinessFilters(r.db.WithContext(ctx).Table("user_business_logs"), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []UserBusinessLogRow
	if err := query.Select("user_business_logs.*, users.username AS user_username, users.email AS user_email, users.display_name AS user_display_name").
		Joins("JOIN users ON users.id = user_business_logs.user_id").
		Order("user_business_logs.id DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repository) FrontendErrorLogs(ctx context.Context, filters FrontendErrorFilters, limit, offset int) ([]FrontendErrorLog, int64, error) {
	query := r.applyFrontendFilters(r.db.WithContext(ctx).Model(&FrontendErrorLog{}), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []FrontendErrorLog
	if err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repository) BackendRuntimeLogs(ctx context.Context, filters BackendRuntimeFilters, limit, offset int) ([]BackendRuntimeLog, int64, error) {
	query := r.applyBackendFilters(r.db.WithContext(ctx).Model(&BackendRuntimeLog{}), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []BackendRuntimeLog
	if err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repository) applyUserSecurityFilters(db *gorm.DB, filters UserSecurityFilters) *gorm.DB {
	if filters.UserID > 0 {
		db = db.Where("user_id = ?", filters.UserID)
	}
	if strings.TrimSpace(filters.Username) != "" {
		like := "%" + strings.TrimSpace(filters.Username) + "%"
		db = db.Where("username LIKE ? OR email LIKE ?", like, like)
	}
	if strings.TrimSpace(filters.Action) != "" {
		db = db.Where("action = ?", strings.TrimSpace(filters.Action))
	}
	if strings.TrimSpace(filters.Result) != "" {
		db = db.Where("result = ?", strings.TrimSpace(filters.Result))
	}
	if strings.TrimSpace(filters.RequestID) != "" {
		db = db.Where("request_id = ?", strings.TrimSpace(filters.RequestID))
	}
	if strings.TrimSpace(filters.IP) != "" {
		db = db.Where("ip = ?", strings.TrimSpace(filters.IP))
	}
	if filters.DateFrom != nil {
		db = db.Where("created_at >= ?", *filters.DateFrom)
	}
	if filters.DateTo != nil {
		db = db.Where("created_at <= ?", *filters.DateTo)
	}
	return db
}

func (r *Repository) applyUserBusinessFilters(db *gorm.DB, filters UserBusinessFilters) *gorm.DB {
	if filters.UserID > 0 {
		db = db.Where("user_id = ?", filters.UserID)
	}
	if strings.TrimSpace(filters.Module) != "" {
		db = db.Where("module = ?", strings.TrimSpace(filters.Module))
	}
	if strings.TrimSpace(filters.Action) != "" {
		db = db.Where("action = ?", strings.TrimSpace(filters.Action))
	}
	if strings.TrimSpace(filters.ObjectType) != "" {
		db = db.Where("object_type = ?", strings.TrimSpace(filters.ObjectType))
	}
	if strings.TrimSpace(filters.ObjectID) != "" {
		db = db.Where("object_id = ?", strings.TrimSpace(filters.ObjectID))
	}
	if strings.TrimSpace(filters.RequestID) != "" {
		db = db.Where("request_id = ?", strings.TrimSpace(filters.RequestID))
	}
	if filters.DateFrom != nil {
		db = db.Where("created_at >= ?", *filters.DateFrom)
	}
	if filters.DateTo != nil {
		db = db.Where("created_at <= ?", *filters.DateTo)
	}
	return db
}

func (r *Repository) applyFrontendFilters(db *gorm.DB, filters FrontendErrorFilters) *gorm.DB {
	if strings.TrimSpace(filters.SourceApp) != "" {
		db = db.Where("source_app = ?", strings.TrimSpace(filters.SourceApp))
	}
	if strings.TrimSpace(filters.PagePath) != "" {
		db = db.Where("page_path LIKE ?", "%"+strings.TrimSpace(filters.PagePath)+"%")
	}
	if strings.TrimSpace(filters.ErrorType) != "" {
		db = db.Where("error_type = ?", strings.TrimSpace(filters.ErrorType))
	}
	if strings.TrimSpace(filters.APIPath) != "" {
		db = db.Where("api_path LIKE ?", "%"+strings.TrimSpace(filters.APIPath)+"%")
	}
	if filters.HTTPStatus > 0 {
		db = db.Where("http_status = ?", filters.HTTPStatus)
	}
	if strings.TrimSpace(filters.RequestID) != "" {
		db = db.Where("request_id = ?", strings.TrimSpace(filters.RequestID))
	}
	if filters.DateFrom != nil {
		db = db.Where("created_at >= ?", *filters.DateFrom)
	}
	if filters.DateTo != nil {
		db = db.Where("created_at <= ?", *filters.DateTo)
	}
	return db
}

func (r *Repository) applyBackendFilters(db *gorm.DB, filters BackendRuntimeFilters) *gorm.DB {
	if strings.TrimSpace(filters.Level) != "" {
		db = db.Where("level = ?", strings.TrimSpace(filters.Level))
	}
	if strings.TrimSpace(filters.Category) != "" {
		db = db.Where("category = ?", strings.TrimSpace(filters.Category))
	}
	if filters.Status > 0 {
		db = db.Where("status = ?", filters.Status)
	}
	if strings.TrimSpace(filters.RequestID) != "" {
		db = db.Where("request_id = ?", strings.TrimSpace(filters.RequestID))
	}
	if strings.TrimSpace(filters.RequestPath) != "" {
		db = db.Where("request_path LIKE ?", "%"+strings.TrimSpace(filters.RequestPath)+"%")
	}
	if filters.DateFrom != nil {
		db = db.Where("created_at >= ?", *filters.DateFrom)
	}
	if filters.DateTo != nil {
		db = db.Where("created_at <= ?", *filters.DateTo)
	}
	return db
}

func (r *Repository) queryDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return r.db
}
