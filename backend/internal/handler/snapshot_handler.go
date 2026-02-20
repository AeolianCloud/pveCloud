package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"pvecloud/backend/internal/middleware"
	"pvecloud/backend/internal/service"
	"pvecloud/backend/pkg/response"
)

// SnapshotHandler 处理快照增删查与恢复接口。
type SnapshotHandler struct {
	service *service.InstanceService
}

// NewSnapshotHandler 创建快照处理器。
func NewSnapshotHandler(s *service.InstanceService) *SnapshotHandler {
	return &SnapshotHandler{service: s}
}

// RegisterRoutes 注册快照路由。
func (h *SnapshotHandler) RegisterRoutes(user *gin.RouterGroup) {
	user.GET("/instances/:id/snapshots", h.List)
	user.POST("/instances/:id/snapshots", h.Create)
	user.DELETE("/instances/:id/snapshots/:name", h.Delete)
	user.POST("/instances/:id/snapshots/:name/restore", h.Restore)
}

// List 查询快照列表。
func (h *SnapshotHandler) List(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	instanceID := parseUintParam(c, "id")
	list, err := h.service.ListSnapshots(c.Request.Context(), userID, instanceID)
	if err != nil {
		if service.IsForbidden(err) {
			response.Error(c, http.StatusForbidden, 40361, err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, 40061, err.Error())
		return
	}
	response.OK(c, list)
}

// Create 创建快照。
func (h *SnapshotHandler) Create(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	instanceID := parseUintParam(c, "id")
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		response.Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	result, err := h.service.CreateSnapshot(c.Request.Context(), userID, instanceID, req.Name)
	if err != nil {
		if service.IsForbidden(err) {
			response.Error(c, http.StatusForbidden, 40362, err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, 40062, err.Error())
		return
	}
	response.OK(c, result)
}

// Delete 删除快照。
func (h *SnapshotHandler) Delete(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	instanceID := parseUintParam(c, "id")
	name := c.Param("name")
	result, err := h.service.DeleteSnapshot(c.Request.Context(), userID, instanceID, name)
	if err != nil {
		if service.IsForbidden(err) {
			response.Error(c, http.StatusForbidden, 40363, err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, 40063, err.Error())
		return
	}
	response.OK(c, result)
}

// Restore 恢复快照。
func (h *SnapshotHandler) Restore(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	instanceID := parseUintParam(c, "id")
	name := c.Param("name")
	result, err := h.service.RestoreSnapshot(c.Request.Context(), userID, instanceID, name)
	if err != nil {
		if service.IsForbidden(err) {
			response.Error(c, http.StatusForbidden, 40364, err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, 40064, err.Error())
		return
	}
	response.OK(c, result)
}
