package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	mysqlerr "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	domaininstance "github.com/AeolianCloud/pveCloud/server/internal/domain/instance"
	domainorder "github.com/AeolianCloud/pveCloud/server/internal/domain/order"
	domainpayment "github.com/AeolianCloud/pveCloud/server/internal/domain/payment"
	domainwallet "github.com/AeolianCloud/pveCloud/server/internal/domain/wallet"
	integrationpayment "github.com/AeolianCloud/pveCloud/server/internal/integration/payment"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	mysqlinstance "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/instance"
	mysqlorder "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/order"
	mysqlpayment "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/payment"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	mysqlwallet "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/wallet"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/usecase/paymentalert"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
	webwallet "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/wallet"
)

type Service struct {
	db        *gorm.DB
	orders    *mysqlorder.Repository
	payments  *mysqlpayment.Repository
	wallets   *mysqlwallet.Repository
	instances *mysqlinstance.Repository
	lifecycle config.InstanceLifecycleConfig
	adapters  integrationpayment.Registry
	alerts    *paymentalert.Recorder
	wallet    *webwallet.Service
}

func NewService(db *gorm.DB, lifecycle config.InstanceLifecycleConfig, registries ...integrationpayment.Registry) *Service {
	registry := integrationpayment.NewSDKRegistry()
	if len(registries) > 0 && registries[0] != nil {
		registry = registries[0]
	}
	return &Service{db: db, orders: mysqlorder.NewRepository(db), payments: mysqlpayment.NewRepository(db), wallets: mysqlwallet.NewRepository(db), instances: mysqlinstance.NewRepository(db), lifecycle: lifecycle, adapters: registry}
}

func (s *Service) SetAlertRecorder(alerts *paymentalert.Recorder) *Service {
	s.alerts = alerts
	return s
}

func (s *Service) SetWalletService(wallet *webwallet.Service) *Service {
	s.wallet = wallet
	return s
}

