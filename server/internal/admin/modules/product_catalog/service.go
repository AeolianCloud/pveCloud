package productcatalog

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/support"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
)

const productCatalogObjectType = "product_catalog"

type ProductCatalogService struct {
	db           *gorm.DB
	auditService *AdminAuditService
}

func NewProductCatalogService(db *gorm.DB, auditService *AdminAuditService) *ProductCatalogService {
	if auditService == nil {
		auditService = NewAdminAuditService(db)
	}
	return &ProductCatalogService{db: db, auditService: auditService}
}

func (s *ProductCatalogService) Products(ctx context.Context, query admindto.ProductListQuery) (admindto.PageResponse[admindto.ProductItem], error) {
	page, perPage := support.NormalizePage(query.Page, query.PerPage)
	db := s.db.WithContext(ctx).Model(&models.Product{})
	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("product_no LIKE ? OR slug LIKE ? OR name LIKE ?", like, like, like)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return admindto.PageResponse[admindto.ProductItem]{}, err
	}
	var products []models.Product
	if err := db.Order("sort_order ASC, id DESC").Limit(perPage).Offset((page - 1) * perPage).Find(&products).Error; err != nil {
		return admindto.PageResponse[admindto.ProductItem]{}, err
	}
	items := make([]admindto.ProductItem, 0, len(products))
	for _, product := range products {
		items = append(items, productItem(product))
	}
	return support.PageResponse(items, total, page, perPage), nil
}

func (s *ProductCatalogService) CreateProduct(ctx context.Context, operatorID uint64, req admindto.ProductRequest) (admindto.ProductItem, error) {
	product := productFromRequest(req)
	if product.ProductNo == "" {
		product.ProductNo = generatedNo("PROD")
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&product).Error; err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "product.create", textutil.Uint64String(product.ID), nil, productAudit(product), "创建产品")
	}); err != nil {
		return admindto.ProductItem{}, err
	}
	return productItem(product), nil
}

func (s *ProductCatalogService) UpdateProduct(ctx context.Context, operatorID uint64, id uint64, req admindto.ProductRequest) (admindto.ProductItem, error) {
	var updated models.Product
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		current, err := findForUpdate[models.Product](tx, id)
		if err != nil {
			return err
		}
		updates := productFromRequest(req)
		updates.ProductNo = strings.TrimSpace(req.ProductNo)
		if updates.ProductNo == "" {
			updates.ProductNo = current.ProductNo
		}
		if err := tx.Model(&models.Product{}).Where("id = ?", id).Updates(productUpdateMap(updates)).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).First(&updated).Error; err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "product.update", textutil.Uint64String(id), productAudit(current), productAudit(updated), "更新产品")
	}); err != nil {
		return admindto.ProductItem{}, err
	}
	return productItem(updated), nil
}

func (s *ProductCatalogService) UpdateProductStatus(ctx context.Context, operatorID uint64, id uint64, req admindto.ProductStatusRequest) (admindto.ProductItem, error) {
	var updated models.Product
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		current, err := findForUpdate[models.Product](tx, id)
		if err != nil {
			return err
		}
		if err := tx.Model(&models.Product{}).Where("id = ?", id).Update("status", req.Status).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).First(&updated).Error; err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "product.status.update", textutil.Uint64String(id), productAudit(current), productAudit(updated), "更新产品状态")
	}); err != nil {
		return admindto.ProductItem{}, err
	}
	return productItem(updated), nil
}

func (s *ProductCatalogService) Plans(ctx context.Context, query admindto.ProductPlanListQuery) (admindto.PageResponse[admindto.ProductPlanItem], error) {
	page, perPage := support.NormalizePage(query.Page, query.PerPage)
	db := s.db.WithContext(ctx).Model(&models.ProductPlan{})
	if query.ProductID > 0 {
		db = db.Where("product_id = ?", query.ProductID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("plan_no LIKE ? OR code LIKE ? OR name LIKE ?", like, like, like)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return admindto.PageResponse[admindto.ProductPlanItem]{}, err
	}
	var plans []models.ProductPlan
	if err := db.Order("sort_order ASC, id DESC").Limit(perPage).Offset((page - 1) * perPage).Find(&plans).Error; err != nil {
		return admindto.PageResponse[admindto.ProductPlanItem]{}, err
	}
	items := make([]admindto.ProductPlanItem, 0, len(plans))
	for _, plan := range plans {
		items = append(items, planItem(plan))
	}
	return support.PageResponse(items, total, page, perPage), nil
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
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&plan).Error; err != nil {
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
	var updated models.ProductPlan
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		current, err := findForUpdate[models.ProductPlan](tx, id)
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
		if err := tx.Model(&models.ProductPlan{}).Where("id = ?", id).Updates(planUpdateMap(updates)).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).First(&updated).Error; err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "product_plan.update", textutil.Uint64String(id), planAudit(current), planAudit(updated), "更新服务器套餐")
	}); err != nil {
		return admindto.ProductPlanItem{}, err
	}
	return planItem(updated), nil
}

