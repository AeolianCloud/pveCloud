package productcatalog

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
)

type ProductCatalogHandler struct {
	service *ProductCatalogService
}

func NewProductCatalogHandler(service *ProductCatalogService) *ProductCatalogHandler {
	return &ProductCatalogHandler{service: service}
}

func (h *ProductCatalogHandler) Show(c *gin.Context) {
	result, err := h.service.Show(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}
