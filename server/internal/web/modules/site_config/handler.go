package siteconfig

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
)

/**
 * SiteConfigHandler 处理 Web 公开站点配置接口。
 */
type SiteConfigHandler struct {
	service *SiteConfigService
}

/**
 * NewSiteConfigHandler 创建 Web 公开站点配置接口处理器。
 */
func NewSiteConfigHandler(service *SiteConfigService) *SiteConfigHandler {
	return &SiteConfigHandler{service: service}
}

/**
 * Show 读取 Web 公开站点基础展示配置。
 *
 * @route GET /api/site-config
 */
func (h *SiteConfigHandler) Show(c *gin.Context) {
	result, err := h.service.Show(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}
