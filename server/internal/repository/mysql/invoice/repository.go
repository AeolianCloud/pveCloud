package invoice

import (
	"context"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domaininvoice "github.com/AeolianCloud/pveCloud/server/internal/domain/invoice"
	mysqlorder "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/order"
)

type Repository struct{ db *gorm.DB }

type EligibleOrderFilters struct {
	UserID   uint64
	Keyword  string
	DateFrom string
	DateTo   string
}

type ListFilters struct {
	UserID       uint64
	Status       string
	InvoiceNo    string
	OrderNo      string
	UserKeyword  string
	TitleKeyword string
	DateFrom     string
	DateTo       string
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) CreateApplication(ctx context.Context, db *gorm.DB, app *Application) error {
	return r.queryDB(db).WithContext(ctx).Create(app).Error
}

func (r *Repository) CreateApplicationOrder(ctx context.Context, db *gorm.DB, row *ApplicationOrder) error {
	return r.queryDB(db).WithContext(ctx).Create(row).Error
}

func (r *Repository) FindByUserClientToken(ctx context.Context, userID uint64, token string) (Application, error) {
	return r.findByUserClientToken(ctx, nil, userID, token)
}

func (r *Repository) FindByUserClientTokenInTx(ctx context.Context, db *gorm.DB, userID uint64, token string) (Application, error) {
	return r.findByUserClientToken(ctx, db, userID, token)
}

func (r *Repository) findByUserClientToken(ctx context.Context, db *gorm.DB, userID uint64, token string) (Application, error) {
	var row Application
	err := r.queryDB(db).WithContext(ctx).Where("user_id = ? AND client_token = ?", userID, strings.TrimSpace(token)).First(&row).Error
	return row, err
}

func (r *Repository) ApplicationForUpdate(ctx context.Context, db *gorm.DB, invoiceNo string) (Application, error) {
	var row Application
	err := r.queryDB(db).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("invoice_no = ?", strings.TrimSpace(invoiceNo)).
		First(&row).Error
	return row, err
}

