package siteconfig

import (
	"context"
	"sort"
	"strconv"
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

type SiteConfigService struct {
	db *gorm.DB
}

func NewSiteConfigService(db *gorm.DB) *SiteConfigService {
	return &SiteConfigService{db: db}
}

func (s *SiteConfigService) Show(ctx context.Context) (webdto.SiteConfigResponse, error) {
	var configs []models.SystemConfig
	if err := s.db.WithContext(ctx).
		Where(
			"config_key IN ? OR config_key LIKE ?",
			[]string{
				siteNameKey,
				siteLogoURLKey,
				loginCaptchaEnabledKey,
				registerCaptchaEnabledKey,
				passwordResetRequestCaptchaKey,
				passwordResetConfirmCaptchaKey,
			},
			"real_name.%",
		).
		Find(&configs).Error; err != nil {
		return webdto.SiteConfigResponse{}, err
	}

	result := webdto.SiteConfigResponse{SiteName: defaultSiteName, RealName: defaultRealNameConfig()}
	values := map[string]string{}
	secrets := map[string]bool{}
	for _, config := range configs {
		value := ""
		if config.ConfigValue != nil {
			value = strings.TrimSpace(*config.ConfigValue)
		}
		values[config.ConfigKey] = value
		if config.IsSecret && value != "" {
			secrets[config.ConfigKey] = true
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
	result.RealName = publicRealNameConfig(values, secrets)
	return result, nil
}

func defaultRealNameConfig() webdto.RealNameConfig {
	return webdto.RealNameConfig{
		RequiredForOrder:  true,
		AllowedProviders:  []string{},
		DefaultProvider:   "",
		ResubmitEnabled:   true,
		MaxSubmitAttempts: 3,
	}
}

func publicRealNameConfig(values map[string]string, secrets map[string]bool) webdto.RealNameConfig {
	config := defaultRealNameConfig()
	entryEnabled := strings.EqualFold(values["real_name.enabled"], "true")
	if value, ok := values["real_name.required_for_order"]; ok {
		config.RequiredForOrder = strings.EqualFold(value, "true")
	}
	allowed := filterSupportedProviders(parseCSVConfigValue(textPtr(values["real_name.allowed_providers"]), []string{"alipay", "wechat"}))
	available := make([]string, 0, len(allowed))
	for _, provider := range allowed {
		if secrets["real_name.identity_digest_secret"] && providerComplete(provider, values, secrets) {
			available = append(available, provider)
		}
	}
	sort.Strings(available)
	manualEnabled := true
	if value, ok := values["real_name.manual_review_enabled"]; ok {
		manualEnabled = strings.EqualFold(value, "true")
	}
	if len(available) == 0 && manualEnabled {
		available = append(available, "manual")
	}
	config.AllowedProviders = available
	config.Enabled = entryEnabled && len(available) > 0
	config.DefaultProvider = strings.ToLower(strings.TrimSpace(values["real_name.default_provider"]))
	if !containsString(available, config.DefaultProvider) {
		if len(available) > 0 {
			config.DefaultProvider = available[0]
		} else {
			config.DefaultProvider = ""
		}
	}
	if value, ok := values["real_name.resubmit_enabled"]; ok {
		config.ResubmitEnabled = strings.EqualFold(value, "true")
	}
	config.MaxSubmitAttempts = parseIntConfigValue(textPtr(values["real_name.max_submit_attempts"]), config.MaxSubmitAttempts)
	config.ReviewNotice = values["real_name.review_notice"]
	return config
}

func providerComplete(provider string, values map[string]string, secrets map[string]bool) bool {
	switch provider {
	case "alipay":
		return strings.EqualFold(values["real_name.alipay.enabled"], "true") &&
			strings.TrimSpace(values["real_name.alipay.app_id"]) != "" &&
			strings.TrimSpace(values["real_name.alipay.gateway_url"]) != "" &&
			strings.TrimSpace(values["real_name.alipay.return_url"]) != "" &&
			(strings.TrimSpace(values["real_name.alipay.notify_url"]) != "" || strings.TrimSpace(values["real_name.callback_base_url"]) != "") &&
			secrets["real_name.alipay.app_private_key"] &&
			secrets["real_name.alipay.alipay_public_key"]
	case "wechat":
		return strings.EqualFold(values["real_name.wechat.enabled"], "true") &&
			strings.TrimSpace(values["real_name.wechat.region"]) != "" &&
			strings.TrimSpace(values["real_name.wechat.endpoint"]) != "" &&
			strings.TrimSpace(values["real_name.wechat.rule_id"]) != "" &&
			strings.TrimSpace(values["real_name.wechat.redirect_url"]) != "" &&
			secrets["real_name.wechat.secret_id"] &&
			secrets["real_name.wechat.secret_key"]
	default:
		return false
	}
}

func parseBoolConfigValue(value *string) bool {
	return value != nil && strings.EqualFold(strings.TrimSpace(*value), "true")
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
		if (provider == "alipay" || provider == "wechat") && !containsString(result, provider) {
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

func textPtr(value string) *string {
	return &value
}
