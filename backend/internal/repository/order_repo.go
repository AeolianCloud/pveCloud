package repository

import (
	"context"

	"gorm.io/gorm"
	"pvecloud/backend/internal/model"
)

// OrderRepository 封装订单写入与查询。
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单仓储。
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create 创建订单。
func (r *OrderRepository) Create(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// GetByID 查询单个订单。
func (r *OrderRepository) GetByID(ctx context.Context, id uint) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// Update 更新订单。
func (r *OrderRepository) Update(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// ListByUser 查询用户订单，支持状态过滤。
func (r *OrderRepository) ListByUser(ctx context.Context, userID uint, status string) ([]model.Order, error) {
	var orders []model.Order
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC").Find(&orders).Error
	return orders, err
}

// ListForAdmin 查询后台订单列表，支持用户和状态过滤。
func (r *OrderRepository) ListForAdmin(ctx context.Context, userID uint, status string) ([]model.Order, error) {
	var orders []model.Order
	query := r.db.WithContext(ctx).Model(&model.Order{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC").Find(&orders).Error
	return orders, err
}

// DB 暴露底层 DB 供复合事务复用。
func (r *OrderRepository) DB() *gorm.DB {
	return r.db
}