func (s *Service) Create(ctx context.Context, userID uint64, orderNo string, req webdto.PaymentCreateRequest) (webdto.PaymentStatus, error) {
	provider := strings.TrimSpace(req.Provider)
	method := strings.TrimSpace(req.Method)
	clientToken := strings.TrimSpace(req.ClientToken)
	if !domainpayment.ProviderSupportsMethod(provider, method) {
		return webdto.PaymentStatus{}, apperrors.ErrValidation.WithMessage("支付方式与供应商不匹配")
	}
	if provider == domainpayment.ProviderWallet {
		return s.createWalletPayment(ctx, userID, orderNo, provider, method, clientToken)
	}
	paymentConfig, err := s.paymentConfig(ctx)
	if err != nil {
		return webdto.PaymentStatus{}, err
	}
	if !paymentConfig.enabled || !paymentConfig.providerEnabled(provider) {
		return webdto.PaymentStatus{}, apperrors.ErrConflict.WithMessage("支付渠道未启用")
	}
	providerConfig, err := paymentConfig.providerConfig(provider, method)
	if err != nil {
		return webdto.PaymentStatus{}, apperrors.ErrConflict.WithMessage("支付渠道配置不完整")
	}
	adapter, err := s.adapters.Adapter(provider)
	if err != nil {
		return webdto.PaymentStatus{}, apperrors.ErrConflict.WithMessage("支付渠道未启用")
	}
	order, err := s.orders.UserOrder(ctx, userID, strings.TrimSpace(orderNo))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return webdto.PaymentStatus{}, apperrors.ErrNotFound.WithMessage("订单不存在")
	}
	if err != nil {
		return webdto.PaymentStatus{}, err
	}
	if existing, err := s.payments.PaymentByIdempotency(ctx, order.ID, provider, method, clientToken); err == nil {
		return s.statusFromPayment(ctx, existing)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return webdto.PaymentStatus{}, err
	}
	if order.Status != domainorder.StatusPending || order.PaymentStatus != domainorder.PaymentStatusUnpaid {
		return webdto.PaymentStatus{}, apperrors.ErrConflict.WithMessage("当前订单不可支付")
	}
	now := time.Now()
	row := mysqlpayment.PaymentTransaction{
		PaymentNo:   fmt.Sprintf("PAY-%d", now.UnixNano()),
		OrderID:     order.ID,
		OrderNo:     order.OrderNo,
		UserID:      userID,
		Provider:    provider,
		Method:      method,
		Status:      domainpayment.StatusPending,
		ClientToken: clientToken,
		AmountCents: order.TotalAmountCents,
		Currency:    order.Currency,
		ExpiresAt:   now.Add(time.Duration(paymentConfig.expireMinutes) * time.Minute).Truncate(time.Millisecond),
	}
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error { return s.payments.CreatePayment(ctx, tx, &row) }); err != nil {
		if existing, findErr := s.payments.PaymentByIdempotency(ctx, order.ID, provider, method, clientToken); findErr == nil {
			return s.statusFromPayment(ctx, existing)
		}
		return webdto.PaymentStatus{}, err
	}
	channelResult, err := adapter.CreatePayment(ctx, providerConfig, integrationpayment.CreatePaymentRequest{
		PaymentNo:   row.PaymentNo,
		OrderNo:     order.OrderNo,
		Subject:     order.ProductName,
		AmountCents: order.TotalAmountCents,
		Currency:    order.Currency,
		Method:      method,
		ExpiresAt:   row.ExpiresAt,
	})
	if err != nil {
		message := truncateString(err.Error(), 500)
		_ = mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
			return s.payments.UpdatePayment(ctx, tx, row.ID, map[string]any{"status": domainpayment.StatusFailed, "failed_at": time.Now().Truncate(time.Millisecond), "last_error_code": "CHANNEL_CREATE_FAILED", "last_error_message": message})
		})
		s.recordAlert(ctx, paymentalert.Event{Event: paymentalert.EventPaymentCreateFailed, PaymentNo: row.PaymentNo, OrderNo: row.OrderNo, Provider: row.Provider, Method: row.Method, Status: domainpayment.StatusFailed, ErrorCode: "CHANNEL_CREATE_FAILED", ErrorMessage: message})
		return webdto.PaymentStatus{}, apperrors.ErrExternalUnavailable.WithMessage("支付渠道下单失败")
	}
	row.UpstreamTradeNo = optionalPtr(channelResult.UpstreamTradeNo)
	row.UpstreamPrepayID = optionalPtr(channelResult.UpstreamPrepayID)
	row.RedirectURL = optionalPtr(channelResult.RedirectURL)
	row.QRCodeURL = optionalPtr(channelResult.QRCodeURL)
	if err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		return s.payments.UpdatePayment(ctx, tx, row.ID, map[string]any{"upstream_trade_no": row.UpstreamTradeNo, "upstream_prepay_id": row.UpstreamPrepayID, "redirect_url": row.RedirectURL, "qr_code_url": row.QRCodeURL, "query_summary": nullableString(channelResult.Summary), "last_error_code": nil, "last_error_message": nil})
	}); err != nil {
		return webdto.PaymentStatus{}, err
	}
	return s.statusFromPayment(ctx, row)
}

func (s *Service) Get(ctx context.Context, userID uint64, paymentNo string) (webdto.PaymentStatus, error) {
	payment, err := s.payments.UserPaymentByNo(ctx, userID, strings.TrimSpace(paymentNo))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return webdto.PaymentStatus{}, apperrors.ErrNotFound.WithMessage("支付不存在")
	}
	if err != nil {
		return webdto.PaymentStatus{}, err
	}
	return s.statusFromPayment(ctx, payment)
}

