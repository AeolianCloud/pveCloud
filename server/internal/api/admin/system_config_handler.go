package admin

import (
	"github.com/gin-gonic/gin"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/dto/admin"
	"github.com/AeolianCloud/pveCloud/server/internal/middleware"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/pkg/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/pkg/response"
	"github.com/AeolianCloud/pveCloud/server/internal/pkg/validator"
	"github.com/AeolianCloud/pveCloud/server/internal/services"
)

/**
 * SystemConfigHandler 处理系统配置接口。
 */
type SystemConfigHandler struct {
	systemConfigService *services.SystemConfigService
}

/**
 * NewSystemConfigHandler 创建系统配置接口处理器。
 *
 * @param systemConfigService 系统配置服务
 * @return *SystemConfigHandler 系统配置接口处理器
 */
func NewSystemConfigHandler(systemConfigService *services.SystemConfigService) *SystemConfigHandler {
	return &SystemConfigHandler{systemConfigService: systemConfigService}
}

/**
 * Configs 按分组查询系统配置。
 *
 * @route GET /admin-api/system-configs
 * @response 200 {"code":0,"message":"成功","data":[]}
 * @auth admin jwt, permission system:update
 */
func (h *SystemConfigHandler) Configs(c *gin.Context) {
	var query admindto.SystemConfigListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}
	result, err := h.systemConfigService.Configs(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

/**
 * Update 更新系统配置。
 *
 * @route PATCH /admin-api/system-configs/{id}
 * @request {"config_value":"pveCloud"}
 * @response 200 {"code":0,"message":"成功","data":{"id":1}}
 * @auth admin jwt, permission system:update
 */
func (h *SystemConfigHandler) Update(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	var req admindto.SystemConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}
	operatorID, ok := middleware.CurrentAdminID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	result, err := h.systemConfigService.Update(c.Request.Context(), operatorID, id, req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}
