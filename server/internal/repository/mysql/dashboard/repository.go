package dashboard

import (
	"context"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

type MetricQuery struct {
	Key   string
	Title string
	Table string
	Where string
	Unit  string
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Count(ctx context.Context, table string, where string) (int64, error) {
	var value int64
	if err := r.db.WithContext(ctx).Table(table).Where(where).Count(&value).Error; err != nil {
		return 0, err
	}
	return value, nil
}
