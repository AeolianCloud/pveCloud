package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"pvecloud/backend/internal/middleware"
	"pvecloud/backend/internal/service"
	"pvecloud/backend/pkg/response"
)

// WalletHandler 处理钱包余额、充值、流水接口。
type WalletHandler struct {
	service *service.BillingService
}

// NewWalletHandler 创建钱包处理器。
func NewWalletHandler(s *service.BillingService) *WalletHandler {
	return &WalletHandler{service: s}
}

// RegisterRoutes 注册钱包路由。
func (h *WalletHandler) RegisterRoutes(user *gin.RouterGroup) {
	user.GET("/wallet", h.GetWallet)
	user.POST("/wallet/recharge", h.Recharge)
	user.GET("/wallet/logs", h.Logs)
}

// GetWallet 查询钱包信息。
func (h *WalletHandler) GetWallet(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	data, err := h.service.GetWallet(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40021, err.Error())
		return
	}
	response.OK(c, data)
}

// Recharge 钱包充值。
func (h *WalletHandler) Recharge(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if err := h.service.Recharge(c.Request.Context(), userID, req.Amount); err != nil {
		response.Error(c, http.StatusBadRequest, 40022, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "充值成功"})
}

// Logs 查询钱包流水。
func (h *WalletHandler) Logs(c *gin.Context) {
	userID := middleware.UserIDFromContext(c)
	start := c.Query("start")
	end := c.Query("end")
	data, err := h.service.Logs(c.Request.Context(), userID, start, end)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40023, err.Error())
		return
	}
	response.OK(c, data)
}
