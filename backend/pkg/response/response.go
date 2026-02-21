package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"pvecloud/backend/pkg/response/errcode"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Success 请求成功
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    errcode.OK.Int(),
		Message: errcode.OK.Msg(),
		Data:    data,
	})
}

// Fail 业务失败，HTTP 200 + 业务错误码
func Fail(c *gin.Context, code errcode.Code) {
	c.JSON(http.StatusOK, Response{
		Code:    code.Int(),
		Message: code.Msg(),
	})
}

// FailMsg 业务失败，允许自定义提示文字
func FailMsg(c *gin.Context, code errcode.Code, msg string) {
	c.JSON(http.StatusOK, Response{
		Code:    code.Int(),
		Message: msg,
	})
}

// BadRequest 参数错误 400
func BadRequest(c *gin.Context, msg string) {
	// 统一规范：参数错误也使用 HTTP 200 返回，由 code 区分业务失败
	// 这样前端只需要处理一套 JSON 响应结构即可。
	c.JSON(http.StatusOK, Response{
		Code:    errcode.InvalidParams.Int(),
		Message: msg,
	})
}

// Unauthorized 未授权 401
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    errcode.Unauthorized.Int(),
		Message: msg,
	})
}

// Forbidden 无权限 403
func Forbidden(c *gin.Context, msg string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    errcode.Forbidden.Int(),
		Message: msg,
	})
}

// InternalError 服务器错误 500
func InternalError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    errcode.ServerError.Int(),
		Message: msg,
	})
}

// NotFound 接口不存在 404
// 说明：用于 NoRoute/NoMethod，确保不返回 HTML。
func NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, Response{
		Code:    errcode.NotFound.Int(),
		Message: errcode.NotFound.Msg(),
	})
}
