package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/api/admin"
)

func RegisterAdminRoutes(group *gin.RouterGroup) {
	systemHandler := admin.NewSystemHandler()

	group.GET("/ping", systemHandler.Ping)
}
