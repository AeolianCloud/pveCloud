package realname

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/cache"
	realnameintegration "github.com/AeolianCloud/pveCloud/server/internal/platform/integrations/realname"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/web/dto"
)

const (
	statusUnverified = "unverified"
	statusPending    = "pending"
	statusApproved   = "approved"
	statusRejected   = "rejected"

	providerAlipay = "alipay"
	providerWechat = "wechat"
	providerManual = "manual"

	digestVersionHMACSHA256 = "hmac-sha256-v1"
	digestVersionLegacy     = "sha256-legacy"

	providerStatusCreated  = "created"
	providerStatusApproved = "approved"
	providerStatusRejected = "rejected"
	providerStatusPending  = "pending"

	providerStatusCreating = "creating"

	legacyDigestCacheTTL = 7 * 24 * time.Hour
	callbackReplayTTL    = 15 * time.Minute
	callbackAcceptWindow = 10 * time.Minute
)

var idCardPattern = regexp.MustCompile(`^[0-9]{17}[0-9Xx]$`)

type RealNameService struct {
	db             *gorm.DB
	redis          *cache.Redis
	providerClient *realnameintegration.Client
}

type SyncApplicationHook func(tx *gorm.DB, before models.UserRealNameApplication, after models.UserRealNameApplication) error

func NewRealNameService(db *gorm.DB, redis *cache.Redis) *RealNameService {
	return &RealNameService{
		db:             db,
		redis:          redis,
		providerClient: realnameintegration.NewClient(&http.Client{Timeout: 10 * time.Second}),
	}
}

func (s *RealNameService) Status(ctx context.Context, userID uint64) (webdto.RealNameStatusResponse, error) {
	config, _, err := s.config(ctx, s.db)
	if err != nil {
		return webdto.RealNameStatusResponse{}, err
	}
	latest, ok, err := s.latest(ctx, s.db, userID)
	if err != nil {
		return webdto.RealNameStatusResponse{}, err
	}
	if !ok {
		return webdto.RealNameStatusResponse{Status: statusUnverified, Config: config.Public()}, nil
	}
	summary := applicationSummary(latest)
	return webdto.RealNameStatusResponse{Status: latest.Status, Application: &summary, Config: config.Public()}, nil
}

