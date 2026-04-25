package admin

import (
	"github.com/gin-gonic/gin"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/dto/admin"
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

	result, err := h.authService.Login(c.Request.Context(), req, c.ClientIP())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}
