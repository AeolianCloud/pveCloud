package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

/**
 * RequestIDKey 是 Gin 上下文中保存请求 ID 的键。
 */
const RequestIDKey = "request_id"

/**
 * RequestID 为每个请求读取或生成 X-Request-ID。
 *
 * @return gin.HandlerFunc Gin 中间件
 */
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// 客户端没有传入请求 ID 时由服务端生成，方便日志串联和问题排查。
			requestID = newRequestID()
		}

		c.Set(RequestIDKey, requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func newRequestID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(buf[:])
}
