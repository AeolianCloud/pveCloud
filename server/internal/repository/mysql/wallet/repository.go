package wallet

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct{ db *gorm.DB }

type AccountFilters struct {
	WalletNo    string
	Status      string
	UserKeyword string
}

type LedgerFilters struct {
	WalletNo    string
	UserKeyword string
	Direction   string
	EntryType   string
	RelatedNo   string
	DateFrom    string
	DateTo      string
}

type RechargeFilters struct {
	WalletNo    string
	UserKeyword string
	Provider    string
	Method      string
	Status      string
	RechargeNo  string
	DateFrom    string
	DateTo      string
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) AccountByUserCurrency(ctx context.Context, userID uint64, currency string) (Account, error) {
	var row Account
	err := r.db.WithContext(ctx).Where("user_id = ? AND currency = ?", userID, currency).First(&row).Error
	return row, err
}

func (r *Repository) AccountByUserCurrencyForUpdate(ctx context.Context, db *gorm.DB, userID uint64, currency string) (Account, error) {
	var row Account
	err := r.queryDB(db).WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND currency = ?", userID, currency).First(&row).Error
	return row, err
}

func (r *Repository) AccountByNo(ctx context.Context, walletNo string) (Account, error) {
	var row Account
	err := r.db.WithContext(ctx).Where("wallet_no = ?", strings.TrimSpace(walletNo)).First(&row).Error
	return row, err
}

func (r *Repository) AccountByNoForUpdate(ctx context.Context, db *gorm.DB, walletNo string) (Account, error) {
	var row Account
	err := r.queryDB(db).WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("wallet_no = ?", strings.TrimSpace(walletNo)).First(&row).Error
	return row, err
}

func (r *Repository) CreateAccount(ctx context.Context, db *gorm.DB, row *Account) error {
	return r.queryDB(db).WithContext(ctx).Create(row).Error
}

