package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	apperrors "github.com/AeolianCloud/pveCloud/server/internal/pkg/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/pkg/response"
)

func Recover(log *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID, _ := c.Get(RequestIDKey)
		log.Error("panic recovered", "request_id", requestID, "panic", recovered)
		response.Error(c, apperrors.ErrInternal)
	})
}