func (s *ProductCatalogService) UpdatePlanStatus(ctx context.Context, operatorID uint64, id uint64, req admindto.ProductPlanStatusRequest) (admindto.ProductPlanItem, error) {
	var updated models.ProductPlan
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		current, err := findForUpdate[models.ProductPlan](tx, id)
		if err != nil {
			return err
		}
		if err := tx.Model(&models.ProductPlan{}).Where("id = ?", id).Update("status", req.Status).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).First(&updated).Error; err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "product_plan.status.update", textutil.Uint64String(id), planAudit(current), planAudit(updated), "更新服务器套餐状态")
	}); err != nil {
		return admindto.ProductPlanItem{}, err
	}
	return planItem(updated), nil
}

func (s *ProductCatalogService) UpdatePlanPrices(ctx context.Context, operatorID uint64, id uint64, req admindto.PlanPriceListRequest) ([]admindto.PlanPriceItem, error) {
	var prices []models.PlanPrice
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := findForUpdate[models.ProductPlan](tx, id); err != nil {
			return err
		}
		var before []models.PlanPrice
		if err := tx.Where("plan_id = ?", id).Order("sort_order ASC, id ASC").Find(&before).Error; err != nil {
			return err
		}
		if err := tx.Where("plan_id = ?", id).Delete(&models.PlanPrice{}).Error; err != nil {
			return err
		}
		prices = make([]models.PlanPrice, 0, len(req.Prices))
		seen := map[string]bool{}
		for _, input := range req.Prices {
			cycle := strings.TrimSpace(input.BillingCycle)
			if seen[cycle] {
				return apperrors.ErrValidation.WithMessage("计费周期不能重复")
			}
			seen[cycle] = true
			prices = append(prices, models.PlanPrice{PlanID: id, BillingCycle: cycle, PriceCents: input.PriceCents, OriginalPriceCents: input.OriginalPriceCents, Currency: input.Currency, Status: input.Status, SortOrder: input.SortOrder})
		}
		if len(prices) > 0 {
			if err := tx.Create(&prices).Error; err != nil {
				return err
			}
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
	var prices []models.PlanPrice
	if err := s.db.WithContext(ctx).Where("plan_id = ?", id).Order("sort_order ASC, id ASC").Find(&prices).Error; err != nil {
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
	var regions []models.SalesRegion
	if err := s.db.WithContext(ctx).Table("sales_regions AS regions").
		Select("regions.*").
		Joins("JOIN plan_regions AS rel ON rel.region_id = regions.id").
		Where("rel.plan_id = ?", id).
		Order("rel.sort_order ASC, regions.sort_order ASC, regions.id ASC").
		Find(&regions).Error; err != nil {
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
	var templates []models.ServerOSTemplate
	if err := s.db.WithContext(ctx).Table("server_os_templates AS templates").
		Select("templates.*").
		Joins("JOIN plan_os_templates AS rel ON rel.template_id = templates.id").
		Where("rel.plan_id = ?", id).
		Order("rel.sort_order ASC, templates.sort_order ASC, templates.id ASC").
		Find(&templates).Error; err != nil {
		return nil, err
	}
	items := make([]admindto.ServerOSTemplateItem, 0, len(templates))
	for _, template := range templates {
		items = append(items, templateItem(template))
	}
	return items, nil
}

func (s *ProductCatalogService) SalesRegions(ctx context.Context, query admindto.SalesRegionListQuery) ([]admindto.SalesRegionItem, error) {
	db := s.db.WithContext(ctx).Model(&models.SalesRegion{})
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("region_no LIKE ? OR code LIKE ? OR name LIKE ?", like, like, like)
	}
	var regions []models.SalesRegion
	if err := db.Order("sort_order ASC, id DESC").Find(&regions).Error; err != nil {
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
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&region).Error; err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "sales_region.create", textutil.Uint64String(region.ID), nil, regionAudit(region), "创建销售地域")
	}); err != nil {
		return admindto.SalesRegionItem{}, err
	}
	return regionItem(region), nil
}

