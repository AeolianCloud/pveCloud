package productcatalog

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	mysqlcatalog "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/catalog"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/support"
)

const productCatalogObjectType = "product_catalog"

type ProductCatalogService struct {
	db           *gorm.DB
	catalog      *mysqlcatalog.Repository
	auditService *AdminAuditService
}

func NewProductCatalogService(db *gorm.DB, auditService *AdminAuditService) *ProductCatalogService {
	if auditService == nil {
		auditService = NewAdminAuditService(db)
	}
	return &ProductCatalogService{db: db, catalog: mysqlcatalog.NewRepository(db), auditService: auditService}
}

func (s *ProductCatalogService) Products(ctx context.Context, query admindto.ProductListQuery) (admindto.PageResponse[admindto.ProductItem], error) {
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	products, total, err := s.catalog.Products(ctx, mysqlcatalog.ProductListFilters{
		Type:    query.Type,
		Status:  query.Status,
		Keyword: query.Keyword,
	}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.ProductItem]{}, err
	}
	items := make([]admindto.ProductItem, 0, len(products))
	for _, product := range products {
		items = append(items, productItem(product))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func (s *ProductCatalogService) CreateProduct(ctx context.Context, operatorID uint64, req admindto.ProductRequest) (admindto.ProductItem, error) {
	product := productFromRequest(req)
	if product.ProductNo == "" {
		product.ProductNo = generatedNo("PROD")
	}
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		if err := s.catalog.CreateProduct(ctx, tx, &product); err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "product.create", textutil.Uint64String(product.ID), nil, productAudit(product), "创建产品")
	}); err != nil {
		return admindto.ProductItem{}, err
	}
	return productItem(product), nil
}

func (s *ProductCatalogService) UpdateProduct(ctx context.Context, operatorID uint64, id uint64, req admindto.ProductRequest) (admindto.ProductItem, error) {
	var updated mysqlcatalog.Product
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.findProductForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		updates := productFromRequest(req)
		updates.ProductNo = strings.TrimSpace(req.ProductNo)
		if updates.ProductNo == "" {
			updates.ProductNo = current.ProductNo
		}
		if strings.TrimSpace(req.Status) == "" {
			updates.Status = current.Status
		}
		if err := s.catalog.UpdateProduct(ctx, tx, id, productUpdateMap(updates)); err != nil {
			return err
		}
		updated, err = s.catalog.FindProductByID(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "product.update", textutil.Uint64String(id), productAudit(current), productAudit(updated), "更新产品")
	}); err != nil {
		return admindto.ProductItem{}, err
	}
	return productItem(updated), nil
}

func (s *ProductCatalogService) UpdateProductStatus(ctx context.Context, operatorID uint64, id uint64, req admindto.ProductStatusRequest) (admindto.ProductItem, error) {
	var updated mysqlcatalog.Product
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.findProductForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		if err := s.catalog.UpdateProductStatus(ctx, tx, id, req.Status); err != nil {
			return err
		}
		updated, err = s.catalog.FindProductByID(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "product.status.update", textutil.Uint64String(id), productAudit(current), productAudit(updated), "更新产品状态")
	}); err != nil {
		return admindto.ProductItem{}, err
	}
	return productItem(updated), nil
}

func (s *ProductCatalogService) Plans(ctx context.Context, query admindto.ProductPlanListQuery) (admindto.PageResponse[admindto.ProductPlanItem], error) {
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	plans, total, err := s.catalog.ProductPlans(ctx, mysqlcatalog.ProductPlanListFilters{
		ProductID: query.ProductID,
		Status:    query.Status,
		Keyword:   query.Keyword,
	}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.ProductPlanItem]{}, err
	}
	items := make([]admindto.ProductPlanItem, 0, len(plans))
	for _, plan := range plans {
		items = append(items, planItem(plan))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func (s *ProductCatalogService) CreatePlan(ctx context.Context, operatorID uint64, req admindto.ProductPlanRequest) (admindto.ProductPlanItem, error) {
	if err := s.ensureProductExists(ctx, req.ProductID); err != nil {
		return admindto.ProductPlanItem{}, err
	}
	plan := planFromRequest(req)
	if plan.PlanNo == "" {
		plan.PlanNo = generatedNo("PLAN")
	}
	if plan.PublicIPCount == 0 {
		plan.PublicIPCount = 1
	}
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		if err := s.catalog.CreatePlan(ctx, tx, &plan); err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "product_plan.create", textutil.Uint64String(plan.ID), nil, planAudit(plan), "创建服务器套餐")
	}); err != nil {
		return admindto.ProductPlanItem{}, err
	}
	return planItem(plan), nil
}

