package dto

/**
 * SiteConfigResponse 表示 Web 公开站点基础展示配置。
 */
type SiteConfigResponse struct {
	SiteName                           string `json:"site_name"`
	LogoURL                            string `json:"logo_url"`
	LoginCaptchaEnabled                bool   `json:"login_captcha_enabled"`
	RegisterCaptchaEnabled             bool   `json:"register_captcha_enabled"`
	PasswordResetRequestCaptchaEnabled bool   `json:"password_reset_request_captcha_enabled"`
	PasswordResetConfirmCaptchaEnabled bool   `json:"password_reset_confirm_captcha_enabled"`
}
