package routes

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap"
	"github.com/AeolianCloud/pveCloud/server/internal/middleware"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/pkg/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/pkg/response"
)

/**
 * NewRouter 创建 API 进程的 Gin 路由树。
 *
 * @param app API 应用依赖容器
 * @return *gin.Engine Gin 路由引擎
 */
func NewRouter(app *bootstrap.App) *gin.Engine {
	if app.Config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(
		middleware.RequestID(),
		middleware.AccessLog(app.Logger),
		middleware.Recover(app.Logger),
		middleware.CORS(),
	)

	router.GET("/healthz", healthz(app))
	RegisterWebRoutes(router.Group("/api"))
	RegisterAdminRoutes(router.Group("/admin-api"), app)

	return router
}

func healthz(app *bootstrap.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		sqlDB, err := app.DB.DB()
		if err != nil {
			response.Error(c, apperrors.ErrInternal.WithMessage("数据库连接异常"))
			return
		}

		// 健康检查只做轻量 Ping，不读取业务表，避免探活接口影响正常业务负载。
		if err := sqlDB.PingContext(ctx); err != nil {
			response.Error(c, apperrors.ErrInternal.WithMessage("数据库连接异常"))
			return
		}
		if err := app.Redis.Client().Ping(ctx).Err(); err != nil {
			response.Error(c, apperrors.ErrInternal.WithMessage("Redis 连接异常"))
			return
		}

		c.JSON(http.StatusOK, response.Envelope{
			Code:    0,
			Message: "成功",
			Data: gin.H{
				"app":      app.Config.App.Name,
				"env":      app.Config.App.Env,
				"database": "正常",
				"redis":    "正常",
				"time":     time.Now().Format(time.RFC3339),
			},
		})
	}
}
