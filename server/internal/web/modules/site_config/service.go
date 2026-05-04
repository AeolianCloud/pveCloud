package siteconfig

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
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
)

/**
 * SiteConfigService 处理 Web 公开站点配置读取。
 */
type SiteConfigService struct {
	db *gorm.DB
}

/**
 * NewSiteConfigService 创建 Web 公开站点配置服务。
 */
func NewSiteConfigService(db *gorm.DB) *SiteConfigService {
	return &SiteConfigService{db: db}
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
			},
		).
		Find(&configs).Error; err != nil {
		return webdto.SiteConfigResponse{}, err
	}

	result := webdto.SiteConfigResponse{SiteName: defaultSiteName}
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
		}
	}

	return result, nil
}

func parseBoolConfigValue(value *string) bool {
	if value == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(*value), "true")
}
