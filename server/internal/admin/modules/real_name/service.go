package realname

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/support"
	domainrealname "github.com/AeolianCloud/pveCloud/server/internal/domain/realname"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/cache"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
)

const (
	realNameObjectType   = "user_real_name"
	realNameSyncAction   = "real_name.sync"
	realNameReviewAction = "real_name.review"
	providerManual       = "manual"
)

type RealNameService struct {
	db           *gorm.DB
	auditService *AdminAuditService
	syncService  *domainrealname.RealNameService
}

func NewRealNameService(db *gorm.DB, redis *cache.Redis, auditService *AdminAuditService) *RealNameService {
	if auditService == nil {
		auditService = NewAdminAuditService(db)
	}
	return &RealNameService{
		db:           db,
		auditService: auditService,
		syncService:  domainrealname.NewRealNameService(db, redis),
	}
}

func (s *RealNameService) Applications(ctx context.Context, query admindto.RealNameApplicationListQuery) (admindto.PageResponse[admindto.RealNameApplicationItem], error) {
	page, perPage := support.NormalizePage(query.Page, query.PerPage)
	db := s.applicationDB(ctx)
	db, err := applyFilters(db, query)
	if err != nil {
		return admindto.PageResponse[admindto.RealNameApplicationItem]{}, err
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return admindto.PageResponse[admindto.RealNameApplicationItem]{}, err
	}
	var rows []applicationRow
	if err := db.Select(applicationSelect()).Order("applications.id DESC").Limit(perPage).Offset((page - 1) * perPage).Scan(&rows).Error; err != nil {
		return admindto.PageResponse[admindto.RealNameApplicationItem]{}, err
	}
	items := make([]admindto.RealNameApplicationItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.item())
	}
	return support.PageResponse(items, total, page, perPage), nil
}

func (s *RealNameService) Detail(ctx context.Context, id uint64) (admindto.RealNameApplicationItem, error) {
	var row applicationRow
	err := s.applicationDB(ctx).Select(applicationSelect()).Where("applications.id = ?", id).Scan(&row).Error
	if err != nil {
		return admindto.RealNameApplicationItem{}, err
	}
	if row.ID == 0 {
		return admindto.RealNameApplicationItem{}, apperrors.ErrNotFound.WithMessage("实名申请不存在")
	}
	return row.item(), nil
}

func (s *RealNameService) Sync(ctx context.Context, operatorID uint64, id uint64) (admindto.RealNameApplicationItem, error) {
	_, err := s.syncService.SyncApplicationByID(ctx, id, func(tx *gorm.DB, before models.UserRealNameApplication, after models.UserRealNameApplication) error {
		return s.auditService.Record(ctx, tx, AdminAuditWriteInput{
			AdminID:    &operatorID,
			Action:     realNameSyncAction,
			ObjectType: realNameObjectType,
			ObjectID:   textutil.Uint64String(id),
			BeforeData: auditSnapshot(before),
			AfterData:  auditSnapshot(after),
			Remark:     "同步实名供应商结果",
		})
	})
	if err != nil {
		_ = s.recordSyncFailureAudit(ctx, operatorID, id, err)
		return admindto.RealNameApplicationItem{}, err
	}
	return s.Detail(ctx, id)
}

func (s *RealNameService) Review(ctx context.Context, operatorID uint64, id uint64, req admindto.RealNameReviewRequest) (admindto.RealNameApplicationItem, error) {
	var updated models.UserRealNameApplication
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current models.UserRealNameApplication
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrNotFound.WithMessage("实名申请不存在")
			}
			return err
		}
		if current.VerificationProvider == nil || *current.VerificationProvider != providerManual {
			return apperrors.ErrConflict.WithMessage("当前实名申请不支持人工审核")
		}
		if current.Status != "pending" {
			return apperrors.ErrConflict.WithMessage("当前实名申请不是待审核状态")
		}
		before := current
		now := time.Now()
		updates := map[string]any{
			"status":                  req.Status,
			"provider_finished_at":    now,
			"provider_status":         req.Status,
			"provider_result_code":    nil,
			"provider_result_message": nil,
			"provider_trace_id":       nil,
		}
		hasReviewColumns, err := hasManualReviewColumns(ctx, tx)
		if err != nil {
			return err
		}
		if hasReviewColumns {
			updates["review_admin_id"] = operatorID
			updates["reviewed_at"] = now
		}
		if req.Status == "approved" {
			updates["reject_reason"] = nil
		} else {
			reason := strings.TrimSpace(req.Reason)
			if reason == "" {
				return apperrors.ErrValidation.WithMessage("拒绝原因不能为空")
			}
			updates["reject_reason"] = reason
		}
		if err := tx.Model(&current).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).First(&updated).Error; err != nil {
			return err
		}
		return s.auditService.Record(ctx, tx, AdminAuditWriteInput{
			AdminID:    &operatorID,
			Action:     realNameReviewAction,
			ObjectType: realNameObjectType,
			ObjectID:   textutil.Uint64String(id),
			BeforeData: auditSnapshot(before),
			AfterData:  auditSnapshot(updated),
			Remark:     "人工实名审核",
		})
	})
	if err != nil {
		return admindto.RealNameApplicationItem{}, err
	}
	return s.Detail(ctx, id)
}

