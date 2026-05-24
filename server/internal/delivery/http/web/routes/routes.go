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
	group.GET("/site-logo/:id", routes.SiteConfig.Logo)
	group.POST("/real-name/provider-callbacks/:provider", routes.RealName.ProviderCallback)
	group.POST("/payment-callbacks/:provider", routes.Payment.Callback)
	group.GET("/server-catalog", routes.ProductCatalog.Show)
	group.GET("/auth/login-captcha", routes.Auth.LoginCaptcha)
	group.GET("/auth/register-captcha", routes.Auth.RegisterCaptcha)
	group.GET("/auth/password-reset-request-captcha", routes.Auth.PasswordResetRequestCaptcha)
	group.GET("/auth/password-reset-confirm-captcha", routes.Auth.PasswordResetConfirmCaptcha)
	group.POST("/auth/login", routes.Auth.Login)
	group.POST("/auth/register", routes.Auth.Register)
	group.POST("/auth/password-reset/request", routes.Auth.RequestPasswordReset)
	group.POST("/auth/password-reset/confirm", routes.Auth.ConfirmPasswordReset)
	group.POST("/client-logs/errors", routes.ClientLogs.Create)

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
	protected.POST("/orders/:order_no/payments", routes.Payment.Create)
	protected.GET("/payments/:payment_no", routes.Payment.Show)
	protected.GET("/wallet", routes.Wallet.Show)
	protected.GET("/wallet/ledger", routes.Wallet.Ledger)
	protected.POST("/wallet/recharges", routes.Wallet.CreateRecharge)
	protected.GET("/wallet/recharges/:recharge_no", routes.Wallet.Recharge)
	protected.GET("/instances", routes.Instance.List)
	protected.GET("/instances/:instance_no", routes.Instance.Detail)
	protected.POST("/instances/:instance_no/start", routes.Instance.Start)
	protected.POST("/instances/:instance_no/stop", routes.Instance.Stop)
	protected.POST("/instances/:instance_no/renewal-orders", routes.Instance.CreateRenewalOrder)
	protected.GET("/tickets", routes.Ticket.List)
	protected.POST("/tickets", routes.Ticket.Create)
	protected.GET("/tickets/:ticket_no", routes.Ticket.Detail)
	protected.POST("/tickets/:ticket_no/messages", routes.Ticket.Reply)
	protected.POST("/tickets/:ticket_no/close", routes.Ticket.Close)
	protected.GET("/tickets/:ticket_no/attachments/:file_id/download", routes.Ticket.Download)
}