func (s *Service) HandleCallback(ctx context.Context, provider string, httpReq *http.Request) error {
	paymentConfig, err := s.paymentConfig(ctx)
	if err != nil {
		return err
	}
	providerConfig, err := paymentConfig.providerConfig(provider, "")
	if err != nil {
		return apperrors.ErrConflict.WithMessage("支付渠道配置不完整")
	}
	adapter, err := s.adapters.Adapter(provider)
	if err != nil {
		return apperrors.ErrConflict.WithMessage("支付渠道未启用")
	}
	parsed, err := adapter.ParseNotification(ctx, providerConfig, httpReq)
	if err != nil {
		if errors.Is(err, integrationpayment.ErrInvalidSignature) {
			s.recordAlert(ctx, paymentalert.Event{Event: paymentalert.EventPaymentCallbackSignatureFailed, Provider: provider, Status: domainpayment.StatusFailed, ErrorCode: "CALLBACK_SIGNATURE_FAILED", ErrorMessage: err.Error()})
			return apperrors.ErrExternalUnavailable.WithMessage("支付回调验签失败")
		}
		return err
	}
	req := webdto.PaymentCallbackRequest{PaymentNo: parsed.PaymentNo, Provider: provider, UpstreamTradeNo: parsed.UpstreamTradeNo, AmountCents: parsed.AmountCents, Status: parsed.Status, Summary: parsed.Summary}
	if strings.TrimSpace(req.PaymentNo) == "" && strings.TrimSpace(req.UpstreamTradeNo) == "" {
		return apperrors.ErrValidation.WithMessage("缺少支付编号或渠道交易号")
	}
	return mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		payment, err := s.callbackPayment(ctx, tx, provider, req)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if s.wallet != nil {
				return s.wallet.ApplyRechargeNotification(ctx, tx, req)
			}
			return apperrors.ErrNotFound.WithMessage("支付不存在")
		}
		if err != nil {
			return err
		}
		if payment.Provider != provider || payment.AmountCents != req.AmountCents {
			return apperrors.ErrConflict.WithMessage("支付回调金额或供应商不一致")
		}
		if payment.Status == domainpayment.StatusPaid {
			return nil
		}
		if payment.Status != domainpayment.StatusPending {
			return apperrors.ErrConflict.WithMessage("当前支付状态不可回调确认")
		}
		if req.Status != domainpayment.StatusPaid {
			return s.markPaymentNonPaid(ctx, tx, payment, req)
		}
		return s.applyPaid(ctx, tx, payment, req)
	})
}

func (s *Service) createWalletPayment(ctx context.Context, userID uint64, orderNo string, provider string, method string, clientToken string) (webdto.PaymentStatus, error) {
	if method != domainpayment.MethodWalletBalance {
		return webdto.PaymentStatus{}, apperrors.ErrValidation.WithMessage("支付方式与供应商不匹配")
	}
	if !s.walletEnabled(ctx) {
		return webdto.PaymentStatus{}, apperrors.ErrConflict.WithMessage("钱包未启用")
	}
	order, err := s.orders.UserOrder(ctx, userID, strings.TrimSpace(orderNo))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return webdto.PaymentStatus{}, apperrors.ErrNotFound.WithMessage("订单不存在")
	}
	if err != nil {
		return webdto.PaymentStatus{}, err
	}
	if existing, err := s.payments.PaymentByIdempotency(ctx, order.ID, provider, method, clientToken); err == nil {
		return s.statusFromPayment(ctx, existing)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return webdto.PaymentStatus{}, err
	}
	var created mysqlpayment.PaymentTransaction
	err = mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		lockedOrder, err := s.orders.OrderForUpdate(ctx, tx, order.OrderNo)
		if err != nil {
			return err
		}
		if lockedOrder.UserID != userID {
			return apperrors.ErrNotFound.WithMessage("订单不存在")
		}
		if existing, err := s.payments.PaymentByIdempotency(ctx, lockedOrder.ID, provider, method, clientToken); err == nil {
			created = existing
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if lockedOrder.Status != domainorder.StatusPending || lockedOrder.PaymentStatus != domainorder.PaymentStatusUnpaid {
			return apperrors.ErrConflict.WithMessage("当前订单不可支付")
		}
		account, err := s.ensureWalletAccountForUpdate(ctx, tx, userID)
		if err != nil {
			return err
		}
		if account.Status != domainwallet.AccountStatusActive {
			return apperrors.ErrConflict.WithMessage("钱包不可用")
		}
		if account.AvailableBalanceCents < lockedOrder.TotalAmountCents {
			return apperrors.ErrConflict.WithMessage("钱包余额不足")
		}
		now := time.Now().Truncate(time.Millisecond)
		payment := mysqlpayment.PaymentTransaction{PaymentNo: fmt.Sprintf("PAY-%d", time.Now().UnixNano()), OrderID: lockedOrder.ID, OrderNo: lockedOrder.OrderNo, UserID: userID, Provider: provider, Method: method, Status: domainpayment.StatusPending, ClientToken: clientToken, AmountCents: lockedOrder.TotalAmountCents, Currency: lockedOrder.Currency, ExpiresAt: now, PaidAt: &now}
		if err := s.payments.CreatePayment(ctx, tx, &payment); err != nil {
			return err
		}
		before := account.AvailableBalanceCents
		after := before - lockedOrder.TotalAmountCents
		if err := s.wallets.UpdateAccount(ctx, tx, account.ID, map[string]any{"available_balance_cents": after, "total_spent_cents": account.TotalSpentCents + lockedOrder.TotalAmountCents}); err != nil {
			return err
		}
		summary := walletLedgerSummary(map[string]any{"payment_no": payment.PaymentNo, "order_no": lockedOrder.OrderNo})
		entry := mysqlwallet.LedgerEntry{EntryNo: fmt.Sprintf("WLE-%d", time.Now().UnixNano()), WalletID: account.ID, WalletNo: account.WalletNo, UserID: userID, Direction: domainwallet.DirectionDebit, EntryType: domainwallet.EntryTypePayment, AmountCents: lockedOrder.TotalAmountCents, BalanceBeforeCents: before, BalanceAfterCents: after, Currency: account.Currency, RelatedType: domainwallet.RelatedTypePayment, RelatedNo: payment.PaymentNo, IdempotencyKey: "payment:" + payment.PaymentNo, Summary: &summary}
		if err := s.wallets.CreateLedgerEntry(ctx, tx, &entry); err != nil {
			return err
		}
		// 钱包余额支付没有外部回调，扣款、支付状态和订单生效必须在同一事务里原子推进。
		if err := s.applyPaid(ctx, tx, payment, webdto.PaymentCallbackRequest{PaymentNo: payment.PaymentNo, Provider: provider, AmountCents: payment.AmountCents, Status: domainpayment.StatusPaid, Summary: summary}); err != nil {
			return err
		}
		payment.Status = domainpayment.StatusPaid
		payment.PaidAt = &now
		created = payment
		return nil
	})
	if err != nil {
		if existing, findErr := s.payments.PaymentByIdempotency(ctx, order.ID, provider, method, clientToken); findErr == nil {
			return s.statusFromPayment(ctx, existing)
		}
		return webdto.PaymentStatus{}, err
	}
	return s.statusFromPayment(ctx, created)
}