func (s *ProductCatalogService) UpdatePlan(ctx context.Context, operatorID uint64, id uint64, req admindto.ProductPlanRequest) (admindto.ProductPlanItem, error) {
	if err := s.ensureProductExists(ctx, req.ProductID); err != nil {
		return admindto.ProductPlanItem{}, err
	}
	var updated mysqlcatalog.ProductPlan
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.findPlanForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		updates := planFromRequest(req)
		if updates.PlanNo == "" {
			updates.PlanNo = current.PlanNo
		}
		if updates.PublicIPCount == 0 {
			updates.PublicIPCount = 1
		}
		if err := s.catalog.UpdatePlan(ctx, tx, id, planUpdateMap(updates)); err != nil {
			return err
		}
		updated, err = s.catalog.FindPlanByID(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "product_plan.update", textutil.Uint64String(id), planAudit(current), planAudit(updated), "更新服务器套餐")
	}); err != nil {
		return admindto.ProductPlanItem{}, err
	}
	return planItem(updated), nil
}

func (s *ProductCatalogService) UpdatePlanStatus(ctx context.Context, operatorID uint64, id uint64, req admindto.ProductPlanStatusRequest) (admindto.ProductPlanItem, error) {
	var updated mysqlcatalog.ProductPlan
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.findPlanForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		if err := s.catalog.UpdatePlanStatus(ctx, tx, id, req.Status); err != nil {
			return err
		}
		updated, err = s.catalog.FindPlanByID(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "product_plan.status.update", textutil.Uint64String(id), planAudit(current), planAudit(updated), "更新服务器套餐状态")
	}); err != nil {
		return admindto.ProductPlanItem{}, err
	}
	return planItem(updated), nil
}

func (s *ProductCatalogService) UpdatePlanPrices(ctx context.Context, operatorID uint64, id uint64, req admindto.PlanPriceListRequest) ([]admindto.PlanPriceItem, error) {
	var prices []mysqlcatalog.PlanPrice
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		if _, err := s.findPlanForUpdate(ctx, tx, id); err != nil {
			return err
		}
		before, err := s.catalog.PlanPricesByPlanID(ctx, tx, id)
		if err != nil {
			return err
		}
		if err := s.catalog.DeletePlanPrices(ctx, tx, id); err != nil {
			return err
		}
		prices = make([]mysqlcatalog.PlanPrice, 0, len(req.Prices))
		seen := map[string]bool{}
		for _, input := range req.Prices {
			cycle := strings.TrimSpace(input.BillingCycle)
			if seen[cycle] {
				return apperrors.ErrValidation.WithMessage("计费周期不能重复")
			}
			if input.OriginalPriceCents != nil && *input.OriginalPriceCents < input.PriceCents {
				return apperrors.ErrValidation.WithMessage("划线价不能低于售价")
			}
			seen[cycle] = true
			prices = append(prices, mysqlcatalog.PlanPrice{PlanID: id, BillingCycle: cycle, PriceCents: input.PriceCents, OriginalPriceCents: input.OriginalPriceCents, Currency: input.Currency, Status: input.Status, SortOrder: input.SortOrder})
		}
		if err := s.catalog.CreatePlanPrices(ctx, tx, prices); err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "product_plan.prices.update", textutil.Uint64String(id), priceAuditList(before), priceAuditList(prices), "更新套餐价格")
	}); err != nil {
		return nil, err
	}
	items := make([]admindto.PlanPriceItem, 0, len(prices))
	for _, price := range prices {
		items = append(items, priceItem(price))
	}
	return items, nil
}

