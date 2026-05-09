package realname

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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
)

const (
	statusPending  = "pending"
	statusApproved = "approved"
	statusRejected = "rejected"

	providerAlipay = "alipay"
	providerWechat = "wechat"

	providerStatusApproved = "approved"
	providerStatusRejected = "rejected"
	providerStatusPending  = "pending"
)

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

func (s *RealNameService) SyncApplicationByID(ctx context.Context, id uint64, hook SyncApplicationHook) (models.UserRealNameApplication, error) {
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

func (s *RealNameService) applicationByID(ctx context.Context, id uint64) (models.UserRealNameApplication, bool, error) {
	var app models.UserRealNameApplication
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.UserRealNameApplication{}, false, nil
	}
	return app, err == nil, err
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
	config.AllowedProviders = filterSupportedProviders(config.AllowedProviders)
	config.AvailableProviders = config.availableProviders(secretRows)
	if !containsString(config.AvailableProviders, config.DefaultProvider) {
		if len(config.AvailableProviders) > 0 {
			config.DefaultProvider = config.AvailableProviders[0]
		} else {
			config.DefaultProvider = ""
		}
	}
	return config, secretRows, nil
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
		RequiredForOrder:  true,
		AllowedProviders:  []string{providerAlipay, providerWechat},
		DefaultProvider:   providerAlipay,
		ResubmitEnabled:   true,
		MaxSubmitAttempts: 3,
		Alipay:            providerConfig{Provider: providerAlipay, GatewayURL: "https://openapi.alipay.com/gateway.do"},
		Wechat:            providerConfig{Provider: providerWechat, Region: "ap-guangzhou", Endpoint: "faceid.tencentcloudapi.com"},
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

func (c realNameConfig) Provider(provider string) providerConfig {
	switch provider {
	case providerAlipay:
		return c.Alipay
	case providerWechat:
		return c.Wechat
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
		if (provider == providerAlipay || provider == providerWechat) && !containsString(result, provider) {
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
