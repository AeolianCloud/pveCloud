package response

import "github.com/gin-gonic/gin"

// Payload 统一 API 返回结构，确保前后端协议稳定一致。
type Payload struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// OK 返回标准成功响应。
// OK 是一个成功响应的辅助函数，用于返回标准化的成功响应格式
// 参数:
//
//	c - gin.Context的指针，用于处理HTTP请求和响应
//	data - 需要返回给客户端的数据，可以是任意类型
func OK(c *gin.Context, data interface{}) {
	c.JSON(200, Payload{Code: 0, Message: "ok", Data: data})
}

// Error 返回标准错误响应。
// Error 函数用于返回错误响应
// @param c gin.Context 请求上下文，用于响应客户端
// @param httpStatus int HTTP状态码，表示响应的HTTP状态
// @param code int 业务错误码，用于标识具体的错误类型
// @param message string 错误信息，向客户端描述具体的错误内容
func Error(c *gin.Context, httpStatus int, code int, message string) {
	// 使用JSON格式向客户端返回错误响应
	// Payload结构包含Code、Message和Data字段，其中Data设为nil表示无有效数据
	c.JSON(httpStatus, Payload{Code: code, Message: message, Data: nil})
}
