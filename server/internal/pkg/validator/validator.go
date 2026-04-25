package validator

import (
	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

/**
 * Struct 校验结构体字段标签。
 *
 * @param value 待校验的 DTO 或结构体实例
 * @return error 校验失败原因
 */
func Struct(value interface{}) error {
	return validate.Struct(value)
}

/**
 * Var 校验单个变量。
 *
 * @param field 待校验变量
 * @param tag go-playground/validator 标签，例如 required、email 或 oneof
 * @return error 校验失败原因
 */
func Var(field interface{}, tag string) error {
	return validate.Var(field, tag)
}