func (s *RealNameService) Submit(ctx context.Context, userID uint64, req webdto.RealNameSubmitRequest) (webdto.RealNameSubmitResponse, error) {
	realName := strings.TrimSpace(req.RealName)
	idType := strings.TrimSpace(req.IDType)
	idNumber := strings.ToUpper(strings.TrimSpace(req.IDNumber))
	if !idCardPattern.MatchString(idNumber) {
		return webdto.RealNameSubmitResponse{}, apperrors.ErrValidation.WithMessage("身份证号码格式错误")
	}

	var cfg realNameConfig
	var provider providerConfig
	var created models.UserRealNameApplication
	var legacyDigest string
	var providerName string
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user models.User
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", userID).First(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrUnauthorized
		}
		if err != nil {
			return err
		}

		currentConfig, secretRows, err := s.config(ctx, tx)
		if err != nil {
			return err
		}
		cfg = currentConfig
		if !cfg.Enabled {
			return apperrors.ErrForbidden.WithMessage("实名功能暂未开放")
		}

		resolvedProvider, err := cfg.ResolveProvider(req.Provider)
		if err != nil {
			return err
		}
		providerName = resolvedProvider
		isManual := providerName == providerManual
		if !isManual {
			if strings.TrimSpace(cfg.IdentityDigestSecret) == "" {
				return apperrors.ErrForbidden.WithMessage("实名摘要密钥尚未配置")
			}
			provider = cfg.Provider(resolvedProvider)
			if !provider.Complete(secretRows, cfg.CallbackBaseURL) {
				return apperrors.ErrForbidden.WithMessage("实名供应商配置不完整")
			}
		}

		latest, ok, err := s.latestForUpdate(ctx, tx, userID)
		if err != nil {
			return err
		}
		attempt := uint(1)
		if ok {
			attempt = latest.SubmitAttempt + 1
			switch latest.Status {
			case statusPending:
				return apperrors.ErrConflict.WithMessage("实名核验中，请勿重复提交")
			case statusApproved:
				return apperrors.ErrConflict.WithMessage("实名已通过，不能重复提交")
			case statusRejected:
				if !cfg.ResubmitEnabled {
					return apperrors.ErrForbidden.WithMessage("实名核验失败后暂不允许重新提交")
				}
				if int(attempt) > cfg.MaxSubmitAttempts {
					return apperrors.ErrForbidden.WithMessage("实名提交次数已达上限")
				}
			}
		}

		digest := ""
		digestVersion := ""
		if strings.TrimSpace(cfg.IdentityDigestSecret) != "" {
			digest = hmacIDNumber(idType, idNumber, cfg.IdentityDigestSecret)
			digestVersion = digestVersionHMACSHA256
			legacyDigest = legacyIDNumberDigest(idType, idNumber)
			if err := s.ensureNoDuplicateApproved(ctx, tx, userID, digest, legacyDigest); err != nil {
				return err
			}
		}

		applicationNo, err := applicationNo()
		if err != nil {
			return err
		}
		var providerStatus *string
		if !isManual {
			statusValue := providerStatusCreating
			providerStatus = &statusValue
		}
		var responseDigestValue *string
		if digest != "" {
			value := responseDigest(providerName, applicationNo, digest)
			responseDigestValue = &value
		}
		created = models.UserRealNameApplication{
			ApplicationNo:          applicationNo,
			UserID:                 userID,
			RealName:               realName,
			IDType:                 idType,
			IDNumberDigest:         textutil.StringPtr(digest),
			IDNumberDigestVersion:  textutil.StringPtr(digestVersion),
			IDNumberMasked:         maskIDNumber(idNumber),
			VerificationProvider:   &providerName,
			ProviderStatus:         providerStatus,
			ProviderResponseDigest: responseDigestValue,
			Status:                 statusPending,
			SubmitAttempt:          attempt,
		}
		return tx.Create(&created).Error
	}); err != nil {
		return webdto.RealNameSubmitResponse{}, err
	}

	if providerName == providerManual {
		return webdto.RealNameSubmitResponse{
			Application: applicationSummary(created),
			ProviderAction: webdto.RealNameProviderAction{
				Provider:   providerManual,
				ActionType: "manual_review",
			},
		}, nil
	}

	s.cacheLegacyDigest(ctx, created.ApplicationNo, legacyDigest)
	session, err := s.providerClient.CreateSession(ctx, provider.IntegrationConfig(cfg.CallbackBaseURL), realnameintegration.CreateSessionInput{
		ApplicationNo: created.ApplicationNo,
		RealName:      realName,
		IDType:        idType,
		IDNumber:      idNumber,
	})
	if err != nil {
		if updateErr := s.markProviderCreationFailed(ctx, created.ID, err); updateErr != nil {
			return webdto.RealNameSubmitResponse{}, updateErr
		}
		return webdto.RealNameSubmitResponse{}, apperrors.ErrRealNameProviderUnavailable.WithMessage("实名供应商暂不可用，请稍后重试")
	}
	if err := s.bindProviderSession(ctx, created.ID, session); err != nil {
		return webdto.RealNameSubmitResponse{}, err
	}
	if updated, ok, err := s.applicationByID(ctx, created.ID); err == nil && ok {
		created = updated
	}
	return webdto.RealNameSubmitResponse{
		Application: applicationSummary(created),
		ProviderAction: webdto.RealNameProviderAction{
			Provider:    session.Provider,
			ActionType:  session.ActionType,
			RedirectURL: session.RedirectURL,
			ExpiresAt:   session.ExpiresAt,
		},
	}, nil
}

func (s *RealNameService) Sync(ctx context.Context, userID uint64, req webdto.RealNameSyncRequest) (webdto.RealNameStatusResponse, error) {
	app, err := s.userApplicationForSync(ctx, userID, req.ApplicationNo)
	if err != nil {
		return webdto.RealNameStatusResponse{}, err
	}
	app, err = s.syncApplicationByID(ctx, app.ID, nil, true)
	if err != nil {
		return webdto.RealNameStatusResponse{}, err
	}
	config, _, err := s.config(ctx, s.db)
	if err != nil {
		return webdto.RealNameStatusResponse{}, err
	}
	summary := applicationSummary(app)
	return webdto.RealNameStatusResponse{Status: app.Status, Application: &summary, Config: config.Public()}, nil
}

