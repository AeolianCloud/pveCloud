package productcatalog

import (
	"context"

	"gorm.io/gorm"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/web/dto"
)

type ProductCatalogService struct {
	db *gorm.DB
}

func NewProductCatalogService(db *gorm.DB) *ProductCatalogService {
	return &ProductCatalogService{db: db}
}

func (s *ProductCatalogService) Show(ctx context.Context) (webdto.ServerCatalogResponse, error) {
	var products []models.Product
	if err := s.db.WithContext(ctx).
		Where("type = ? AND status = ? AND visible = 1", "server", "active").
		Order("sort_order ASC, id ASC").
		Find(&products).Error; err != nil {
		return webdto.ServerCatalogResponse{}, err
	}
	if len(products) == 0 {
		return webdto.ServerCatalogResponse{Products: []webdto.ServerCatalogProduct{}}, nil
	}

	productIDs := make([]uint64, 0, len(products))
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
	}

	var plans []models.ProductPlan
	if err := s.db.WithContext(ctx).
		Where("product_id IN ? AND status IN ? AND visible = 1", productIDs, []string{"active", "sold_out"}).
		Order("is_featured DESC, sort_order ASC, id ASC").
		Find(&plans).Error; err != nil {
		return webdto.ServerCatalogResponse{}, err
	}
	if len(plans) == 0 {
		return webdto.ServerCatalogResponse{Products: catalogProducts(products, nil, nil, nil, nil)}, nil
	}

	planIDs := make([]uint64, 0, len(plans))
	for _, plan := range plans {
		planIDs = append(planIDs, plan.ID)
	}

	prices, err := s.planPrices(ctx, planIDs)
	if err != nil {
		return webdto.ServerCatalogResponse{}, err
	}
	regions, err := s.planRegions(ctx, planIDs)
	if err != nil {
		return webdto.ServerCatalogResponse{}, err
	}
	templates, err := s.planTemplates(ctx, planIDs)
	if err != nil {
		return webdto.ServerCatalogResponse{}, err
	}

	return webdto.ServerCatalogResponse{Products: catalogProducts(products, plans, prices, regions, templates)}, nil
}

func (s *ProductCatalogService) planPrices(ctx context.Context, planIDs []uint64) (map[uint64][]webdto.ServerCatalogPlanPrice, error) {
	var prices []models.PlanPrice
	if err := s.db.WithContext(ctx).
		Where("plan_id IN ? AND status = ?", planIDs, "active").
		Order("sort_order ASC, id ASC").
		Find(&prices).Error; err != nil {
		return nil, err
	}
	result := make(map[uint64][]webdto.ServerCatalogPlanPrice)
	for _, price := range prices {
		result[price.PlanID] = append(result[price.PlanID], webdto.ServerCatalogPlanPrice{BillingCycle: price.BillingCycle, PriceCents: price.PriceCents, OriginalPriceCents: price.OriginalPriceCents, Currency: price.Currency})
	}
	return result, nil
}

func (s *ProductCatalogService) planRegions(ctx context.Context, planIDs []uint64) (map[uint64][]webdto.ServerCatalogRegion, error) {
	type row struct {
		PlanID   uint64
		RegionNo string
		Code     string
		Name     string
		Country  *string
		City     *string
		Summary  *string
	}
	var rows []row
	if err := s.db.WithContext(ctx).Table("plan_regions AS rel").
		Select("rel.plan_id, regions.region_no, regions.code, regions.name, regions.country, regions.city, regions.summary").
		Joins("JOIN sales_regions AS regions ON regions.id = rel.region_id").
		Where("rel.plan_id IN ? AND rel.status = ? AND regions.status = ? AND regions.visible = 1", planIDs, "active", "active").
		Order("rel.sort_order ASC, regions.sort_order ASC, regions.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[uint64][]webdto.ServerCatalogRegion)
	for _, row := range rows {
		result[row.PlanID] = append(result[row.PlanID], webdto.ServerCatalogRegion{RegionNo: row.RegionNo, Code: row.Code, Name: row.Name, Country: row.Country, City: row.City, Summary: row.Summary})
	}
	return result, nil
}

func (s *ProductCatalogService) planTemplates(ctx context.Context, planIDs []uint64) (map[uint64][]webdto.ServerCatalogOSTemplate, error) {
	type row struct {
		PlanID       uint64
		TemplateNo   string
		Code         string
		Name         string
		OSFamily     string
		Distribution string
		Version      string
		Architecture string
		Summary      *string
	}
	var rows []row
	if err := s.db.WithContext(ctx).Table("plan_os_templates AS rel").
		Select("rel.plan_id, templates.template_no, templates.code, templates.name, templates.os_family, templates.distribution, templates.version, templates.architecture, templates.summary").
		Joins("JOIN server_os_templates AS templates ON templates.id = rel.template_id").
		Where("rel.plan_id IN ? AND rel.status = ? AND templates.status = ? AND templates.visible = 1", planIDs, "active", "active").
		Order("rel.sort_order ASC, templates.sort_order ASC, templates.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[uint64][]webdto.ServerCatalogOSTemplate)
	for _, row := range rows {
		result[row.PlanID] = append(result[row.PlanID], webdto.ServerCatalogOSTemplate{TemplateNo: row.TemplateNo, Code: row.Code, Name: row.Name, OSFamily: row.OSFamily, Distribution: row.Distribution, Version: row.Version, Architecture: row.Architecture, Summary: row.Summary})
	}
	return result, nil
}

func catalogProducts(products []models.Product, plans []models.ProductPlan, prices map[uint64][]webdto.ServerCatalogPlanPrice, regions map[uint64][]webdto.ServerCatalogRegion, templates map[uint64][]webdto.ServerCatalogOSTemplate) []webdto.ServerCatalogProduct {
	plansByProduct := make(map[uint64][]webdto.ServerCatalogPlan)
	for _, plan := range plans {
		planPrices := prices[plan.ID]
		planRegions := regions[plan.ID]
		planTemplates := templates[plan.ID]
		if len(planPrices) == 0 || len(planRegions) == 0 || len(planTemplates) == 0 {
			continue
		}
		plansByProduct[plan.ProductID] = append(plansByProduct[plan.ProductID], webdto.ServerCatalogPlan{PlanNo: plan.PlanNo, Code: plan.Code, Name: plan.Name, Summary: plan.Summary, CPUCores: plan.CPUCores, MemoryMB: plan.MemoryMB, SystemDiskGB: plan.SystemDiskGB, DataDiskGB: plan.DataDiskGB, BandwidthMbps: plan.BandwidthMbps, TrafficGB: plan.TrafficGB, PublicIPCount: plan.PublicIPCount, Virtualization: plan.Virtualization, Architecture: plan.Architecture, IsFeatured: plan.IsFeatured, Status: plan.Status, Prices: planPrices, Regions: planRegions, OSTemplates: planTemplates})
	}

	items := make([]webdto.ServerCatalogProduct, 0, len(products))
	for _, product := range products {
		if len(plansByProduct[product.ID]) == 0 {
			continue
		}
		items = append(items, webdto.ServerCatalogProduct{ProductNo: product.ProductNo, Slug: product.Slug, Name: product.Name, Summary: product.Summary, Description: product.Description, Plans: plansByProduct[product.ID]})
	}
	return items
}
