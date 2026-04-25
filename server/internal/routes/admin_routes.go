package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/api/admin"
	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap"
	"github.com/AeolianCloud/pveCloud/server/internal/services"
)

/**
 * RegisterAdminRoutes 注册管理端 API 路由。
 *
 * @param group 管理端路由分组
 * @param app API 应用依赖容器
 */
func RegisterAdminRoutes(group *gin.RouterGroup, app *bootstrap.App) {
	systemHandler := admin.NewSystemHandler()
	authService := services.NewAdminAuthService(app.DB, app.Config.JWT)
	authHandler := admin.NewAuthHandler(authService)

	group.GET("/ping", systemHandler.Ping)
	group.POST("/auth/login", authHandler.Login)
}
