package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"pvecloud/backend/internal/middleware"
	"pvecloud/backend/internal/service"
	"pvecloud/backend/pkg/response"
)

// AuthHandler 处理注册、登录、刷新、登出接口。
type AuthHandler struct {
	userService *service.UserService
	limiter     *middleware.LoginRateLimiter
}

// NewAuthHandler 创建认证处理器。
func NewAuthHandler(userService *service.UserService, limiter *middleware.LoginRateLimiter) *AuthHandler {
	return &AuthHandler{userService: userService, limiter: limiter}
}

// RegisterRoutes 注册认证相关路由。
func (h *AuthHandler) RegisterRoutes(pub *gin.RouterGroup, user *gin.RouterGroup) {
	pub.POST("/register", h.Register)
	pub.POST("/verify-email", h.VerifyEmail)
	pub.POST("/login", h.limiter.Middleware(), h.Login)
	user.POST("/refresh-token", h.RefreshToken)
	user.POST("/logout", h.Logout)
}

// Register 创建用户账号。
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if err := h.userService.Register(c.Request.Context(), req.Email, req.Password); err != nil {
		response.Error(c, http.StatusBadRequest, 40002, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "注册成功，请前往邮箱完成验证"})
}

// VerifyEmail 消费验证 token，完成邮箱验证。
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Token == "" {
		response.Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if err := h.userService.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		response.Error(c, http.StatusBadRequest, 40004, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "邮箱验证成功"})
}

// Login 校验账号密码并返回 token 对。
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}

	access, refresh, user, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
	ip := c.ClientIP()
	if err != nil {
		h.limiter.RecordFailure(ip)
		response.Error(c, http.StatusUnauthorized, 40111, err.Error())
		return
	}
	h.limiter.RecordSuccess(ip)
	response.OK(c, gin.H{"access_token": access, "refresh_token": refresh, "user": gin.H{"id": user.ID, "email": user.Email, "role": user.Role}})
}

// RefreshToken 基于 refresh_token 签发新 access_token。
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	access, err := h.userService.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, 40112, "refresh token 无效或已过期")
		return
	}
	response.OK(c, gin.H{"access_token": access})
}

// Logout 将当前 access token 加入黑名单，并撤销传入的 refresh token。
func (h *AuthHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.ShouldBindJSON(&req)

	authHeader := c.GetHeader("Authorization")
	accessToken := strings.TrimPrefix(authHeader, "Bearer ")
	if err := h.userService.Logout(c.Request.Context(), accessToken, req.RefreshToken); err != nil {
		response.Error(c, http.StatusBadRequest, 40003, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "已退出登录"})
}
