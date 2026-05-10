package catalog

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/usecase/web/catalog"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
)

type Handler struct {
	service *catalog.ServerCatalogService
}

func NewHandler(service *catalog.ServerCatalogService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Show(c *gin.Context) {
	result, err := h.service.Show(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, serverCatalogResponse(result))
}

func serverCatalogResponse(catalogResult catalog.ServerCatalog) webdto.ServerCatalogResponse {
	products := make([]webdto.ServerCatalogProduct, 0, len(catalogResult.Products))
	for _, product := range catalogResult.Products {
		products = append(products, webdto.ServerCatalogProduct{
			ProductNo:   product.ProductNo,
			Slug:        product.Slug,
			Name:        product.Name,
			Summary:     product.Summary,
			Description: product.Description,
			Plans:       serverCatalogPlans(product.Plans),
		})
	}
	return webdto.ServerCatalogResponse{Products: products}
}

func serverCatalogPlans(plans []catalog.ServerCatalogPlan) []webdto.ServerCatalogPlan {
	items := make([]webdto.ServerCatalogPlan, 0, len(plans))
	for _, plan := range plans {
		items = append(items, webdto.ServerCatalogPlan{
			PlanNo:         plan.PlanNo,
			Code:           plan.Code,
			Name:           plan.Name,
			Summary:        plan.Summary,
			CPUCores:       plan.CPUCores,
			MemoryMB:       plan.MemoryMB,
			SystemDiskGB:   plan.SystemDiskGB,
			DataDiskGB:     plan.DataDiskGB,
			BandwidthMbps:  plan.BandwidthMbps,
			TrafficGB:      plan.TrafficGB,
			PublicIPCount:  plan.PublicIPCount,
			Virtualization: plan.Virtualization,
			Architecture:   plan.Architecture,
			IsFeatured:     plan.IsFeatured,
			Status:         plan.Status,
			Prices:         serverCatalogPlanPrices(plan.Prices),
			Regions:        serverCatalogRegions(plan.Regions),
			OSTemplates:    serverCatalogOSTemplates(plan.OSTemplates),
			NetworkTypes:   serverCatalogNetworkTypes(plan.NetworkTypes),
		})
	}
	return items
}

func serverCatalogPlanPrices(prices []catalog.ServerCatalogPlanPrice) []webdto.ServerCatalogPlanPrice {
	items := make([]webdto.ServerCatalogPlanPrice, 0, len(prices))
	for _, price := range prices {
		items = append(items, webdto.ServerCatalogPlanPrice{
			BillingCycle:       price.BillingCycle,
			PriceCents:         price.PriceCents,
			OriginalPriceCents: price.OriginalPriceCents,
			Currency:           price.Currency,
		})
	}
	return items
}

func serverCatalogRegions(regions []catalog.ServerCatalogRegion) []webdto.ServerCatalogRegion {
	items := make([]webdto.ServerCatalogRegion, 0, len(regions))
	for _, region := range regions {
		items = append(items, webdto.ServerCatalogRegion{
			RegionNo: region.RegionNo,
			Code:     region.Code,
			Name:     region.Name,
			Country:  region.Country,
			City:     region.City,
			Summary:  region.Summary,
		})
	}
	return items
}

func serverCatalogOSTemplates(templates []catalog.ServerCatalogOSTemplate) []webdto.ServerCatalogOSTemplate {
	items := make([]webdto.ServerCatalogOSTemplate, 0, len(templates))
	for _, template := range templates {
		items = append(items, webdto.ServerCatalogOSTemplate{
			TemplateNo:   template.TemplateNo,
			Code:         template.Code,
			Name:         template.Name,
			OSFamily:     template.OSFamily,
			Distribution: template.Distribution,
			Version:      template.Version,
			Architecture: template.Architecture,
			Summary:      template.Summary,
		})
	}
	return items
}

func serverCatalogNetworkTypes(networkTypes []catalog.ServerCatalogNetworkType) []webdto.ServerCatalogNetworkType {
	items := make([]webdto.ServerCatalogNetworkType, 0, len(networkTypes))
	for _, item := range networkTypes {
		items = append(items, webdto.ServerCatalogNetworkType{
			NetworkTypeNo: item.NetworkTypeNo,
			Code:          item.Code,
			Name:          item.Name,
			Summary:       item.Summary,
		})
	}
	return items
}
