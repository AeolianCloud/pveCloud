package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"pvecloud/backend/internal/middleware"
	"pvecloud/backend/internal/service"
	"pvecloud/backend/pkg/response"
)

// InstanceHandler 处理实例列表、详情、开关机、控制台等接口。
type InstanceHandler struct {
	service *service.InstanceService
}

// NewInstanceHandler 创建实例处理器。
func NewInstanceHandler(s *service.InstanceService) *InstanceHandler {
	return &InstanceHandler{service: s}
}

// RegisterRoutes 注册实例路由。
func (h *InstanceHandler) RegisterRoutes(user *gin.RouterGroup) {
	user.GET("/instances", h.List)
	user.GET("/instances/:id", h.Detail)
	user.POST("/instances/:id/start", h.Start)
	user.POST("/instances/:id/stop", h.Stop)
	user.POST("/instances/:id/reboot", h.Reboot)
	user.POST("/instances/:id/console", h.Console)
}

// List 查询用户实例列表。
func (h *InstanceHandler) List(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	list, err := h.service.ListUserInstances(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40051, err.Error())
		return
	}
	response.OK(c, list)
}

// Detail 查询实例详情。
func (h *InstanceHandler) Detail(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	instanceID := parseUintParam(c, "id")
	item, err := h.service.GetUserInstance(c.Request.Context(), userID, instanceID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40052, err.Error())
		return
	}
	response.OK(c, item)
}

// Start 执行开机。
func (h *InstanceHandler) Start(c *gin.Context) {
	h.operate(c, "start")
}

// Stop 执行关机。
func (h *InstanceHandler) Stop(c *gin.Context) {
	h.operate(c, "stop")
}

// Reboot 执行重启。
func (h *InstanceHandler) Reboot(c *gin.Context) {
	h.operate(c, "reboot")
}

func (h *InstanceHandler) operate(c *gin.Context, action string) {
	userID := middleware.UserIDFromContext(c)
	instanceID := parseUintParam(c, "id")
	result, err := h.service.Operate(c.Request.Context(), userID, instanceID, action)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40053, err.Error())
		return
	}
	response.OK(c, result)
}

// Console 获取 VNC 控制台访问信息。
func (h *InstanceHandler) Console(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	instanceID := parseUintParam(c, "id")
	info, err := h.service.GetConsole(c.Request.Context(), userID, instanceID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40054, err.Error())
		return
	}
	response.OK(c, info)
}
