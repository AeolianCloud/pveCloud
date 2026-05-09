package siteconfig

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/usecase/web/siteconfig"
)

type Handler struct {
	service *siteconfig.SiteConfigService
}

func NewHandler(service *siteconfig.SiteConfigService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Show(c *gin.Context) {
	result, err := h.service.Show(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, siteConfigResponse(result))
}

func siteConfigResponse(config siteconfig.SiteConfig) webdto.SiteConfigResponse {
	return webdto.SiteConfigResponse{
		SiteName:                           config.SiteName,
		LogoURL:                            config.LogoURL,
		LoginCaptchaEnabled:                config.LoginCaptchaEnabled,
		RegisterCaptchaEnabled:             config.RegisterCaptchaEnabled,
		PasswordResetRequestCaptchaEnabled: config.PasswordResetRequestCaptchaEnabled,
		PasswordResetConfirmCaptchaEnabled: config.PasswordResetConfirmCaptchaEnabled,
		RealName: webdto.RealNameConfig{
			Enabled:           config.RealName.Enabled,
			RequiredForOrder:  config.RealName.RequiredForOrder,
			AllowedProviders:  config.RealName.AllowedProviders,
			DefaultProvider:   config.RealName.DefaultProvider,
			ResubmitEnabled:   config.RealName.ResubmitEnabled,
			MaxSubmitAttempts: config.RealName.MaxSubmitAttempts,
			ReviewNotice:      config.RealName.ReviewNotice,
		},
	}
}