func (s *Service) ApplyPaidForAdmin(ctx context.Context, paymentNo string) error {
	payment, err := s.payments.PaymentByNo(ctx, strings.TrimSpace(paymentNo))
	if err != nil {
		return err
	}
	paymentConfig, err := s.paymentConfig(ctx)
	if err != nil {
		return err
	}
	providerConfig, err := paymentConfig.providerConfig(payment.Provider, payment.Method)
	if err != nil {
		return apperrors.ErrConflict.WithMessage("支付渠道配置不完整")
	}
	adapter, err := s.adapters.Adapter(payment.Provider)
	if err != nil {
		return apperrors.ErrConflict.WithMessage("支付渠道未启用")
	}
	query, err := adapter.QueryPayment(ctx, providerConfig, integrationpayment.QueryPaymentRequest{PaymentNo: payment.PaymentNo, UpstreamTradeNo: valueOf(payment.UpstreamTradeNo), Method: payment.Method})
	if err != nil {
		return apperrors.ErrExternalUnavailable.WithMessage("支付渠道状态同步失败")
	}
	return mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		payment, err := s.payments.PaymentForUpdate(ctx, tx, strings.TrimSpace(paymentNo))
		if err != nil {
			return err
		}
		if payment.Status == domainpayment.StatusPaid {
			return nil
		}
		callback := webdto.PaymentCallbackRequest{PaymentNo: payment.PaymentNo, Provider: payment.Provider, UpstreamTradeNo: query.UpstreamTradeNo, AmountCents: query.AmountCents, Status: query.Status, Summary: query.Summary}
		if callback.AmountCents == 0 {
			callback.AmountCents = payment.AmountCents
		}
		if callback.Status != domainpayment.StatusPaid {
			return s.markPaymentNonPaid(ctx, tx, payment, callback)
		}
		return s.applyPaid(ctx, tx, payment, callback)
	})
}