func (r *Repository) UpdateApplication(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Model(&Application{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) UpdateOrderStatusSnapshot(ctx context.Context, db *gorm.DB, invoiceID uint64, status string) error {
	return r.queryDB(db).WithContext(ctx).
		Model(&ApplicationOrder{}).
		Where("invoice_id = ?", invoiceID).
		Update("status_snapshot", status).Error
}

func (r *Repository) UserDetail(ctx context.Context, userID uint64, invoiceNo string) (ApplicationRow, error) {
	var row ApplicationRow
	err := r.baseDetailQuery(ctx).
		Where("invoice_applications.user_id = ? AND invoice_applications.invoice_no = ?", userID, strings.TrimSpace(invoiceNo)).
		Take(&row).Error
	return row, err
}

func (r *Repository) Detail(ctx context.Context, invoiceNo string) (ApplicationRow, error) {
	var row ApplicationRow
	err := r.baseDetailQuery(ctx).
		Where("invoice_applications.invoice_no = ?", strings.TrimSpace(invoiceNo)).
		Take(&row).Error
	return row, err
}

func (r *Repository) Orders(ctx context.Context, invoiceID uint64) ([]ApplicationOrder, error) {
	var rows []ApplicationOrder
	err := r.db.WithContext(ctx).
		Where("invoice_id = ?", invoiceID).
		Order("id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *Repository) UserList(ctx context.Context, filters ListFilters, limit, offset int) ([]ApplicationRow, int64, error) {
	query := r.applyListFilters(r.baseDetailQuery(ctx), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []ApplicationRow
	err := query.Order("invoice_applications.created_at DESC, invoice_applications.id DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) AdminList(ctx context.Context, filters ListFilters, limit, offset int) ([]ApplicationRow, int64, error) {
	query := r.applyListFilters(r.baseDetailQuery(ctx), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []ApplicationRow
	err := query.Order("invoice_applications.created_at DESC, invoice_applications.id DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) EligibleOrders(ctx context.Context, filters EligibleOrderFilters, limit, offset int) ([]EligibleOrderRow, int64, error) {
	query := r.applyEligibleFilters(r.eligibleBaseQuery(ctx), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []EligibleOrderRow
	err := query.Select(`orders.id, orders.order_no, orders.order_type, orders.related_instance_no,
			orders.total_amount_cents, orders.currency, orders.payment_status, orders.paid_at,
			orders.product_name, orders.plan_name, 0 AS invoice_occupied`).
		Order("orders.paid_at DESC, orders.id DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) LockUserOrdersByNos(ctx context.Context, db *gorm.DB, userID uint64, orderNos []string) ([]mysqlorder.Order, error) {
	var rows []mysqlorder.Order
	if len(orderNos) == 0 {
		return rows, nil
	}
	err := r.queryDB(db).WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND order_no IN ?", userID, orderNos).
		Order("id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *Repository) ActiveOrderCount(ctx context.Context, db *gorm.DB, orderIDs []uint64) (int64, error) {
	if len(orderIDs) == 0 {
		return 0, nil
	}
	var count int64
	err := r.queryDB(db).WithContext(ctx).
		Model(&ApplicationOrder{}).
		Where("order_id IN ? AND status_snapshot IN ?", orderIDs, []string{domaininvoice.StatusPending, domaininvoice.StatusProcessing, domaininvoice.StatusIssued}).
		Count(&count).Error
	return count, err
}

func (r *Repository) HasActiveOrderInvoice(ctx context.Context, db *gorm.DB, orderID uint64) (bool, error) {
	count, err := r.ActiveOrderCount(ctx, db, []uint64{orderID})
	return count > 0, err
}

func (r *Repository) RefundBlockingOrderCount(ctx context.Context, db *gorm.DB, orderIDs []uint64) (int64, error) {
	if len(orderIDs) == 0 {
		return 0, nil
	}
	var count int64
	err := r.queryDB(db).WithContext(ctx).
		Table("refund_transactions").
		Where("order_id IN ? AND status IN ?", orderIDs, []string{"pending", "succeeded"}).
		Count(&count).Error
	return count, err
}

func (r *Repository) FileReferenceExists(ctx context.Context, fileID uint64, invoiceID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("file_attachment_references").
		Where("file_id = ? AND ref_type = ? AND ref_id = ?", fileID, domaininvoice.FileRefType, uint64String(invoiceID)).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) baseDetailQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("invoice_applications").
		Select(`invoice_applications.*, users.username, users.email AS user_email, users.display_name,
			(SELECT COUNT(*) FROM invoice_application_orders WHERE invoice_application_orders.invoice_id = invoice_applications.id) AS order_count`).
		Joins("JOIN users ON users.id = invoice_applications.user_id")
}

func (r *Repository) eligibleBaseQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("orders").
		Where("orders.currency = ?", "CNY").
		Where("orders.payment_status IN ?", []string{"paid", "manual_confirmed"}).
		Where("orders.status NOT IN ?", []string{"cancelled", "closed"}).
		Where(`NOT EXISTS (
			SELECT 1 FROM invoice_application_orders iao
			WHERE iao.order_id = orders.id AND iao.status_snapshot IN ?
		)`, []string{domaininvoice.StatusPending, domaininvoice.StatusProcessing, domaininvoice.StatusIssued}).
		Where(`NOT EXISTS (
			SELECT 1 FROM refund_transactions rt
			WHERE rt.order_id = orders.id AND rt.status IN ?
		)`, []string{"pending", "succeeded"})
}

func (r *Repository) applyEligibleFilters(db *gorm.DB, filters EligibleOrderFilters) *gorm.DB {
	if filters.UserID > 0 {
		db = db.Where("orders.user_id = ?", filters.UserID)
	}
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("orders.order_no LIKE ? OR orders.product_name LIKE ? OR orders.plan_name LIKE ?", like, like, like)
	}
	if strings.TrimSpace(filters.DateFrom) != "" {
		db = db.Where("orders.paid_at >= ?", strings.TrimSpace(filters.DateFrom))
	}
	if strings.TrimSpace(filters.DateTo) != "" {
		db = db.Where("orders.paid_at <= ?", strings.TrimSpace(filters.DateTo))
	}
	return db
}

func (r *Repository) applyListFilters(db *gorm.DB, filters ListFilters) *gorm.DB {
	if filters.UserID > 0 {
		db = db.Where("invoice_applications.user_id = ?", filters.UserID)
	}
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("invoice_applications.status = ?", strings.TrimSpace(filters.Status))
	}
	if value := strings.TrimSpace(filters.InvoiceNo); value != "" {
		db = db.Where("invoice_applications.invoice_no LIKE ?", "%"+value+"%")
	}
	if value := strings.TrimSpace(filters.OrderNo); value != "" {
		db = db.Where(`EXISTS (
			SELECT 1 FROM invoice_application_orders iao
			WHERE iao.invoice_id = invoice_applications.id AND iao.order_no LIKE ?
		)`, "%"+value+"%")
	}
	if keyword := strings.TrimSpace(filters.UserKeyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("users.username LIKE ? OR users.email LIKE ? OR users.display_name LIKE ?", like, like, like)
	}
	if keyword := strings.TrimSpace(filters.TitleKeyword); keyword != "" {
		db = db.Where("invoice_applications.title LIKE ?", "%"+keyword+"%")
	}
	if strings.TrimSpace(filters.DateFrom) != "" {
		db = db.Where("invoice_applications.created_at >= ?", strings.TrimSpace(filters.DateFrom))
	}
	if strings.TrimSpace(filters.DateTo) != "" {
		db = db.Where("invoice_applications.created_at <= ?", strings.TrimSpace(filters.DateTo))
	}
	return db
}

func (r *Repository) queryDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return r.db
}

func uint64String(value uint64) string {
	return strconv.FormatUint(value, 10)
}
