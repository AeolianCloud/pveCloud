package wallet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	mysqlerr "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	domainpayment "github.com/AeolianCloud/pveCloud/server/internal/domain/payment"
	domainwallet "github.com/AeolianCloud/pveCloud/server/internal/domain/wallet"
	integrationpayment "github.com/AeolianCloud/pveCloud/server/internal/integration/payment"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	mysqlwallet "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/wallet"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
)

type Service struct {
	db       *gorm.DB
	wallets  *mysqlwallet.Repository
	adapters integrationpayment.Registry
}

func NewService(db *gorm.DB, registries ...integrationpayment.Registry) *Service {
	registry := integrationpayment.NewSDKRegistry()
	if len(registries) > 0 && registries[0] != nil {
		registry = registries[0]
	}
	return &Service{db: db, wallets: mysqlwallet.NewRepository(db), adapters: registry}
}

func (s *Service) Get(ctx context.Context, userID uint64) (webdto.WalletSummary, error) {
	account, err := s.ensureAccount(ctx, nil, userID)
	if err != nil {
		return webdto.WalletSummary{}, err
	}
	return walletSummary(account), nil
}

func (s *Service) Ledger(ctx context.Context, userID uint64, query webdto.WalletLedgerQuery) (webdto.PageResponse[webdto.WalletLedgerItem], error) {
	page, perPage := normalizePage(query.Page, query.PerPage)
	rows, total, err := s.wallets.ListUserLedger(ctx, userID, mysqlwallet.LedgerFilters{Direction: query.Direction, EntryType: query.EntryType, RelatedNo: query.RelatedNo, DateFrom: query.DateFrom, DateTo: query.DateTo}, perPage, (page-1)*perPage)
	if err != nil {
		return webdto.PageResponse[webdto.WalletLedgerItem]{}, err
	}
	items := make([]webdto.WalletLedgerItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, ledgerItem(row))
	}
	return webPageResponse(items, total, page, perPage), nil
}

