package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"pvecloud/backend/internal/middleware"
)

func New(db *gorm.DB, log *zap.Logger, jwtSecret string) *gin.Engine {
	r := gin.New()

	// 全局中间件
	r.Use(middleware.Recovery(log))
	r.Use(middleware.CORS())
	r.Use(middleware.Logger(log))

	// 健康检查（不需要鉴权）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 公开路由
	auth := r.Group("/api/v1/auth")
	{
		// 登录、注册等接口在此注册
		_ = auth
	}

	// JWT 保护路由
	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(jwtSecret))
	{
		// 业务接口在此注册
		_ = api
	}

	return r
}