func (s *Service) applyPaid(ctx context.Context, tx *gorm.DB, payment mysqlpayment.PaymentTransaction, req webdto.PaymentCallbackRequest) error {
	order, err := s.orders.OrderForUpdate(ctx, tx, payment.OrderNo)
	if err != nil {
		return err
	}
	if order.TotalAmountCents != payment.AmountCents || order.Currency != payment.Currency {
		return apperrors.ErrConflict.WithMessage("支付金额与订单不一致")
	}
	now := time.Now().Truncate(time.Millisecond)
	summary := callbackSummary(req)
	updates := map[string]any{"status": domainpayment.StatusPaid, "paid_at": now, "callback_summary": summary}
	if strings.TrimSpace(req.UpstreamTradeNo) != "" {
		updates["upstream_trade_no"] = strings.TrimSpace(req.UpstreamTradeNo)
	}
	if err := s.payments.UpdatePayment(ctx, tx, payment.ID, updates); err != nil {
		return err
	}
	orderUpdates := map[string]any{"payment_status": domainorder.PaymentStatusPaid, "paid_at": now, "payment_provider": payment.Provider}
	if tradeNo := firstNonEmpty(req.UpstreamTradeNo, valueOf(payment.UpstreamTradeNo)); tradeNo != "" {
		orderUpdates["payment_trade_no"] = tradeNo
	}
	if err := s.orders.Update(ctx, tx, order.ID, orderUpdates); err != nil {
		return err
	}
	if order.OrderType == domainorder.TypeRenewal {
		return s.applyRenewal(ctx, tx, order, payment, now)
	}
	return s.enqueueProvision(ctx, tx, order, payment, now)
}

func (s *Service) applyRenewal(ctx context.Context, tx *gorm.DB, order mysqlorder.Order, payment mysqlpayment.PaymentTransaction, now time.Time) error {
	if order.RelatedInstanceNo == nil || strings.TrimSpace(*order.RelatedInstanceNo) == "" {
		return apperrors.ErrConflict.WithMessage("续费订单未关联实例")
	}
	instance, err := s.instances.InstanceForUpdate(ctx, tx, *order.RelatedInstanceNo)
	if err != nil {
		return err
	}
	months, ok := domainorder.BillingCycleMonths(order.BillingCycle)
	if !ok {
		return apperrors.ErrValidation.WithMessage("续费周期不支持")
	}
	before := instance.ExpiresAt
	after := renewalExpiresAt(now, before, months)
	if err := s.instances.UpdateInstance(ctx, tx, instance.ID, renewalInstanceUpdates(s.lifecycle, after)); err != nil {
		return err
	}
	if err := s.enqueueLifecycleTasks(ctx, tx, instance.InstanceNo, after); err != nil {
		return err
	}
	if err := s.orders.Update(ctx, tx, order.ID, map[string]any{"status": domainorder.StatusFulfilled}); err != nil {
		return err
	}
	effectType := domainpayment.EffectTypeRenewalExtension
	instanceNo := instance.InstanceNo
	return s.payments.CreateEffect(ctx, tx, &mysqlpayment.PaymentEffect{EffectNo: fmt.Sprintf("EFF-%d", time.Now().UnixNano()), PaymentID: payment.ID, PaymentNo: payment.PaymentNo, OrderID: order.ID, OrderNo: order.OrderNo, OrderType: order.OrderType, EffectType: effectType, Status: domainpayment.EffectStatusActive, InstanceID: &instance.ID, InstanceNo: &instanceNo, BeforeExpiresAt: before, AfterExpiresAt: &after, AppliedAt: now})
}

func (s *Service) enqueueProvision(ctx context.Context, tx *gorm.DB, order mysqlorder.Order, payment mysqlpayment.PaymentTransaction, now time.Time) error {
	objectType := "order"
	objectNo := order.OrderNo
	key := domaininstance.TaskTypePaymentProvision + ":" + payment.PaymentNo
	payload, _ := json.Marshal(map[string]string{"payment_no": payment.PaymentNo, "order_no": order.OrderNo})
	task := mysqlinstance.Task{TaskNo: fmt.Sprintf("TASK-%d", time.Now().UnixNano()), TaskType: domaininstance.TaskTypePaymentProvision, IdempotencyKey: &key, Status: domaininstance.TaskStatusPending, ObjectType: &objectType, ObjectNo: &objectNo, Payload: stringPtr(string(payload)), MaxAttempts: 10, ScheduledAt: now}
	return s.instances.CreateTaskIgnoreDuplicate(ctx, tx, &task)
}