func (s *RealNameService) ProviderCallback(ctx context.Context, provider string, request realnameintegration.CallbackRequest) error {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider != providerAlipay && provider != providerWechat {
		return apperrors.ErrValidation.WithMessage("实名供应商不支持")
	}
	if provider == providerWechat {
		return apperrors.ErrValidation.WithMessage("微信实名回调暂未开放，请使用服务端同步查询")
	}
	config, secretRows, err := s.config(ctx, s.db)
	if err != nil {
		return err
	}
	providerConfig := config.Provider(provider)
	if !providerConfig.Complete(secretRows, config.CallbackBaseURL) {
		return apperrors.ErrRealNameProviderUnavailable.WithMessage("实名供应商配置不完整")
	}
	callback, err := s.providerClient.ParseCallback(ctx, providerConfig.IntegrationConfig(config.CallbackBaseURL), request)
	if err != nil {
		if realnameintegration.IsInvalidCallback(err) {
			return apperrors.ErrValidation.WithMessage(err.Error())
		}
		return apperrors.ErrRealNameProviderUnavailable.WithMessage("实名供应商回调暂不可用")
	}
	if callback.Timestamp != nil {
		age := time.Since(*callback.Timestamp)
		if age > callbackAcceptWindow || age < -callbackAcceptWindow {
			return apperrors.ErrValidation.WithMessage("实名供应商回调已过期")
		}
	}
	replayAllowed, err := s.allowCallbackReplay(ctx, provider, callback.ReplayKey, callback.PayloadDigest)
	if err != nil {
		return err
	}
	if !replayAllowed {
		return nil
	}
	replayCommitted := false
	defer func() {
		if !replayCommitted {
			s.forgetCallbackReplay(ctx, provider, callback.ReplayKey)
		}
	}()
	app, err := s.applicationByProviderSession(ctx, provider, callback.ProviderApplicationID)
	if err != nil {
		return err
	}
	_, err = s.syncApplicationByID(ctx, app.ID, nil, false)
	if realnameintegration.IsUnavailable(err) {
		return apperrors.ErrRealNameProviderUnavailable.WithMessage("实名供应商结果暂不可确认")
	}
	if err != nil {
		return err
	}
	replayCommitted = true
	return nil
}

func (s *RealNameService) RequireApprovedForOrder(ctx context.Context, userID uint64) error {
	config, _, err := s.config(ctx, s.db)
	if err != nil {
		return err
	}
	if !config.RequiredForOrder {
		return nil
	}
	latest, ok, err := s.latest(ctx, s.db, userID)
	if err != nil {
		return err
	}
	if !ok || latest.Status != statusApproved {
		return apperrors.ErrForbidden.WithMessage("请先完成实名认证后再购买机器")
	}
	return nil
}

func (s *RealNameService) SyncApplicationByID(ctx context.Context, id uint64, hook SyncApplicationHook) (models.UserRealNameApplication, error) {
	return s.syncApplicationByID(ctx, id, hook, false)
}

func (s *RealNameService) syncApplicationByID(ctx context.Context, id uint64, hook SyncApplicationHook, allowUnavailable bool) (models.UserRealNameApplication, error) {
	app, ok, err := s.applicationByID(ctx, id)
	if err != nil {
		return models.UserRealNameApplication{}, err
	}
	if !ok {
		return models.UserRealNameApplication{}, apperrors.ErrNotFound.WithMessage("实名申请不存在")
	}
	if app.VerificationProvider == nil || app.ProviderApplicationID == nil {
		return models.UserRealNameApplication{}, apperrors.ErrConflict.WithMessage("实名申请缺少供应商会话")
	}
	result, err := s.queryApplicationResult(ctx, app)
	if err != nil {
		if allowUnavailable && realnameintegration.IsUnavailable(err) {
			return app, nil
		}
		return models.UserRealNameApplication{}, err
	}
	var updated models.UserRealNameApplication
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current models.UserRealNameApplication
		lockErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&current).Error
		if errors.Is(lockErr, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("实名申请不存在")
		}
		if lockErr != nil {
			return lockErr
		}
		before := current
		if err := s.applyProviderResult(ctx, tx, &current, result); err != nil {
			return err
		}
		updated = current
		if hook != nil {
			return hook(tx, before, updated)
		}
		return nil
	})
	if err != nil {
		return models.UserRealNameApplication{}, err
	}
	return updated, nil
}