func (s *Service) CreateRecharge(ctx context.Context, userID uint64, req webdto.WalletRechargeCreateRequest) (webdto.WalletRechargeStatus, error) {
	provider := strings.TrimSpace(req.Provider)
	method := strings.TrimSpace(req.Method)
	clientToken := strings.TrimSpace(req.ClientToken)
	if !domainpayment.ProviderSupportsMethod(provider, method) || provider == domainpayment.ProviderWallet {
		return webdto.WalletRechargeStatus{}, apperrors.ErrValidation.WithMessage("充值支付方式与供应商不匹配")
	}
	cfg, err := s.config(ctx)
	if err != nil {
		return webdto.WalletRechargeStatus{}, err
	}
	if !cfg.walletEnabled {
		return webdto.WalletRechargeStatus{}, apperrors.ErrConflict.WithMessage("钱包未启用")
	}
	if req.AmountCents < cfg.rechargeMinCents || req.AmountCents > cfg.rechargeMaxCents {
		return webdto.WalletRechargeStatus{}, apperrors.ErrValidation.WithMessage("充值金额超出允许范围")
	}
	if !cfg.paymentEnabled || !cfg.providerEnabled(provider) {
		return webdto.WalletRechargeStatus{}, apperrors.ErrConflict.WithMessage("支付渠道未启用")
	}
	providerConfig, err := cfg.providerConfig(provider, method)
	if err != nil {
		return webdto.WalletRechargeStatus{}, apperrors.ErrConflict.WithMessage("支付渠道配置不完整")
	}
	adapter, err := s.adapters.Adapter(provider)
	if err != nil {
		return webdto.WalletRechargeStatus{}, apperrors.ErrConflict.WithMessage("支付渠道未启用")
	}

	var recharge mysqlwallet.Recharge
	created := false
	err = mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		account, err := s.ensureAccount(ctx, tx, userID)
		if err != nil {
			return err
		}
		if existing, err := s.wallets.RechargeByIdempotencyForUpdate(ctx, tx, account.ID, provider, method, clientToken); err == nil {
			recharge = existing
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		now := time.Now()
		recharge = mysqlwallet.Recharge{
			RechargeNo:  fmt.Sprintf("RCH-%d", now.UnixNano()),
			WalletID:    account.ID,
			WalletNo:    account.WalletNo,
			UserID:      userID,
			Provider:    provider,
			Method:      method,
			Status:      domainwallet.RechargeStatusPending,
			ClientToken: clientToken,
			AmountCents: req.AmountCents,
			Currency:    domainwallet.CurrencyCNY,
			ExpiresAt:   now.Add(time.Duration(cfg.expireMinutes) * time.Minute).Truncate(time.Millisecond),
		}
		if err := s.wallets.CreateRecharge(ctx, tx, &recharge); err != nil {
			if isDuplicate(err) {
				existing, findErr := s.wallets.RechargeByIdempotencyForUpdate(ctx, tx, account.ID, provider, method, clientToken)
				if findErr == nil {
					recharge = existing
					return nil
				}
			}
			return err
		}
		created = true
		return nil
	})
	if err != nil {
		return webdto.WalletRechargeStatus{}, err
	}
	if !created {
		return rechargeStatus(recharge), nil
	}

	result, err := adapter.CreatePayment(ctx, providerConfig, integrationpayment.CreatePaymentRequest{
		PaymentNo:   recharge.RechargeNo,
		OrderNo:     recharge.RechargeNo,
		Subject:     "钱包充值",
		AmountCents: recharge.AmountCents,
		Currency:    recharge.Currency,
		Method:      recharge.Method,
		ExpiresAt:   recharge.ExpiresAt,
	})
	if err != nil {
		message := truncateString(err.Error(), 500)
		_ = mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
			return s.wallets.UpdateRecharge(ctx, tx, recharge.ID, map[string]any{"status": domainwallet.RechargeStatusFailed, "failed_at": time.Now().Truncate(time.Millisecond), "last_error_code": "CHANNEL_CREATE_FAILED", "last_error_message": message})
		})
		return webdto.WalletRechargeStatus{}, apperrors.ErrExternalUnavailable.WithMessage("充值渠道下单失败")
	}
	recharge.UpstreamTradeNo = optionalPtr(result.UpstreamTradeNo)
	recharge.UpstreamPrepayID = optionalPtr(result.UpstreamPrepayID)
	recharge.RedirectURL = optionalPtr(result.RedirectURL)
	recharge.QRCodeURL = optionalPtr(result.QRCodeURL)
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		return s.wallets.UpdateRecharge(ctx, tx, recharge.ID, map[string]any{"upstream_trade_no": recharge.UpstreamTradeNo, "upstream_prepay_id": recharge.UpstreamPrepayID, "redirect_url": recharge.RedirectURL, "qr_code_url": recharge.QRCodeURL, "query_summary": optionalPtr(result.Summary), "last_error_code": nil, "last_error_message": nil})
	}); err != nil {
		return webdto.WalletRechargeStatus{}, err
	}
	return rechargeStatus(recharge), nil
}

func (s *Service) GetRecharge(ctx context.Context, userID uint64, rechargeNo string) (webdto.WalletRechargeStatus, error) {
	recharge, err := s.wallets.UserRechargeByNo(ctx, userID, strings.TrimSpace(rechargeNo))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return webdto.WalletRechargeStatus{}, apperrors.ErrNotFound.WithMessage("充值不存在")
	}
	if err != nil {
		return webdto.WalletRechargeStatus{}, err
	}
	return rechargeStatus(recharge), nil
}

