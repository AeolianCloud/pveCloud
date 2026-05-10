package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/app/api"
	"github.com/AeolianCloud/pveCloud/server/internal/delivery/http/web/middleware"
)

/**
 * RegisterWebRoutes 注册用户端公开 API 路由。
 */
func RegisterWebRoutes(group *gin.RouterGroup, app *api.App) {
	routes := app.Routes.Web
	group.Use(middleware.RequestContext())

	group.GET("/site-config", routes.SiteConfig.Show)
	group.POST("/real-name/provider-callbacks/:provider", routes.RealName.ProviderCallback)
	group.GET("/server-catalog", routes.ProductCatalog.Show)
	group.GET("/auth/login-captcha", routes.Auth.LoginCaptcha)
	group.GET("/auth/register-captcha", routes.Auth.RegisterCaptcha)
	group.GET("/auth/password-reset-request-captcha", routes.Auth.PasswordResetRequestCaptcha)
	group.GET("/auth/password-reset-confirm-captcha", routes.Auth.PasswordResetConfirmCaptcha)
	group.POST("/auth/login", routes.Auth.Login)
	group.POST("/auth/register", routes.Auth.Register)
	group.POST("/auth/password-reset/request", routes.Auth.RequestPasswordReset)
	group.POST("/auth/password-reset/confirm", routes.Auth.ConfirmPasswordReset)

	protected := group.Group("")
	protected.Use(routes.AuthMiddleware)
	protected.GET("/auth/me", routes.Auth.Me)
	protected.POST("/auth/logout", routes.Auth.Logout)
	protected.POST("/auth/refresh", routes.Auth.Refresh)
	protected.PATCH("/user/profile", routes.UserProfile.UpdateProfile)
	protected.POST("/user/password", routes.UserProfile.ChangePassword)
	protected.GET("/user/real-name", routes.RealName.Status)
	protected.POST("/user/real-name", routes.RealName.Submit)
	protected.POST("/user/real-name/sync", routes.RealName.Sync)
	protected.POST("/orders", routes.Order.Create)
	protected.GET("/orders", routes.Order.List)
	protected.GET("/orders/:order_no", routes.Order.Detail)
	protected.POST("/orders/:order_no/cancel", routes.Order.Cancel)
}