func (s *Service) markPaymentNonPaid(ctx context.Context, tx *gorm.DB, payment mysqlpayment.PaymentTransaction, req webdto.PaymentCallbackRequest) error {
	status := req.Status
	if status != domainpayment.StatusClosed && status != domainpayment.StatusFailed && status != domainpayment.StatusRefunded {
		return apperrors.ErrValidation.WithMessage("支付回调状态不支持")
	}
	now := time.Now().Truncate(time.Millisecond)
	updates := map[string]any{"status": status, "callback_summary": callbackSummary(req)}
	if status == domainpayment.StatusClosed {
		updates["closed_at"] = now
	}
	if status == domainpayment.StatusFailed {
		updates["failed_at"] = now
	}
	return s.payments.UpdatePayment(ctx, tx, payment.ID, updates)
}

func (s *Service) callbackPayment(ctx context.Context, tx *gorm.DB, provider string, req webdto.PaymentCallbackRequest) (mysqlpayment.PaymentTransaction, error) {
	if strings.TrimSpace(req.PaymentNo) != "" {
		return s.payments.PaymentForUpdate(ctx, tx, strings.TrimSpace(req.PaymentNo))
	}
	return s.payments.PaymentByUpstreamTradeForUpdate(ctx, tx, provider, strings.TrimSpace(req.UpstreamTradeNo))
}

func (s *Service) statusFromPayment(ctx context.Context, payment mysqlpayment.PaymentTransaction) (webdto.PaymentStatus, error) {
	order, err := s.orders.FindByOrderNo(ctx, payment.OrderNo)
	if err != nil {
		return webdto.PaymentStatus{}, err
	}
	return webdto.PaymentStatus{PaymentNo: payment.PaymentNo, OrderNo: payment.OrderNo, Provider: payment.Provider, Method: payment.Method, AmountCents: payment.AmountCents, Currency: payment.Currency, Status: payment.Status, ExpiresAt: payment.ExpiresAt, RedirectURL: payment.RedirectURL, QRCodeURL: payment.QRCodeURL, PaidAt: payment.PaidAt, OrderStatus: order.Status, OrderPaymentStatus: order.PaymentStatus, RelatedInstanceNo: order.RelatedInstanceNo, LastErrorMessage: payment.LastErrorMessage}, nil
}

