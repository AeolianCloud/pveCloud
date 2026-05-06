package siteconfig

import (
	"context"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/bootstrap"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/web/dto"
)

const (
	defaultSiteName                = "pveCloud"
	siteNameKey                    = "site.name"
	siteLogoURLKey                 = "site.logo_url"
	loginCaptchaEnabledKey         = "web.auth.login_captcha_enabled"
	registerCaptchaEnabledKey      = "web.auth.register_captcha_enabled"
	passwordResetRequestCaptchaKey = "web.auth.password_reset_request_captcha_enabled"
	passwordResetConfirmCaptchaKey = "web.auth.password_reset_confirm_captcha_enabled"
	realNameEnabledKey             = "real_name.enabled"
	realNameRequiredForOrderKey    = "real_name.required_for_order"
	realNameResubmitEnabledKey     = "real_name.resubmit_enabled"
	realNameMaxSubmitAttemptsKey   = "real_name.max_submit_attempts"
	realNameFrontRequiredKey       = "real_name.id_card_front_required"
	realNameBackRequiredKey        = "real_name.id_card_back_required"
	realNameHoldRequiredKey        = "real_name.hold_card_required"
	realNameImageMaxSizeMBKey      = "real_name.image_max_size_mb"
	realNameAllowedImageTypesKey   = "real_name.allowed_image_types"
	realNameReviewNoticeKey        = "real_name.review_notice"
)

/**
 * SiteConfigService 处理 Web 公开站点配置读取。
 */
type SiteConfigService struct {
	db      *gorm.DB
	storage bootstrap.StorageConfig
}

/**
 * NewSiteConfigService 创建 Web 公开站点配置服务。
 */
func NewSiteConfigService(db *gorm.DB, storage bootstrap.StorageConfig) *SiteConfigService {
	return &SiteConfigService{db: db, storage: storage}
}

/**
 * Show 读取 Web 公开站点基础展示配置。
 */
func (s *SiteConfigService) Show(ctx context.Context) (webdto.SiteConfigResponse, error) {
	var configs []models.SystemConfig
	if err := s.db.WithContext(ctx).
		Where(
			"config_key IN ? AND is_secret = 0",
			[]string{
				siteNameKey,
				siteLogoURLKey,
				loginCaptchaEnabledKey,
				registerCaptchaEnabledKey,
				passwordResetRequestCaptchaKey,
				passwordResetConfirmCaptchaKey,
				realNameEnabledKey,
				realNameRequiredForOrderKey,
				realNameResubmitEnabledKey,
				realNameMaxSubmitAttemptsKey,
				realNameFrontRequiredKey,
				realNameBackRequiredKey,
				realNameHoldRequiredKey,
				realNameImageMaxSizeMBKey,
				realNameAllowedImageTypesKey,
				realNameReviewNoticeKey,
			},
		).
		Find(&configs).Error; err != nil {
		return webdto.SiteConfigResponse{}, err
	}

	result := webdto.SiteConfigResponse{SiteName: defaultSiteName, RealName: defaultRealNameConfig()}
	for _, config := range configs {
		value := ""
		if config.ConfigValue != nil {
			value = strings.TrimSpace(*config.ConfigValue)
		}
		switch config.ConfigKey {
		case siteNameKey:
			if value != "" {
				result.SiteName = value
			}
		case siteLogoURLKey:
			result.LogoURL = value
		case loginCaptchaEnabledKey:
			result.LoginCaptchaEnabled = parseBoolConfigValue(config.ConfigValue)
		case registerCaptchaEnabledKey:
			result.RegisterCaptchaEnabled = parseBoolConfigValue(config.ConfigValue)
		case passwordResetRequestCaptchaKey:
			result.PasswordResetRequestCaptchaEnabled = parseBoolConfigValue(config.ConfigValue)
		case passwordResetConfirmCaptchaKey:
			result.PasswordResetConfirmCaptchaEnabled = parseBoolConfigValue(config.ConfigValue)
		case realNameEnabledKey:
			result.RealName.Enabled = parseBoolConfigValue(config.ConfigValue)
		case realNameRequiredForOrderKey:
			result.RealName.RequiredForOrder = parseBoolConfigValue(config.ConfigValue)
		case realNameResubmitEnabledKey:
			result.RealName.ResubmitEnabled = parseBoolConfigValue(config.ConfigValue)
		case realNameMaxSubmitAttemptsKey:
			result.RealName.MaxSubmitAttempts = parseIntConfigValue(config.ConfigValue, result.RealName.MaxSubmitAttempts)
		case realNameFrontRequiredKey:
			result.RealName.IDCardFrontRequired = parseBoolConfigValue(config.ConfigValue)
		case realNameBackRequiredKey:
			result.RealName.IDCardBackRequired = parseBoolConfigValue(config.ConfigValue)
		case realNameHoldRequiredKey:
			result.RealName.HoldCardRequired = parseBoolConfigValue(config.ConfigValue)
		case realNameImageMaxSizeMBKey:
			result.RealName.ImageMaxSizeMB = parseIntConfigValue(config.ConfigValue, result.RealName.ImageMaxSizeMB)
		case realNameAllowedImageTypesKey:
			result.RealName.AllowedImageTypes = parseCSVConfigValue(config.ConfigValue, result.RealName.AllowedImageTypes)
		case realNameReviewNoticeKey:
			result.RealName.ReviewNotice = value
		}
	}

	result.RealName = effectiveRealNameConfig(result.RealName, s.storage)
	return result, nil
}

func defaultRealNameConfig() webdto.RealNameConfig {
	return webdto.RealNameConfig{
		RequiredForOrder:    true,
		ResubmitEnabled:     true,
		MaxSubmitAttempts:   3,
		IDCardFrontRequired: true,
		IDCardBackRequired:  true,
		ImageMaxSizeMB:      5,
		AllowedImageTypes:   []string{"image/jpeg", "image/png", "image/webp"},
	}
}

func parseBoolConfigValue(value *string) bool {
	if value == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(*value), "true")
}

func parseIntConfigValue(value *string, fallback int) int {
	if value == nil {
		return fallback
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(*value))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseCSVConfigValue(value *string, fallback []string) []string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return fallback
	}
	parts := strings.Split(*value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			result = append(result, item)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}

func effectiveRealNameConfig(config webdto.RealNameConfig, storage bootstrap.StorageConfig) webdto.RealNameConfig {
	if storage.MaxSize > 0 {
		storageMaxMB := int(storage.MaxSize / (1024 * 1024))
		if storageMaxMB <= 0 {
			storageMaxMB = 1
		}
		if config.ImageMaxSizeMB <= 0 || config.ImageMaxSizeMB > storageMaxMB {
			config.ImageMaxSizeMB = storageMaxMB
		}
	}
	config.AllowedImageTypes = intersectStrings(config.AllowedImageTypes, storage.AllowedTypes)
	return config
}

func intersectStrings(left []string, right []string) []string {
	rightSet := make(map[string]struct{}, len(right))
	for _, item := range right {
		trimmed := strings.ToLower(strings.TrimSpace(item))
		if trimmed != "" {
			rightSet[trimmed] = struct{}{}
		}
	}
	result := make([]string, 0, len(left))
	seen := map[string]struct{}{}
	for _, item := range left {
		trimmed := strings.ToLower(strings.TrimSpace(item))
		if trimmed == "" {
			continue
		}
		if _, ok := rightSet[trimmed]; !ok {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}
