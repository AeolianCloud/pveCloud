package catalog

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

type ProductListFilters struct {
	Type    string
	Status  string
	Keyword string
}

type ProductPlanListFilters struct {
	ProductID uint64
	Status    string
	Keyword   string
}

type SalesRegionListFilters struct {
	Status  string
	Keyword string
}

type ServerOSTemplateListFilters struct {
	Status  string
	Keyword string
}

type PlanRegionRow struct {
	PlanID   uint64
	RegionNo string
	Code     string
	Name     string
	Country  *string
	City     *string
	Summary  *string
}

type PlanOSTemplateRow struct {
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

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Products(ctx context.Context, filters ProductListFilters, limit int, offset int) ([]Product, int64, error) {
	query := r.applyProductFilters(r.db.WithContext(ctx).Model(&Product{}), filters)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var products []Product
	if err := query.Order("sort_order ASC, id DESC").Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *Repository) ProductPlans(ctx context.Context, filters ProductPlanListFilters, limit int, offset int) ([]ProductPlan, int64, error) {
	query := r.applyProductPlanFilters(r.db.WithContext(ctx).Model(&ProductPlan{}), filters)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var plans []ProductPlan
	if err := query.Order("sort_order ASC, id DESC").Limit(limit).Offset(offset).Find(&plans).Error; err != nil {
		return nil, 0, err
	}
	return plans, total, nil
}

func (r *Repository) SalesRegions(ctx context.Context, filters SalesRegionListFilters) ([]SalesRegion, error) {
	query := r.applySalesRegionFilters(r.db.WithContext(ctx).Model(&SalesRegion{}), filters)
	var regions []SalesRegion
	if err := query.Order("sort_order ASC, id DESC").Find(&regions).Error; err != nil {
		return nil, err
	}
	return regions, nil
}

func (r *Repository) ServerOSTemplates(ctx context.Context, filters ServerOSTemplateListFilters) ([]ServerOSTemplate, error) {
	query := r.applyServerOSTemplateFilters(r.db.WithContext(ctx).Model(&ServerOSTemplate{}), filters)
	var templates []ServerOSTemplate
	if err := query.Order("sort_order ASC, id DESC").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (r *Repository) CreateProduct(ctx context.Context, db *gorm.DB, product *Product) error {
	return r.queryDB(db).WithContext(ctx).Create(product).Error
}

func (r *Repository) CreatePlan(ctx context.Context, db *gorm.DB, plan *ProductPlan) error {
	return r.queryDB(db).WithContext(ctx).Create(plan).Error
}

func (r *Repository) CreateSalesRegion(ctx context.Context, db *gorm.DB, region *SalesRegion) error {
	return r.queryDB(db).WithContext(ctx).Create(region).Error
}

func (r *Repository) CreateServerOSTemplate(ctx context.Context, db *gorm.DB, template *ServerOSTemplate) error {
	return r.queryDB(db).WithContext(ctx).Create(template).Error
}

func (r *Repository) FindProductByID(ctx context.Context, db *gorm.DB, id uint64) (Product, error) {
	var product Product
	err := r.queryDB(db).WithContext(ctx).Where("id = ?", id).First(&product).Error
	return product, err
}

func (r *Repository) FindProductByIDForUpdate(ctx context.Context, db *gorm.DB, id uint64) (Product, error) {
	var product Product
	err := r.queryDB(db).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&product).Error
	return product, err
}

func (r *Repository) FindPlanByID(ctx context.Context, db *gorm.DB, id uint64) (ProductPlan, error) {
	var plan ProductPlan
	err := r.queryDB(db).WithContext(ctx).Where("id = ?", id).First(&plan).Error
	return plan, err
}

func (r *Repository) FindPlanByIDForUpdate(ctx context.Context, db *gorm.DB, id uint64) (ProductPlan, error) {
	var plan ProductPlan
	err := r.queryDB(db).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&plan).Error
	return plan, err
}

func (r *Repository) FindSalesRegionByIDForUpdate(ctx context.Context, db *gorm.DB, id uint64) (SalesRegion, error) {
	var region SalesRegion
	err := r.queryDB(db).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&region).Error
	return region, err
}

func (r *Repository) FindServerOSTemplateByIDForUpdate(ctx context.Context, db *gorm.DB, id uint64) (ServerOSTemplate, error) {
	var template ServerOSTemplate
	err := r.queryDB(db).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&template).Error
	return template, err
}

