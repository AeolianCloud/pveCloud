package order

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct{ db *gorm.DB }

type ListFilters struct {
	UserID      uint64
	Status      string
	OrderNo     string
	UserKeyword string
	DateFrom    string
	DateTo      string
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) CatalogSelection(ctx context.Context, planNo, cycle, regionNo, templateNo, networkTypeNo string) (CatalogSelection, error) {
	var row CatalogSelection
	err := r.db.WithContext(ctx).Table("product_plans AS plans").
		Select(`products.product_no, products.type AS product_type, products.name AS product_name, products.summary AS product_summary,
			plans.plan_no, plans.code AS plan_code, plans.name AS plan_name, plans.summary AS plan_summary, plans.cpu_cores, plans.memory_mb, plans.system_disk_gb, plans.data_disk_gb, plans.bandwidth_mbps, plans.traffic_gb, plans.public_ip_count, plans.virtualization, plans.architecture,
			prices.billing_cycle, prices.price_cents, prices.original_price_cents, prices.currency,
			regions.region_no, regions.code AS region_code, regions.name AS region_name,
			network_types.network_type_no, network_types.code AS network_type_code, network_types.name AS network_type_name,
			templates.template_no, templates.code AS template_code, templates.name AS template_name, templates.os_family, templates.distribution AS os_distribution, templates.version AS os_version, templates.architecture AS os_architecture`).
		Joins("JOIN products ON products.id = plans.product_id").
		Joins("JOIN plan_prices AS prices ON prices.plan_id = plans.id AND prices.billing_cycle = ? AND prices.status = ?", cycle, "active").
		Joins("JOIN plan_regions AS plan_regions ON plan_regions.plan_id = plans.id AND plan_regions.status = ?", "active").
		Joins("JOIN sales_regions AS regions ON regions.id = plan_regions.region_id AND regions.region_no = ? AND regions.status = ? AND regions.visible = 1", regionNo, "active").
		Joins("JOIN plan_network_types AS plan_network_types ON plan_network_types.plan_id = plans.id AND plan_network_types.status = ?", "active").
		Joins("JOIN network_types AS network_types ON network_types.id = plan_network_types.network_type_id AND network_types.network_type_no = ? AND network_types.status = ? AND network_types.visible = 1", networkTypeNo, "active").
		Joins("JOIN plan_os_templates AS plan_templates ON plan_templates.plan_id = plans.id AND plan_templates.status = ?", "active").
		Joins("JOIN server_os_templates AS templates ON templates.id = plan_templates.template_id AND templates.template_no = ? AND templates.status = ? AND templates.visible = 1", templateNo, "active").
		Where("plans.plan_no = ? AND plans.status = ? AND plans.visible = 1 AND products.type = ? AND products.status = ? AND products.visible = 1", planNo, "active", "server", "active").
		Take(&row).Error
	return row, err
}

func (r *Repository) Create(ctx context.Context, db *gorm.DB, order *Order) error {
	return r.queryDB(db).WithContext(ctx).Create(order).Error
}

func (r *Repository) FindByUserClientToken(ctx context.Context, userID uint64, token string) (Order, error) {
	var order Order
	err := r.db.WithContext(ctx).Where("user_id = ? AND client_token = ?", userID, token).First(&order).Error
	return order, err
}

func (r *Repository) UserOrder(ctx context.Context, userID uint64, orderNo string) (Order, error) {
	var order Order
	err := r.db.WithContext(ctx).Where("user_id = ? AND order_no = ?", userID, orderNo).First(&order).Error
	return order, err
}

func (r *Repository) OrderForUpdate(ctx context.Context, db *gorm.DB, orderNo string) (Order, error) {
	var order Order
	err := r.queryDB(db).WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_no = ?", orderNo).First(&order).Error
	return order, err
}

func (r *Repository) Update(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Model(&Order{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) List(ctx context.Context, filters ListFilters, limit, offset int) ([]OrderRow, int64, error) {
	query := r.applyFilters(r.db.WithContext(ctx).Table("orders").Joins("JOIN users ON users.id = orders.user_id"), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []OrderRow
	if err := query.Select("orders.*, users.username, users.email, users.display_name").Order("orders.created_at DESC, orders.id DESC").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repository) Detail(ctx context.Context, orderNo string) (OrderRow, error) {
	var row OrderRow
	err := r.db.WithContext(ctx).Table("orders").Select("orders.*, users.username, users.email, users.display_name").Joins("JOIN users ON users.id = orders.user_id").Where("orders.order_no = ?", orderNo).Take(&row).Error
	return row, err
}

func (r *Repository) queryDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return r.db
}

func (r *Repository) applyFilters(db *gorm.DB, filters ListFilters) *gorm.DB {
	if filters.UserID > 0 {
		db = db.Where("orders.user_id = ?", filters.UserID)
	}
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("orders.status = ?", strings.TrimSpace(filters.Status))
	}
	if strings.TrimSpace(filters.OrderNo) != "" {
		db = db.Where("orders.order_no = ?", strings.TrimSpace(filters.OrderNo))
	}
	if keyword := strings.TrimSpace(filters.UserKeyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("users.username LIKE ? OR users.email LIKE ? OR users.display_name LIKE ?", like, like, like)
	}
	if strings.TrimSpace(filters.DateFrom) != "" {
		db = db.Where("orders.created_at >= ?", strings.TrimSpace(filters.DateFrom))
	}
	if strings.TrimSpace(filters.DateTo) != "" {
		db = db.Where("orders.created_at <= ?", strings.TrimSpace(filters.DateTo))
	}
	return db
}
