// internal/handler/menu/menu.go
// 菜单 HTTP 处理器：
// - /menus/my：按当前用户权限裁剪后的菜单树（侧边栏渲染用）
// - /menus：菜单管理 CRUD（仅 super_admin）
package menu

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	svcmenu "pvecloud/backend/internal/service/menu"
	"pvecloud/backend/pkg/response"
	"pvecloud/backend/pkg/response/errcode"
)

// Handler 菜单处理器。
type Handler struct {
	svc *svcmenu.Service
}

// New 创建菜单处理器。
func New(svc *svcmenu.Service) *Handler {
	return &Handler{svc: svc}
}

// MyTree 获取当前用户可见菜单树。
// GET /api/v1/menus/my
func (h *Handler) MyTree(c *gin.Context) {
	userIDAny, ok := c.Get("user_id")
	if !ok {
		// JWT 中间件保证存在；这里兜底处理，避免 panic。
		response.Unauthorized(c, errcode.Unauthorized.Msg())
		return
	}
	userID, _ := userIDAny.(uint)

	tree, err := h.svc.ListTreeForUser(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, tree)
}

// List 获取完整菜单树（菜单管理页使用，仅 super_admin）。
// GET /api/v1/menus
func (h *Handler) List(c *gin.Context) {
	tree, err := h.svc.ListTreeAll()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, tree)
}

// Create 创建菜单（仅 super_admin）。
// POST /api/v1/menus
func (h *Handler) Create(c *gin.Context) {
	var req svcmenu.CreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, errcode.InvalidParams, err.Error())
		return
	}

	menu, err := h.svc.Create(&req)
	if err != nil {
		switch {
		case errors.Is(err, svcmenu.ErrParentNotFound):
			response.FailMsg(c, errcode.InvalidParams, "父菜单不存在")
		case errors.Is(err, svcmenu.ErrInvalidPath):
			response.FailMsg(c, errcode.InvalidParams, "path 必须以 / 开头或为空（目录节点）")
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, menu)
}

// Update 更新菜单（仅 super_admin）。
// PUT /api/v1/menus/:id
func (h *Handler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}

	var req svcmenu.UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, errcode.InvalidParams, err.Error())
		return
	}

	menu, err := h.svc.Update(id, &req)
	if err != nil {
		switch {
		case errors.Is(err, svcmenu.ErrMenuNotFound):
			response.Fail(c, errcode.NotFound)
		case errors.Is(err, svcmenu.ErrParentNotFound):
			response.FailMsg(c, errcode.InvalidParams, "父菜单不存在")
		case errors.Is(err, svcmenu.ErrInvalidPath):
			response.FailMsg(c, errcode.InvalidParams, "path 必须以 / 开头或为空（目录节点）")
		case errors.Is(err, svcmenu.ErrInvalidParent):
			response.FailMsg(c, errcode.InvalidParams, "parent_id 不能等于自身")
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Success(c, menu)
}

// Delete 删除菜单（软删除，仅 super_admin）。
// DELETE /api/v1/menus/:id
func (h *Handler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Fail(c, errcode.InvalidParams)
		return
	}

	if err := h.svc.Delete(id); err != nil {
		if errors.Is(err, svcmenu.ErrMenuNotFound) {
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

