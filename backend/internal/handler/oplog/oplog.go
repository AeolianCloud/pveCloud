// internal/handler/oplog/oplog.go
// 操作日志 HTTP 处理器。
package oplog

import (
	"github.com/gin-gonic/gin"
	svcoplog "pvecloud/backend/internal/service/oplog"
	"pvecloud/backend/pkg/pagination"
	"pvecloud/backend/pkg/response"
	"pvecloud/backend/pkg/response/errcode"
)

// Handler 操作日志处理器。
type Handler struct {
	svc *svcoplog.Service
}

// New 创建操作日志处理器。
func New(svc *svcoplog.Service) *Handler {
	return &Handler{svc: svc}
}

// List 分页查询操作日志。
// GET /api/v1/op-logs?page_num=1&page_size=20&username=xxx&module=admin&action=delete
func (h *Handler) List(c *gin.Context) {
	var req svcoplog.ListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}

	logs, total, err := h.svc.List(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, pagination.NewResult(&req.Page, total, logs))
}
