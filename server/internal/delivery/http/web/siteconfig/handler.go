package siteconfig

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
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

func (h *Handler) Logo(c *gin.Context) {
	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, apperrors.ErrValidation.WithMessage("Logo ID 格式错误"))
		return
	}

	path, mimeType, filename, err := h.service.PublicLogoPath(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.Header("Content-Type", mimeType)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename*=UTF-8''%s", urlEncodeFilename(filename)))
	c.File(path)
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

func urlEncodeFilename(value string) string {
	return strings.ReplaceAll(url.QueryEscape(value), "+", "%20")
}
