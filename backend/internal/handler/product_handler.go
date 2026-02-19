package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"pvecloud/backend/internal/model"
	"pvecloud/backend/internal/service"
	"pvecloud/backend/pkg/response"
)

// ProductHandler 处理商品前后台接口。
type ProductHandler struct {
	service *service.ProductService
}

// NewProductHandler 创建商品处理器。
func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{service: productService}
}

// RegisterRoutes 注册商品相关路由。
func (h *ProductHandler) RegisterRoutes(pub *gin.RouterGroup, admin *gin.RouterGroup) {
	pub.GET("/products", h.ListPublic)
	pub.GET("/products/:id", h.GetPublicDetail)

	admin.GET("/products", h.ListAdmin)
	admin.POST("/products", h.Create)
	admin.PUT("/products/:id", h.Update)
	admin.DELETE("/products/:id", h.Delete)
}

// ListPublic 查询前台商品列表。
func (h *ProductHandler) ListPublic(c *gin.Context) {
	items, err := h.service.ListPublicProducts(c.Request.Context(), 0)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, err.Error())
		return
	}
	response.OK(c, items)
}

// GetPublicDetail 查询前台商品详情。
func (h *ProductHandler) GetPublicDetail(c *gin.Context) {
	id := parseUintParam(c, "id")
	item, err := h.service.GetDetail(c.Request.Context(), id, false)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40011, err.Error())
		return
	}
	response.OK(c, item)
}

// ListAdmin 查询后台商品。
func (h *ProductHandler) ListAdmin(c *gin.Context) {
	items, err := h.service.ListAdminProducts(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, err.Error())
		return
	}
	response.OK(c, items)
}

// Create 创建商品和定价。
func (h *ProductHandler) Create(c *gin.Context) {
	var req struct {
		Product model.Product        `json:"product"`
		Prices  []model.ProductPrice `json:"prices"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if err := h.service.Create(c.Request.Context(), &req.Product, req.Prices); err != nil {
		response.Error(c, http.StatusBadRequest, 40012, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "商品创建成功"})
}

// Update 编辑商品和定价。
func (h *ProductHandler) Update(c *gin.Context) {
	id := parseUintParam(c, "id")
	var req struct {
		Product model.Product        `json:"product"`
		Prices  []model.ProductPrice `json:"prices"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	req.Product.ID = id
	if err := h.service.Update(c.Request.Context(), &req.Product, req.Prices); err != nil {
		response.Error(c, http.StatusBadRequest, 40013, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "商品更新成功"})
}

// Delete 删除商品。
func (h *ProductHandler) Delete(c *gin.Context) {
	id := parseUintParam(c, "id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, http.StatusBadRequest, 40014, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "商品删除成功"})
}