func (s *ProductCatalogService) PlanPrices(ctx context.Context, id uint64) ([]admindto.PlanPriceItem, error) {
	if err := s.ensurePlanExists(ctx, id); err != nil {
		return nil, err
	}
	prices, err := s.catalog.PlanPrices(ctx, id)
	if err != nil {
		return nil, err
	}
	items := make([]admindto.PlanPriceItem, 0, len(prices))
	for _, price := range prices {
		items = append(items, priceItem(price))
	}
	return items, nil
}

func (s *ProductCatalogService) UpdatePlanRegions(ctx context.Context, operatorID uint64, id uint64, req admindto.PlanRelationRequest) (admindto.PlanRelationResponse, error) {
	return s.updatePlanRelations(ctx, operatorID, id, req.IDs, "region")
}

func (s *ProductCatalogService) PlanRegions(ctx context.Context, id uint64) ([]admindto.SalesRegionItem, error) {
	if err := s.ensurePlanExists(ctx, id); err != nil {
		return nil, err
	}
	regions, err := s.catalog.PlanRegions(ctx, id)
	if err != nil {
		return nil, err
	}
	items := make([]admindto.SalesRegionItem, 0, len(regions))
	for _, region := range regions {
		items = append(items, regionItem(region))
	}
	return items, nil
}

func (s *ProductCatalogService) UpdatePlanOSTemplates(ctx context.Context, operatorID uint64, id uint64, req admindto.PlanRelationRequest) (admindto.PlanRelationResponse, error) {
	return s.updatePlanRelations(ctx, operatorID, id, req.IDs, "os_template")
}

func (s *ProductCatalogService) PlanOSTemplates(ctx context.Context, id uint64) ([]admindto.ServerOSTemplateItem, error) {
	if err := s.ensurePlanExists(ctx, id); err != nil {
		return nil, err
	}
	templates, err := s.catalog.PlanOSTemplates(ctx, id)
	if err != nil {
		return nil, err
	}
	items := make([]admindto.ServerOSTemplateItem, 0, len(templates))
	for _, template := range templates {
		items = append(items, templateItem(template))
	}
	return items, nil
}

func (s *ProductCatalogService) SalesRegions(ctx context.Context, query admindto.SalesRegionListQuery) ([]admindto.SalesRegionItem, error) {
	regions, err := s.catalog.SalesRegions(ctx, mysqlcatalog.SalesRegionListFilters{Status: query.Status, Keyword: query.Keyword})
	if err != nil {
		return nil, err
	}
	items := make([]admindto.SalesRegionItem, 0, len(regions))
	for _, region := range regions {
		items = append(items, regionItem(region))
	}
	return items, nil
}

func (s *ProductCatalogService) CreateSalesRegion(ctx context.Context, operatorID uint64, req admindto.SalesRegionRequest) (admindto.SalesRegionItem, error) {
	region := regionFromRequest(req)
	if region.RegionNo == "" {
		region.RegionNo = generatedNo("REG")
	}
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		if err := s.catalog.CreateSalesRegion(ctx, tx, &region); err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "sales_region.create", textutil.Uint64String(region.ID), nil, regionAudit(region), "创建销售地域")
	}); err != nil {
		return admindto.SalesRegionItem{}, err
	}
	return regionItem(region), nil
}

func (s *ProductCatalogService) UpdateSalesRegion(ctx context.Context, operatorID uint64, id uint64, req admindto.SalesRegionRequest) (admindto.SalesRegionItem, error) {
	var updated mysqlcatalog.SalesRegion
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.findSalesRegionForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		updates := regionFromRequest(req)
		if updates.RegionNo == "" {
			updates.RegionNo = current.RegionNo
		}
		if err := s.catalog.UpdateSalesRegion(ctx, tx, id, regionUpdateMap(updates)); err != nil {
			return err
		}
		updated, err = s.catalog.FindSalesRegionByIDForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "sales_region.update", textutil.Uint64String(id), regionAudit(current), regionAudit(updated), "更新销售地域")
	}); err != nil {
		return admindto.SalesRegionItem{}, err
	}
	return regionItem(updated), nil
}