func (s *RealNameService) queryApplicationResult(ctx context.Context, app models.UserRealNameApplication) (providerResult, error) {
	if app.VerificationProvider == nil || app.ProviderApplicationID == nil {
		return providerResult{}, apperrors.ErrConflict.WithMessage("实名申请缺少供应商会话")
	}
	config, secretRows, err := s.config(ctx, s.db)
	if err != nil {
		return providerResult{}, err
	}
	providerConfig := config.Provider(*app.VerificationProvider)
	if !providerConfig.Complete(secretRows, config.CallbackBaseURL) {
		return providerResult{}, apperrors.ErrRealNameProviderUnavailable.WithMessage("实名供应商配置不完整")
	}
	result, err := s.providerClient.QueryResult(ctx, providerConfig.IntegrationConfig(config.CallbackBaseURL), *app.ProviderApplicationID)
	if err != nil {
		return providerResult{}, err
	}
	return providerResult{
		ProviderStatus: result.ProviderStatus,
		FinalStatus:    result.FinalStatus,
		ResultCode:     result.ResultCode,
		ResultMessage:  result.ResultMessage,
		ResponseDigest: result.ResponseDigest,
		TraceID:        result.TraceID,
	}, nil
}

func (s *RealNameService) applyProviderResult(ctx context.Context, tx *gorm.DB, app *models.UserRealNameApplication, result providerResult) error {
	if app.Status != statusPending {
		return nil
	}
	now := time.Now()
	updates := map[string]any{
		"provider_status":          result.ProviderStatus,
		"provider_result_code":     textutil.NormalizeOptionalString(&result.ResultCode),
		"provider_result_message":  textutil.NormalizeOptionalString(&result.ResultMessage),
		"provider_response_digest": textutil.NormalizeOptionalString(&result.ResponseDigest),
		"provider_trace_id":        textutil.NormalizeOptionalString(&result.TraceID),
	}
	switch result.FinalStatus {
	case statusApproved:
		legacyDigest := s.loadLegacyDigest(ctx, app.ApplicationNo)
		digest := ""
		if app.IDNumberDigest != nil {
			digest = *app.IDNumberDigest
		}
		if strings.TrimSpace(digest) == "" {
			updates["status"] = statusRejected
			updates["reject_reason"] = "实名申请缺少证件摘要"
			updates["provider_status"] = providerStatusRejected
			updates["provider_finished_at"] = now
			if err := tx.Model(app).Updates(updates).Error; err != nil {
				return err
			}
			app.Status = statusRejected
			reason := "实名申请缺少证件摘要"
			app.RejectReason = &reason
			app.ProviderFinishedAt = &now
			return nil
		}
		if err := s.ensureNoDuplicateApproved(ctx, tx, app.UserID, digest, legacyDigest); err != nil {
			updates["status"] = statusRejected
			updates["reject_reason"] = "证件号码已被其它用户实名"
			updates["provider_status"] = providerStatusRejected
			updates["provider_finished_at"] = now
			if err := tx.Model(app).Updates(updates).Error; err != nil {
				return err
			}
			app.Status = statusRejected
			reason := "证件号码已被其它用户实名"
			app.RejectReason = &reason
			app.ProviderFinishedAt = &now
			return nil
		}
		updates["status"] = statusApproved
		updates["reject_reason"] = nil
		updates["provider_finished_at"] = now
		updates["provider_status"] = providerStatusApproved
	case statusRejected:
		updates["status"] = statusRejected
		updates["reject_reason"] = result.UserMessage()
		updates["provider_finished_at"] = now
		updates["provider_status"] = providerStatusRejected
	default:
		updates["status"] = statusPending
		updates["provider_status"] = providerStatusPending
	}
	if err := tx.Model(app).Updates(updates).Error; err != nil {
		if result.FinalStatus == statusApproved && isDuplicateApprovedDigest(err) {
			updates["status"] = statusRejected
			updates["reject_reason"] = "证件号码已被其它用户实名"
			updates["provider_finished_at"] = now
			updates["provider_status"] = providerStatusRejected
			if rejectErr := tx.Model(app).Updates(updates).Error; rejectErr != nil {
				return rejectErr
			}
			return tx.Where("id = ?", app.ID).First(app).Error
		}
		return err
	}
	if err := tx.Where("id = ?", app.ID).First(app).Error; err != nil {
		return err
	}
	return nil
}

