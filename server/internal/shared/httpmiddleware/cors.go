package httpmiddleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/**
 * CORS 设置本地联调需要的跨域响应头。
 *
 * @return gin.HandlerFunc Gin 中间件
 */
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 当前阶段允许携带凭据的前后端联调；生产环境应在部署配置中收紧来源。
		c.Header("Access-Control-Allow-Origin", c.GetHeader("Origin"))
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Vary", "Origin")

		if c.Request.Method == http.MethodOptions {
			// 预检请求不进入业务路由，避免无意义的鉴权和日志噪声。
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
