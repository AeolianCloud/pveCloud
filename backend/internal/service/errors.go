package service

import (
	"errors"
	"fmt"
)

// ErrForbidden 统一表示权限不足类错误，供 handler 映射为 HTTP 403。
var ErrForbidden = errors.New("forbidden")

// WrapForbidden 构造带上下文信息的权限错误。
func WrapForbidden(message string) error {
	return fmt.Errorf("%w: %s", ErrForbidden, message)
}

// IsForbidden 判断错误是否为权限错误。
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}