func (s *ProductCatalogService) ServerOSTemplates(ctx context.Context, query admindto.ServerOSTemplateListQuery) ([]admindto.ServerOSTemplateItem, error) {
	templates, err := s.catalog.ServerOSTemplates(ctx, mysqlcatalog.ServerOSTemplateListFilters{Status: query.Status, Keyword: query.Keyword})
	if err != nil {
		return nil, err
	}
	items := make([]admindto.ServerOSTemplateItem, 0, len(templates))
	for _, template := range templates {
		items = append(items, templateItem(template))
	}
	return items, nil
}

func (s *ProductCatalogService) CreateServerOSTemplate(ctx context.Context, operatorID uint64, req admindto.ServerOSTemplateRequest) (admindto.ServerOSTemplateItem, error) {
	template := templateFromRequest(req)
	if template.TemplateNo == "" {
		template.TemplateNo = generatedNo("TPL")
	}
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		if err := s.catalog.CreateServerOSTemplate(ctx, tx, &template); err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "server_os_template.create", textutil.Uint64String(template.ID), nil, templateAudit(template), "创建服务器系统模板")
	}); err != nil {
		return admindto.ServerOSTemplateItem{}, err
	}
	return templateItem(template), nil
}

func (s *ProductCatalogService) UpdateServerOSTemplate(ctx context.Context, operatorID uint64, id uint64, req admindto.ServerOSTemplateRequest) (admindto.ServerOSTemplateItem, error) {
	var updated mysqlcatalog.ServerOSTemplate
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		current, err := s.findServerOSTemplateForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		updates := templateFromRequest(req)
		if updates.TemplateNo == "" {
			updates.TemplateNo = current.TemplateNo
		}
		if err := s.catalog.UpdateServerOSTemplate(ctx, tx, id, templateUpdateMap(updates)); err != nil {
			return err
		}
		updated, err = s.catalog.FindServerOSTemplateByIDForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "server_os_template.update", textutil.Uint64String(id), templateAudit(current), templateAudit(updated), "更新服务器系统模板")
	}); err != nil {
		return admindto.ServerOSTemplateItem{}, err
	}
	return templateItem(updated), nil
}

func (s *ProductCatalogService) updatePlanRelations(ctx context.Context, operatorID uint64, planID uint64, ids []uint64, relationType string) (admindto.PlanRelationResponse, error) {
	uniqueIDs := uniqueUint64(ids)
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		if _, err := s.findPlanForUpdate(ctx, tx, planID); err != nil {
			return err
		}
		if relationType == "region" {
			if err := s.ensureIDsExist(ctx, tx, &mysqlcatalog.SalesRegion{}, uniqueIDs); err != nil {
				return err
			}
			before, err := s.catalog.PlanRegionRelations(ctx, tx, planID)
			if err != nil {
				return err
			}
			if err := s.catalog.DeletePlanRegions(ctx, tx, planID); err != nil {
				return err
			}
			relations := make([]mysqlcatalog.PlanRegion, 0, len(uniqueIDs))
			for index, id := range uniqueIDs {
				relations = append(relations, mysqlcatalog.PlanRegion{PlanID: planID, RegionID: id, Status: "active", SortOrder: index + 1})
			}
			if err := s.catalog.CreatePlanRegions(ctx, tx, relations); err != nil {
				return err
			}
			return s.record(ctx, tx, operatorID, "product_plan.regions.update", textutil.Uint64String(planID), before, relations, "更新套餐销售地域")
		}
		if err := s.ensureIDsExist(ctx, tx, &mysqlcatalog.ServerOSTemplate{}, uniqueIDs); err != nil {
			return err
		}
		before, err := s.catalog.PlanOSTemplateRelations(ctx, tx, planID)
		if err != nil {
			return err
		}
		if err := s.catalog.DeletePlanOSTemplates(ctx, tx, planID); err != nil {
			return err
		}
		relations := make([]mysqlcatalog.PlanOSTemplate, 0, len(uniqueIDs))
		for index, id := range uniqueIDs {
			relations = append(relations, mysqlcatalog.PlanOSTemplate{PlanID: planID, TemplateID: id, Status: "active", SortOrder: index + 1})
		}
		if err := s.catalog.CreatePlanOSTemplates(ctx, tx, relations); err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "product_plan.os_templates.update", textutil.Uint64String(planID), before, relations, "更新套餐系统模板")
	}); err != nil {
		return admindto.PlanRelationResponse{}, err
	}
	return admindto.PlanRelationResponse{PlanID: planID, RelatedIDs: uniqueIDs}, nil
}

