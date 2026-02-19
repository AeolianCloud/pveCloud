package repository

import (
	"context"

	"gorm.io/gorm"
	"pvecloud/backend/internal/model"
)

// UserRepository 封装用户数据访问，隔离 SQL 细节。
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储。
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户记录并初始化钱包。
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		wallet := &model.Wallet{UserID: user.ID, Balance: 0, FrozenBalance: 0}
		return tx.Create(wallet).Error
	})
}

// GetByEmail 通过邮箱查询用户。
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID 通过用户 ID 查询用户。
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户字段。
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// List 查询用户列表，支持关键词搜索。
func (r *UserRepository) List(ctx context.Context, keyword string) ([]model.User, error) {
	var users []model.User
	query := r.db.WithContext(ctx).Model(&model.User{})
	if keyword != "" {
		query = query.Where("email LIKE ?", "%"+keyword+"%")
	}
	err := query.Order("id DESC").Find(&users).Error
	return users, err
}
