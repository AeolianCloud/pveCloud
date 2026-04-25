package errors

import (
	"errors"
	"net/http"
)

/**
 * AppError 表示带业务错误码和 HTTP 状态码的应用错误。
 */
type AppError struct {
	Code       int
	Message    string
	HTTPStatus int
}

/**
 * New 创建应用错误。
 *
 * @param code 业务错误码
 * @param message 中文错误信息
 * @param httpStatus HTTP 状态码
 * @return *AppError 应用错误
 */
func New(code int, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus}
}

/**
 * Error 返回错误消息文本。
 *
 * @return string 中文错误信息
 */
func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

/**
 * WithMessage 基于现有错误码创建带自定义消息的新错误。
 *
 * @param message 自定义中文错误信息
 * @return *AppError 新应用错误
 */
func (e *AppError) WithMessage(message string) *AppError {
	if e == nil {
		return ErrInternal.WithMessage(message)
	}
	return &AppError{Code: e.Code, Message: message, HTTPStatus: e.HTTPStatus}
}

var (
	/** ErrValidation 表示请求参数或基础校验失败。 */
	ErrValidation = New(40001, "参数错误", http.StatusBadRequest)
	/** ErrUnauthorized 表示用户或管理员未登录、token 无效或 token 过期。 */
	ErrUnauthorized = New(40101, "未登录或登录已过期", http.StatusUnauthorized)
	/** ErrForbidden 表示当前主体没有执行该操作的权限。 */
	ErrForbidden = New(40301, "无权限", http.StatusForbidden)
	/** ErrNotFound 表示请求资源不存在或不可见。 */
	ErrNotFound = New(40401, "资源不存在", http.StatusNotFound)
	/** ErrConflict 表示业务状态冲突或重复提交。 */
	ErrConflict = New(40901, "状态冲突", http.StatusConflict)
	/** ErrInternal 表示服务端内部错误。 */
	ErrInternal = New(50001, "系统内部错误", http.StatusInternalServerError)
)

/**
 * From 将任意错误转换为应用错误。
 *
 * @param err 任意错误
 * @return *AppError 应用错误
 */
func From(err error) *AppError {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return ErrInternal
}
