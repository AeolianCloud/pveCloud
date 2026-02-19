package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"pvecloud/backend/internal/middleware"
	"pvecloud/backend/internal/service"
	"pvecloud/backend/pkg/response"
)

// TicketHandler 处理用户工单与后台工单管理接口。
type TicketHandler struct {
	service *service.TicketService
}

// NewTicketHandler 创建工单处理器。
func NewTicketHandler(s *service.TicketService) *TicketHandler {
	return &TicketHandler{service: s}
}

// RegisterRoutes 注册工单相关路由。
func (h *TicketHandler) RegisterRoutes(user *gin.RouterGroup, admin *gin.RouterGroup) {
	user.POST("/tickets", h.Create)
	user.GET("/tickets", h.UserList)
	user.POST("/tickets/:id/replies", h.UserReply)
	user.GET("/tickets/:id/replies", h.Replies)

	admin.GET("/tickets", h.AdminList)
	admin.POST("/tickets/:id/replies", h.AdminReply)
	admin.POST("/tickets/:id/close", h.Close)
	admin.POST("/tickets/:id/processing", h.Processing)
}

// Create 用户提交工单。
func (h *TicketHandler) Create(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	var req struct {
		Title      string `json:"title"`
		Content    string `json:"content"`
		Priority   string `json:"priority"`
		InstanceID *uint  `json:"instance_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Title == "" || req.Content == "" {
		response.Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	ticket, err := h.service.CreateTicket(c.Request.Context(), userID, req.Title, req.Content, req.Priority, req.InstanceID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40071, err.Error())
		return
	}
	response.OK(c, gin.H{"ticket_id": ticket.ID})
}

// UserList 查询用户工单。
func (h *TicketHandler) UserList(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	status := c.Query("status")
	list, err := h.service.ListUserTickets(c.Request.Context(), userID, status)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40072, err.Error())
		return
	}
	response.OK(c, list)
}

// UserReply 用户在自己的工单下回复。
func (h *TicketHandler) UserReply(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	ticketID := parseUintParam(c, "id")
	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Content == "" {
		response.Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if err := h.service.AddReply(c.Request.Context(), userID, ticketID, false, req.Content); err != nil {
		response.Error(c, http.StatusBadRequest, 40073, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "回复成功"})
}

// AdminReply 管理员回复工单。
func (h *TicketHandler) AdminReply(c *gin.Context) {
	adminID := middleware.UserIDFromContext(c)
	ticketID := parseUintParam(c, "id")
	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Content == "" {
		response.Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if err := h.service.AddReply(c.Request.Context(), adminID, ticketID, true, req.Content); err != nil {
		response.Error(c, http.StatusBadRequest, 40074, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "回复成功"})
}

// Replies 查询工单回复线程。
func (h *TicketHandler) Replies(c *gin.Context) {
	ticketID := parseUintParam(c, "id")
	replies, err := h.service.ListReplies(c.Request.Context(), ticketID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40075, err.Error())
		return
	}
	response.OK(c, replies)
}

// AdminList 查询后台工单列表。
func (h *TicketHandler) AdminList(c *gin.Context) {
	status := c.Query("status")
	list, err := h.service.ListAdminTickets(c.Request.Context(), status)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40076, err.Error())
		return
	}
	response.OK(c, list)
}

// Close 关闭工单。
func (h *TicketHandler) Close(c *gin.Context) {
	ticketID := parseUintParam(c, "id")
	if err := h.service.ChangeStatus(c.Request.Context(), ticketID, "closed"); err != nil {
		response.Error(c, http.StatusBadRequest, 40077, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "工单已关闭"})
}

// Processing 将工单设为处理中。
func (h *TicketHandler) Processing(c *gin.Context) {
	ticketID := parseUintParam(c, "id")
	if err := h.service.ChangeStatus(c.Request.Context(), ticketID, "processing"); err != nil {
		response.Error(c, http.StatusBadRequest, 40078, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "工单处理中"})
}