func (r *Repository) UpdateProduct(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Model(&Product{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) UpdatePlan(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Model(&ProductPlan{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) UpdateSalesRegion(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Model(&SalesRegion{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) UpdateServerOSTemplate(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Model(&ServerOSTemplate{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) UpdateProductStatus(ctx context.Context, db *gorm.DB, id uint64, status string) error {
	return r.queryDB(db).WithContext(ctx).Model(&Product{}).Where("id = ?", id).Update("status", status).Error
}

func (r *Repository) UpdatePlanStatus(ctx context.Context, db *gorm.DB, id uint64, status string) error {
	return r.queryDB(db).WithContext(ctx).Model(&ProductPlan{}).Where("id = ?", id).Update("status", status).Error
}

func (r *Repository) DeletePlanPrices(ctx context.Context, db *gorm.DB, planID uint64) error {
	return r.queryDB(db).WithContext(ctx).Where("plan_id = ?", planID).Delete(&PlanPrice{}).Error
}

func (r *Repository) DeletePlanRegions(ctx context.Context, db *gorm.DB, planID uint64) error {
	return r.queryDB(db).WithContext(ctx).Where("plan_id = ?", planID).Delete(&PlanRegion{}).Error
}

func (r *Repository) DeletePlanOSTemplates(ctx context.Context, db *gorm.DB, planID uint64) error {
	return r.queryDB(db).WithContext(ctx).Where("plan_id = ?", planID).Delete(&PlanOSTemplate{}).Error
}

func (r *Repository) CreatePlanPrices(ctx context.Context, db *gorm.DB, prices []PlanPrice) error {
	if len(prices) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Create(&prices).Error
}

func (r *Repository) CreatePlanRegions(ctx context.Context, db *gorm.DB, regions []PlanRegion) error {
	if len(regions) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Create(&regions).Error
}

func (r *Repository) CreatePlanOSTemplates(ctx context.Context, db *gorm.DB, templates []PlanOSTemplate) error {
	if len(templates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Create(&templates).Error
}

func (r *Repository) PlanPrices(ctx context.Context, planID uint64) ([]PlanPrice, error) {
	return r.PlanPricesByPlanID(ctx, nil, planID)
}

func (r *Repository) PlanPricesByPlanID(ctx context.Context, db *gorm.DB, planID uint64) ([]PlanPrice, error) {
	var prices []PlanPrice
	if err := r.queryDB(db).WithContext(ctx).Where("plan_id = ?", planID).Order("sort_order ASC, id ASC").Find(&prices).Error; err != nil {
		return nil, err
	}
	return prices, nil
}

func (r *Repository) PlanRegionRelations(ctx context.Context, db *gorm.DB, planID uint64) ([]PlanRegion, error) {
	var rows []PlanRegion
	if err := r.queryDB(db).WithContext(ctx).Where("plan_id = ?", planID).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) PlanOSTemplateRelations(ctx context.Context, db *gorm.DB, planID uint64) ([]PlanOSTemplate, error) {
	var rows []PlanOSTemplate
	if err := r.queryDB(db).WithContext(ctx).Where("plan_id = ?", planID).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) PlanRegionRows(ctx context.Context, planID uint64) ([]PlanRegionRow, error) {
	var rows []PlanRegionRow
	if err := r.db.WithContext(ctx).Table("plan_regions AS rel").
		Select("rel.plan_id, regions.region_no, regions.code, regions.name, regions.country, regions.city, regions.summary").
		Joins("JOIN sales_regions AS regions ON regions.id = rel.region_id").
		Where("rel.plan_id = ?", planID).
		Order("rel.sort_order ASC, regions.sort_order ASC, regions.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) PlanRegions(ctx context.Context, planID uint64) ([]SalesRegion, error) {
	var regions []SalesRegion
	if err := r.db.WithContext(ctx).Table("sales_regions AS regions").
		Select("regions.*").
		Joins("JOIN plan_regions AS rel ON rel.region_id = regions.id").
		Where("rel.plan_id = ?", planID).
		Order("rel.sort_order ASC, regions.sort_order ASC, regions.id ASC").
		Find(&regions).Error; err != nil {
		return nil, err
	}
	return regions, nil
}

func (r *Repository) PlanOSTemplateRows(ctx context.Context, planID uint64) ([]PlanOSTemplateRow, error) {
	var rows []PlanOSTemplateRow
	if err := r.db.WithContext(ctx).Table("plan_os_templates AS rel").
		Select("rel.plan_id, templates.template_no, templates.code, templates.name, templates.os_family, templates.distribution, templates.version, templates.architecture, templates.summary").
		Joins("JOIN server_os_templates AS templates ON templates.id = rel.template_id").
		Where("rel.plan_id = ?", planID).
		Order("rel.sort_order ASC, templates.sort_order ASC, templates.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) PlanOSTemplates(ctx context.Context, planID uint64) ([]ServerOSTemplate, error) {
	var templates []ServerOSTemplate
	if err := r.db.WithContext(ctx).Table("server_os_templates AS templates").
		Select("templates.*").
		Joins("JOIN plan_os_templates AS rel ON rel.template_id = templates.id").
		Where("rel.plan_id = ?", planID).
		Order("rel.sort_order ASC, templates.sort_order ASC, templates.id ASC").
		Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (r *Repository) CountByID(ctx context.Context, db *gorm.DB, model any, id uint64) (int64, error) {
	var count int64
	if err := r.queryDB(db).WithContext(ctx).Model(model).Where("id = ?", id).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) CountByIDs(ctx context.Context, db *gorm.DB, model any, ids []uint64) (int64, error) {
	var count int64
	if err := r.queryDB(db).WithContext(ctx).Model(model).Where("id IN ?", ids).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) ActiveServerProducts(ctx context.Context) ([]Product, error) {
	var products []Product
	if err := r.db.WithContext(ctx).
		Where("type = ? AND status = ? AND visible = 1", "server", "active").
		Order("sort_order ASC, id ASC").
		Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *Repository) VisibleServerPlans(ctx context.Context, productIDs []uint64) ([]ProductPlan, error) {
	var plans []ProductPlan
	if err := r.db.WithContext(ctx).
		Where("product_id IN ? AND status IN ? AND visible = 1", productIDs, []string{"active", "sold_out"}).
		Order("is_featured DESC, sort_order ASC, id ASC").
		Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *Repository) ActivePlanPrices(ctx context.Context, planIDs []uint64) ([]PlanPrice, error) {
	var prices []PlanPrice
	if err := r.db.WithContext(ctx).
		Where("plan_id IN ? AND status = ?", planIDs, "active").
		Order("sort_order ASC, id ASC").
		Find(&prices).Error; err != nil {
		return nil, err
	}
	return prices, nil
}

func (r *Repository) ActivePlanRegions(ctx context.Context, planIDs []uint64) ([]PlanRegionRow, error) {
	var rows []PlanRegionRow
	if err := r.db.WithContext(ctx).Table("plan_regions AS rel").
		Select("rel.plan_id, regions.region_no, regions.code, regions.name, regions.country, regions.city, regions.summary").
		Joins("JOIN sales_regions AS regions ON regions.id = rel.region_id").
		Where("rel.plan_id IN ? AND rel.status = ? AND regions.status = ? AND regions.visible = 1", planIDs, "active", "active").
		Order("rel.sort_order ASC, regions.sort_order ASC, regions.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) ActivePlanOSTemplates(ctx context.Context, planIDs []uint64) ([]PlanOSTemplateRow, error) {
	var rows []PlanOSTemplateRow
	if err := r.db.WithContext(ctx).Table("plan_os_templates AS rel").
		Select("rel.plan_id, templates.template_no, templates.code, templates.name, templates.os_family, templates.distribution, templates.version, templates.architecture, templates.summary").
		Joins("JOIN server_os_templates AS templates ON templates.id = rel.template_id").
		Where("rel.plan_id IN ? AND rel.status = ? AND templates.status = ? AND templates.visible = 1", planIDs, "active", "active").
		Order("rel.sort_order ASC, templates.sort_order ASC, templates.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) applyProductFilters(db *gorm.DB, filters ProductListFilters) *gorm.DB {
	if strings.TrimSpace(filters.Type) != "" {
		db = db.Where("type = ?", strings.TrimSpace(filters.Type))
	}
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("status = ?", strings.TrimSpace(filters.Status))
	}
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("product_no LIKE ? OR slug LIKE ? OR name LIKE ?", like, like, like)
	}
	return db
}

func (r *Repository) applyProductPlanFilters(db *gorm.DB, filters ProductPlanListFilters) *gorm.DB {
	if filters.ProductID > 0 {
		db = db.Where("product_id = ?", filters.ProductID)
	}
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("status = ?", strings.TrimSpace(filters.Status))
	}
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("plan_no LIKE ? OR code LIKE ? OR name LIKE ?", like, like, like)
	}
	return db
}

func (r *Repository) applySalesRegionFilters(db *gorm.DB, filters SalesRegionListFilters) *gorm.DB {
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("status = ?", strings.TrimSpace(filters.Status))
	}
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("region_no LIKE ? OR code LIKE ? OR name LIKE ?", like, like, like)
	}
	return db
}

func (r *Repository) applyServerOSTemplateFilters(db *gorm.DB, filters ServerOSTemplateListFilters) *gorm.DB {
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("status = ?", strings.TrimSpace(filters.Status))
	}
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("template_no LIKE ? OR code LIKE ? OR name LIKE ?", like, like, like)
	}
	return db
}

func (r *Repository) queryDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return r.db
}
