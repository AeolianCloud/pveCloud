// internal/handler/auth/auth.go
// 认证相关 HTTP 处理器：登录、获取当前用户信息。
package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	svcauth "pvecloud/backend/internal/service/auth"
	"pvecloud/backend/pkg/response"
	"pvecloud/backend/pkg/response/errcode"
)

// Handler 认证处理器。
type Handler struct {
	svc *svcauth.Service
}

// New 创建认证处理器。
func New(svc *svcauth.Service) *Handler {
	return &Handler{svc: svc}
}

// loginReq 登录请求体。
type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 处理登录请求。
// POST /api/v1/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "用户名和密码不能为空")
		return
	}

	opt := svcauth.LoginOption{
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	result, err := h.svc.Login(req.Username, req.Password, opt)
	if err != nil {
		switch {
		case errors.Is(err, svcauth.ErrUserNotFound):
			response.Fail(c, errcode.UserNotFound)
		case errors.Is(err, svcauth.ErrPasswordWrong):
			response.Fail(c, errcode.PasswordWrong)
		case errors.Is(err, svcauth.ErrUserDisabled):
			response.Fail(c, errcode.UserDisabled)
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, result)
}

// Profile 获取当前登录用户信息（含角色和权限）。
// GET /api/v1/profile
func (h *Handler) Profile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := h.svc.GetUserByID(userID.(uint))
	if err != nil {
		response.Fail(c, errcode.UserNotFound)
		return
	}

	response.Success(c, user)
}
