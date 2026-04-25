package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/api/admin"
)

/**
 * RegisterAdminRoutes 注册管理端 API 路由。
 *
 * @param group 管理端路由分组
 */
func RegisterAdminRoutes(group *gin.RouterGroup) {
	systemHandler := admin.NewSystemHandler()

	group.GET("/ping", systemHandler.Ping)
}
