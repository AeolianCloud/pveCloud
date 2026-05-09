package catalog

import (
	"context"

	domaincatalog "github.com/AeolianCloud/pveCloud/server/internal/domain/catalog"
	mysqlcatalog "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/catalog"
)

type ServerCatalogService struct {
	products *mysqlcatalog.Repository
}

type ServerCatalog struct {
	Products []ServerCatalogProduct
}

type ServerCatalogProduct struct {
	ProductNo   string
	Slug        string
	Name        string
	Summary     *string
	Description *string
	Plans       []ServerCatalogPlan
}

type ServerCatalogPlan struct {
	PlanNo         string
	Code           string
	Name           string
	Summary        *string
	CPUCores       int
	MemoryMB       int
	SystemDiskGB   int
	DataDiskGB     int
	BandwidthMbps  int
	TrafficGB      *int
	PublicIPCount  int
	Virtualization string
	Architecture   string
	IsFeatured     bool
	Status         string
	Prices         []ServerCatalogPlanPrice
	Regions        []ServerCatalogRegion
	OSTemplates    []ServerCatalogOSTemplate
}

type ServerCatalogPlanPrice struct {
	BillingCycle       string
	PriceCents         uint64
	OriginalPriceCents *uint64
	Currency           string
}

type ServerCatalogRegion struct {
	RegionNo string
	Code     string
	Name     string
	Country  *string
	City     *string
	Summary  *string
}

type ServerCatalogOSTemplate struct {
	TemplateNo   string
	Code         string
	Name         string
	OSFamily     string
	Distribution string
	Version      string
	Architecture string
	Summary      *string
}

func NewServerCatalogService(products *mysqlcatalog.Repository) *ServerCatalogService {
	return &ServerCatalogService{products: products}
}

func (s *ServerCatalogService) Show(ctx context.Context) (ServerCatalog, error) {
	products, err := s.products.ActiveServerProducts(ctx)
	if err != nil {
		return ServerCatalog{}, err
	}
	if len(products) == 0 {
		return ServerCatalog{Products: []ServerCatalogProduct{}}, nil
	}

	productIDs := make([]uint64, 0, len(products))
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
	}

	plans, err := s.products.VisibleServerPlans(ctx, productIDs)
	if err != nil {
		return ServerCatalog{}, err
	}
	if len(plans) == 0 {
		return ServerCatalog{Products: catalogProducts(products, nil, nil, nil, nil)}, nil
	}

	planIDs := make([]uint64, 0, len(plans))
	for _, plan := range plans {
		planIDs = append(planIDs, plan.ID)
	}

	prices, err := s.planPrices(ctx, planIDs)
	if err != nil {
		return ServerCatalog{}, err
	}
	regions, err := s.planRegions(ctx, planIDs)
	if err != nil {
		return ServerCatalog{}, err
	}
	templates, err := s.planTemplates(ctx, planIDs)
	if err != nil {
		return ServerCatalog{}, err
	}

	return ServerCatalog{Products: catalogProducts(products, plans, prices, regions, templates)}, nil
}

func (s *ServerCatalogService) planPrices(ctx context.Context, planIDs []uint64) (map[uint64][]ServerCatalogPlanPrice, error) {
	prices, err := s.products.ActivePlanPrices(ctx, planIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[uint64][]ServerCatalogPlanPrice)
	for _, price := range prices {
		result[price.PlanID] = append(result[price.PlanID], ServerCatalogPlanPrice{BillingCycle: price.BillingCycle, PriceCents: price.PriceCents, OriginalPriceCents: price.OriginalPriceCents, Currency: price.Currency})
	}
	return result, nil
}

func (s *ServerCatalogService) planRegions(ctx context.Context, planIDs []uint64) (map[uint64][]ServerCatalogRegion, error) {
	rows, err := s.products.ActivePlanRegions(ctx, planIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[uint64][]ServerCatalogRegion)
	for _, row := range rows {
		result[row.PlanID] = append(result[row.PlanID], ServerCatalogRegion{RegionNo: row.RegionNo, Code: row.Code, Name: row.Name, Country: row.Country, City: row.City, Summary: row.Summary})
	}
	return result, nil
}

func (s *ServerCatalogService) planTemplates(ctx context.Context, planIDs []uint64) (map[uint64][]ServerCatalogOSTemplate, error) {
	rows, err := s.products.ActivePlanOSTemplates(ctx, planIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[uint64][]ServerCatalogOSTemplate)
	for _, row := range rows {
		result[row.PlanID] = append(result[row.PlanID], ServerCatalogOSTemplate{TemplateNo: row.TemplateNo, Code: row.Code, Name: row.Name, OSFamily: row.OSFamily, Distribution: row.Distribution, Version: row.Version, Architecture: row.Architecture, Summary: row.Summary})
	}
	return result, nil
}

func catalogProducts(products []mysqlcatalog.Product, plans []mysqlcatalog.ProductPlan, prices map[uint64][]ServerCatalogPlanPrice, regions map[uint64][]ServerCatalogRegion, templates map[uint64][]ServerCatalogOSTemplate) []ServerCatalogProduct {
	plansByProduct := make(map[uint64][]ServerCatalogPlan)
	for _, plan := range plans {
		planPrices := prices[plan.ID]
		planRegions := regions[plan.ID]
		planTemplates := templates[plan.ID]
		if !domaincatalog.HasRenderablePlanParts(len(planPrices), len(planRegions), len(planTemplates)) {
			continue
		}
		plansByProduct[plan.ProductID] = append(plansByProduct[plan.ProductID], ServerCatalogPlan{PlanNo: plan.PlanNo, Code: plan.Code, Name: plan.Name, Summary: plan.Summary, CPUCores: plan.CPUCores, MemoryMB: plan.MemoryMB, SystemDiskGB: plan.SystemDiskGB, DataDiskGB: plan.DataDiskGB, BandwidthMbps: plan.BandwidthMbps, TrafficGB: plan.TrafficGB, PublicIPCount: plan.PublicIPCount, Virtualization: plan.Virtualization, Architecture: plan.Architecture, IsFeatured: plan.IsFeatured, Status: plan.Status, Prices: planPrices, Regions: planRegions, OSTemplates: planTemplates})
	}

	items := make([]ServerCatalogProduct, 0, len(products))
	for _, product := range products {
		if len(plansByProduct[product.ID]) == 0 {
			continue
		}
		items = append(items, ServerCatalogProduct{ProductNo: product.ProductNo, Slug: product.Slug, Name: product.Name, Summary: product.Summary, Description: product.Description, Plans: plansByProduct[product.ID]})
	}
	return items
}
