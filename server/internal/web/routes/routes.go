package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/platform/bootstrap"
	"github.com/AeolianCloud/pveCloud/server/internal/web/middleware"
	webauth "github.com/AeolianCloud/pveCloud/server/internal/web/modules/auth"
	productcatalog "github.com/AeolianCloud/pveCloud/server/internal/web/modules/product_catalog"
	realname "github.com/AeolianCloud/pveCloud/server/internal/web/modules/real_name"
	siteconfig "github.com/AeolianCloud/pveCloud/server/internal/web/modules/site_config"
	userprofile "github.com/AeolianCloud/pveCloud/server/internal/web/modules/user_profile"
)

/**
 * RegisterWebRoutes 注册用户端公开 API 路由。
 */
func RegisterWebRoutes(group *gin.RouterGroup, app *bootstrap.App) {
	siteConfigService := siteconfig.NewSiteConfigService(app.DB)
	siteConfigHandler := siteconfig.NewSiteConfigHandler(siteConfigService)
	authService := webauth.NewUserAuthService(app.DB, app.Redis, app.Config.JWT, app.Config.Mail)
	authHandler := webauth.NewUserAuthHandler(authService)
	userProfileService := userprofile.NewUserProfileService(app.DB)
	userProfileHandler := userprofile.NewUserProfileHandler(userProfileService)
	productCatalogService := productcatalog.NewProductCatalogService(app.DB)
	productCatalogHandler := productcatalog.NewProductCatalogHandler(productCatalogService)
	realNameService := realname.NewRealNameService(app.DB, app.Config.Storage)
	realNameHandler := realname.NewRealNameHandler(realNameService)

	group.GET("/site-config", siteConfigHandler.Show)
	group.GET("/server-catalog", productCatalogHandler.Show)
	group.GET("/auth/login-captcha", authHandler.LoginCaptcha)
	group.GET("/auth/register-captcha", authHandler.RegisterCaptcha)
	group.GET("/auth/password-reset-request-captcha", authHandler.PasswordResetRequestCaptcha)
	group.GET("/auth/password-reset-confirm-captcha", authHandler.PasswordResetConfirmCaptcha)
	group.POST("/auth/login", authHandler.Login)
	group.POST("/auth/register", authHandler.Register)
	group.POST("/auth/password-reset/request", authHandler.RequestPasswordReset)
	group.POST("/auth/password-reset/confirm", authHandler.ConfirmPasswordReset)

	protected := group.Group("")
	protected.Use(middleware.UserAuth(app.Config.JWT, app.DB))
	protected.GET("/auth/me", authHandler.Me)
	protected.POST("/auth/logout", authHandler.Logout)
	protected.POST("/auth/refresh", authHandler.Refresh)
	protected.PATCH("/user/profile", userProfileHandler.UpdateProfile)
	protected.POST("/user/password", userProfileHandler.ChangePassword)
	protected.POST("/user/real-name/files", realNameHandler.UploadFile)
	protected.GET("/user/real-name", realNameHandler.Status)
	protected.POST("/user/real-name", realNameHandler.Submit)
}
