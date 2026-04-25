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
 * AuthHandler 处理管理端认证接口。
 */
type AuthHandler struct {
	authService *services.AdminAuthService
}

/**
 * NewAuthHandler 创建管理端认证接口处理器。
 *
 * @param authService 管理端认证服务
 * @return *AuthHandler 管理端认证接口处理器
 */
func NewAuthHandler(authService *services.AdminAuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

/**
 * Login 处理管理员登录。
 *
 * @route POST /admin-api/auth/login
 * @request {"username":"admin","password":"password"}
 * @response 200 {"code":0,"message":"成功","data":{"access_token":"...","token_type":"Bearer","expires_in":28800,"admin":{"id":1,"username":"admin","display_name":"超级管理员","status":"active"},"role_ids":[1],"permission_codes":["dashboard:view"]}}
 * @auth 无需登录
 */
func (h *AuthHandler) Login(c *gin.Context) {
	var req admindto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}

	result, err := h.authService.Login(c.Request.Context(), req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

/**
 * Me 返回当前管理员认证态。
 *
 * @route GET /admin-api/auth/me
 * @response 200 {"code":0,"message":"成功","data":{"admin":{"id":1,"username":"admin","display_name":"超级管理员","status":"active"},"role_ids":[1],"permission_codes":["dashboard:view"],"menus":[],"session":{"session_id":"adm_xxx","issued_at":"2026-04-26T00:00:00Z","expires_at":"2026-04-26T08:00:00Z"}}}
 * @auth admin jwt
 */
func (h *AuthHandler) Me(c *gin.Context) {
	admin, ok := middleware.CurrentAdmin(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	session, ok := middleware.CurrentAdminSession(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}

	response.Success(c, h.authService.Me(
		admin,
		middleware.CurrentAdminRoleIDs(c),
		middleware.CurrentAdminPermissionCodes(c),
		session,
	))
}

/**
 * Logout 吊销当前管理员会话。
 *
 * @route POST /admin-api/auth/logout
 * @response 200 {"code":0,"message":"成功","data":{}}
 * @auth admin jwt
 */
func (h *AuthHandler) Logout(c *gin.Context) {
	adminID, ok := middleware.CurrentAdminID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	session, ok := middleware.CurrentAdminSession(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}

	if err := h.authService.Logout(c.Request.Context(), adminID, session.SessionID, c.ClientIP(), c.Request.UserAgent()); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{})
}

/**
 * Refresh 轮换当前管理员 token。
 *
 * @route POST /admin-api/auth/refresh
 * @response 200 {"code":0,"message":"成功","data":{"access_token":"...","token_type":"Bearer","expires_in":28800,"admin":{"id":1,"username":"admin","display_name":"超级管理员","status":"active"},"role_ids":[1],"permission_codes":["dashboard:view"],"session":{"session_id":"adm_xxx","issued_at":"2026-04-26T00:00:00Z","expires_at":"2026-04-26T08:00:00Z"}}}
 * @auth admin jwt
 */
func (h *AuthHandler) Refresh(c *gin.Context) {
	adminID, ok := middleware.CurrentAdminID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	session, ok := middleware.CurrentAdminSession(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}

	result, err := h.authService.Refresh(c.Request.Context(), adminID, session.SessionID, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}
