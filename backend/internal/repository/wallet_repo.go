package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"pvecloud/backend/internal/model"
)

// WalletRepository 封装钱包余额与流水操作。
type WalletRepository struct {
	db *gorm.DB
}

// NewWalletRepository 创建钱包仓储。
func NewWalletRepository(db *gorm.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

// GetByUserID 查询用户钱包。
func (r *WalletRepository) GetByUserID(ctx context.Context, userID uint) (*model.Wallet, error) {
	var wallet model.Wallet
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// ChangeBalance 在事务中修改余额并插入流水；支持 recharge/consume/refund。
func (r *WalletRepository) ChangeBalance(ctx context.Context, userID uint, delta float64, logType string, orderID *uint, remark string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return r.changeBalanceWithTx(ctx, tx, userID, delta, logType, orderID, remark)
	})
}

// ChangeBalanceWithTx 使用外部事务修改余额并写入流水，便于复合事务中保证原子性。
func (r *WalletRepository) ChangeBalanceWithTx(ctx context.Context, tx *gorm.DB, userID uint, delta float64, logType string, orderID *uint, remark string) error {
	return r.changeBalanceWithTx(ctx, tx, userID, delta, logType, orderID, remark)
}

// ListLogs 查询钱包流水，支持时间范围过滤。
func (r *WalletRepository) ListLogs(ctx context.Context, userID uint, start string, end string) ([]model.WalletLog, error) {
	var logs []model.WalletLog
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if start != "" {
		query = query.Where("created_at >= ?", start)
	}
	if end != "" {
		query = query.Where("created_at <= ?", end)
	}
	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, err
}

func (r *WalletRepository) changeBalanceWithTx(ctx context.Context, tx *gorm.DB, userID uint, delta float64, logType string, orderID *uint, remark string) error {
	var wallet model.Wallet
	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		return err
	}
	if wallet.Balance+delta < 0 {
		return errors.New("insufficient balance")
	}
	wallet.Balance += delta
	if err := tx.WithContext(ctx).Save(&wallet).Error; err != nil {
		return err
	}
	entry := model.WalletLog{UserID: userID, Type: logType, Amount: delta, OrderID: orderID, Remark: remark}
	return tx.WithContext(ctx).Create(&entry).Error
}
