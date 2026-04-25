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
	RegisterAdminRoutes(router.Group("/admin-api"))

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

		if err := sqlDB.PingContext(ctx); err != nil {
			response.Error(c, apperrors.ErrInternal.WithMessage("数据库连接异常"))
			return
		}

		c.JSON(http.StatusOK, response.Envelope{
			Code:    0,
			Message: "ok",
			Data: gin.H{
				"app":      app.Config.App.Name,
				"env":      app.Config.App.Env,
				"database": "ok",
				"time":     time.Now().Format(time.RFC3339),
			},
		})
	}
}
