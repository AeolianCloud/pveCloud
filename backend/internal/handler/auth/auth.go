// internal/handler/auth/auth.go
// 认证相关 HTTP 处理器：登录、获取当前用户信息。
package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	"pvecloud/backend/internal/security"
	svcauth "pvecloud/backend/internal/service/auth"
	"pvecloud/backend/pkg/response"
	"pvecloud/backend/pkg/response/errcode"
)

// Handler 认证处理器。
type Handler struct {
	svc   *svcauth.Service
	guard *security.LoginGuard
}

// New 创建认证处理器。
func New(svc *svcauth.Service, guard *security.LoginGuard) *Handler {
	return &Handler{svc: svc, guard: guard}
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
		// 统一规范：参数错误使用业务错误码 + HTTP 200 返回，避免前端出现多套错误处理逻辑
		response.FailMsg(c, errcode.InvalidParams, "用户名和密码不能为空")
		return
	}

	// 登录限流 + 防爆破（失败次数锁定）
	if h.guard != nil {
		if err := h.guard.PreCheck(req.Username, c.ClientIP()); err != nil {
		// 对外只暴露“请求过于频繁/稍后再试”，避免泄露过多风控细节
		response.FailMsg(c, errcode.TooManyReqs, err.Error())
		return
		}
	}

	opt := svcauth.LoginOption{
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	result, err := h.svc.Login(req.Username, req.Password, opt)
	if err != nil {
		// 失败计数：对“用户不存在/密码错误/账号禁用”都算失败一次（防止枚举用户）
		if h.guard != nil {
			switch {
			case errors.Is(err, svcauth.ErrUserNotFound),
				errors.Is(err, svcauth.ErrPasswordWrong),
				errors.Is(err, svcauth.ErrUserDisabled):
				_ = h.guard.RecordFailure(req.Username, c.ClientIP())
			}
		}

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

	// 登录成功：清理失败计数与锁定
	if h.guard != nil {
		_ = h.guard.RecordSuccess(req.Username, c.ClientIP())
	}
	response.Success(c, result)
}

// refreshReq 刷新 Token 请求体。
type refreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Refresh 刷新 Access Token，并旋转 Refresh Token。
// POST /api/v1/auth/refresh
func (h *Handler) Refresh(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, errcode.InvalidParams, "refresh_token 不能为空")
		return
	}

	opt := svcauth.LoginOption{
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	result, err := h.svc.Refresh(req.RefreshToken, opt)
	if err != nil {
		// Refresh 失败统一视为登录过期/无效，让前端回到登录页重新登录
		response.Unauthorized(c, errcode.TokenExpired.Msg())
		return
	}

	response.Success(c, result)
}

// Logout 退出登录：撤销当前会话，使 token 立即失效。
// POST /api/v1/auth/logout
func (h *Handler) Logout(c *gin.Context) {
	sessionIDAny, ok := c.Get("session_id")
	if !ok {
		response.Unauthorized(c, errcode.Unauthorized.Msg())
		return
	}

	sessionID, _ := sessionIDAny.(uint)
	if err := h.svc.Logout(sessionID); err != nil {
		// 会话不存在/已撤销：也按成功处理，保持幂等（前端只需要清本地 token）
		response.Success(c, nil)
		return
	}

	response.Success(c, nil)
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
