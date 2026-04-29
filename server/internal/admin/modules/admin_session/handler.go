package adminsession

import (
	"strings"

	"github.com/gin-gonic/gin"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/middleware"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
)

/**
 * AdminSessionHandler 处理管理员会话接口。
 */
type AdminSessionHandler struct {
	adminSessionService *AdminSessionService
}

/**
 * NewAdminSessionHandler 创建管理员会话接口处理器。
 *
 * @param adminSessionService 管理员会话服务
 * @return *AdminSessionHandler 管理员会话接口处理器
 */
func NewAdminSessionHandler(adminSessionService *AdminSessionService) *AdminSessionHandler {
	return &AdminSessionHandler{adminSessionService: adminSessionService}
}

/**
 * List 分页查询管理员会话。
 *
 * @route GET /admin-api/admin-sessions
 * @response 200 {"code":0,"message":"成功","data":{"list":[],"total":0,"page":1,"per_page":15,"last_page":0}}
 * @auth admin jwt, permission admin-session:view
 */
func (h *AdminSessionHandler) List(c *gin.Context) {
	var query admindto.AdminSessionListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}

	currentSession, _ := middleware.CurrentAdminSession(c)
	result, err := h.adminSessionService.List(c.Request.Context(), query, currentSession.SessionID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

/**
 * Update 吊销指定管理员会话。
 *
 * @route PATCH /admin-api/admin-sessions/{session_id}
 * @request {"status":"revoked"}
 * @response 200 {"code":0,"message":"成功","data":{}}
 * @auth admin jwt, permission admin-session:revoke
 */
func (h *AdminSessionHandler) Update(c *gin.Context) {
	sessionID, ok := adminSessionPathID(c)
	if !ok {
		return
	}

	var req admindto.AdminSessionUpdateRequest
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
	currentSession, ok := middleware.CurrentAdminSession(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	if err := h.adminSessionService.Revoke(
		c.Request.Context(),
		operatorID,
		currentSession.SessionID,
		sessionID,
		c.ClientIP(),
		c.Request.UserAgent(),
	); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{})
}

func adminSessionPathID(c *gin.Context) (string, bool) {
	sessionID := strings.TrimSpace(c.Param("session_id"))
	if sessionID == "" || len(sessionID) > 64 {
		response.Error(c, apperrors.ErrValidation.WithMessage("会话 ID 格式错误"))
		return "", false
	}
	return sessionID, true
}
