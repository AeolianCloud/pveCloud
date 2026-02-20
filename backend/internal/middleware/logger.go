package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		fields := []zap.Field{
			zap.Int("状态码", c.Writer.Status()),
			zap.String("方法", c.Request.Method),
			zap.String("路径", path),
			zap.String("查询", query),
			zap.String("IP", c.ClientIP()),
			zap.Duration("耗时", time.Since(start)),
			zap.String("UA", c.Request.UserAgent()),
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				log.Error("请求处理出错", append(fields, zap.Error(e.Err))...)
			}
			return
		}

		status := c.Writer.Status()
		switch {
		case status >= 500:
			log.Error("服务器错误", fields...)
		case status >= 400:
			log.Warn("客户端错误", fields...)
		default:
			log.Info("请求完成", fields...)
		}
	}
}
