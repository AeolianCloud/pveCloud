package siteconfig

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/web/dto"
)

const (
	defaultSiteName = "pveCloud"
	siteNameKey     = "site.name"
	siteLogoURLKey  = "site.logo_url"
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
		Where("config_key IN ? AND is_secret = 0", []string{siteNameKey, siteLogoURLKey}).
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
		}
	}

	return result, nil
}