func (s *ProductCatalogService) UpdateSalesRegion(ctx context.Context, operatorID uint64, id uint64, req admindto.SalesRegionRequest) (admindto.SalesRegionItem, error) {
	var updated models.SalesRegion
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		current, err := findForUpdate[models.SalesRegion](tx, id)
		if err != nil {
			return err
		}
		updates := regionFromRequest(req)
		if updates.RegionNo == "" {
			updates.RegionNo = current.RegionNo
		}
		if err := tx.Model(&models.SalesRegion{}).Where("id = ?", id).Updates(regionUpdateMap(updates)).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).First(&updated).Error; err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "sales_region.update", textutil.Uint64String(id), regionAudit(current), regionAudit(updated), "更新销售地域")
	}); err != nil {
		return admindto.SalesRegionItem{}, err
	}
	return regionItem(updated), nil
}

func (s *ProductCatalogService) ServerOSTemplates(ctx context.Context, query admindto.ServerOSTemplateListQuery) ([]admindto.ServerOSTemplateItem, error) {
	db := s.db.WithContext(ctx).Model(&models.ServerOSTemplate{})
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("template_no LIKE ? OR code LIKE ? OR name LIKE ?", like, like, like)
	}
	var templates []models.ServerOSTemplate
	if err := db.Order("sort_order ASC, id DESC").Find(&templates).Error; err != nil {
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
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&template).Error; err != nil {
			return err
		}
		return s.record(ctx, tx, operatorID, "server_os_template.create", textutil.Uint64String(template.ID), nil, templateAudit(template), "创建服务器系统模板")
	}); err != nil {
		return admindto.ServerOSTemplateItem{}, err
	}
	return templateItem(template), nil
}

func (s *ProductCatalogService) UpdateServerOSTemplate(ctx context.Context, operatorID uint64, id uint64, req admindto.ServerOSTemplateRequest) (admindto.ServerOSTemplateItem, error) {
	var updated models.ServerOSTemplate
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		current, err := findForUpdate[models.ServerOSTemplate](tx, id)
		if err != nil {
			return err
		}
		updates := templateFromRequest(req)
		if updates.TemplateNo == "" {
			updates.TemplateNo = current.TemplateNo
		}
		if err := tx.Model(&models.ServerOSTemplate{}).Where("id = ?", id).Updates(templateUpdateMap(updates)).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).First(&updated).Error; err != nil {
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
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := findForUpdate[models.ProductPlan](tx, planID); err != nil {
			return err
		}
		if relationType == "region" {
			if err := ensureIDsExist[models.SalesRegion](tx, uniqueIDs); err != nil {
				return err
			}
			var before []models.PlanRegion
			if err := tx.Where("plan_id = ?", planID).Find(&before).Error; err != nil {
				return err
			}
			if err := tx.Where("plan_id = ?", planID).Delete(&models.PlanRegion{}).Error; err != nil {
				return err
			}
			relations := make([]models.PlanRegion, 0, len(uniqueIDs))
			for index, id := range uniqueIDs {
				relations = append(relations, models.PlanRegion{PlanID: planID, RegionID: id, Status: "active", SortOrder: index + 1})
			}
			if len(relations) > 0 {
				if err := tx.Create(&relations).Error; err != nil {
					return err
				}
			}
			return s.record(ctx, tx, operatorID, "product_plan.regions.update", textutil.Uint64String(planID), before, relations, "更新套餐销售地域")
		}
		if err := ensureIDsExist[models.ServerOSTemplate](tx, uniqueIDs); err != nil {
			return err
		}
		var before []models.PlanOSTemplate
		if err := tx.Where("plan_id = ?", planID).Find(&before).Error; err != nil {
			return err
		}
		if err := tx.Where("plan_id = ?", planID).Delete(&models.PlanOSTemplate{}).Error; err != nil {
			return err
		}
		relations := make([]models.PlanOSTemplate, 0, len(uniqueIDs))
		for index, id := range uniqueIDs {
			relations = append(relations, models.PlanOSTemplate{PlanID: planID, TemplateID: id, Status: "active", SortOrder: index + 1})
		}
		if len(relations) > 0 {
			if err := tx.Create(&relations).Error; err != nil {
				return err
			}
		}
		return s.record(ctx, tx, operatorID, "product_plan.os_templates.update", textutil.Uint64String(planID), before, relations, "更新套餐系统模板")
	}); err != nil {
		return admindto.PlanRelationResponse{}, err
	}
	return admindto.PlanRelationResponse{PlanID: planID, RelatedIDs: uniqueIDs}, nil
}

func (s *ProductCatalogService) ensureProductExists(ctx context.Context, id uint64) error {
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.Product{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return apperrors.ErrNotFound.WithMessage("产品不存在")
	}
	return nil
}

