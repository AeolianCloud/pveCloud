package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/app/api"
	adminroutes "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/routes"
	webroutes "github.com/AeolianCloud/pveCloud/server/internal/delivery/http/web/routes"
	mysqllogs "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/logs"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	httpmiddleware "github.com/AeolianCloud/pveCloud/server/internal/shared/httpmiddleware"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
)

/**
 * NewRouter 创建 API 进程的 Gin 路由树。
 *
 * @param app API 应用依赖容器
 * @return *gin.Engine Gin 路由引擎
 */
func NewRouter(app *api.App) *gin.Engine {
	if app.Config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(
		httpmiddleware.RequestID(),
		httpmiddleware.AccessLog(app.Logger, func(c *gin.Context, log httpmiddleware.BackendRuntimeLogInput) {
			if app.Logs == nil {
				return
			}
			_ = app.LogRecorder.BackendRuntime(c.Request.Context(), nil, mysqllogs.BackendRuntimeLog{
				Level:         log.Level,
				Category:      log.Category,
				RequestID:     stringPtr(log.RequestID),
				RequestMethod: stringPtr(log.RequestMethod),
				RequestPath:   stringPtr(log.RequestPath),
				Status:        intPtr(log.Status),
				LatencyMS:     int64Ptr(log.LatencyMS),
				ClientIP:      stringPtr(log.ClientIP),
				Message:       log.Message,
			})
		}),
		httpmiddleware.Recover(app.Logger, func(c *gin.Context, message string) {
			if app.Logs == nil {
				return
			}
			_ = app.LogRecorder.BackendRuntime(c.Request.Context(), nil, mysqllogs.BackendRuntimeLog{
				Level:         "error",
				Category:      "panic",
				RequestID:     requestIDFromContext(c),
				RequestMethod: stringPtr(c.Request.Method),
				RequestPath:   stringPtr(c.Request.URL.Path),
				ClientIP:      stringPtr(c.ClientIP()),
				Message:       message,
			})
		}),
		httpmiddleware.CORS(),
	)

	router.GET("/healthz", healthz(app))
	webroutes.RegisterWebRoutes(router.Group("/api"), app)
	adminroutes.RegisterAdminRoutes(router.Group("/admin-api"), app)

	return router
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func intPtr(value int) *int {
	return &value
}

func int64Ptr(value int64) *int64 {
	return &value
}

func requestIDFromContext(c *gin.Context) *string {
	value, _ := c.Get(httpmiddleware.RequestIDKey)
	text, _ := value.(string)
	if text == "" {
		return nil
	}
	return &text
}

func healthz(app *api.App) gin.HandlerFunc {
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
			Data:    gin.H{"status": "ok"},
		})
	}
}