func (s *RealNameService) markProviderCreationFailed(ctx context.Context, id uint64, providerErr error) error {
	now := time.Now()
	providerStatus := providerStatusRejected
	resultCode := "CREATE_FAILED"
	resultMessage := "实名供应商暂不可用，请稍后重试"
	_ = providerErr
	return s.db.WithContext(ctx).Model(&models.UserRealNameApplication{}).Where("id = ?", id).Updates(map[string]any{
		"status":                  statusRejected,
		"reject_reason":           resultMessage,
		"provider_status":         &providerStatus,
		"provider_result_code":    &resultCode,
		"provider_result_message": &resultMessage,
		"provider_finished_at":    &now,
		"provider_trace_id":       nil,
	}).Error
}

func (s *RealNameService) bindProviderSession(ctx context.Context, id uint64, session realnameintegration.Session) error {
	now := time.Now()
	providerStatus := providerStatusCreated
	responseDigest := strings.TrimSpace(session.ResponseDigest)
	traceID := strings.TrimSpace(session.TraceID)
	return s.db.WithContext(ctx).Model(&models.UserRealNameApplication{}).Where("id = ?", id).Updates(map[string]any{
		"provider_application_id":  session.ProviderApplicationID,
		"provider_status":          &providerStatus,
		"provider_started_at":      &now,
		"provider_response_digest": textutil.NormalizeOptionalString(&responseDigest),
		"provider_trace_id":        textutil.NormalizeOptionalString(&traceID),
	}).Error
}

func (s *RealNameService) userApplicationForSync(ctx context.Context, userID uint64, applicationNo string) (models.UserRealNameApplication, error) {
	query := s.db.WithContext(ctx).Where("user_id = ?", userID)
	trimmedNo := strings.TrimSpace(applicationNo)
	if trimmedNo != "" {
		query = query.Where("application_no = ?", trimmedNo)
	} else {
		query = query.Where("status = ?", statusPending).Order("id DESC")
	}
	var app models.UserRealNameApplication
	if err := query.First(&app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.UserRealNameApplication{}, apperrors.ErrNotFound.WithMessage("实名申请不存在")
		}
		return models.UserRealNameApplication{}, err
	}
	if app.Status != statusPending {
		return models.UserRealNameApplication{}, apperrors.ErrConflict.WithMessage("当前实名申请无需同步")
	}
	return app, nil
}

func (s *RealNameService) applicationByProviderSession(ctx context.Context, provider string, providerApplicationID string) (models.UserRealNameApplication, error) {
	var app models.UserRealNameApplication
	err := s.db.WithContext(ctx).
		Where("verification_provider = ? AND provider_application_id = ?", provider, providerApplicationID).
		First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.UserRealNameApplication{}, apperrors.ErrNotFound.WithMessage("实名申请不存在")
	}
	return app, err
}

func (s *RealNameService) applicationByID(ctx context.Context, id uint64) (models.UserRealNameApplication, bool, error) {
	var app models.UserRealNameApplication
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.UserRealNameApplication{}, false, nil
	}
	return app, err == nil, err
}

func (s *RealNameService) cacheLegacyDigest(ctx context.Context, applicationNo string, legacyDigest string) {
	if s.redis == nil || strings.TrimSpace(applicationNo) == "" || strings.TrimSpace(legacyDigest) == "" {
		return
	}
	_ = s.redis.Client().Set(ctx, s.redis.Key("web", "real_name", "legacy_digest", applicationNo), legacyDigest, legacyDigestCacheTTL).Err()
}