func (s *ProductCatalogService) ensurePlanExists(ctx context.Context, id uint64) error {
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.ProductPlan{}).Where("id = ?", id).Count(&count).Error; err != nil {
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

func findForUpdate[T any](tx *gorm.DB, id uint64) (T, error) {
	var item T
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return item, apperrors.ErrNotFound.WithMessage("资源不存在")
	}
	return item, err
}

func ensureIDsExist[T any](tx *gorm.DB, ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	var count int64
	if err := tx.Model(new(T)).Where("id IN ?", ids).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(ids)) {
		return apperrors.ErrValidation.WithMessage("关联资源不存在")
	}
	return nil
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

func productFromRequest(req admindto.ProductRequest) models.Product {
	productType := strings.TrimSpace(req.Type)
	if productType == "" {
		productType = "server"
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "draft"
	}
	return models.Product{ProductNo: strings.TrimSpace(req.ProductNo), Type: productType, Slug: strings.TrimSpace(req.Slug), Name: strings.TrimSpace(req.Name), Summary: textutil.NormalizeOptionalString(req.Summary), Description: textutil.NormalizeOptionalString(req.Description), Status: status, Visible: req.Visible, SortOrder: req.SortOrder}
}

func productUpdateMap(product models.Product) map[string]any {
	return map[string]any{"product_no": product.ProductNo, "type": product.Type, "slug": product.Slug, "name": product.Name, "summary": product.Summary, "description": product.Description, "status": product.Status, "visible": product.Visible, "sort_order": product.SortOrder}
}

func planFromRequest(req admindto.ProductPlanRequest) models.ProductPlan {
	return models.ProductPlan{PlanNo: strings.TrimSpace(req.PlanNo), ProductID: req.ProductID, Code: strings.TrimSpace(req.Code), Name: strings.TrimSpace(req.Name), Summary: textutil.NormalizeOptionalString(req.Summary), CPUCores: req.CPUCores, MemoryMB: req.MemoryMB, SystemDiskGB: req.SystemDiskGB, DataDiskGB: req.DataDiskGB, BandwidthMbps: req.BandwidthMbps, TrafficGB: req.TrafficGB, PublicIPCount: req.PublicIPCount, Virtualization: strings.TrimSpace(req.Virtualization), Architecture: strings.TrimSpace(req.Architecture), IsFeatured: req.IsFeatured, Status: strings.TrimSpace(req.Status), Visible: req.Visible, SortOrder: req.SortOrder}
}

func planUpdateMap(plan models.ProductPlan) map[string]any {
	return map[string]any{"plan_no": plan.PlanNo, "product_id": plan.ProductID, "code": plan.Code, "name": plan.Name, "summary": plan.Summary, "cpu_cores": plan.CPUCores, "memory_mb": plan.MemoryMB, "system_disk_gb": plan.SystemDiskGB, "data_disk_gb": plan.DataDiskGB, "bandwidth_mbps": plan.BandwidthMbps, "traffic_gb": plan.TrafficGB, "public_ip_count": plan.PublicIPCount, "virtualization": plan.Virtualization, "architecture": plan.Architecture, "is_featured": plan.IsFeatured, "status": plan.Status, "visible": plan.Visible, "sort_order": plan.SortOrder}
}

func regionFromRequest(req admindto.SalesRegionRequest) models.SalesRegion {
	return models.SalesRegion{RegionNo: strings.TrimSpace(req.RegionNo), Code: strings.TrimSpace(req.Code), Name: strings.TrimSpace(req.Name), Country: textutil.NormalizeOptionalString(req.Country), City: textutil.NormalizeOptionalString(req.City), Summary: textutil.NormalizeOptionalString(req.Summary), Status: strings.TrimSpace(req.Status), Visible: req.Visible, SortOrder: req.SortOrder}
}

func regionUpdateMap(region models.SalesRegion) map[string]any {
	return map[string]any{"region_no": region.RegionNo, "code": region.Code, "name": region.Name, "country": region.Country, "city": region.City, "summary": region.Summary, "status": region.Status, "visible": region.Visible, "sort_order": region.SortOrder}
}

func templateFromRequest(req admindto.ServerOSTemplateRequest) models.ServerOSTemplate {
	return models.ServerOSTemplate{TemplateNo: strings.TrimSpace(req.TemplateNo), Code: strings.TrimSpace(req.Code), Name: strings.TrimSpace(req.Name), OSFamily: strings.TrimSpace(req.OSFamily), Distribution: strings.TrimSpace(req.Distribution), Version: strings.TrimSpace(req.Version), Architecture: strings.TrimSpace(req.Architecture), Summary: textutil.NormalizeOptionalString(req.Summary), Status: strings.TrimSpace(req.Status), Visible: req.Visible, SortOrder: req.SortOrder}
}

