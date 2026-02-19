package service

import (
	"context"
	"errors"

	"pvecloud/backend/internal/repository"
)

var errRechargeTooSmall = errors.New("最低充值金额为 10 元")

// BillingService 封装充值、扣费、退款和流水查询逻辑。
type BillingService struct {
	walletRepo *repository.WalletRepository
}

// NewBillingService 创建计费服务。
func NewBillingService(walletRepo *repository.WalletRepository) *BillingService {
	return &BillingService{walletRepo: walletRepo}
}

// GetWallet 查询钱包余额。
func (s *BillingService) GetWallet(ctx context.Context, userID uint) (interface{}, error) {
	return s.walletRepo.GetByUserID(ctx, userID)
}

// Recharge 执行充值并写入流水。
func (s *BillingService) Recharge(ctx context.Context, userID uint, amount float64) error {
	if amount < 10 {
		return errRechargeTooSmall
	}
	return s.walletRepo.ChangeBalance(ctx, userID, amount, "recharge", nil, "用户主动充值")
}

// Consume 扣款并写入流水。
func (s *BillingService) Consume(ctx context.Context, userID uint, amount float64, orderID *uint, remark string) error {
	if amount <= 0 {
		return errors.New("扣费金额必须大于0")
	}
	return s.walletRepo.ChangeBalance(ctx, userID, -amount, "consume", orderID, remark)
}

// Refund 退款并写入流水。
func (s *BillingService) Refund(ctx context.Context, userID uint, amount float64, orderID *uint, remark string) error {
	if amount <= 0 {
		return errors.New("退款金额必须大于0")
	}
	return s.walletRepo.ChangeBalance(ctx, userID, amount, "refund", orderID, remark)
}

// Logs 查询流水记录。
func (s *BillingService) Logs(ctx context.Context, userID uint, start string, end string) (interface{}, error) {
	return s.walletRepo.ListLogs(ctx, userID, start, end)
}
