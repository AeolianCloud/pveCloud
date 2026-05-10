package productcatalog

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/httputil"
	"github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/middleware"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	productcatalogusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/productcatalog"
)

type ProductCatalogHandler struct {
	service *productcatalogusecase.ProductCatalogService
}

func NewProductCatalogHandler(service *productcatalogusecase.ProductCatalogService) *ProductCatalogHandler {
	return &ProductCatalogHandler{service: service}
}

func (h *ProductCatalogHandler) Products(c *gin.Context) {
	var query admindto.ProductListQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.Products(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) CreateProduct(c *gin.Context) {
	var req admindto.ProductRequest
	if !bindJSON(c, &req) {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.CreateProduct(c.Request.Context(), operatorID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) UpdateProduct(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	var req admindto.ProductRequest
	if !bindJSON(c, &req) {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.UpdateProduct(c.Request.Context(), operatorID, id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) UpdateProductStatus(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	var req admindto.ProductStatusRequest
	if !bindJSON(c, &req) {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.UpdateProductStatus(c.Request.Context(), operatorID, id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) DeleteProduct(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteProduct(c.Request.Context(), operatorID, id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *ProductCatalogHandler) Plans(c *gin.Context) {
	var query admindto.ProductPlanListQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.Plans(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) CreatePlan(c *gin.Context) {
	var req admindto.ProductPlanRequest
	if !bindJSON(c, &req) {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.CreatePlan(c.Request.Context(), operatorID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) UpdatePlan(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	var req admindto.ProductPlanRequest
	if !bindJSON(c, &req) {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.UpdatePlan(c.Request.Context(), operatorID, id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) UpdatePlanStatus(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	var req admindto.ProductPlanStatusRequest
	if !bindJSON(c, &req) {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.UpdatePlanStatus(c.Request.Context(), operatorID, id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) DeletePlan(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	if err := h.service.DeletePlan(c.Request.Context(), operatorID, id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *ProductCatalogHandler) UpdatePlanPrices(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	var req admindto.PlanPriceListRequest
	if !bindJSON(c, &req) {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.UpdatePlanPrices(c.Request.Context(), operatorID, id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) PlanPrices(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	result, err := h.service.PlanPrices(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) UpdatePlanRegions(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	var req admindto.PlanRelationRequest
	if !bindJSON(c, &req) {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.UpdatePlanRegions(c.Request.Context(), operatorID, id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) PlanRegions(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	result, err := h.service.PlanRegions(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) UpdatePlanOSTemplates(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	var req admindto.PlanRelationRequest
	if !bindJSON(c, &req) {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.UpdatePlanOSTemplates(c.Request.Context(), operatorID, id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) PlanOSTemplates(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	result, err := h.service.PlanOSTemplates(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) UpdatePlanNetworkTypes(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	var req admindto.PlanRelationRequest
	if !bindJSON(c, &req) {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.UpdatePlanNetworkTypes(c.Request.Context(), operatorID, id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) PlanNetworkTypes(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	result, err := h.service.PlanNetworkTypes(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) SalesRegions(c *gin.Context) {
	var query admindto.SalesRegionListQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.SalesRegions(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) CreateSalesRegion(c *gin.Context) {
	var req admindto.SalesRegionRequest
	if !bindJSON(c, &req) {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.CreateSalesRegion(c.Request.Context(), operatorID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) UpdateSalesRegion(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	var req admindto.SalesRegionRequest
	if !bindJSON(c, &req) {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.UpdateSalesRegion(c.Request.Context(), operatorID, id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) DeleteSalesRegion(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteSalesRegion(c.Request.Context(), operatorID, id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *ProductCatalogHandler) ServerOSTemplates(c *gin.Context) {
	var query admindto.ServerOSTemplateListQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.ServerOSTemplates(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) CreateServerOSTemplate(c *gin.Context) {
	var req admindto.ServerOSTemplateRequest
	if !bindJSON(c, &req) {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.CreateServerOSTemplate(c.Request.Context(), operatorID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) UpdateServerOSTemplate(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	var req admindto.ServerOSTemplateRequest
	if !bindJSON(c, &req) {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.UpdateServerOSTemplate(c.Request.Context(), operatorID, id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) DeleteServerOSTemplate(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteServerOSTemplate(c.Request.Context(), operatorID, id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *ProductCatalogHandler) NetworkTypes(c *gin.Context) {
	var query admindto.NetworkTypeListQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.NetworkTypes(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) CreateNetworkType(c *gin.Context) {
	var req admindto.NetworkTypeRequest
	if !bindJSON(c, &req) {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.CreateNetworkType(c.Request.Context(), operatorID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) UpdateNetworkType(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	var req admindto.NetworkTypeRequest
	if !bindJSON(c, &req) {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.UpdateNetworkType(c.Request.Context(), operatorID, id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProductCatalogHandler) DeleteNetworkType(c *gin.Context) {
	id, ok := httputil.AdminPathID(c)
	if !ok {
		return
	}
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteNetworkType(c.Request.Context(), operatorID, id); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

func bindQuery(c *gin.Context, target any) bool {
	if err := c.ShouldBindQuery(target); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return false
	}
	if err := validator.Struct(target); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return false
	}
	return true
}

func bindJSON(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return false
	}
	if err := validator.Struct(target); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return false
	}
	return true
}

func currentAdminID(c *gin.Context) (uint64, bool) {
	operatorID, ok := middleware.CurrentAdminID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return 0, false
	}
	return operatorID, true
}
