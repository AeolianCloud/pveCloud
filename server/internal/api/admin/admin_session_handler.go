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
 * AdminSessionHandler 处理管理端登录会话接口。
 */
type AdminSessionHandler struct {
	adminSessionService *services.AdminSessionService
}

/**
 * NewAdminSessionHandler 创建管理端登录会话接口处理器。
 *
 * @param adminSessionService 管理端登录会话服务
 * @return *AdminSessionHandler 管理端登录会话接口处理器
 */
func NewAdminSessionHandler(adminSessionService *services.AdminSessionService) *AdminSessionHandler {
	return &AdminSessionHandler{adminSessionService: adminSessionService}
}

/**
 * Sessions 分页查询管理端登录会话。
 *
 * @route GET /admin-api/admin-sessions
 * @response 200 {"code":0,"message":"成功","data":{"list":[],"total":0,"page":1,"per_page":15,"last_page":0}}
 * @auth admin jwt, permission admin:manage
 */
func (h *AdminSessionHandler) Sessions(c *gin.Context) {
	var query admindto.AdminSessionListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}

	result, err := h.adminSessionService.Sessions(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

/**
 * Revoke 吊销指定活跃会话。
 *
 * @route POST /admin-api/admin-sessions/{id}/revoke
 * @response 200 {"code":0,"message":"成功","data":{}}
 * @auth admin jwt, permission admin:manage
 */
func (h *AdminSessionHandler) Revoke(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	operatorID, ok := middleware.CurrentAdminID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	session, ok := middleware.CurrentAdminSession(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	if err := h.adminSessionService.Revoke(c.Request.Context(), operatorID, session.SessionID, id, c.ClientIP(), c.Request.UserAgent()); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{})
}