func (s *RealNameService) loadLegacyDigest(ctx context.Context, applicationNo string) string {
	if s.redis == nil || strings.TrimSpace(applicationNo) == "" {
		return ""
	}
	value, err := s.redis.Client().Get(ctx, s.redis.Key("web", "real_name", "legacy_digest", applicationNo)).Result()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func (s *RealNameService) allowCallbackReplay(ctx context.Context, provider string, replayKey string, payloadDigest string) (bool, error) {
	if s.redis == nil || strings.TrimSpace(replayKey) == "" {
		return true, nil
	}
	key := s.redis.Key("web", "real_name", "callback", provider, replayKey)
	ok, err := s.redis.Client().SetNX(ctx, key, payloadDigest, callbackReplayTTL).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (s *RealNameService) forgetCallbackReplay(ctx context.Context, provider string, replayKey string) {
	if s.redis == nil || strings.TrimSpace(replayKey) == "" {
		return
	}
	_ = s.redis.Client().Del(ctx, s.redis.Key("web", "real_name", "callback", provider, replayKey)).Err()
}

func (s *RealNameService) ensureNoDuplicateApproved(ctx context.Context, tx *gorm.DB, userID uint64, digest string, legacyDigest string) error {
	digests := []string{digest}
	if legacyDigest != "" && legacyDigest != digest {
		digests = append(digests, legacyDigest)
	}
	var duplicate int64
	if err := tx.WithContext(ctx).Model(&models.UserRealNameApplication{}).
		Where("id_number_digest IN ? AND status = ? AND user_id <> ?", digests, statusApproved, userID).
		Count(&duplicate).Error; err != nil {
		return err
	}
	if duplicate > 0 {
		return apperrors.ErrConflict.WithMessage("该证件号码已完成实名")
	}
	return nil
}

func (s *RealNameService) latest(ctx context.Context, db *gorm.DB, userID uint64) (models.UserRealNameApplication, bool, error) {
	var app models.UserRealNameApplication
	err := db.WithContext(ctx).Where("user_id = ?", userID).Order("id DESC").First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.UserRealNameApplication{}, false, nil
	}
	return app, err == nil, err
}

func (s *RealNameService) latestForUpdate(ctx context.Context, db *gorm.DB, userID uint64) (models.UserRealNameApplication, bool, error) {
	var app models.UserRealNameApplication
	err := db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).Order("id DESC").First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.UserRealNameApplication{}, false, nil
	}
	return app, err == nil, err
}

func (s *RealNameService) config(ctx context.Context, db *gorm.DB) (realNameConfig, map[string]bool, error) {
	config := defaultRealNameConfig()
	var rows []models.SystemConfig
	if err := db.WithContext(ctx).Where("config_key LIKE ?", "real_name.%").Find(&rows).Error; err != nil {
		return config, nil, err
	}
	secretRows := make(map[string]bool)
	values := make(map[string]string)
	for _, row := range rows {
		value := ""
		if row.ConfigValue != nil {
			value = strings.TrimSpace(*row.ConfigValue)
		}
		values[row.ConfigKey] = value
		if row.IsSecret && value != "" {
			secretRows[row.ConfigKey] = true
		}
	}
	config.Apply(values)
	config.Finalize(secretRows)
	return config, secretRows, nil
}

func applicationSummary(app models.UserRealNameApplication) webdto.RealNameApplicationSummary {
	return webdto.RealNameApplicationSummary{
		ApplicationNo:        app.ApplicationNo,
		RealName:             app.RealName,
		IDType:               app.IDType,
		IDNumberMasked:       app.IDNumberMasked,
		VerificationProvider: app.VerificationProvider,
		ProviderStatus:       app.ProviderStatus,
		Status:               app.Status,
		FailureReason:        app.RejectReason,
		SubmitAttempt:        app.SubmitAttempt,
		CreatedAt:            app.CreatedAt,
		VerifiedAt:           app.ProviderFinishedAt,
	}
}

func hmacIDNumber(idType, idNumber, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strings.ToLower(strings.TrimSpace(idType)) + ":" + strings.ToUpper(strings.TrimSpace(idNumber))))
	return hex.EncodeToString(mac.Sum(nil))
}

func legacyIDNumberDigest(idType, idNumber string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(idType)) + ":" + strings.ToUpper(strings.TrimSpace(idNumber))))
	return hex.EncodeToString(sum[:])
}

func maskIDNumber(value string) string {
	if len(value) <= 8 {
		return value
	}
	return value[:4] + strings.Repeat("*", len(value)-8) + value[len(value)-4:]
}

func applicationNo() (string, error) {
	random, err := randomHex(4)
	if err != nil {
		return "", err
	}
	return "RN" + time.Now().Format("20060102150405") + strings.ToUpper(random), nil
}

