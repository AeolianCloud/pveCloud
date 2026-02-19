package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"pvecloud/backend/internal/middleware"
	"pvecloud/backend/internal/service"
	"pvecloud/backend/pkg/response"
)

// TaskHandler 处理任务状态查询接口。
type TaskHandler struct {
	service *service.OrderService
}

// NewTaskHandler 创建任务处理器。
func NewTaskHandler(s *service.OrderService) *TaskHandler {
	return &TaskHandler{service: s}
}

// RegisterRoutes 注册任务路由。
func (h *TaskHandler) RegisterRoutes(user *gin.RouterGroup) {
	user.GET("/tasks/:id/status", h.Status)
}

// Status 返回任务当前状态。
func (h *TaskHandler) Status(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	taskID := parseUintParam(c, "id")
	task, err := h.service.GetTaskStatus(c.Request.Context(), userID, taskID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40041, err.Error())
		return
	}
	response.OK(c, task)
}
