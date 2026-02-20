package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"pvecloud/backend/pkg/response/errcode"
)

func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("系统异常已恢复",
					zap.Any("错误", err),
					zap.String("路径", c.Request.URL.Path),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    errcode.ServerError.Int(),
					"message": errcode.ServerError.Msg(),
				})
			}
		}()
		c.Next()
	}
}
