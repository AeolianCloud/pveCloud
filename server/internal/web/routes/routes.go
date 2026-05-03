package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/platform/bootstrap"
	"github.com/AeolianCloud/pveCloud/server/internal/web/middleware"
	webauth "github.com/AeolianCloud/pveCloud/server/internal/web/modules/auth"
	siteconfig "github.com/AeolianCloud/pveCloud/server/internal/web/modules/site_config"
)

/**
 * RegisterWebRoutes 注册用户端公开 API 路由。
 */
func RegisterWebRoutes(group *gin.RouterGroup, app *bootstrap.App) {
	siteConfigService := siteconfig.NewSiteConfigService(app.DB)
	siteConfigHandler := siteconfig.NewSiteConfigHandler(siteConfigService)
	authService := webauth.NewUserAuthService(app.DB, app.Config.JWT)
	authHandler := webauth.NewUserAuthHandler(authService)

	group.GET("/site-config", siteConfigHandler.Show)
	group.POST("/auth/login", authHandler.Login)

	protected := group.Group("")
	protected.Use(middleware.UserAuth(app.Config.JWT, app.DB))
	protected.GET("/auth/me", authHandler.Me)
	protected.POST("/auth/logout", authHandler.Logout)
	protected.POST("/auth/refresh", authHandler.Refresh)
}
