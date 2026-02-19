package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"pvecloud/backend/internal/pveclient"
	"pvecloud/backend/internal/repository"
	"pvecloud/backend/internal/security"
	"pvecloud/backend/pkg/response"
)

// AdminHandler 提供后台仪表盘、用户管理、节点监控等接口。
type AdminHandler struct {
	userRepo   *repository.UserRepository
	orderRepo  *repository.OrderRepository
	tokenStore *security.TokenStore
	pve        pveclient.PVEClient
}

// NewAdminHandler 创建后台处理器。
func NewAdminHandler(
	userRepo *repository.UserRepository,
	orderRepo *repository.OrderRepository,
	tokenStore *security.TokenStore,
	pve pveclient.PVEClient,
) *AdminHandler {
	return &AdminHandler{userRepo: userRepo, orderRepo: orderRepo, tokenStore: tokenStore, pve: pve}
}

// RegisterRoutes 注册后台管理路由。
func (h *AdminHandler) RegisterRoutes(admin *gin.RouterGroup) {
	admin.GET("/users", h.Users)
	admin.POST("/users/:id/toggle-status", h.ToggleUserStatus)
	admin.POST("/users/:id/force-logout", h.ForceLogoutUser)
	admin.GET("/dashboard", h.Dashboard)
	admin.GET("/nodes", h.Nodes)
}

// Users 查询后台用户列表。
func (h *AdminHandler) Users(c *gin.Context) {
	keyword := c.Query("keyword")
	users, err := h.userRepo.List(c.Request.Context(), keyword)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40081, err.Error())
		return
	}
	response.OK(c, users)
}

// ToggleUserStatus 切换用户状态；禁用用户时自动触发强制下线。
func (h *AdminHandler) ToggleUserStatus(c *gin.Context) {
	id := parseUintParam(c, "id")
	user, err := h.userRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40082, err.Error())
		return
	}

	if user.Status == "active" {
		user.Status = "disabled"
	} else {
		user.Status = "active"
	}

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		response.Error(c, http.StatusBadRequest, 40083, err.Error())
		return
	}

	if user.Status == "disabled" && h.tokenStore != nil {
		_ = h.tokenStore.ForceLogoutUser(c.Request.Context(), user.ID)
	}

	response.OK(c, gin.H{"message": "状态已更新", "status": user.Status})
}

// ForceLogoutUser 仅执行强制下线，不变更账号状态。
func (h *AdminHandler) ForceLogoutUser(c *gin.Context) {
	id := parseUintParam(c, "id")
	if h.tokenStore == nil {
		response.OK(c, gin.H{"message": "token store not enabled, skip"})
		return
	}
	if err := h.tokenStore.ForceLogoutUser(c.Request.Context(), id); err != nil {
		response.Error(c, http.StatusBadRequest, 40084, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "用户已强制下线"})
}

// Dashboard 返回后台关键统计指标。
func (h *AdminHandler) Dashboard(c *gin.Context) {
	users, _ := h.userRepo.List(c.Request.Context(), "")
	orders, _ := h.orderRepo.ListForAdmin(c.Request.Context(), 0, "")

	pendingTickets := 0
	activeInstances := 0
	todayRevenue := 0.0
	for _, order := range orders {
		if order.Status == "pending" {
			pendingTickets++
		}
		if order.Status == "active" {
			activeInstances++
			todayRevenue += order.Amount
		}
	}

	response.OK(c, gin.H{
		"total_users":      len(users),
		"active_instances": activeInstances,
		"today_revenue":    todayRevenue,
		"pending_tickets":  pendingTickets,
	})
}

// Nodes 查询节点资源使用率。
func (h *AdminHandler) Nodes(c *gin.Context) {
	nodes := []string{"node-a", "node-b"}
	list := make([]interface{}, 0, len(nodes))
	for _, node := range nodes {
		status, err := h.pve.GetNodeStatus(c.Request.Context(), node)
		if err == nil {
			list = append(list, status)
		}
	}
	response.OK(c, list)
}
