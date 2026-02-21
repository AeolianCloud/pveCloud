// internal/handler/admin/admin.go
// 管理员账号 HTTP 处理器：列表、新建、更新、状态切换、删除。
package admin

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	svcadmin "pvecloud/backend/internal/service/admin"
	"pvecloud/backend/pkg/pagination"
	"pvecloud/backend/pkg/response"
	"pvecloud/backend/pkg/response/errcode"
)

// Handler 管理员处理器。
type Handler struct {
	svc *svcadmin.Service
}

// New 创建管理员处理器。
func New(svc *svcadmin.Service) *Handler {
	return &Handler{svc: svc}
}

// List 获取管理员列表（分页 + 关键词搜索）。
// GET /api/v1/admin-users?page_num=1&page_size=20&keyword=xxx
func (h *Handler) List(c *gin.Context) {
	var req svcadmin.ListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}

	users, total, err := h.svc.List(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, pagination.NewResult(&req.Page, total, users))
}

// Create 新建管理员账号。
// POST /api/v1/admin-users
func (h *Handler) Create(c *gin.Context) {
	var req svcadmin.CreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, errcode.InvalidParams, err.Error())
		return
	}

	user, err := h.svc.Create(&req)
	if err != nil {
		switch {
		case errors.Is(err, svcadmin.ErrUsernameExists):
			response.Fail(c, errcode.UserAlreadyExists)
		case errors.Is(err, svcadmin.ErrRoleNotFound):
			response.FailMsg(c, errcode.InvalidParams, "指定角色不存在")
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, user)
}

// Update 更新管理员信息（昵称、邮箱、角色、密码）。
// PUT /api/v1/admin-users/:id
func (h *Handler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}

	var req svcadmin.UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, errcode.InvalidParams, err.Error())
		return
	}

	user, err := h.svc.Update(id, &req)
	if err != nil {
		if errors.Is(err, svcadmin.ErrUserNotFound) {
			response.Fail(c, errcode.UserNotFound)
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, user)
}

// statusReq 切换状态请求体。
type statusReq struct {
	Status int8 `json:"status" binding:"oneof=0 1"`
}

// SetStatus 启用或禁用管理员账号。
// PATCH /api/v1/admin-users/:id/status
func (h *Handler) SetStatus(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}

	var req statusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, errcode.InvalidParams, "status 只能为 0 或 1")
		return
	}

	if err := h.svc.SetStatus(id, req.Status); err != nil {
		if errors.Is(err, svcadmin.ErrUserNotFound) {
			response.Fail(c, errcode.UserNotFound)
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, nil)
}

// Delete 软删除管理员账号。
// DELETE /api/v1/admin-users/:id
func (h *Handler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}

	if err := h.svc.Delete(id); err != nil {
		if errors.Is(err, svcadmin.ErrUserNotFound) {
			response.Fail(c, errcode.UserNotFound)
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
