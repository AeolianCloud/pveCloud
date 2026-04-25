package errors

import (
	"errors"
	"net/http"
)

type AppError struct {
	Code       int
	Message    string
	HTTPStatus int
}

func New(code int, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus}
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *AppError) WithMessage(message string) *AppError {
	if e == nil {
		return ErrInternal.WithMessage(message)
	}
	return &AppError{Code: e.Code, Message: message, HTTPStatus: e.HTTPStatus}
}

var (
	ErrValidation   = New(40001, "参数错误", http.StatusBadRequest)
	ErrUnauthorized = New(40101, "未登录或登录已过期", http.StatusUnauthorized)
	ErrForbidden    = New(40301, "无权限", http.StatusForbidden)
	ErrNotFound     = New(40401, "资源不存在", http.StatusNotFound)
	ErrConflict     = New(40901, "状态冲突", http.StatusConflict)
	ErrInternal     = New(50001, "系统内部错误", http.StatusInternalServerError)
)

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
