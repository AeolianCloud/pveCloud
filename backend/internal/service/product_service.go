package service

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"pvecloud/backend/internal/model"
	"pvecloud/backend/internal/repository"
)

var (
	errProductNotFound = errors.New("商品不存在")
	errProductOffline  = errors.New("商品已下架")
)

// ProductService 负责前后台商品查询与管理逻辑。
type ProductService struct {
	repo *repository.ProductRepository
}

// NewProductService 创建商品服务。
func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// ListPublicProducts 返回前台可见商品。
func (s *ProductService) ListPublicProducts(ctx context.Context, regionID uint) ([]map[string]interface{}, error) {
	products, err := s.repo.ListProducts(ctx, regionID, false)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(products))
	for _, product := range products {
		prices, _ := s.repo.ListPrices(ctx, product.ID)
		result = append(result, map[string]interface{}{"product": product, "prices": prices})
	}
	return result, nil
}

// GetDetail 查询单个商品和所有计费周期定价。
func (s *ProductService) GetDetail(ctx context.Context, id uint, allowNonPublished bool) (map[string]interface{}, error) {
	product, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errProductNotFound
		}
		return nil, err
	}
	if !allowNonPublished && product.Status != "published" {
		return nil, errProductOffline
	}
	prices, err := s.repo.ListPrices(ctx, id)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"product": product, "prices": prices}, nil
}

// Create 创建商品和定价。
func (s *ProductService) Create(ctx context.Context, product *model.Product, prices []model.ProductPrice) error {
	if product.Status == "" {
		product.Status = "draft"
	}
	if err := s.repo.CreateProduct(ctx, product); err != nil {
		return err
	}
	return s.repo.ReplacePrices(ctx, product.ID, prices)
}

// Update 更新商品与定价。
func (s *ProductService) Update(ctx context.Context, product *model.Product, prices []model.ProductPrice) error {
	if err := s.repo.UpdateProduct(ctx, product); err != nil {
		return err
	}
	return s.repo.ReplacePrices(ctx, product.ID, prices)
}

// Delete 删除商品。
func (s *ProductService) Delete(ctx context.Context, id uint) error {
	return s.repo.DeleteProduct(ctx, id)
}

// ListAdminProducts 返回后台商品列表。
func (s *ProductService) ListAdminProducts(ctx context.Context) ([]model.Product, error) {
	return s.repo.ListProducts(ctx, 0, true)
}
