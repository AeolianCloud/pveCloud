package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
)

/**
 * Envelope 表示统一 API 响应包裹结构。
 */
type Envelope struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

/**
 * Success 输出 HTTP 200 成功响应。
 *
 * @param c Gin 请求上下文
 * @param data 响应数据
 */
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Envelope{
		Code:    0,
		Message: "成功",
		Data:    data,
	})
}

/**
 * Created 输出 HTTP 201 创建成功响应。
 *
 * @param c Gin 请求上下文
 * @param data 响应数据
 */
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Envelope{
		Code:    0,
		Message: "成功",
		Data:    data,
	})
}

/**
 * Error 输出统一业务错误响应。
 *
 * @param c Gin 请求上下文
 * @param err 业务错误或普通错误
 */
func Error(c *gin.Context, err error) {
	appErr := apperrors.From(err)
	if appErr == nil {
		appErr = apperrors.ErrInternal
	}

	c.JSON(appErr.HTTPStatus, Envelope{
		Code:    appErr.Code,
		Message: appErr.Message,
		Data:    nil,
	})
}
