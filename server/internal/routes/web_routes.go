package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/api/web"
)

func RegisterWebRoutes(group *gin.RouterGroup) {
	systemHandler := web.NewSystemHandler()

	group.GET("/ping", systemHandler.Ping)
}