func (r *Repository) UpdateAccount(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Model(&Account{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) CreateLedgerEntry(ctx context.Context, db *gorm.DB, row *LedgerEntry) error {
	return r.queryDB(db).WithContext(ctx).Create(row).Error
}

func (r *Repository) LedgerByIdempotency(ctx context.Context, db *gorm.DB, walletID uint64, key string) (LedgerEntry, error) {
	var row LedgerEntry
	err := r.queryDB(db).WithContext(ctx).
		Where("wallet_id = ? AND idempotency_key = ?", walletID, strings.TrimSpace(key)).First(&row).Error
	return row, err
}

func (r *Repository) CreateRecharge(ctx context.Context, db *gorm.DB, row *Recharge) error {
	return r.queryDB(db).WithContext(ctx).Create(row).Error
}

func (r *Repository) RechargeByNo(ctx context.Context, rechargeNo string) (Recharge, error) {
	var row Recharge
	err := r.db.WithContext(ctx).Where("recharge_no = ?", strings.TrimSpace(rechargeNo)).First(&row).Error
	return row, err
}

func (r *Repository) UserRechargeByNo(ctx context.Context, userID uint64, rechargeNo string) (Recharge, error) {
	var row Recharge
	err := r.db.WithContext(ctx).Where("user_id = ? AND recharge_no = ?", userID, strings.TrimSpace(rechargeNo)).First(&row).Error
	return row, err
}

func (r *Repository) RechargeByIdempotency(ctx context.Context, walletID uint64, provider, method, token string) (Recharge, error) {
	var row Recharge
	err := r.db.WithContext(ctx).
		Where("wallet_id = ? AND provider = ? AND method = ? AND client_token = ?", walletID, provider, method, strings.TrimSpace(token)).
		First(&row).Error
	return row, err
}

func (r *Repository) RechargeByIdempotencyForUpdate(ctx context.Context, db *gorm.DB, walletID uint64, provider, method, token string) (Recharge, error) {
	var row Recharge
	err := r.queryDB(db).WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("wallet_id = ? AND provider = ? AND method = ? AND client_token = ?", walletID, provider, method, strings.TrimSpace(token)).
		First(&row).Error
	return row, err
}

func (r *Repository) RechargeForUpdate(ctx context.Context, db *gorm.DB, rechargeNo string) (Recharge, error) {
	var row Recharge
	err := r.queryDB(db).WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("recharge_no = ?", strings.TrimSpace(rechargeNo)).First(&row).Error
	return row, err
}

func (r *Repository) RechargeByUpstreamTradeForUpdate(ctx context.Context, db *gorm.DB, provider, upstreamTradeNo string) (Recharge, error) {
	var row Recharge
	err := r.queryDB(db).WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("provider = ? AND upstream_trade_no = ?", provider, strings.TrimSpace(upstreamTradeNo)).First(&row).Error
	return row, err
}

func (r *Repository) UpdateRecharge(ctx context.Context, db *gorm.DB, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.queryDB(db).WithContext(ctx).Model(&Recharge{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) ListAccounts(ctx context.Context, filters AccountFilters, limit, offset int) ([]AccountRow, int64, error) {
	query := r.applyAccountFilters(r.db.WithContext(ctx).Table("wallet_accounts").Joins("JOIN users ON users.id = wallet_accounts.user_id"), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []AccountRow
	err := query.Select("wallet_accounts.*, users.username, users.email, users.display_name").
		Order("wallet_accounts.created_at DESC, wallet_accounts.id DESC").Limit(limit).Offset(offset).Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) ListLedger(ctx context.Context, filters LedgerFilters, limit, offset int) ([]LedgerRow, int64, error) {
	query := r.applyLedgerFilters(r.db.WithContext(ctx).Table("wallet_ledger_entries").Joins("JOIN users ON users.id = wallet_ledger_entries.user_id"), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []LedgerRow
	err := query.Select("wallet_ledger_entries.*, users.username, users.email, users.display_name").
		Order("wallet_ledger_entries.created_at DESC, wallet_ledger_entries.id DESC").Limit(limit).Offset(offset).Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) ListUserLedger(ctx context.Context, userID uint64, filters LedgerFilters, limit, offset int) ([]LedgerEntry, int64, error) {
	query := r.applyUserLedgerFilters(r.db.WithContext(ctx).Model(&LedgerEntry{}).Where("user_id = ?", userID), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []LedgerEntry
	err := query.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&rows).Error
	return rows, total, err
}

func (r *Repository) ListRecharges(ctx context.Context, filters RechargeFilters, limit, offset int) ([]RechargeRow, int64, error) {
	query := r.applyRechargeFilters(r.db.WithContext(ctx).Table("wallet_recharges").Joins("JOIN users ON users.id = wallet_recharges.user_id"), filters)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []RechargeRow
	err := query.Select("wallet_recharges.*, users.username, users.email, users.display_name").
		Order("wallet_recharges.created_at DESC, wallet_recharges.id DESC").Limit(limit).Offset(offset).Scan(&rows).Error
	return rows, total, err
}

func (r *Repository) ListRecentLedger(ctx context.Context, walletNo string, limit int) ([]LedgerEntry, error) {
	var rows []LedgerEntry
	err := r.db.WithContext(ctx).Where("wallet_no = ?", strings.TrimSpace(walletNo)).
		Order("created_at DESC, id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *Repository) ListRecentRecharges(ctx context.Context, walletNo string, limit int) ([]Recharge, error) {
	var rows []Recharge
	err := r.db.WithContext(ctx).Where("wallet_no = ?", strings.TrimSpace(walletNo)).
		Order("created_at DESC, id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *Repository) queryDB(db *gorm.DB) *gorm.DB {
	if db != nil {
		return db
	}
	return r.db
}

func (r *Repository) applyAccountFilters(db *gorm.DB, filters AccountFilters) *gorm.DB {
	if strings.TrimSpace(filters.WalletNo) != "" {
		db = db.Where("wallet_accounts.wallet_no = ?", strings.TrimSpace(filters.WalletNo))
	}
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("wallet_accounts.status = ?", strings.TrimSpace(filters.Status))
	}
	return applyUserKeyword(db, filters.UserKeyword)
}

func (r *Repository) applyLedgerFilters(db *gorm.DB, filters LedgerFilters) *gorm.DB {
	if strings.TrimSpace(filters.WalletNo) != "" {
		db = db.Where("wallet_ledger_entries.wallet_no = ?", strings.TrimSpace(filters.WalletNo))
	}
	if strings.TrimSpace(filters.Direction) != "" {
		db = db.Where("wallet_ledger_entries.direction = ?", strings.TrimSpace(filters.Direction))
	}
	if strings.TrimSpace(filters.EntryType) != "" {
		db = db.Where("wallet_ledger_entries.entry_type = ?", strings.TrimSpace(filters.EntryType))
	}
	if strings.TrimSpace(filters.RelatedNo) != "" {
		db = db.Where("wallet_ledger_entries.related_no = ?", strings.TrimSpace(filters.RelatedNo))
	}
	if strings.TrimSpace(filters.DateFrom) != "" {
		db = db.Where("wallet_ledger_entries.created_at >= ?", strings.TrimSpace(filters.DateFrom))
	}
	if strings.TrimSpace(filters.DateTo) != "" {
		db = db.Where("wallet_ledger_entries.created_at <= ?", strings.TrimSpace(filters.DateTo))
	}
	return applyUserKeyword(db, filters.UserKeyword)
}

func (r *Repository) applyUserLedgerFilters(db *gorm.DB, filters LedgerFilters) *gorm.DB {
	if strings.TrimSpace(filters.Direction) != "" {
		db = db.Where("direction = ?", strings.TrimSpace(filters.Direction))
	}
	if strings.TrimSpace(filters.EntryType) != "" {
		db = db.Where("entry_type = ?", strings.TrimSpace(filters.EntryType))
	}
	if strings.TrimSpace(filters.RelatedNo) != "" {
		db = db.Where("related_no = ?", strings.TrimSpace(filters.RelatedNo))
	}
	if strings.TrimSpace(filters.DateFrom) != "" {
		db = db.Where("created_at >= ?", strings.TrimSpace(filters.DateFrom))
	}
	if strings.TrimSpace(filters.DateTo) != "" {
		db = db.Where("created_at <= ?", strings.TrimSpace(filters.DateTo))
	}
	return db
}

func (r *Repository) applyRechargeFilters(db *gorm.DB, filters RechargeFilters) *gorm.DB {
	if strings.TrimSpace(filters.WalletNo) != "" {
		db = db.Where("wallet_recharges.wallet_no = ?", strings.TrimSpace(filters.WalletNo))
	}
	if strings.TrimSpace(filters.Provider) != "" {
		db = db.Where("wallet_recharges.provider = ?", strings.TrimSpace(filters.Provider))
	}
	if strings.TrimSpace(filters.Method) != "" {
		db = db.Where("wallet_recharges.method = ?", strings.TrimSpace(filters.Method))
	}
	if strings.TrimSpace(filters.Status) != "" {
		db = db.Where("wallet_recharges.status = ?", strings.TrimSpace(filters.Status))
	}
	if strings.TrimSpace(filters.RechargeNo) != "" {
		db = db.Where("wallet_recharges.recharge_no = ?", strings.TrimSpace(filters.RechargeNo))
	}
	if strings.TrimSpace(filters.DateFrom) != "" {
		db = db.Where("wallet_recharges.created_at >= ?", strings.TrimSpace(filters.DateFrom))
	}
	if strings.TrimSpace(filters.DateTo) != "" {
		db = db.Where("wallet_recharges.created_at <= ?", strings.TrimSpace(filters.DateTo))
	}
	return applyUserKeyword(db, filters.UserKeyword)
}

func applyUserKeyword(db *gorm.DB, keyword string) *gorm.DB {
	if trimmed := strings.TrimSpace(keyword); trimmed != "" {
		like := "%" + trimmed + "%"
		return db.Where("users.username LIKE ? OR users.email LIKE ? OR users.display_name LIKE ?", like, like, like)
	}
	return db
}
