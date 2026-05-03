package dto

/**
 * SiteConfigResponse 表示 Web 公开站点基础展示配置。
 */
type SiteConfigResponse struct {
	SiteName string `json:"site_name"`
	LogoURL  string `json:"logo_url"`
}
