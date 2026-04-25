package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	apperrors "github.com/AeolianCloud/pveCloud/server/internal/pkg/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/pkg/response"
)

/**
 * Recover 捕获请求链路中的 panic 并转换为统一错误响应。
 *
 * @param log 结构化日志记录器
 * @return gin.HandlerFunc Gin 中间件
 */
func Recover(log *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID, _ := c.Get(RequestIDKey)
		log.Error("已恢复 panic", "request_id", requestID, "panic", recovered)
		response.Error(c, apperrors.ErrInternal)
	})
}
