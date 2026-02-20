package handler

import (
	"net/http"
	"strconv"

	"pvecloud/backend/internal/middleware"
	"pvecloud/backend/internal/service"
	"pvecloud/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// OrderHandler 处理下单、订单查询、续费接口。
type OrderHandler struct {
	service *service.OrderService
}

// NewOrderHandler 创建订单处理器。
func NewOrderHandler(s *service.OrderService) *OrderHandler {
	return &OrderHandler{service: s}
}

// RegisterRoutes 注册订单路由。
// RegisterRoutes 注册订单相关的路由
// user: 用户路由组
// admin: 管理员路由组
func (h *OrderHandler) RegisterRoutes(user *gin.RouterGroup, admin *gin.RouterGroup) {

	// 用户订单相关路由
	user.POST("/orders", h.Create)          // 创建订单
	user.GET("/orders", h.List)             // 获取订单列表
	user.GET("/orders/:id", h.Detail)       // 获取订单详情
	user.POST("/orders/:id/renew", h.Renew) // 续费订单

	// 管理员订单相关路由
	admin.GET("/orders", h.AdminList) // 获取所有订单列表（管理员视角）
}

// Create 用户下单。
func (h *OrderHandler) Create(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	var req service.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	order, task, err := h.service.CreateOrder(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40031, err.Error())
		return
	}
	response.OK(c, gin.H{"order_id": order.ID, "task_id": task.ID})
}

// List 查询用户订单列表。
func (h *OrderHandler) List(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	status := c.Query("status")
	orders, err := h.service.ListOrders(c.Request.Context(), userID, status)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40032, err.Error())
		return
	}
	response.OK(c, orders)
}

// Detail 查询订单详情。
func (h *OrderHandler) Detail(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	id := parseUintParam(c, "id")
	order, err := h.service.GetOrderDetail(c.Request.Context(), userID, id)
	if err != nil {
		if service.IsForbidden(err) {
			response.Error(c, http.StatusForbidden, 40331, err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, 40033, err.Error())
		return
	}
	response.OK(c, order)
}

// Renew 执行订单续费。
func (h *OrderHandler) Renew(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	id := parseUintParam(c, "id")
	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if err := h.service.RenewOrder(c.Request.Context(), userID, id, req.Amount); err != nil {
		if service.IsForbidden(err) {
			response.Error(c, http.StatusForbidden, 40332, err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, 40034, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "续费成功"})
}

// AdminList 后台订单列表。
func (h *OrderHandler) AdminList(c *gin.Context) {
	status := c.Query("status")
	dateRange := c.Query("date_range")
	userID, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)
	orders, err := h.service.ListAdminOrders(c.Request.Context(), uint(userID), status, dateRange)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40035, err.Error())
		return
	}
	response.OK(c, orders)
}
