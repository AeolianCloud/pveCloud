package systemconfig

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	mysqlrealname "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/realname"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ConfigRows(ctx context.Context, groupName string) ([]SystemConfig, error) {
	db := r.db.WithContext(ctx).Model(&SystemConfig{})
	if groupName != "" {
		db = db.Where("group_name = ?", groupName)
	}

	var configs []SystemConfig
	if err := db.Order("group_name ASC, id ASC").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

func (r *Repository) PublicSiteConfigRows(ctx context.Context, keys []string, prefix string) ([]SystemConfig, error) {
	var configs []SystemConfig
	if err := r.db.WithContext(ctx).
		Where("config_key IN ? OR config_key LIKE ?", keys, prefix+"%").
		Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

func (r *Repository) ValueByKey(ctx context.Context, key string) (*string, bool, error) {
	var config SystemConfig
	err := r.db.WithContext(ctx).
		Select("config_value").
		Where("config_key = ? AND is_secret = 0", key).
		First(&config).Error
	if err == gorm.ErrRecordNotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return config.ConfigValue, true, nil
}

func (r *Repository) FindByIDForUpdate(ctx context.Context, db *gorm.DB, id uint64) (SystemConfig, error) {
	var config SystemConfig
	err := r.queryDB(db).
		WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&config).Error
	return config, err
}

func (r *Repository) UpdateValueAndReload(ctx context.Context, db *gorm.DB, id uint64, value string) (SystemConfig, error) {
	targetDB := r.queryDB(db).WithContext(ctx)
	if err := targetDB.Model(&SystemConfig{}).Where("id = ?", id).Update("config_value", value).Error; err != nil {
		return SystemConfig{}, err
	}

	var updated SystemConfig
	if err := targetDB.Where("id = ?", id).First(&updated).Error; err != nil {
		return SystemConfig{}, err
	}
	return updated, nil
}

func (r *Repository) RealNameConfigRows(ctx context.Context, db *gorm.DB) ([]SystemConfig, error) {
	var configs []SystemConfig
	if err := r.queryDB(db).
		WithContext(ctx).
		Where("config_key LIKE ?", "real_name.%").
		Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

func (r *Repository) ConfigRowsByPrefix(ctx context.Context, db *gorm.DB, prefix string) ([]SystemConfig, error) {
	var configs []SystemConfig
	if err := r.queryDB(db).
		WithContext(ctx).
		Where("config_key LIKE ?", prefix+"%").
		Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

func (r *Repository) CountRealNameApplicationsByDigestVersion(ctx context.Context, db *gorm.DB, version string) (int64, error) {
	var count int64
	if err := r.queryDB(db).
		WithContext(ctx).
		Model(&mysqlrealname.UserRealNameApplication{}).
		Where("id_number_digest_version = ?", version).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) queryDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return r.db
}
