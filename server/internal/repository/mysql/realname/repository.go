package realname

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

type ApplicationListFilters struct {
	Keyword        string
	Status         string
	IDType         string
	Provider       string
	ProviderStatus string
	DateFrom       *time.Time
	DateToOpen     *time.Time
}

type ApplicationListRow struct {
	ID                    uint64
	ApplicationNo         string
	RealName              string
	IDType                string
	IDNumberMasked        string
	VerificationProvider  *string
	ProviderApplicationID *string
	ProviderStatus        *string
	ProviderResultCode    *string
	ProviderResultMessage *string
	ProviderTraceID       *string
	Status                string
	SubmitAttempt         uint
	FailureReason         *string `gorm:"column:reject_reason"`
	ProviderStartedAt     *time.Time
	ProviderFinishedAt    *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
	UserID                uint64
	Username              string
	Email                 string
	DisplayName           *string
	UserStatus            string
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Applications(ctx context.Context, filters ApplicationListFilters, limit int, offset int) ([]ApplicationListRow, int64, error) {
	query := r.applyApplicationListFilters(r.applicationDB(ctx), filters)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []ApplicationListRow
	if err := query.Select(applicationSelect()).
		Order("applications.id DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repository) ApplicationDetailRow(ctx context.Context, id uint64) (ApplicationListRow, error) {
	var row ApplicationListRow
	err := r.applicationDB(ctx).
		Select(applicationSelect()).
		Where("applications.id = ?", id).
		Scan(&row).Error
	return row, err
}

func (r *Repository) FindApplicationByID(ctx context.Context, db *gorm.DB, id uint64) (UserRealNameApplication, error) {
	var app UserRealNameApplication
	err := r.queryDB(db).WithContext(ctx).Where("id = ?", id).First(&app).Error
	return app, err
}

func (r *Repository) FindApplicationByIDForUpdate(ctx context.Context, db *gorm.DB, id uint64) (UserRealNameApplication, error) {
	var app UserRealNameApplication
	err := r.queryDB(db).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&app).Error
	return app, err
}

func (r *Repository) LatestApplication(ctx context.Context, db *gorm.DB, userID uint64) (UserRealNameApplication, error) {
	var app UserRealNameApplication
	err := r.queryDB(db).WithContext(ctx).
		Where("user_id = ?", userID).
		Order("id DESC").
		First(&app).Error
	return app, err
}

func (r *Repository) LatestApplicationForUpdate(ctx context.Context, db *gorm.DB, userID uint64) (UserRealNameApplication, error) {
	var app UserRealNameApplication
	err := r.queryDB(db).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		Order("id DESC").
		First(&app).Error
	return app, err
}

func (r *Repository) CreateApplication(ctx context.Context, db *gorm.DB, app *UserRealNameApplication) error {
	return r.queryDB(db).WithContext(ctx).Create(app).Error
}

func (r *Repository) UpdateApplication(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).
		Model(&UserRealNameApplication{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *Repository) UpdateApplicationAndReload(ctx context.Context, db *gorm.DB, app *UserRealNameApplication, updates map[string]any) error {
	if err := r.queryDB(db).WithContext(ctx).Model(app).Updates(updates).Error; err != nil {
		return err
	}
	return r.queryDB(db).WithContext(ctx).Where("id = ?", app.ID).First(app).Error
}

func (r *Repository) CountApprovedApplicationsByDigests(ctx context.Context, db *gorm.DB, userID uint64, digests []string, approvedStatus string) (int64, error) {
	var count int64
	if len(digests) == 0 {
		return 0, nil
	}
	if err := r.queryDB(db).WithContext(ctx).
		Model(&UserRealNameApplication{}).
		Where("id_number_digest IN ? AND status = ? AND user_id <> ?", digests, approvedStatus, userID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) UserApplicationForSync(ctx context.Context, userID uint64, applicationNo string, pendingStatus string) (UserRealNameApplication, error) {
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if strings.TrimSpace(applicationNo) != "" {
		query = query.Where("application_no = ?", strings.TrimSpace(applicationNo))
	} else {
		query = query.Where("status = ?", pendingStatus).Order("id DESC")
	}
	var app UserRealNameApplication
	err := query.First(&app).Error
	return app, err
}

func (r *Repository) ApplicationByProviderSession(ctx context.Context, provider string, providerApplicationID string) (UserRealNameApplication, error) {
	var app UserRealNameApplication
	err := r.db.WithContext(ctx).
		Where("verification_provider = ? AND provider_application_id = ?", provider, providerApplicationID).
		First(&app).Error
	return app, err
}

func (r *Repository) HasManualReviewColumns(ctx context.Context, db *gorm.DB) (bool, error) {
	var count int64
	err := r.queryDB(db).WithContext(ctx).Raw(`
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'user_real_name_applications'
		  AND COLUMN_NAME IN ('review_admin_id', 'reviewed_at')
	`).Scan(&count).Error
	return count == 2, err
}

func (r *Repository) applicationDB(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("user_real_name_applications AS applications").
		Joins("JOIN users ON users.id = applications.user_id")
}

func (r *Repository) applyApplicationListFilters(db *gorm.DB, filters ApplicationListFilters) *gorm.DB {
	if strings.TrimSpace(filters.Keyword) != "" {
		keyword := "%" + strings.TrimSpace(filters.Keyword) + "%"
		db = db.Where("users.username LIKE ? OR users.email LIKE ? OR applications.real_name LIKE ? OR applications.application_no LIKE ?", keyword, keyword, keyword, keyword)
	}
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("applications.status = ?", strings.TrimSpace(filters.Status))
	}
	if strings.TrimSpace(filters.IDType) != "" {
		db = db.Where("applications.id_type = ?", strings.TrimSpace(filters.IDType))
	}
	if strings.TrimSpace(filters.Provider) != "" {
		db = db.Where("applications.verification_provider = ?", strings.TrimSpace(filters.Provider))
	}
	if strings.TrimSpace(filters.ProviderStatus) != "" {
		db = db.Where("applications.provider_status = ?", strings.TrimSpace(filters.ProviderStatus))
	}
	if filters.DateFrom != nil {
		db = db.Where("applications.created_at >= ?", *filters.DateFrom)
	}
	if filters.DateToOpen != nil {
		db = db.Where("applications.created_at < ?", *filters.DateToOpen)
	}
	return db
}

func (r *Repository) queryDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return r.db
}

func applicationSelect() string {
	return `applications.id, applications.application_no, applications.real_name, applications.id_type, applications.id_number_masked,
		applications.verification_provider, applications.provider_application_id, applications.provider_status, applications.provider_result_code,
		applications.provider_result_message, applications.provider_trace_id, applications.status, applications.submit_attempt,
		applications.reject_reason, applications.provider_started_at, applications.provider_finished_at, applications.created_at, applications.updated_at,
		users.id AS user_id, users.username, users.email, users.display_name, users.status AS user_status`
}
