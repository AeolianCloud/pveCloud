package httpmiddleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

/**
 * AccessLog 记录每个 HTTP 请求的基础访问日志。
 *
 * @param log 结构化日志记录器
 * @return gin.HandlerFunc Gin 中间件
 */
func AccessLog(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		// 请求结束后再记录日志，才能拿到最终状态码和完整耗时。
		requestID, _ := c.Get(RequestIDKey)
		log.Info(
			"HTTP 请求",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