func (s *ProductCatalogService) ensureProductExists(ctx context.Context, id uint64) error {
	count, err := s.catalog.CountByID(ctx, nil, &mysqlcatalog.Product{}, id)
	if err != nil {
		return err
	}
	if count == 0 {
		return apperrors.ErrNotFound.WithMessage("产品不存在")
	}
	return nil
}

func (s *ProductCatalogService) ensurePlanExists(ctx context.Context, id uint64) error {
	count, err := s.catalog.CountByID(ctx, nil, &mysqlcatalog.ProductPlan{}, id)
	if err != nil {
		return err
	}
	if count == 0 {
		return apperrors.ErrNotFound.WithMessage("套餐不存在")
	}
	return nil
}

func (s *ProductCatalogService) record(ctx context.Context, tx *gorm.DB, operatorID uint64, action string, objectID string, before any, after any, remark string) error {
	return s.auditService.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: action, ObjectType: productCatalogObjectType, ObjectID: objectID, BeforeData: before, AfterData: after, Remark: remark})
}

func generatedNo(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func (s *ProductCatalogService) ensureIDsExist(ctx context.Context, tx *gorm.DB, model any, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	count, err := s.catalog.CountByIDs(ctx, tx, model, ids)
	if err != nil {
		return err
	}
	if count != int64(len(ids)) {
		return apperrors.ErrValidation.WithMessage("关联资源不存在")
	}
	return nil
}

func (s *ProductCatalogService) findProductForUpdate(ctx context.Context, tx *gorm.DB, id uint64) (mysqlcatalog.Product, error) {
	product, err := s.catalog.FindProductByIDForUpdate(ctx, tx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return mysqlcatalog.Product{}, apperrors.ErrNotFound.WithMessage("资源不存在")
	}
	return product, err
}

func (s *ProductCatalogService) findPlanForUpdate(ctx context.Context, tx *gorm.DB, id uint64) (mysqlcatalog.ProductPlan, error) {
	plan, err := s.catalog.FindPlanByIDForUpdate(ctx, tx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return mysqlcatalog.ProductPlan{}, apperrors.ErrNotFound.WithMessage("资源不存在")
	}
	return plan, err
}

func (s *ProductCatalogService) findSalesRegionForUpdate(ctx context.Context, tx *gorm.DB, id uint64) (mysqlcatalog.SalesRegion, error) {
	region, err := s.catalog.FindSalesRegionByIDForUpdate(ctx, tx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return mysqlcatalog.SalesRegion{}, apperrors.ErrNotFound.WithMessage("资源不存在")
	}
	return region, err
}

func (s *ProductCatalogService) findServerOSTemplateForUpdate(ctx context.Context, tx *gorm.DB, id uint64) (mysqlcatalog.ServerOSTemplate, error) {
	template, err := s.catalog.FindServerOSTemplateByIDForUpdate(ctx, tx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return mysqlcatalog.ServerOSTemplate{}, apperrors.ErrNotFound.WithMessage("资源不存在")
	}
	return template, err
}

func uniqueUint64(ids []uint64) []uint64 {
	seen := map[uint64]bool{}
	result := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}
	return result
}

func productFromRequest(req admindto.ProductRequest) mysqlcatalog.Product {
	productType := strings.TrimSpace(req.Type)
	if productType == "" {
		productType = "server"
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "draft"
	}
	return mysqlcatalog.Product{ProductNo: strings.TrimSpace(req.ProductNo), Type: productType, Slug: strings.TrimSpace(req.Slug), Name: strings.TrimSpace(req.Name), Summary: textutil.NormalizeOptionalString(req.Summary), Description: textutil.NormalizeOptionalString(req.Description), Status: status, Visible: req.Visible, SortOrder: req.SortOrder}
}