func (s *Service) ApplyRechargeNotification(ctx context.Context, tx *gorm.DB, req webdto.PaymentCallbackRequest) error {
	recharge, err := s.callbackRecharge(ctx, tx, req.Provider, req)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.ErrNotFound.WithMessage("支付或充值不存在")
	}
	if err != nil {
		return err
	}
	if recharge.Provider != req.Provider || recharge.AmountCents != req.AmountCents {
		return apperrors.ErrConflict.WithMessage("充值回调金额或供应商不一致")
	}
	if recharge.Status == domainwallet.RechargeStatusPaid {
		return nil
	}
	if recharge.Status != domainwallet.RechargeStatusPending {
		return apperrors.ErrConflict.WithMessage("当前充值状态不可回调确认")
	}
	if req.Status != domainpayment.StatusPaid {
		return s.markRechargeNonPaid(ctx, tx, recharge, req)
	}
	account, err := s.wallets.AccountByNoForUpdate(ctx, tx, recharge.WalletNo)
	if err != nil {
		return err
	}
	now := time.Now().Truncate(time.Millisecond)
	before := account.AvailableBalanceCents
	after := before + recharge.AmountCents
	summary := callbackSummary(req)
	if err := s.wallets.UpdateAccount(ctx, tx, account.ID, map[string]any{"available_balance_cents": after, "total_recharged_cents": account.TotalRechargedCents + recharge.AmountCents}); err != nil {
		return err
	}
	// 充值入账的幂等键绑定充值编号；重复回调会在充值状态检查处跳过，唯一键再兜底防止重复流水。
	entry := mysqlwallet.LedgerEntry{EntryNo: fmt.Sprintf("WLE-%d", time.Now().UnixNano()), WalletID: account.ID, WalletNo: account.WalletNo, UserID: account.UserID, Direction: domainwallet.DirectionCredit, EntryType: domainwallet.EntryTypeRecharge, AmountCents: recharge.AmountCents, BalanceBeforeCents: before, BalanceAfterCents: after, Currency: account.Currency, RelatedType: domainwallet.RelatedTypeRecharge, RelatedNo: recharge.RechargeNo, IdempotencyKey: "recharge:" + recharge.RechargeNo, Summary: &summary}
	if err := s.wallets.CreateLedgerEntry(ctx, tx, &entry); err != nil {
		return err
	}
	updates := map[string]any{"status": domainwallet.RechargeStatusPaid, "paid_at": now, "callback_summary": summary}
	if strings.TrimSpace(req.UpstreamTradeNo) != "" {
		updates["upstream_trade_no"] = strings.TrimSpace(req.UpstreamTradeNo)
	}
	return s.wallets.UpdateRecharge(ctx, tx, recharge.ID, updates)
}

func (s *Service) ensureAccount(ctx context.Context, tx *gorm.DB, userID uint64) (mysqlwallet.Account, error) {
	if tx != nil {
		account, err := s.wallets.AccountByUserCurrencyForUpdate(ctx, tx, userID, domainwallet.CurrencyCNY)
		if err == nil {
			return account, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return mysqlwallet.Account{}, err
		}
		account = mysqlwallet.Account{WalletNo: fmt.Sprintf("WAL-%d", time.Now().UnixNano()), UserID: userID, Currency: domainwallet.CurrencyCNY, Status: domainwallet.AccountStatusActive}
		if err := s.wallets.CreateAccount(ctx, tx, &account); err != nil {
			if isDuplicate(err) {
				return s.wallets.AccountByUserCurrencyForUpdate(ctx, tx, userID, domainwallet.CurrencyCNY)
			}
			return mysqlwallet.Account{}, err
		}
		return account, nil
	}
	var account mysqlwallet.Account
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(dbtx *gorm.DB) error {
		var err error
		account, err = s.ensureAccount(ctx, dbtx, userID)
		return err
	})
	return account, err
}

func (s *Service) callbackRecharge(ctx context.Context, tx *gorm.DB, provider string, req webdto.PaymentCallbackRequest) (mysqlwallet.Recharge, error) {
	if strings.TrimSpace(req.PaymentNo) != "" {
		return s.wallets.RechargeForUpdate(ctx, tx, strings.TrimSpace(req.PaymentNo))
	}
	return s.wallets.RechargeByUpstreamTradeForUpdate(ctx, tx, provider, strings.TrimSpace(req.UpstreamTradeNo))
}

func (s *Service) markRechargeNonPaid(ctx context.Context, tx *gorm.DB, recharge mysqlwallet.Recharge, req webdto.PaymentCallbackRequest) error {
	now := time.Now().Truncate(time.Millisecond)
	status := domainwallet.RechargeStatusFailed
	updates := map[string]any{"status": status, "callback_summary": callbackSummary(req), "failed_at": now}
	if req.Status == domainpayment.StatusClosed || req.Status == domainpayment.StatusRefunded {
		status = domainwallet.RechargeStatusClosed
		updates["status"] = status
		updates["closed_at"] = now
		delete(updates, "failed_at")
	}
	return s.wallets.UpdateRecharge(ctx, tx, recharge.ID, updates)
}

type configSnapshot struct {
	walletEnabled    bool
	rechargeMinCents uint64
	rechargeMaxCents uint64
	paymentEnabled   bool
	expireMinutes    int
	alipayEnabled    bool
	wechatEnabled    bool
	values           map[string]string
}

func (c configSnapshot) providerEnabled(provider string) bool {
	switch provider {
	case domainpayment.ProviderAlipay:
		return c.alipayEnabled
	case domainpayment.ProviderWechat:
		return c.wechatEnabled
	default:
		return false
	}
}

