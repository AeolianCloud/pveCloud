package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/modules/audit"
	httpmiddleware "github.com/AeolianCloud/pveCloud/server/internal/shared/httpmiddleware"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
)

func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID, _ := c.Get(httpmiddleware.RequestIDKey)
		ctx := audit.WithRequestContext(c.Request.Context(), audit.RequestContext{
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