func productUpdateMap(product mysqlcatalog.Product) map[string]any {
	return map[string]any{"product_no": product.ProductNo, "type": product.Type, "slug": product.Slug, "name": product.Name, "summary": product.Summary, "description": product.Description, "status": product.Status, "visible": product.Visible, "sort_order": product.SortOrder}
}

func planFromRequest(req admindto.ProductPlanRequest) mysqlcatalog.ProductPlan {
	return mysqlcatalog.ProductPlan{PlanNo: strings.TrimSpace(req.PlanNo), ProductID: req.ProductID, Code: strings.TrimSpace(req.Code), Name: strings.TrimSpace(req.Name), Summary: textutil.NormalizeOptionalString(req.Summary), CPUCores: req.CPUCores, MemoryMB: req.MemoryMB, SystemDiskGB: req.SystemDiskGB, DataDiskGB: req.DataDiskGB, BandwidthMbps: req.BandwidthMbps, TrafficGB: req.TrafficGB, PublicIPCount: req.PublicIPCount, Virtualization: strings.TrimSpace(req.Virtualization), Architecture: strings.TrimSpace(req.Architecture), IsFeatured: req.IsFeatured, Status: strings.TrimSpace(req.Status), Visible: req.Visible, SortOrder: req.SortOrder}
}

func planUpdateMap(plan mysqlcatalog.ProductPlan) map[string]any {
	return map[string]any{"plan_no": plan.PlanNo, "product_id": plan.ProductID, "code": plan.Code, "name": plan.Name, "summary": plan.Summary, "cpu_cores": plan.CPUCores, "memory_mb": plan.MemoryMB, "system_disk_gb": plan.SystemDiskGB, "data_disk_gb": plan.DataDiskGB, "bandwidth_mbps": plan.BandwidthMbps, "traffic_gb": plan.TrafficGB, "public_ip_count": plan.PublicIPCount, "virtualization": plan.Virtualization, "architecture": plan.Architecture, "is_featured": plan.IsFeatured, "status": plan.Status, "visible": plan.Visible, "sort_order": plan.SortOrder}
}

func regionFromRequest(req admindto.SalesRegionRequest) mysqlcatalog.SalesRegion {
	return mysqlcatalog.SalesRegion{RegionNo: strings.TrimSpace(req.RegionNo), Code: strings.TrimSpace(req.Code), Name: strings.TrimSpace(req.Name), Country: textutil.NormalizeOptionalString(req.Country), City: textutil.NormalizeOptionalString(req.City), Summary: textutil.NormalizeOptionalString(req.Summary), Status: strings.TrimSpace(req.Status), Visible: req.Visible, SortOrder: req.SortOrder}
}

func regionUpdateMap(region mysqlcatalog.SalesRegion) map[string]any {
	return map[string]any{"region_no": region.RegionNo, "code": region.Code, "name": region.Name, "country": region.Country, "city": region.City, "summary": region.Summary, "status": region.Status, "visible": region.Visible, "sort_order": region.SortOrder}
}

func templateFromRequest(req admindto.ServerOSTemplateRequest) mysqlcatalog.ServerOSTemplate {
	return mysqlcatalog.ServerOSTemplate{TemplateNo: strings.TrimSpace(req.TemplateNo), Code: strings.TrimSpace(req.Code), Name: strings.TrimSpace(req.Name), OSFamily: strings.TrimSpace(req.OSFamily), Distribution: strings.TrimSpace(req.Distribution), Version: strings.TrimSpace(req.Version), Architecture: strings.TrimSpace(req.Architecture), Summary: textutil.NormalizeOptionalString(req.Summary), Status: strings.TrimSpace(req.Status), Visible: req.Visible, SortOrder: req.SortOrder}
}

func templateUpdateMap(template mysqlcatalog.ServerOSTemplate) map[string]any {
	return map[string]any{"template_no": template.TemplateNo, "code": template.Code, "name": template.Name, "os_family": template.OSFamily, "distribution": template.Distribution, "version": template.Version, "architecture": template.Architecture, "summary": template.Summary, "status": template.Status, "visible": template.Visible, "sort_order": template.SortOrder}
}