func templateUpdateMap(template models.ServerOSTemplate) map[string]any {
	return map[string]any{"template_no": template.TemplateNo, "code": template.Code, "name": template.Name, "os_family": template.OSFamily, "distribution": template.Distribution, "version": template.Version, "architecture": template.Architecture, "summary": template.Summary, "status": template.Status, "visible": template.Visible, "sort_order": template.SortOrder}
}

func productItem(product models.Product) admindto.ProductItem {
	return admindto.ProductItem{ID: product.ID, ProductNo: product.ProductNo, Type: product.Type, Slug: product.Slug, Name: product.Name, Summary: product.Summary, Description: product.Description, Status: product.Status, Visible: product.Visible, SortOrder: product.SortOrder, CreatedAt: product.CreatedAt, UpdatedAt: product.UpdatedAt}
}

func planItem(plan models.ProductPlan) admindto.ProductPlanItem {
	return admindto.ProductPlanItem{ID: plan.ID, PlanNo: plan.PlanNo, ProductID: plan.ProductID, Code: plan.Code, Name: plan.Name, Summary: plan.Summary, CPUCores: plan.CPUCores, MemoryMB: plan.MemoryMB, SystemDiskGB: plan.SystemDiskGB, DataDiskGB: plan.DataDiskGB, BandwidthMbps: plan.BandwidthMbps, TrafficGB: plan.TrafficGB, PublicIPCount: plan.PublicIPCount, Virtualization: plan.Virtualization, Architecture: plan.Architecture, IsFeatured: plan.IsFeatured, Status: plan.Status, Visible: plan.Visible, SortOrder: plan.SortOrder, CreatedAt: plan.CreatedAt, UpdatedAt: plan.UpdatedAt}
}

func priceItem(price models.PlanPrice) admindto.PlanPriceItem {
	return admindto.PlanPriceItem{ID: price.ID, PlanID: price.PlanID, BillingCycle: price.BillingCycle, PriceCents: price.PriceCents, OriginalPriceCents: price.OriginalPriceCents, Currency: price.Currency, Status: price.Status, SortOrder: price.SortOrder, CreatedAt: price.CreatedAt, UpdatedAt: price.UpdatedAt}
}

func regionItem(region models.SalesRegion) admindto.SalesRegionItem {
	return admindto.SalesRegionItem{ID: region.ID, RegionNo: region.RegionNo, Code: region.Code, Name: region.Name, Country: region.Country, City: region.City, Summary: region.Summary, Status: region.Status, Visible: region.Visible, SortOrder: region.SortOrder, CreatedAt: region.CreatedAt, UpdatedAt: region.UpdatedAt}
}

func templateItem(template models.ServerOSTemplate) admindto.ServerOSTemplateItem {
	return admindto.ServerOSTemplateItem{ID: template.ID, TemplateNo: template.TemplateNo, Code: template.Code, Name: template.Name, OSFamily: template.OSFamily, Distribution: template.Distribution, Version: template.Version, Architecture: template.Architecture, Summary: template.Summary, Status: template.Status, Visible: template.Visible, SortOrder: template.SortOrder, CreatedAt: template.CreatedAt, UpdatedAt: template.UpdatedAt}
}

func productAudit(product models.Product) map[string]any {
	return map[string]any{"id": product.ID, "product_no": product.ProductNo, "slug": product.Slug, "name": product.Name, "status": product.Status, "visible": product.Visible}
}
func planAudit(plan models.ProductPlan) map[string]any {
	return map[string]any{"id": plan.ID, "plan_no": plan.PlanNo, "product_id": plan.ProductID, "code": plan.Code, "name": plan.Name, "status": plan.Status, "visible": plan.Visible}
}
func regionAudit(region models.SalesRegion) map[string]any {
	return map[string]any{"id": region.ID, "region_no": region.RegionNo, "code": region.Code, "name": region.Name, "status": region.Status, "visible": region.Visible}
}
func templateAudit(template models.ServerOSTemplate) map[string]any {
	return map[string]any{"id": template.ID, "template_no": template.TemplateNo, "code": template.Code, "name": template.Name, "status": template.Status, "visible": template.Visible}
}
func priceAuditList(prices []models.PlanPrice) []map[string]any {
	items := make([]map[string]any, 0, len(prices))
	for _, price := range prices {
		items = append(items, map[string]any{"id": price.ID, "billing_cycle": price.BillingCycle, "price_cents": price.PriceCents, "original_price_cents": price.OriginalPriceCents, "currency": price.Currency, "status": price.Status})
	}
	return items
}
