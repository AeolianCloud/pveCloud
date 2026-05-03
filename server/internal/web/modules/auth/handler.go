package auth

import (
	"github.com/gin-gonic/gin"

	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/web/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/web/middleware"
)

/**
 * UserAuthHandler 处理用户端认证接口。
 */
type UserAuthHandler struct {
	service *UserAuthService
}

/**
 * NewUserAuthHandler 创建用户端认证接口处理器。
 */
func NewUserAuthHandler(service *UserAuthService) *UserAuthHandler {
	return &UserAuthHandler{service: service}
}

/**
 * Login 处理用户端登录。
 */
func (h *UserAuthHandler) Login(c *gin.Context) {
	var req webdto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}
	result, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

/**
 * Me 返回当前用户端认证态。
 */
func (h *UserAuthHandler) Me(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	session, ok := middleware.CurrentUserSession(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	response.Success(c, h.service.Me(user, session))
}

/**
 * Logout 吊销当前用户端会话。
 */
func (h *UserAuthHandler) Logout(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	session, ok := middleware.CurrentUserSession(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	if err := h.service.Logout(c.Request.Context(), userID, session.SessionID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{})
}

/**
 * Refresh 轮换当前用户端 token。
 */
func (h *UserAuthHandler) Refresh(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	session, ok := middleware.CurrentUserSession(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	result, err := h.service.Refresh(c.Request.Context(), userID, session.SessionID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}