func hasManualReviewColumns(ctx context.Context, tx *gorm.DB) (bool, error) {
	var count int64
	err := tx.WithContext(ctx).Raw(`
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'user_real_name_applications'
		  AND COLUMN_NAME IN ('review_admin_id', 'reviewed_at')
	`).Scan(&count).Error
	return count == 2, err
}

func (s *RealNameService) recordSyncFailureAudit(ctx context.Context, operatorID uint64, id uint64, syncErr error) error {
	app, detailErr := s.applicationByID(ctx, id)
	if detailErr != nil {
		return detailErr
	}
	message := "同步实名供应商结果失败"
	if syncErr != nil && strings.TrimSpace(syncErr.Error()) != "" {
		message = message + "：" + strings.TrimSpace(syncErr.Error())
	}
	return s.auditService.Record(ctx, nil, AdminAuditWriteInput{
		AdminID:    &operatorID,
		Action:     realNameSyncAction,
		ObjectType: realNameObjectType,
		ObjectID:   textutil.Uint64String(id),
		BeforeData: auditSnapshot(app),
		AfterData:  auditSnapshot(app),
		Remark:     message,
	})
}

func (s *RealNameService) applicationByID(ctx context.Context, id uint64) (models.UserRealNameApplication, error) {
	var app models.UserRealNameApplication
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.UserRealNameApplication{}, apperrors.ErrNotFound.WithMessage("实名申请不存在")
	}
	return app, err
}

func applyFilters(db *gorm.DB, query admindto.RealNameApplicationListQuery) (*gorm.DB, error) {
	if query.Keyword != "" {
		keyword := "%" + strings.TrimSpace(query.Keyword) + "%"
		db = db.Where("users.username LIKE ? OR users.email LIKE ? OR applications.real_name LIKE ? OR applications.application_no LIKE ?", keyword, keyword, keyword, keyword)
	}
	if query.Status != "" {
		db = db.Where("applications.status = ?", strings.TrimSpace(query.Status))
	}
	if query.IDType != "" {
		db = db.Where("applications.id_type = ?", strings.TrimSpace(query.IDType))
	}
	if query.Provider != "" {
		db = db.Where("applications.verification_provider = ?", strings.TrimSpace(query.Provider))
	}
	if query.ProviderStatus != "" {
		db = db.Where("applications.provider_status = ?", strings.TrimSpace(query.ProviderStatus))
	}
	if query.DateFrom != "" {
		from, err := time.ParseInLocation("2006-01-02", query.DateFrom, time.Local)
		if err != nil {
			return nil, apperrors.ErrValidation.WithMessage("开始时间格式错误")
		}
		db = db.Where("applications.created_at >= ?", from)
	}
	if query.DateTo != "" {
		to, err := time.ParseInLocation("2006-01-02", query.DateTo, time.Local)
		if err != nil {
			return nil, apperrors.ErrValidation.WithMessage("结束时间格式错误")
		}
		db = db.Where("applications.created_at < ?", to.Add(24*time.Hour))
	}
	return db, nil
}

func (s *RealNameService) applicationDB(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx).
		Table("user_real_name_applications AS applications").
		Joins("JOIN users ON users.id = applications.user_id")
}

func auditSnapshot(app models.UserRealNameApplication) map[string]any {
	return map[string]any{
		"id":                    app.ID,
		"application_no":        app.ApplicationNo,
		"user_id":               app.UserID,
		"status":                app.Status,
		"id_number_masked":      app.IDNumberMasked,
		"verification_provider": app.VerificationProvider,
		"provider_status":       app.ProviderStatus,
	}
}

func applicationSelect() string {
	return `applications.id, applications.application_no, applications.real_name, applications.id_type, applications.id_number_masked,
		applications.verification_provider, applications.provider_application_id, applications.provider_status, applications.provider_result_code,
		applications.provider_result_message, applications.provider_trace_id, applications.status, applications.submit_attempt,
		applications.reject_reason, applications.provider_started_at, applications.provider_finished_at, applications.created_at, applications.updated_at,
		users.id AS user_id, users.username, users.email, users.display_name, users.status AS user_status`
}

type applicationRow struct {
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

func (r applicationRow) item() admindto.RealNameApplicationItem {
	return admindto.RealNameApplicationItem{
		ID:                    r.ID,
		ApplicationNo:         r.ApplicationNo,
		User:                  admindto.RealNameUserSummary{ID: r.UserID, Username: r.Username, Email: r.Email, DisplayName: r.DisplayName, Status: r.UserStatus},
		RealName:              r.RealName,
		IDType:                r.IDType,
		IDNumberMasked:        r.IDNumberMasked,
		VerificationProvider:  r.VerificationProvider,
		ProviderApplicationID: r.ProviderApplicationID,
		ProviderStatus:        r.ProviderStatus,
		ProviderResultCode:    r.ProviderResultCode,
		ProviderResultMessage: r.ProviderResultMessage,
		ProviderTraceID:       r.ProviderTraceID,
		Status:                r.Status,
		SubmitAttempt:         r.SubmitAttempt,
		FailureReason:         r.FailureReason,
		ProviderStartedAt:     r.ProviderStartedAt,
		ProviderFinishedAt:    r.ProviderFinishedAt,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
	}
}

func stringPtr(value string) *string {
	return &value
}
