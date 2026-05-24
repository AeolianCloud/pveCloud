package wallet

import (
	"context"
	"strings"

	"gorm.io/gorm"

	mysqlwallet "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/wallet"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	adminsupport "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/support"
)

type Service struct {
	wallets *mysqlwallet.Repository
}

func NewService(db *gorm.DB) *Service {
	return &Service{wallets: mysqlwallet.NewRepository(db)}
}

func (s *Service) List(ctx context.Context, query admindto.WalletListQuery) (admindto.PageResponse[admindto.WalletItem], error) {
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	rows, total, err := s.wallets.ListAccounts(ctx, mysqlwallet.AccountFilters{WalletNo: query.WalletNo, Status: query.Status, UserKeyword: query.UserKeyword}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.WalletItem]{}, err
	}
	items := make([]admindto.WalletItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, walletItem(row))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func (s *Service) Detail(ctx context.Context, walletNo string) (admindto.WalletDetail, error) {
	rows, _, err := s.wallets.ListAccounts(ctx, mysqlwallet.AccountFilters{WalletNo: strings.TrimSpace(walletNo)}, 1, 0)
	if err != nil {
		return admindto.WalletDetail{}, err
	}
	if len(rows) == 0 {
		return admindto.WalletDetail{}, apperrors.ErrNotFound.WithMessage("钱包不存在")
	}
	ledgerRows, err := s.wallets.ListRecentLedger(ctx, walletNo, 10)
	if err != nil {
		return admindto.WalletDetail{}, err
	}
	rechargeRows, err := s.wallets.ListRecentRecharges(ctx, walletNo, 10)
	if err != nil {
		return admindto.WalletDetail{}, err
	}
	recentLedger := make([]admindto.WalletLedgerItem, 0, len(ledgerRows))
	for _, ledger := range ledgerRows {
		recentLedger = append(recentLedger, ledgerItem(mysqlwallet.LedgerRow{LedgerEntry: ledger}))
	}
	recentRecharges := make([]admindto.WalletRechargeItem, 0, len(rechargeRows))
	for _, recharge := range rechargeRows {
		recentRecharges = append(recentRecharges, rechargeItem(mysqlwallet.RechargeRow{Recharge: recharge}))
	}
	return admindto.WalletDetail{WalletItem: walletItem(rows[0]), RecentLedger: recentLedger, RecentRecharges: recentRecharges}, nil
}

func (s *Service) Ledger(ctx context.Context, query admindto.WalletLedgerListQuery) (admindto.PageResponse[admindto.WalletLedgerItem], error) {
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	rows, total, err := s.wallets.ListLedger(ctx, mysqlwallet.LedgerFilters{WalletNo: query.WalletNo, UserKeyword: query.UserKeyword, Direction: query.Direction, EntryType: query.EntryType, RelatedNo: query.RelatedNo, DateFrom: query.DateFrom, DateTo: query.DateTo}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.WalletLedgerItem]{}, err
	}
	items := make([]admindto.WalletLedgerItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, ledgerItem(row))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func (s *Service) Recharges(ctx context.Context, query admindto.WalletRechargeListQuery) (admindto.PageResponse[admindto.WalletRechargeItem], error) {
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	rows, total, err := s.wallets.ListRecharges(ctx, mysqlwallet.RechargeFilters{WalletNo: query.WalletNo, UserKeyword: query.UserKeyword, Provider: query.Provider, Method: query.Method, Status: query.Status, RechargeNo: query.RechargeNo, DateFrom: query.DateFrom, DateTo: query.DateTo}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.WalletRechargeItem]{}, err
	}
	items := make([]admindto.WalletRechargeItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, rechargeItem(row))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func walletItem(row mysqlwallet.AccountRow) admindto.WalletItem {
	return admindto.WalletItem{WalletNo: row.WalletNo, User: walletUser(row.UserID, row.Username, row.Email, row.DisplayName), Currency: row.Currency, Status: row.Status, AvailableBalanceCents: row.AvailableBalanceCents, TotalRechargedCents: row.TotalRechargedCents, TotalSpentCents: row.TotalSpentCents, TotalRefundedCents: row.TotalRefundedCents, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func ledgerItem(row mysqlwallet.LedgerRow) admindto.WalletLedgerItem {
	return admindto.WalletLedgerItem{EntryNo: row.EntryNo, WalletNo: row.WalletNo, User: walletUser(row.UserID, row.Username, row.Email, row.DisplayName), Direction: row.Direction, EntryType: row.EntryType, AmountCents: row.AmountCents, BalanceBeforeCents: row.BalanceBeforeCents, BalanceAfterCents: row.BalanceAfterCents, Currency: row.Currency, RelatedType: row.RelatedType, RelatedNo: row.RelatedNo, Summary: row.Summary, CreatedAt: row.CreatedAt}
}

func rechargeItem(row mysqlwallet.RechargeRow) admindto.WalletRechargeItem {
	return admindto.WalletRechargeItem{RechargeNo: row.RechargeNo, WalletNo: row.WalletNo, User: walletUser(row.UserID, row.Username, row.Email, row.DisplayName), Provider: row.Provider, Method: row.Method, Status: row.Status, AmountCents: row.AmountCents, Currency: row.Currency, ExpiresAt: row.ExpiresAt, PaidAt: row.PaidAt, ClosedAt: row.ClosedAt, FailedAt: row.FailedAt, CreatedAt: row.CreatedAt}
}

func walletUser(id uint64, username string, email string, displayName *string) admindto.WalletUserSummary {
	return admindto.WalletUserSummary{ID: id, Username: username, Email: email, DisplayName: displayName}
}
