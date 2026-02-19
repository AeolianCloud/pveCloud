package repository

import (
	"context"

	"gorm.io/gorm"
	"pvecloud/backend/internal/model"
)

// ProductRepository 封装商品与价格的数据访问。
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository 创建商品仓储。
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// CreateProduct 创建商品。
func (r *ProductRepository) CreateProduct(ctx context.Context, product *model.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

// UpdateProduct 更新商品信息。
func (r *ProductRepository) UpdateProduct(ctx context.Context, product *model.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

// DeleteProduct 删除商品。
func (r *ProductRepository) DeleteProduct(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Product{}, id).Error
}

// GetProductByID 查询单个商品。
func (r *ProductRepository) GetProductByID(ctx context.Context, id uint) (*model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// ListProducts 查询商品列表，allowNonPublished 用于后台放开草稿/下架查询。
func (r *ProductRepository) ListProducts(ctx context.Context, regionID uint, allowNonPublished bool) ([]model.Product, error) {
	var products []model.Product
	query := r.db.WithContext(ctx).Model(&model.Product{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if !allowNonPublished {
		query = query.Where("status = ?", "published")
	}
	err := query.Order("id DESC").Find(&products).Error
	return products, err
}

// ReplacePrices 使用事务覆盖商品的定价行。
func (r *ProductRepository) ReplacePrices(ctx context.Context, productID uint, prices []model.ProductPrice) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_id = ?", productID).Delete(&model.ProductPrice{}).Error; err != nil {
			return err
		}
		for i := range prices {
			prices[i].ProductID = productID
		}
		if len(prices) == 0 {
			return nil
		}
		return tx.Create(&prices).Error
	})
}

// ListPrices 查询某商品全部计费周期价格。
func (r *ProductRepository) ListPrices(ctx context.Context, productID uint) ([]model.ProductPrice, error) {
	var prices []model.ProductPrice
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).Order("id ASC").Find(&prices).Error
	return prices, err
}
