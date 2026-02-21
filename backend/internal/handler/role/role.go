// internal/handler/role/role.go
// 角色管理 HTTP 处理器：列表、新建、更新、删除、权限分配。
package role

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	svcrole "pvecloud/backend/internal/service/role"
	"pvecloud/backend/pkg/pagination"
	"pvecloud/backend/pkg/response"
	"pvecloud/backend/pkg/response/errcode"
)

// Handler 角色处理器。
type Handler struct {
	svc *svcrole.Service
}

// New 创建角色处理器。
func New(svc *svcrole.Service) *Handler {
	return &Handler{svc: svc}
}

// List 获取角色列表（分页 + 关键词）。
// GET /api/v1/roles?page_num=1&page_size=20&keyword=xxx
func (h *Handler) List(c *gin.Context) {
	var req svcrole.ListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}

	roles, total, err := h.svc.List(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, pagination.NewResult(&req.Page, total, roles))
}

// GetByID 获取角色详情（含权限列表）。
// GET /api/v1/roles/:id
func (h *Handler) GetByID(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}

	role, err := h.svc.GetByID(id)
	if err != nil {
		if errors.Is(err, svcrole.ErrRoleNotFound) {
			response.Fail(c, errcode.NotFound)
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, role)
}

// Create 新建角色。
// POST /api/v1/roles
func (h *Handler) Create(c *gin.Context) {
	var req svcrole.CreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, errcode.InvalidParams, err.Error())
		return
	}

	role, err := h.svc.Create(&req)
	if err != nil {
		if errors.Is(err, svcrole.ErrRoleNameExists) {
			response.FailMsg(c, errcode.InvalidParams, "角色标识已存在")
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, role)
}

// Update 更新角色基本信息。
// PUT /api/v1/roles/:id
func (h *Handler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}

	var req svcrole.UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, errcode.InvalidParams, err.Error())
		return
	}

	role, err := h.svc.Update(id, &req)
	if err != nil {
		if errors.Is(err, svcrole.ErrRoleNotFound) {
			response.Fail(c, errcode.NotFound)
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, role)
}

// Delete 软删除角色。
// DELETE /api/v1/roles/:id
func (h *Handler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}

	if err := h.svc.Delete(id); err != nil {
		if errors.Is(err, svcrole.ErrRoleNotFound) {
			response.Fail(c, errcode.NotFound)
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, nil)
}

// AssignPermissions 替换角色的权限列表。
// PUT /api/v1/roles/:id/permissions
func (h *Handler) AssignPermissions(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}

	var req svcrole.AssignPermissionsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, errcode.InvalidParams, err.Error())
		return
	}

	if err := h.svc.AssignPermissions(id, &req); err != nil {
		if errors.Is(err, svcrole.ErrRoleNotFound) {
			response.Fail(c, errcode.NotFound)
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, nil)
}

// parseID 解析 URL 中的 :id 参数。
func parseID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid id")
	}
	return uint(id), nil
}
