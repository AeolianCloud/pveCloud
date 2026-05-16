package httpmiddleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

type BackendRuntimeRecorder func(ctx *gin.Context, log BackendRuntimeLogInput)

type BackendRuntimeLogInput struct {
	Level         string
	Category      string
	RequestID     string
	RequestMethod string
	RequestPath   string
	Status        int
	LatencyMS     int64
	ClientIP      string
	Message       string
}

func AccessLog(log *slog.Logger, recorder BackendRuntimeRecorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		requestID, _ := c.Get(RequestIDKey)
		elapsed := time.Since(start)
		log.Info(
			"HTTP 请求",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", elapsed.Milliseconds(),
			"client_ip", c.ClientIP(),
		)

		if recorder != nil {
			recorder(c, BackendRuntimeLogInput{
				Level:         levelByStatus(c.Writer.Status()),
				Category:      "access",
				RequestID:     stringValue(requestID),
				RequestMethod: c.Request.Method,
				RequestPath:   c.Request.URL.Path,
				Status:        c.Writer.Status(),
				LatencyMS:     elapsed.Milliseconds(),
				ClientIP:      c.ClientIP(),
				Message:       "HTTP 请求",
			})
		}
	}
}

func levelByStatus(status int) string {
	if status >= 500 {
		return "error"
	}
	if status >= 400 {
		return "warn"
	}
	return "info"
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}
