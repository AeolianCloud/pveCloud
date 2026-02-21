// internal/handler/permission/permission.go
// 权限查询 HTTP 处理器。
package permission

import (
	"github.com/gin-gonic/gin"
	svcperm "pvecloud/backend/internal/service/permission"
	"pvecloud/backend/pkg/response"
)

// Handler 权限处理器。
type Handler struct {
	svc *svcperm.Service
}

// New 创建权限处理器。
func New(svc *svcperm.Service) *Handler {
	return &Handler{svc: svc}
}

// ListGrouped 返回按 group 分组的全部权限列表，供前端权限分配使用。
// GET /api/v1/permissions
func (h *Handler) ListGrouped(c *gin.Context) {
	groups, err := h.svc.ListGrouped()
	if err != nil {
		response.InternalError(c, "查询权限失败")
		return
	}
	response.Success(c, groups)
}