type configSnapshot struct {
	enabled       bool
	expireMinutes int
	alipayEnabled bool
	wechatEnabled bool
	values        map[string]string
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

func (s *Service) paymentConfig(ctx context.Context) (configSnapshot, error) {
	var rows []struct {
		ConfigKey   string  `gorm:"column:config_key"`
		ConfigValue *string `gorm:"column:config_value"`
	}
	if err := s.db.WithContext(ctx).Table("system_configs").Select("config_key, config_value").Where("config_key LIKE ?", "payment.%").Find(&rows).Error; err != nil {
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
	return configSnapshot{enabled: values["payment.enabled"] == "true", expireMinutes: expireMinutes, alipayEnabled: values["payment.alipay.enabled"] == "true", wechatEnabled: values["payment.wechat.enabled"] == "true", values: values}, nil
}

func (s *Service) walletEnabled(ctx context.Context) bool {
	var row struct {
		ConfigValue *string `gorm:"column:config_value"`
	}
	if err := s.db.WithContext(ctx).Table("system_configs").Select("config_value").Where("config_key = ?", "wallet.enabled").Take(&row).Error; err != nil {
		return false
	}
	return valueOf(row.ConfigValue) == "true"
}

func (s *Service) ensureWalletAccountForUpdate(ctx context.Context, tx *gorm.DB, userID uint64) (mysqlwallet.Account, error) {
	account, err := s.wallets.AccountByUserCurrencyForUpdate(ctx, tx, userID, domainwallet.CurrencyCNY)
	if err == nil {
		return account, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return mysqlwallet.Account{}, err
	}
	account = mysqlwallet.Account{WalletNo: fmt.Sprintf("WAL-%d", time.Now().UnixNano()), UserID: userID, Currency: domainwallet.CurrencyCNY, Status: domainwallet.AccountStatusActive}
	if err := s.wallets.CreateAccount(ctx, tx, &account); err != nil {
		// 钱包账户按 user_id+currency 唯一。并发首次余额支付时，另一个事务可能已先创建账户；
		// 这里回查并继续加锁，避免把可恢复的唯一键冲突暴露成支付失败。
		if isDuplicate(err) {
			return s.wallets.AccountByUserCurrencyForUpdate(ctx, tx, userID, domainwallet.CurrencyCNY)
		}
		return mysqlwallet.Account{}, err
	}
	return account, nil
}

func walletLedgerSummary(values map[string]any) string {
	data, _ := json.Marshal(values)
	return string(data)
}

func callbackSummary(req webdto.PaymentCallbackRequest) string {
	if strings.TrimSpace(req.Summary) != "" {
		return req.Summary
	}
	data, _ := json.Marshal(map[string]any{"payment_no": req.PaymentNo, "provider": req.Provider, "upstream_trade_no": req.UpstreamTradeNo, "amount_cents": req.AmountCents, "status": req.Status})
	return string(data)
}

func renewalExpiresAt(now time.Time, currentExpiresAt *time.Time, months int) time.Time {
	base := now
	if currentExpiresAt != nil && currentExpiresAt.After(now) {
		base = *currentExpiresAt
	}
	return base.AddDate(0, months, 0).Truncate(time.Millisecond)
}

func renewalInstanceUpdates(lifecycle config.InstanceLifecycleConfig, nextExpiresAt time.Time) map[string]any {
	updates := map[string]any{"expires_at": nextExpiresAt, "expire_notice_sent_at": nil}
	if lifecycle.AutoReleaseEnabled {
		updates["expire_release_scheduled_at"] = nextExpiresAt.Add(time.Duration(lifecycle.ExpireReleaseAfterSeconds) * time.Second).Truncate(time.Millisecond)
	} else {
		updates["expire_release_scheduled_at"] = nil
	}
	return updates
}

func (s *Service) enqueueLifecycleTasks(ctx context.Context, tx *gorm.DB, instanceNo string, expiresAt time.Time) error {
	payload, _ := json.Marshal(map[string]string{"instance_no": instanceNo, "expires_at": expiresAt.Format(time.RFC3339Nano)})
	objectType := "instance"
	objectNo := strings.TrimSpace(instanceNo)
	noticeKey := "expiry_notice:" + objectNo + ":" + expiresAt.Format(time.RFC3339Nano)
	noticeAt := expiresAt.Add(-time.Duration(s.lifecycle.ExpireNoticeBeforeSeconds) * time.Second).Truncate(time.Millisecond)
	if noticeAt.Before(time.Now()) {
		noticeAt = time.Now().Truncate(time.Millisecond)
	}
	noticeTask := mysqlinstance.Task{TaskNo: fmt.Sprintf("TASK-%d", time.Now().UnixNano()), TaskType: domaininstance.TaskTypeExpiryNotice, IdempotencyKey: &noticeKey, Status: "pending", ObjectType: &objectType, ObjectNo: &objectNo, Payload: stringPtr(string(payload)), MaxAttempts: 10, ScheduledAt: noticeAt}
	if err := s.instances.CreateTaskIgnoreDuplicate(ctx, tx, &noticeTask); err != nil {
		return err
	}
	if !s.lifecycle.AutoReleaseEnabled {
		return nil
	}
	releaseKey := "expiry_release:" + objectNo + ":" + expiresAt.Format(time.RFC3339Nano)
	releaseTask := mysqlinstance.Task{TaskNo: fmt.Sprintf("TASK-%d", time.Now().UnixNano()+1), TaskType: domaininstance.TaskTypeExpiryRelease, IdempotencyKey: &releaseKey, Status: "pending", ObjectType: &objectType, ObjectNo: &objectNo, Payload: stringPtr(string(payload)), MaxAttempts: 10, ScheduledAt: expiresAt.Add(time.Duration(s.lifecycle.ExpireReleaseAfterSeconds) * time.Second).Truncate(time.Millisecond)}
	return s.instances.CreateTaskIgnoreDuplicate(ctx, tx, &releaseTask)
}

func stringPtr(value string) *string { return &value }

func optionalPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return stringPtr(strings.TrimSpace(value))
}

func nullableString(value string) *string {
	return optionalPtr(value)
}

func truncateString(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	return value[:max]
}

func valueOf(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func (s *Service) recordAlert(ctx context.Context, event paymentalert.Event) {
	if s.alerts != nil {
		s.alerts.Record(ctx, event)
	}
}

func isDuplicate(err error) bool {
	var mysqlErr *mysqlerr.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
