package payment

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct{ db *gorm.DB }

type PaymentFilters struct {
	Provider    string
	Method      string
	Status      string
	OrderNo     string
	PaymentNo   string
	UserKeyword string
	DateFrom    string
	DateTo      string
}

type RefundFilters struct {
	Provider  string
	Status    string
	OrderNo   string
	PaymentNo string
	RefundNo  string
	DateFrom  string
	DateTo    string
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) CreatePayment(ctx context.Context, db *gorm.DB, row *PaymentTransaction) error {
	return r.queryDB(db).WithContext(ctx).Create(row).Error
}

func (r *Repository) PaymentByIdempotency(ctx context.Context, orderID uint64, provider, method, token string) (PaymentTransaction, error) {
	var row PaymentTransaction
	err := r.db.WithContext(ctx).
		Where("order_id = ? AND provider = ? AND method = ? AND client_token = ?", orderID, provider, method, token).
		First(&row).Error
	return row, err
}

func (r *Repository) PaymentByNo(ctx context.Context, paymentNo string) (PaymentTransaction, error) {
	var row PaymentTransaction
	err := r.db.WithContext(ctx).Where("payment_no = ?", paymentNo).First(&row).Error
	return row, err
}

func (r *Repository) UserPaymentByNo(ctx context.Context, userID uint64, paymentNo string) (PaymentTransaction, error) {
	var row PaymentTransaction
	err := r.db.WithContext(ctx).Where("user_id = ? AND payment_no = ?", userID, paymentNo).First(&row).Error
	return row, err
}

func (r *Repository) PaymentForUpdate(ctx context.Context, db *gorm.DB, paymentNo string) (PaymentTransaction, error) {
	var row PaymentTransaction
	err := r.queryDB(db).WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("payment_no = ?", paymentNo).First(&row).Error
	return row, err
}

func (r *Repository) PaymentByUpstreamTradeForUpdate(ctx context.Context, db *gorm.DB, provider, upstreamTradeNo string) (PaymentTransaction, error) {
	var row PaymentTransaction
	err := r.queryDB(db).WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("provider = ? AND upstream_trade_no = ?", provider, upstreamTradeNo).First(&row).Error
	return row, err
}

func (r *Repository) UpdatePayment(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Model(&PaymentTransaction{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) CreateRefund(ctx context.Context, db *gorm.DB, row *RefundTransaction) error {
	return r.queryDB(db).WithContext(ctx).Create(row).Error
}

func (r *Repository) RefundByPaymentID(ctx context.Context, paymentID uint64) (RefundTransaction, error) {
	var row RefundTransaction
	err := r.db.WithContext(ctx).Where("payment_id = ?", paymentID).First(&row).Error
	return row, err
}

func (r *Repository) UpdateRefund(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Model(&RefundTransaction{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) CreateEffect(ctx context.Context, db *gorm.DB, row *PaymentEffect) error {
	return r.queryDB(db).WithContext(ctx).Create(row).Error
}

func (r *Repository) EffectByPaymentID(ctx context.Context, db *gorm.DB, paymentID uint64) (PaymentEffect, error) {
	var row PaymentEffect
	err := r.queryDB(db).WithContext(ctx).Where("payment_id = ?", paymentID).First(&row).Error
	return row, err
}

func (r *Repository) EffectByPaymentIDForUpdate(ctx context.Context, db *gorm.DB, paymentID uint64) (PaymentEffect, error) {
	var row PaymentEffect
	err := r.queryDB(db).WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("payment_id = ?", paymentID).First(&row).Error
	return row, err
}

func (r *Repository) UpdateEffect(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Model(&PaymentEffect{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) ListPayments(ctx context.Context, filters PaymentFilters, limit, offset int) ([]PaymentRow, int64, error) {
	query := r.applyPaymentFilters(r.db.WithContext(ctx).Table("payment_transactions").Joins("JOIN users ON users.id = payment_transactions.user_id").Joins("JOIN orders ON orders.id = payment_transactions.order_id"), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []PaymentRow
	err := query.Select("payment_transactions.*, users.username, users.email, users.display_name, orders.status AS order_status, orders.order_type").
		Order("payment_transactions.created_at DESC, payment_transactions.id DESC").Limit(limit).Offset(offset).Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) ListRefunds(ctx context.Context, filters RefundFilters, limit, offset int) ([]RefundRow, int64, error) {
	query := r.applyRefundFilters(r.db.WithContext(ctx).Table("refund_transactions").Joins("JOIN users ON users.id = refund_transactions.user_id"), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []RefundRow
	err := query.Select("refund_transactions.*, users.username, users.email, users.display_name").
		Order("refund_transactions.created_at DESC, refund_transactions.id DESC").Limit(limit).Offset(offset).Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) queryDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return r.db
}

func (r *Repository) applyPaymentFilters(db *gorm.DB, filters PaymentFilters) *gorm.DB {
	if strings.TrimSpace(filters.Provider) != "" {
		db = db.Where("payment_transactions.provider = ?", strings.TrimSpace(filters.Provider))
	}
	if strings.TrimSpace(filters.Method) != "" {
		db = db.Where("payment_transactions.method = ?", strings.TrimSpace(filters.Method))
	}
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("payment_transactions.status = ?", strings.TrimSpace(filters.Status))
	}
	if strings.TrimSpace(filters.OrderNo) != "" {
		db = db.Where("payment_transactions.order_no = ?", strings.TrimSpace(filters.OrderNo))
	}
	if strings.TrimSpace(filters.PaymentNo) != "" {
		db = db.Where("payment_transactions.payment_no = ?", strings.TrimSpace(filters.PaymentNo))
	}
	if keyword := strings.TrimSpace(filters.UserKeyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("users.username LIKE ? OR users.email LIKE ? OR users.display_name LIKE ?", like, like, like)
	}
	if strings.TrimSpace(filters.DateFrom) != "" {
		db = db.Where("payment_transactions.created_at >= ?", strings.TrimSpace(filters.DateFrom))
	}
	if strings.TrimSpace(filters.DateTo) != "" {
		db = db.Where("payment_transactions.created_at <= ?", strings.TrimSpace(filters.DateTo))
	}
	return db
}

func (r *Repository) applyRefundFilters(db *gorm.DB, filters RefundFilters) *gorm.DB {
	if strings.TrimSpace(filters.Provider) != "" {
		db = db.Where("refund_transactions.provider = ?", strings.TrimSpace(filters.Provider))
	}
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("refund_transactions.status = ?", strings.TrimSpace(filters.Status))
	}
	if strings.TrimSpace(filters.OrderNo) != "" {
		db = db.Where("refund_transactions.order_no = ?", strings.TrimSpace(filters.OrderNo))
	}
	if strings.TrimSpace(filters.PaymentNo) != "" {
		db = db.Where("refund_transactions.payment_no = ?", strings.TrimSpace(filters.PaymentNo))
	}
	if strings.TrimSpace(filters.RefundNo) != "" {
		db = db.Where("refund_transactions.refund_no = ?", strings.TrimSpace(filters.RefundNo))
	}
	if strings.TrimSpace(filters.DateFrom) != "" {
		db = db.Where("refund_transactions.created_at >= ?", strings.TrimSpace(filters.DateFrom))
	}
	if strings.TrimSpace(filters.DateTo) != "" {
		db = db.Where("refund_transactions.created_at <= ?", strings.TrimSpace(filters.DateTo))
	}
	return db
}
