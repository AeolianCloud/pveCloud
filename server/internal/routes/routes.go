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
	if app.Config.OpenAPI.Enabled && app.OpenAPISpec != nil {
		// OpenAPI 规范文件已经在启动阶段校验过，这里只负责只读输出。
		router.GET("/openapi.yaml", openAPI(app))
	}
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

		c.JSON(http.StatusOK, response.Envelope{
			Code:    0,
			Message: "成功",
			Data: gin.H{
				"app":      app.Config.App.Name,
				"env":      app.Config.App.Env,
				"database": "正常",
				"time":     time.Now().Format(time.RFC3339),
			},
		})
	}
}

func openAPI(app *bootstrap.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := app.OpenAPISpec.Read()
		if err != nil {
			response.Error(c, apperrors.ErrInternal.WithMessage("OpenAPI 规范文件读取失败"))
			return
		}

		c.Data(http.StatusOK, "application/yaml; charset=utf-8", data)
	}
}
