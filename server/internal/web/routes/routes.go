package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/web/modules/system"
)

/**
 * RegisterWebRoutes 注册用户端 API 路由。
 *
 * @param group 用户端路由分组
 */
func RegisterWebRoutes(group *gin.RouterGroup) {
	systemHandler := system.NewSystemHandler()

	group.GET("/ping", systemHandler.Ping)
}
