package realname

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/AeolianCloud/pveCloud/server/internal/platform/cache"
	mysqlrealname "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/realname"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	realnamesync "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/realname/syncsvc"
	"github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/support"
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
	syncService  *realnamesync.RealNameService
	applications *mysqlrealname.Repository
}

func NewRealNameService(db *gorm.DB, redis *cache.Redis, auditService *AdminAuditService) *RealNameService {
	if auditService == nil {
		auditService = NewAdminAuditService(db)
	}
	return &RealNameService{
		db:           db,
		auditService: auditService,
		syncService:  realnamesync.NewRealNameService(db, redis),
		applications: mysqlrealname.NewRepository(db),
	}
}

func (s *RealNameService) Applications(ctx context.Context, query admindto.RealNameApplicationListQuery) (admindto.PageResponse[admindto.RealNameApplicationItem], error) {
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	filters, err := buildApplicationFilters(query)
	if err != nil {
		return admindto.PageResponse[admindto.RealNameApplicationItem]{}, err
	}
	rows, total, err := s.applications.Applications(ctx, filters, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.RealNameApplicationItem]{}, err
	}
	items := make([]admindto.RealNameApplicationItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, applicationItem(row))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func (s *RealNameService) Detail(ctx context.Context, id uint64) (admindto.RealNameApplicationItem, error) {
	row, err := s.applications.ApplicationDetailRow(ctx, id)
	if err != nil {
		return admindto.RealNameApplicationItem{}, err
	}
	if row.ID == 0 {
		return admindto.RealNameApplicationItem{}, apperrors.ErrNotFound.WithMessage("实名申请不存在")
	}
	return applicationItem(row), nil
}

func (s *RealNameService) Sync(ctx context.Context, operatorID uint64, id uint64) (admindto.RealNameApplicationItem, error) {
	_, err := s.syncService.SyncApplicationByID(ctx, id, func(tx *gorm.DB, before mysqlrealname.UserRealNameApplication, after mysqlrealname.UserRealNameApplication) error {
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
	var updated mysqlrealname.UserRealNameApplication
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.applications.FindApplicationByIDForUpdate(ctx, tx, id)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("实名申请不存在")
		}
		if err != nil {
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
		if err := s.applications.UpdateApplication(ctx, tx, current.ID, updates); err != nil {
			return err
		}
		updated, err = s.applications.FindApplicationByID(ctx, tx, id)
		if err != nil {
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
	return mysqlrealname.NewRepository(tx).HasManualReviewColumns(ctx, tx)
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

func (s *RealNameService) applicationByID(ctx context.Context, id uint64) (mysqlrealname.UserRealNameApplication, error) {
	app, err := s.applications.FindApplicationByID(ctx, nil, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return mysqlrealname.UserRealNameApplication{}, apperrors.ErrNotFound.WithMessage("实名申请不存在")
	}
	return app, err
}

func buildApplicationFilters(query admindto.RealNameApplicationListQuery) (mysqlrealname.ApplicationListFilters, error) {
	filters := mysqlrealname.ApplicationListFilters{
		Keyword:        query.Keyword,
		Status:         query.Status,
		IDType:         query.IDType,
		Provider:       query.Provider,
		ProviderStatus: query.ProviderStatus,
	}
	if query.Keyword != "" {
		filters.Keyword = strings.TrimSpace(query.Keyword)
	}
	if query.DateFrom != "" {
		from, err := time.ParseInLocation("2006-01-02", query.DateFrom, time.Local)
		if err != nil {
			return filters, apperrors.ErrValidation.WithMessage("开始时间格式错误")
		}
		filters.DateFrom = &from
	}
	if query.DateTo != "" {
		to, err := time.ParseInLocation("2006-01-02", query.DateTo, time.Local)
		if err != nil {
			return filters, apperrors.ErrValidation.WithMessage("结束时间格式错误")
		}
		dateToOpen := to.Add(24 * time.Hour)
		filters.DateToOpen = &dateToOpen
	}
	return filters, nil
}

func auditSnapshot(app mysqlrealname.UserRealNameApplication) map[string]any {
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

func applicationItem(r mysqlrealname.ApplicationListRow) admindto.RealNameApplicationItem {
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
