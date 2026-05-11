package siteconfig

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	domainfile "github.com/AeolianCloud/pveCloud/server/internal/domain/file"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	mysqlfile "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/file"
	mysqlsystemconfig "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/systemconfig"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"gorm.io/gorm"
)

const (
	defaultSiteName                = "pveCloud"
	siteNameKey                    = "site.name"
	siteLogoURLKey                 = "site.logo_url"
	loginCaptchaEnabledKey         = "web.auth.login_captcha_enabled"
	registerCaptchaEnabledKey      = "web.auth.register_captcha_enabled"
	passwordResetRequestCaptchaKey = "web.auth.password_reset_request_captcha_enabled"
	passwordResetConfirmCaptchaKey = "web.auth.password_reset_confirm_captcha_enabled"
	realNameConfigPrefix           = "real_name."
)

var adminFileURLPattern = regexp.MustCompile(`^/admin-api/files/(\d+)(?:/download)?$`)

type SiteConfigService struct {
	configs *mysqlsystemconfig.Repository
	files   *mysqlfile.Repository
	storage config.StorageConfig
}

type SiteConfig struct {
	SiteName                           string
	LogoURL                            string
	LoginCaptchaEnabled                bool
	RegisterCaptchaEnabled             bool
	PasswordResetRequestCaptchaEnabled bool
	PasswordResetConfirmCaptchaEnabled bool
	RealName                           RealNameConfig
}

type RealNameConfig struct {
	Enabled           bool
	RequiredForOrder  bool
	AllowedProviders  []string
	DefaultProvider   string
	ResubmitEnabled   bool
	MaxSubmitAttempts int
	ReviewNotice      string
}

func NewSiteConfigService(configs *mysqlsystemconfig.Repository, files *mysqlfile.Repository, storage config.StorageConfig) *SiteConfigService {
	return &SiteConfigService{configs: configs, files: files, storage: storage}
}

func (s *SiteConfigService) Show(ctx context.Context) (SiteConfig, error) {
	configs, err := s.configs.PublicSiteConfigRows(ctx, []string{
		siteNameKey,
		siteLogoURLKey,
		loginCaptchaEnabledKey,
		registerCaptchaEnabledKey,
		passwordResetRequestCaptchaKey,
		passwordResetConfirmCaptchaKey,
	}, realNameConfigPrefix)
	if err != nil {
		return SiteConfig{}, err
	}
	return buildSiteConfig(configs), nil
}

func buildSiteConfig(configs []mysqlsystemconfig.SystemConfig) SiteConfig {
	result := SiteConfig{SiteName: defaultSiteName, RealName: defaultRealNameConfig()}
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
			result.LogoURL = publicLogoURL(value)
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
	return result
}

func (s *SiteConfigService) PublicLogoPath(ctx context.Context, id uint64) (string, string, string, error) {
	configuredID, err := s.currentLogoFileID(ctx)
	if err != nil {
		return "", "", "", err
	}
	if configuredID == 0 || configuredID != id {
		return "", "", "", apperrors.ErrNotFound.WithMessage("站点 Logo 不存在")
	}

	attachment, err := s.files.FindAttachmentByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", "", "", apperrors.ErrNotFound.WithMessage("站点 Logo 不存在")
	}
	if err != nil {
		return "", "", "", err
	}
	if attachment.Status != "active" || !strings.HasPrefix(attachment.MimeType, "image/") {
		return "", "", "", apperrors.ErrNotFound.WithMessage("站点 Logo 不存在")
	}

	absolutePath, err := s.safeStoragePath(attachment.StoragePath)
	if err != nil {
		return "", "", "", err
	}
	if _, err := os.Stat(absolutePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", "", "", apperrors.ErrNotFound.WithMessage("站点 Logo 不存在")
		}
		return "", "", "", err
	}
	return absolutePath, attachment.MimeType, attachment.OriginalName, nil
}

func (s *SiteConfigService) currentLogoFileID(ctx context.Context) (uint64, error) {
	value, ok, err := s.configs.ValueByKey(ctx, siteLogoURLKey)
	if err != nil || !ok || value == nil {
		return 0, err
	}
	return fileIDFromAdminURL(strings.TrimSpace(*value))
}

func fileIDFromAdminURL(value string) (uint64, error) {
	match := adminFileURLPattern.FindStringSubmatch(value)
	if len(match) != 2 {
		return 0, nil
	}
	id, err := strconv.ParseUint(match[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("解析站点 Logo 文件 ID 失败: %w", err)
	}
	return id, nil
}

func publicLogoURL(value string) string {
	id, err := fileIDFromAdminURL(strings.TrimSpace(value))
	if err != nil || id == 0 {
		return value
	}
	return fmt.Sprintf("/api/site-logo/%d", id)
}

func (s *SiteConfigService) safeStoragePath(storagePath string) (string, error) {
	cleanPath := filepath.Clean(strings.TrimSpace(storagePath))
	if !domainfile.IsSafeRelativeStoragePath(cleanPath) {
		return "", apperrors.ErrNotFound.WithMessage("站点 Logo 不存在")
	}
	root, err := filepath.Abs(s.storage.LocalPath)
	if err != nil {
		return "", err
	}
	target, err := filepath.Abs(filepath.Join(root, cleanPath))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", apperrors.ErrNotFound.WithMessage("站点 Logo 不存在")
	}
	return target, nil
}

func defaultRealNameConfig() RealNameConfig {
	return RealNameConfig{
		RequiredForOrder:  true,
		AllowedProviders:  []string{},
		DefaultProvider:   "",
		ResubmitEnabled:   true,
		MaxSubmitAttempts: 3,
	}
}

func publicRealNameConfig(values map[string]string, secrets map[string]bool) RealNameConfig {
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