func productItem(product mysqlcatalog.Product) admindto.ProductItem {
	return admindto.ProductItem{ID: product.ID, ProductNo: product.ProductNo, Type: product.Type, Slug: product.Slug, Name: product.Name, Summary: product.Summary, Description: product.Description, Status: product.Status, Visible: product.Visible, SortOrder: product.SortOrder, CreatedAt: product.CreatedAt, UpdatedAt: product.UpdatedAt}
}

func planItem(plan mysqlcatalog.ProductPlan) admindto.ProductPlanItem {
	return admindto.ProductPlanItem{ID: plan.ID, PlanNo: plan.PlanNo, ProductID: plan.ProductID, Code: plan.Code, Name: plan.Name, Summary: plan.Summary, CPUCores: plan.CPUCores, MemoryMB: plan.MemoryMB, SystemDiskGB: plan.SystemDiskGB, DataDiskGB: plan.DataDiskGB, BandwidthMbps: plan.BandwidthMbps, TrafficGB: plan.TrafficGB, PublicIPCount: plan.PublicIPCount, Virtualization: plan.Virtualization, Architecture: plan.Architecture, IsFeatured: plan.IsFeatured, Status: plan.Status, Visible: plan.Visible, SortOrder: plan.SortOrder, CreatedAt: plan.CreatedAt, UpdatedAt: plan.UpdatedAt}
}

func priceItem(price mysqlcatalog.PlanPrice) admindto.PlanPriceItem {
	return admindto.PlanPriceItem{ID: price.ID, PlanID: price.PlanID, BillingCycle: price.BillingCycle, PriceCents: price.PriceCents, OriginalPriceCents: price.OriginalPriceCents, Currency: price.Currency, Status: price.Status, SortOrder: price.SortOrder, CreatedAt: price.CreatedAt, UpdatedAt: price.UpdatedAt}
}

func regionItem(region mysqlcatalog.SalesRegion) admindto.SalesRegionItem {
	return admindto.SalesRegionItem{ID: region.ID, RegionNo: region.RegionNo, Code: region.Code, Name: region.Name, Country: region.Country, City: region.City, Summary: region.Summary, Status: region.Status, Visible: region.Visible, SortOrder: region.SortOrder, CreatedAt: region.CreatedAt, UpdatedAt: region.UpdatedAt}
}

func templateItem(template mysqlcatalog.ServerOSTemplate) admindto.ServerOSTemplateItem {
	return admindto.ServerOSTemplateItem{ID: template.ID, TemplateNo: template.TemplateNo, Code: template.Code, Name: template.Name, OSFamily: template.OSFamily, Distribution: template.Distribution, Version: template.Version, Architecture: template.Architecture, Summary: template.Summary, Status: template.Status, Visible: template.Visible, SortOrder: template.SortOrder, CreatedAt: template.CreatedAt, UpdatedAt: template.UpdatedAt}
}

func productAudit(product mysqlcatalog.Product) map[string]any {
	return map[string]any{"id": product.ID, "product_no": product.ProductNo, "slug": product.Slug, "name": product.Name, "status": product.Status, "visible": product.Visible}
}
func planAudit(plan mysqlcatalog.ProductPlan) map[string]any {
	return map[string]any{"id": plan.ID, "plan_no": plan.PlanNo, "product_id": plan.ProductID, "code": plan.Code, "name": plan.Name, "status": plan.Status, "visible": plan.Visible}
}
func regionAudit(region mysqlcatalog.SalesRegion) map[string]any {
	return map[string]any{"id": region.ID, "region_no": region.RegionNo, "code": region.Code, "name": region.Name, "status": region.Status, "visible": region.Visible}
}
func templateAudit(template mysqlcatalog.ServerOSTemplate) map[string]any {
	return map[string]any{"id": template.ID, "template_no": template.TemplateNo, "code": template.Code, "name": template.Name, "status": template.Status, "visible": template.Visible}
}
func priceAuditList(prices []mysqlcatalog.PlanPrice) []map[string]any {
	items := make([]map[string]any, 0, len(prices))
	for _, price := range prices {
		items = append(items, map[string]any{"id": price.ID, "billing_cycle": price.BillingCycle, "price_cents": price.PriceCents, "original_price_cents": price.OriginalPriceCents, "currency": price.Currency, "status": price.Status})
	}
	return items
}
