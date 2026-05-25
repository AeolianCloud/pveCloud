package dashboard

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) count(ctx context.Context, table string, apply func(*gorm.DB) *gorm.DB) (int64, error) {
	var value int64
	query := r.db.WithContext(ctx).Table(table)
	if apply != nil {
		query = apply(query)
	}
	if err := query.Count(&value).Error; err != nil {
		return 0, err
	}
	return value, nil
}

func (r *Repository) CountActiveAdmins(ctx context.Context) (int64, error) {
	return r.count(ctx, "admin_users", func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status = ? AND deleted_at IS NULL", "active")
	})
}

func (r *Repository) CountActiveRoles(ctx context.Context) (int64, error) {
	return r.count(ctx, "admin_roles", func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status = ?", "active")
	})
}

func (r *Repository) CountActiveSessions(ctx context.Context, now time.Time) (int64, error) {
	return r.count(ctx, "admin_sessions", func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status = ? AND expires_at > ?", "active", now)
	})
}

func (r *Repository) CountAuditLogsSince(ctx context.Context, since time.Time) (int64, error) {
	return r.count(ctx, "admin_audit_logs", func(tx *gorm.DB) *gorm.DB {
		return tx.Where("created_at >= ?", since)
	})
}

func (r *Repository) CountPendingOrders(ctx context.Context) (int64, error) {
	return r.count(ctx, "orders", func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status = ?", "pending")
	})
}

func (r *Repository) CountOrderErrors(ctx context.Context) (int64, error) {
	return r.count(ctx, "orders", func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status = ?", "error")
	})
}

func (r *Repository) CountInstanceErrors(ctx context.Context) (int64, error) {
	return r.count(ctx, "instances", func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status = ?", "error")
	})
}

func (r *Repository) CountFailedAsyncTasks(ctx context.Context) (int64, error) {
	return r.count(ctx, "async_tasks", func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status = ?", "failed")
	})
}

func (r *Repository) CountPendingTickets(ctx context.Context) (int64, error) {
	return r.count(ctx, "tickets", func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status = ?", "waiting_admin")
	})
}

func (r *Repository) CountInvoiceTodo(ctx context.Context) (int64, error) {
	return r.count(ctx, "invoice_applications", func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status IN ?", []string{"pending", "processing"})
	})
}

func (r *Repository) CountPaymentExceptions(ctx context.Context) (int64, error) {
	failedPayments, err := r.count(ctx, "payment_transactions", func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status = ?", "failed")
	})
	if err != nil {
		return 0, err
	}
	actionableRefunds, err := r.count(ctx, "refund_transactions", func(tx *gorm.DB) *gorm.DB {
		return tx.Where("status IN ?", []string{"pending", "failed"})
	})
	if err != nil {
		return 0, err
	}
	return failedPayments + actionableRefunds, nil
}
