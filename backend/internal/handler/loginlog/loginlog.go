// internal/handler/loginlog/loginlog.go
// 登录日志 HTTP 处理器。
package loginlog

import (
	"github.com/gin-gonic/gin"
	svclog "pvecloud/backend/internal/service/loginlog"
	"pvecloud/backend/pkg/pagination"
	"pvecloud/backend/pkg/response"
	"pvecloud/backend/pkg/response/errcode"
)

// Handler 登录日志处理器。
type Handler struct {
	svc *svclog.Service
}

// New 创建登录日志处理器。
func New(svc *svclog.Service) *Handler {
	return &Handler{svc: svc}
}

// List 分页查询登录日志。
// GET /api/v1/login-logs?page_num=1&page_size=20&username=xxx&status=0
func (h *Handler) List(c *gin.Context) {
	var req svclog.ListReq
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
