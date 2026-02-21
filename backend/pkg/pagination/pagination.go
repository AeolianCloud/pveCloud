// pkg/pagination/pagination.go
// 统一分页工具，供所有列表接口使用。
package pagination

// Page 分页请求参数，在 handler 中通过 c.ShouldBindQuery(&p) 绑定。
type Page struct {
	// 页码，从 1 开始，默认 1
	PageNum int `form:"page_num" json:"page_num"`
	// 每页条数，默认 20，最大 100
	PageSize int `form:"page_size" json:"page_size"`
}

// Normalize 修正非法值，确保参数在合理范围内。
func (p *Page) Normalize() {
	if p.PageNum < 1 {
		p.PageNum = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// Offset 计算 SQL OFFSET 值。
func (p *Page) Offset() int {
	return (p.PageNum - 1) * p.PageSize
}

// Result 分页响应结构，嵌入到 response.Success 的 data 字段。
type Result struct {
	// 当前页码
	PageNum int `json:"page_num"`
	// 每页条数
	PageSize int `json:"page_size"`
	// 总记录数
	Total int64 `json:"total"`
	// 数据列表
	List any `json:"list"`
}

// NewResult 构造分页结果。
func NewResult(p *Page, total int64, list any) *Result {
	return &Result{
		PageNum:  p.PageNum,
		PageSize: p.PageSize,
		Total:    total,
		List:     list,
	}
}
