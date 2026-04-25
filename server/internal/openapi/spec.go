package openapi

import (
	"context"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
)

/**
 * Spec 表示已经通过校验的 OpenAPI 3.x API 规范文件。
 */
type Spec struct {
	path string
}

/**
 * Load 加载并校验 OpenAPI 3.x API 规范文件。
 *
 * @param ctx 校验上下文
 * @param path OpenAPI 规范文件路径
 * @return *Spec 已校验的 OpenAPI 规范引用
 * @return error 加载或校验失败原因
 */
func Load(ctx context.Context, path string) (*Spec, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, err
	}
	// OpenAPI 文件是接口契约的唯一来源，校验失败应在启动阶段暴露出来。
	if err := doc.Validate(ctx); err != nil {
		return nil, err
	}

	return &Spec{path: path}, nil
}

/**
 * Read 读取原始 OpenAPI 3.x API 规范文件内容。
 *
 * @return []byte OpenAPI YAML 文件内容
 * @return error 读取失败原因
 */
func (s *Spec) Read() ([]byte, error) {
	return os.ReadFile(s.path)
}