func randomHex(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func isDuplicateApprovedDigest(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

type realNameConfig struct {
	Enabled              bool
	RequiredForOrder     bool
	AllowedProviders     []string
	AvailableProviders   []string
	DefaultProvider      string
	ManualReviewEnabled  bool
	IdentityDigestSecret string
	CallbackBaseURL      string
	ResubmitEnabled      bool
	MaxSubmitAttempts    int
	ReviewNotice         string
	Alipay               providerConfig
	Wechat               providerConfig
}

func defaultRealNameConfig() realNameConfig {
	return realNameConfig{
		RequiredForOrder:    true,
		AllowedProviders:    []string{providerAlipay, providerWechat},
		DefaultProvider:     providerAlipay,
		ManualReviewEnabled: true,
		ResubmitEnabled:     true,
		MaxSubmitAttempts:   3,
		Alipay:              providerConfig{Provider: providerAlipay, GatewayURL: "https://openapi.alipay.com/gateway.do"},
		Wechat:              providerConfig{Provider: providerWechat, Region: "ap-guangzhou", Endpoint: "faceid.tencentcloudapi.com"},
	}
}

func (c *realNameConfig) Apply(values map[string]string) {
	c.Enabled = parseBool(values["real_name.enabled"])
	if value, ok := values["real_name.required_for_order"]; ok {
		c.RequiredForOrder = parseBool(value)
	}
	c.AllowedProviders = csv(values["real_name.allowed_providers"], c.AllowedProviders)
	if value := strings.TrimSpace(values["real_name.default_provider"]); value != "" {
		c.DefaultProvider = value
	}
	if value, ok := values["real_name.manual_review_enabled"]; ok {
		c.ManualReviewEnabled = parseBool(value)
	}
	c.IdentityDigestSecret = values["real_name.identity_digest_secret"]
	c.CallbackBaseURL = values["real_name.callback_base_url"]
	if value, ok := values["real_name.resubmit_enabled"]; ok {
		c.ResubmitEnabled = parseBool(value)
	}
	if value := strings.TrimSpace(values["real_name.max_submit_attempts"]); value != "" {
		c.MaxSubmitAttempts = positiveInt(value, c.MaxSubmitAttempts)
	}
	c.ReviewNotice = values["real_name.review_notice"]
	c.Alipay.Apply(values)
	c.Wechat.Apply(values)
}

func (c realNameConfig) Public() webdto.RealNameConfig {
	return webdto.RealNameConfig{
		Enabled:           c.Enabled,
		RequiredForOrder:  c.RequiredForOrder,
		AllowedProviders:  append([]string(nil), c.AvailableProviders...),
		DefaultProvider:   c.DefaultProvider,
		ResubmitEnabled:   c.ResubmitEnabled,
		MaxSubmitAttempts: c.MaxSubmitAttempts,
		ReviewNotice:      c.ReviewNotice,
	}
}

func (c *realNameConfig) Finalize(secretRows map[string]bool) {
	c.AllowedProviders = filterSupportedProviders(c.AllowedProviders)
	if strings.TrimSpace(c.IdentityDigestSecret) != "" {
		c.AvailableProviders = c.availableProviders(secretRows)
	}
	if len(c.AvailableProviders) == 0 && c.ManualReviewEnabled {
		c.AvailableProviders = []string{providerManual}
	}
	c.Enabled = c.Enabled && len(c.AvailableProviders) > 0
	if !containsString(c.AvailableProviders, c.DefaultProvider) {
		if len(c.AvailableProviders) > 0 {
			c.DefaultProvider = c.AvailableProviders[0]
		} else {
			c.DefaultProvider = ""
		}
	}
}

func (c realNameConfig) ResolveProvider(requested string) (string, error) {
	provider := strings.ToLower(strings.TrimSpace(requested))
	if provider == "" {
		provider = c.DefaultProvider
	}
	if provider == "" || !containsString(c.AvailableProviders, provider) {
		return "", apperrors.ErrValidation.WithMessage("实名供应商不可用")
	}
	return provider, nil
}

func (c realNameConfig) Provider(provider string) providerConfig {
	switch provider {
	case providerAlipay:
		return c.Alipay
	case providerWechat:
		return c.Wechat
	case providerManual:
		return providerConfig{Provider: providerManual, Enabled: true}
	default:
		return providerConfig{Provider: provider}
	}
}

func (c realNameConfig) availableProviders(secretRows map[string]bool) []string {
	result := make([]string, 0, len(c.AllowedProviders))
	for _, provider := range c.AllowedProviders {
		cfg := c.Provider(provider)
		if cfg.Enabled && cfg.Complete(secretRows, c.CallbackBaseURL) {
			result = append(result, provider)
		}
	}
	sort.Strings(result)
	return result
}

type providerConfig struct {
	Provider        string
	Enabled         bool
	AppID           string
	GatewayURL      string
	AppPrivateKey   string
	AlipayPublicKey string
	ReturnURL       string
	NotifyURL       string
	SecretID        string
	SecretKey       string
	Region          string
	Endpoint        string
	RuleID          string
	RedirectURL     string
}

func (p *providerConfig) Apply(values map[string]string) {
	prefix := "real_name." + p.Provider + "."
	p.Enabled = parseBool(values[prefix+"enabled"])
	switch p.Provider {
	case providerAlipay:
		p.AppID = values[prefix+"app_id"]
		if value := strings.TrimSpace(values[prefix+"gateway_url"]); value != "" {
			p.GatewayURL = value
		}
		p.AppPrivateKey = values[prefix+"app_private_key"]
		p.AlipayPublicKey = values[prefix+"alipay_public_key"]
		p.ReturnURL = values[prefix+"return_url"]
		p.NotifyURL = values[prefix+"notify_url"]
	case providerWechat:
		p.SecretID = values[prefix+"secret_id"]
		p.SecretKey = values[prefix+"secret_key"]
		if value := strings.TrimSpace(values[prefix+"region"]); value != "" {
			p.Region = value
		}
		if value := strings.TrimSpace(values[prefix+"endpoint"]); value != "" {
			p.Endpoint = value
		}
		p.RuleID = values[prefix+"rule_id"]
		p.RedirectURL = values[prefix+"redirect_url"]
	}
}

func (p providerConfig) IntegrationConfig(callbackBaseURL string) realnameintegration.ProviderConfig {
	notifyURL := strings.TrimSpace(p.NotifyURL)
	if notifyURL == "" {
		notifyURL = defaultProviderCallbackURL(callbackBaseURL, p.Provider)
	}
	return realnameintegration.ProviderConfig{
		Provider:        p.Provider,
		AppID:           p.AppID,
		GatewayURL:      p.GatewayURL,
		AppPrivateKey:   p.AppPrivateKey,
		AlipayPublicKey: p.AlipayPublicKey,
		ReturnURL:       p.ReturnURL,
		NotifyURL:       notifyURL,
		CallbackBaseURL: callbackBaseURL,
		SecretID:        p.SecretID,
		SecretKey:       p.SecretKey,
		Region:          p.Region,
		Endpoint:        p.Endpoint,
		RuleID:          p.RuleID,
		RedirectURL:     p.RedirectURL,
	}
}

func (p providerConfig) Complete(secretRows map[string]bool, callbackBaseURL string) bool {
	if !p.Enabled {
		return false
	}
	switch p.Provider {
	case providerAlipay:
		return strings.TrimSpace(p.AppID) != "" &&
			strings.TrimSpace(p.GatewayURL) != "" &&
			strings.TrimSpace(p.ReturnURL) != "" &&
			(strings.TrimSpace(p.NotifyURL) != "" || strings.TrimSpace(callbackBaseURL) != "") &&
			secretRows["real_name.alipay.app_private_key"] &&
			secretRows["real_name.alipay.alipay_public_key"]
	case providerWechat:
		return strings.TrimSpace(p.Region) != "" &&
			strings.TrimSpace(p.Endpoint) != "" &&
			strings.TrimSpace(p.RuleID) != "" &&
			strings.TrimSpace(p.RedirectURL) != "" &&
			secretRows["real_name.wechat.secret_id"] &&
			secretRows["real_name.wechat.secret_key"]
	default:
		return false
	}
}

type providerResult struct {
	ProviderStatus string
	FinalStatus    string
	ResultCode     string
	ResultMessage  string
	ResponseDigest string
	TraceID        string
}

func (r providerResult) UserMessage() string {
	if strings.TrimSpace(r.ResultMessage) != "" {
		return strings.TrimSpace(r.ResultMessage)
	}
	if strings.TrimSpace(r.ResultCode) != "" {
		return "实名供应商核验失败：" + strings.TrimSpace(r.ResultCode)
	}
	return "实名供应商核验失败"
}

func responseDigest(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])
}

func defaultProviderCallbackURL(base string, provider string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		return ""
	}
	return base + "/" + strings.Trim(strings.ToLower(provider), "/")
}

func parseBool(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), "true")
}

func positiveInt(value string, fallback int) int {
	var parsed int
	if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &parsed); err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func csv(value string, fallback []string) []string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.ToLower(strings.TrimSpace(part))
		if item != "" {
			result = append(result, item)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}

func filterSupportedProviders(providers []string) []string {
	result := make([]string, 0, len(providers))
	for _, provider := range providers {
		if (provider == providerAlipay || provider == providerWechat || provider == providerManual) && !containsString(result, provider) {
			result = append(result, provider)
		}
	}
	return result
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(target)) {
			return true
		}
	}
	return false
}