func (c configSnapshot) providerConfig(provider string, method string) (integrationpayment.Config, error) {
	cfg := integrationpayment.Config{Provider: provider, Values: c.values}
	if err := integrationpayment.ValidateProductionConfig(cfg, method); err != nil {
		return integrationpayment.Config{}, err
	}
	return cfg, nil
}

func (s *Service) config(ctx context.Context) (configSnapshot, error) {
	var rows []struct {
		ConfigKey   string  `gorm:"column:config_key"`
		ConfigValue *string `gorm:"column:config_value"`
	}
	if err := s.db.WithContext(ctx).Table("system_configs").Select("config_key, config_value").Where("config_key LIKE ? OR config_key LIKE ?", "wallet.%", "payment.%").Find(&rows).Error; err != nil {
		return configSnapshot{}, err
	}
	values := map[string]string{}
	for _, row := range rows {
		values[row.ConfigKey] = valueOf(row.ConfigValue)
	}
	expireMinutes, _ := strconv.Atoi(values["payment.default_expire_minutes"])
	if expireMinutes <= 0 {
		expireMinutes = 30
	}
	minCents := parseUint(values["wallet.recharge_min_cents"], 100)
	maxCents := parseUint(values["wallet.recharge_max_cents"], 500000)
	if maxCents < minCents {
		maxCents = minCents
	}
	return configSnapshot{walletEnabled: values["wallet.enabled"] == "true", rechargeMinCents: minCents, rechargeMaxCents: maxCents, paymentEnabled: values["payment.enabled"] == "true", expireMinutes: expireMinutes, alipayEnabled: values["payment.alipay.enabled"] == "true", wechatEnabled: values["payment.wechat.enabled"] == "true", values: values}, nil
}

func walletSummary(row mysqlwallet.Account) webdto.WalletSummary {
	return webdto.WalletSummary{WalletNo: row.WalletNo, Currency: row.Currency, Status: row.Status, AvailableBalanceCents: row.AvailableBalanceCents, TotalRechargedCents: row.TotalRechargedCents, TotalSpentCents: row.TotalSpentCents, TotalRefundedCents: row.TotalRefundedCents, CreatedAt: row.CreatedAt}
}

func ledgerItem(row mysqlwallet.LedgerEntry) webdto.WalletLedgerItem {
	return webdto.WalletLedgerItem{EntryNo: row.EntryNo, WalletNo: row.WalletNo, Direction: row.Direction, EntryType: row.EntryType, AmountCents: row.AmountCents, BalanceBeforeCents: row.BalanceBeforeCents, BalanceAfterCents: row.BalanceAfterCents, Currency: row.Currency, RelatedType: row.RelatedType, RelatedNo: row.RelatedNo, Summary: row.Summary, CreatedAt: row.CreatedAt}
}

func rechargeStatus(row mysqlwallet.Recharge) webdto.WalletRechargeStatus {
	return webdto.WalletRechargeStatus{RechargeNo: row.RechargeNo, WalletNo: row.WalletNo, Provider: row.Provider, Method: row.Method, Status: row.Status, AmountCents: row.AmountCents, Currency: row.Currency, ExpiresAt: row.ExpiresAt, PaidAt: row.PaidAt, ClosedAt: row.ClosedAt, FailedAt: row.FailedAt, RedirectURL: row.RedirectURL, QRCodeURL: row.QRCodeURL, LastErrorMessage: row.LastErrorMessage, CreatedAt: row.CreatedAt}
}

func normalizePage(page int, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage
}

func webPageResponse[T any](items []T, total int64, page int, perPage int) webdto.PageResponse[T] {
	lastPage := 0
	if total > 0 {
		lastPage = int(math.Ceil(float64(total) / float64(perPage)))
	}
	return webdto.PageResponse[T]{List: items, Total: total, Page: page, PerPage: perPage, LastPage: lastPage}
}

func callbackSummary(req webdto.PaymentCallbackRequest) string {
	if strings.TrimSpace(req.Summary) != "" {
		return req.Summary
	}
	data, _ := json.Marshal(map[string]any{"payment_no": req.PaymentNo, "provider": req.Provider, "upstream_trade_no": req.UpstreamTradeNo, "amount_cents": req.AmountCents, "status": req.Status})
	return string(data)
}

func parseUint(value string, fallback uint64) uint64 {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil || parsed == 0 {
		return fallback
	}
	return parsed
}

func optionalPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(value)
	return &trimmed
}

func valueOf(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func truncateString(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	return value[:max]
}

func isDuplicate(err error) bool {
	var mysqlErr *mysqlerr.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
