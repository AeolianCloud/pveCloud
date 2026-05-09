package middleware

import (
	"github.com/gin-gonic/gin"

	httpmiddleware "github.com/AeolianCloud/pveCloud/server/internal/shared/httpmiddleware"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/requestcontext"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
)

/**
 * AdminAuditContext 把管理端请求上下文写入 request context，供审计写入统一使用。
 *
 * @return gin.HandlerFunc Gin 中间件
 */
func AdminAuditContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID, _ := c.Get(httpmiddleware.RequestIDKey)
		ctx := requestcontext.WithRequestContext(c.Request.Context(), requestcontext.RequestContext{
			RequestID:     stringValue(requestID),
			RequestMethod: c.Request.Method,
			RequestPath:   c.Request.URL.Path,
			IP:            c.ClientIP(),
			UserAgent:     textutil.TrimTo(c.Request.UserAgent(), 500),
		})
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}
