package payment

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	domaininstance "github.com/AeolianCloud/pveCloud/server/internal/domain/instance"
	domainorder "github.com/AeolianCloud/pveCloud/server/internal/domain/order"
	domainpayment "github.com/AeolianCloud/pveCloud/server/internal/domain/payment"
	integrationpayment "github.com/AeolianCloud/pveCloud/server/internal/integration/payment"
	mysqlinstance "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/instance"
	mysqlorder "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/order"
	mysqlpayment "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/payment"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	adminaudit "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/audit"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	adminsupport "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/support"
	"github.com/AeolianCloud/pveCloud/server/internal/usecase/paymentalert"
	webpayment "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/payment"
)

type AdminAuditService = adminaudit.AdminAuditService
type AdminAuditWriteInput = adminaudit.AdminAuditWriteInput

type Service struct {
	db        *gorm.DB
	orders    *mysqlorder.Repository
	payments  *mysqlpayment.Repository
	instances *mysqlinstance.Repository
	web       *webpayment.Service
	audit     *AdminAuditService
	adapters  integrationpayment.Registry
	alerts    *paymentalert.Recorder
}

func NewService(db *gorm.DB, web *webpayment.Service, audit *AdminAuditService, registries ...integrationpayment.Registry) *Service {
	if audit == nil {
		audit = adminaudit.NewAdminAuditService(db)
	}
	registry := integrationpayment.NewSDKRegistry()
	if len(registries) > 0 && registries[0] != nil {
		registry = registries[0]
	}
	return &Service{db: db, orders: mysqlorder.NewRepository(db), payments: mysqlpayment.NewRepository(db), instances: mysqlinstance.NewRepository(db), web: web, audit: audit, adapters: registry}
}

func (s *Service) SetAlertRecorder(alerts *paymentalert.Recorder) *Service {
	s.alerts = alerts
	return s
}

func (s *Service) List(ctx context.Context, query admindto.PaymentListQuery) (admindto.PageResponse[admindto.PaymentItem], error) {
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	rows, total, err := s.payments.ListPayments(ctx, mysqlpayment.PaymentFilters{Provider: query.Provider, Method: query.Method, Status: query.Status, OrderNo: query.OrderNo, PaymentNo: query.PaymentNo, UserKeyword: query.UserKeyword, DateFrom: query.DateFrom, DateTo: query.DateTo}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.PaymentItem]{}, err
	}
	items := make([]admindto.PaymentItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, paymentItem(row))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func (s *Service) Detail(ctx context.Context, paymentNo string) (admindto.PaymentDetail, error) {
	payment, err := s.payments.PaymentByNo(ctx, strings.TrimSpace(paymentNo))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return admindto.PaymentDetail{}, apperrors.ErrNotFound.WithMessage("支付不存在")
	}
	if err != nil {
		return admindto.PaymentDetail{}, err
	}
	order, err := s.orders.FindByOrderNo(ctx, payment.OrderNo)
	if err != nil {
		return admindto.PaymentDetail{}, err
	}
	item := admindto.PaymentItem{PaymentNo: payment.PaymentNo, OrderNo: payment.OrderNo, User: admindto.OrderUserSummary{ID: payment.UserID}, Provider: payment.Provider, Method: payment.Method, Status: payment.Status, AmountCents: payment.AmountCents, Currency: payment.Currency, ExpiresAt: payment.ExpiresAt, PaidAt: payment.PaidAt, CreatedAt: payment.CreatedAt, OrderStatus: order.Status, OrderType: order.OrderType}
	if rows, _, rowErr := s.payments.ListPayments(ctx, mysqlpayment.PaymentFilters{PaymentNo: payment.PaymentNo}, 1, 0); rowErr == nil && len(rows) == 1 {
		item.User = admindto.OrderUserSummary{ID: rows[0].UserID, Username: rows[0].Username, Email: rows[0].Email, DisplayName: rows[0].DisplayName}
	}
	detail := admindto.PaymentDetail{PaymentItem: item, UpstreamTradeNo: payment.UpstreamTradeNo, LastErrorMessage: payment.LastErrorMessage}
	if refund, err := s.payments.RefundByPaymentID(ctx, payment.ID); err == nil {
		refundItem := refundItem(mysqlpayment.RefundRow{RefundTransaction: refund})
		detail.Refund = &refundItem
	}
	return detail, nil
}

func (s *Service) Refunds(ctx context.Context, query admindto.RefundListQuery) (admindto.PageResponse[admindto.RefundItem], error) {
	page, perPage := adminsupport.NormalizePage(query.Page, query.PerPage)
	rows, total, err := s.payments.ListRefunds(ctx, mysqlpayment.RefundFilters{Provider: query.Provider, Status: query.Status, OrderNo: query.OrderNo, PaymentNo: query.PaymentNo, RefundNo: query.RefundNo, DateFrom: query.DateFrom, DateTo: query.DateTo}, perPage, (page-1)*perPage)
	if err != nil {
		return admindto.PageResponse[admindto.RefundItem]{}, err
	}
	items := make([]admindto.RefundItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, refundItem(row))
	}
	return adminsupport.PageResponse(items, total, page, perPage), nil
}

func (s *Service) Sync(ctx context.Context, operatorID uint64, paymentNo string) (admindto.PaymentDetail, error) {
	if s.web != nil {
		if err := s.web.ApplyPaidForAdmin(ctx, strings.TrimSpace(paymentNo)); err != nil {
			return admindto.PaymentDetail{}, err
		}
	}
	if err := s.audit.Record(ctx, nil, AdminAuditWriteInput{AdminID: &operatorID, Action: "payment.sync", ObjectType: "payment", ObjectID: strings.TrimSpace(paymentNo), Remark: "同步支付渠道状态"}); err != nil {
		return admindto.PaymentDetail{}, err
	}
	return s.Detail(ctx, paymentNo)
}

func (s *Service) CreateRefund(ctx context.Context, operatorID uint64, paymentNo string, req admindto.RefundCreateRequest) (admindto.RefundItem, error) {
	var created mysqlpayment.RefundTransaction
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		payment, err := s.payments.PaymentForUpdate(ctx, tx, strings.TrimSpace(paymentNo))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound.WithMessage("支付不存在")
		}
		if err != nil {
			return err
		}
		if payment.Status != domainpayment.StatusPaid {
			return apperrors.ErrConflict.WithMessage("当前支付状态不可退款")
		}
		if _, err := s.payments.RefundByPaymentID(ctx, payment.ID); err == nil {
			return apperrors.ErrConflict.WithMessage("该支付已存在退款")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		order, err := s.orders.OrderForUpdate(ctx, tx, payment.OrderNo)
		if err != nil {
			return err
		}
		if order.OrderType == domainorder.TypePurchase {
			if instance, err := s.instances.InstanceByOrderID(ctx, order.ID); err == nil && instance.Status != domaininstance.StatusReleased {
				return apperrors.ErrConflict.WithMessage("新购已交付订单需先释放实例")
			}
		}
		refund := mysqlpayment.RefundTransaction{RefundNo: fmt.Sprintf("RF-%d", time.Now().UnixNano()), PaymentID: payment.ID, PaymentNo: payment.PaymentNo, OrderID: payment.OrderID, OrderNo: payment.OrderNo, UserID: payment.UserID, Provider: payment.Provider, Status: domainpayment.RefundStatusPending, AmountCents: payment.AmountCents, Currency: payment.Currency, Reason: strings.TrimSpace(req.Reason), RequestedByAdminID: operatorID, UpstreamTradeNo: payment.UpstreamTradeNo}
		if err := s.payments.CreateRefund(ctx, tx, &refund); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: "payment.refund.create", ObjectType: "refund", ObjectID: refund.RefundNo, AfterData: map[string]any{"payment_no": payment.PaymentNo, "order_no": payment.OrderNo, "amount_cents": payment.AmountCents}, Remark: req.Reason}); err != nil {
			return err
		}
		created = refund
		return nil
	})
	if err != nil {
		return admindto.RefundItem{}, err
	}
	result, err := s.createChannelRefund(ctx, created)
	if err != nil {
		_ = mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
			now := time.Now().Truncate(time.Millisecond)
			message := truncateString(err.Error(), 500)
			if updateErr := s.payments.UpdateRefund(ctx, tx, created.ID, map[string]any{"status": domainpayment.RefundStatusFailed, "failed_at": now, "last_error_code": "CHANNEL_REFUND_FAILED", "last_error_message": message}); updateErr != nil {
				return updateErr
			}
			return s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: "payment.refund.failed", ObjectType: "refund", ObjectID: created.RefundNo, Remark: message})
		})
		s.recordAlert(ctx, paymentalert.Event{Event: paymentalert.EventRefundFailed, PaymentNo: created.PaymentNo, RefundNo: created.RefundNo, OrderNo: created.OrderNo, Provider: created.Provider, Status: domainpayment.RefundStatusFailed, ErrorCode: "CHANNEL_REFUND_FAILED", ErrorMessage: err.Error()})
		return admindto.RefundItem{}, apperrors.ErrExternalUnavailable.WithMessage("支付渠道退款失败")
	}
	if result.Status != domainpayment.RefundStatusSucceeded {
		_ = mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
			return s.payments.UpdateRefund(ctx, tx, created.ID, map[string]any{"query_summary": nullableString(result.Summary), "upstream_refund_no": nullableString(result.UpstreamRefundNo)})
		})
		s.recordAlert(ctx, paymentalert.Event{Event: paymentalert.EventRefundPending, PaymentNo: created.PaymentNo, RefundNo: created.RefundNo, OrderNo: created.OrderNo, Provider: created.Provider, Status: domainpayment.RefundStatusPending})
		return refundItem(mysqlpayment.RefundRow{RefundTransaction: created}), nil
	}
	err = mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		payment, err := s.payments.PaymentForUpdate(ctx, tx, created.PaymentNo)
		if err != nil {
			return err
		}
		order, err := s.orders.OrderForUpdate(ctx, tx, payment.OrderNo)
		if err != nil {
			return err
		}
		created.UpstreamRefundNo = nullableString(result.UpstreamRefundNo)
		if err := s.payments.UpdateRefund(ctx, tx, created.ID, map[string]any{"upstream_refund_no": created.UpstreamRefundNo, "query_summary": nullableString(result.Summary)}); err != nil {
			return err
		}
		if err := s.completeRefund(ctx, tx, order, payment, created); err != nil {
			return err
		}
		return s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: "payment.refund.succeeded", ObjectType: "refund", ObjectID: created.RefundNo, AfterData: map[string]any{"payment_no": payment.PaymentNo, "order_no": payment.OrderNo, "amount_cents": payment.AmountCents}, Remark: req.Reason})
	})
	if err != nil {
		return admindto.RefundItem{}, err
	}
	created.Status = domainpayment.RefundStatusSucceeded
	now := time.Now().Truncate(time.Millisecond)
	created.ChannelConfirmedAt = &now
	created.CompletedAt = &now
	return refundItem(mysqlpayment.RefundRow{RefundTransaction: created}), nil
}

func (s *Service) RetryProvision(ctx context.Context, operatorID uint64, paymentNo string) (admindto.PaymentDetail, error) {
	err := mysqltx.NewManager(s.db).WithinContext(ctx, func(tx *gorm.DB) error {
		payment, err := s.payments.PaymentForUpdate(ctx, tx, strings.TrimSpace(paymentNo))
		if err != nil {
			return err
		}
		order, err := s.orders.OrderForUpdate(ctx, tx, payment.OrderNo)
		if err != nil {
			return err
		}
		if order.OrderType != domainorder.TypePurchase || order.Status != domainorder.StatusError || order.PaymentStatus != domainorder.PaymentStatusPaid {
			return apperrors.ErrConflict.WithMessage("当前订单不可重试交付")
		}
		objectType := "order"
		objectNo := order.OrderNo
		key := domaininstance.TaskTypePaymentProvision + ":" + payment.PaymentNo
		if existing, err := s.instances.TaskByIdempotencyKeyForUpdate(ctx, tx, key); err == nil {
			if existing.Status == domaininstance.TaskStatusFailed || existing.Status == domaininstance.TaskStatusCancelled {
				if err := s.instances.UpdateTask(ctx, tx, existing.ID, map[string]any{"status": domaininstance.TaskStatusPending, "scheduled_at": time.Now(), "locked_by": nil, "locked_until": nil, "last_error_code": nil, "last_error_message": nil, "completed_at": nil}); err != nil {
					return err
				}
				return s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: "payment.provision.retry", ObjectType: "payment", ObjectID: payment.PaymentNo, Remark: "重试支付后自动交付"})
			}
			return apperrors.ErrConflict.WithMessage("自动交付任务已存在，不能重复入队")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		task := mysqlinstance.Task{TaskNo: fmt.Sprintf("TASK-%d", time.Now().UnixNano()), TaskType: domaininstance.TaskTypePaymentProvision, IdempotencyKey: &key, Status: domaininstance.TaskStatusPending, ObjectType: &objectType, ObjectNo: &objectNo, MaxAttempts: 10, ScheduledAt: time.Now()}
		if err := s.instances.CreateTaskIgnoreDuplicate(ctx, tx, &task); err != nil {
			return err
		}
		return s.audit.Record(ctx, tx, AdminAuditWriteInput{AdminID: &operatorID, Action: "payment.provision.retry", ObjectType: "payment", ObjectID: payment.PaymentNo, Remark: "重试支付后自动交付"})
	})
	if err != nil {
		return admindto.PaymentDetail{}, err
	}
	return s.Detail(ctx, paymentNo)
}

func (s *Service) completeRefund(ctx context.Context, tx *gorm.DB, order mysqlorder.Order, payment mysqlpayment.PaymentTransaction, refund mysqlpayment.RefundTransaction) error {
	now := time.Now().Truncate(time.Millisecond)
	if order.OrderType == domainorder.TypeRenewal {
		effect, err := s.payments.EffectByPaymentIDForUpdate(ctx, tx, payment.ID)
		if err != nil {
			return apperrors.ErrConflict.WithMessage("续费支付缺少可回滚生效记录")
		}
		if effect.Status != domainpayment.EffectStatusActive || effect.InstanceNo == nil || effect.BeforeExpiresAt == nil {
			return apperrors.ErrConflict.WithMessage("续费支付生效记录不可回滚")
		}
		instance, err := s.instances.InstanceForUpdate(ctx, tx, *effect.InstanceNo)
		if err != nil {
			return err
		}
		if err := s.instances.UpdateInstance(ctx, tx, instance.ID, map[string]any{"expires_at": effect.BeforeExpiresAt}); err != nil {
			return err
		}
		if err := s.payments.UpdateEffect(ctx, tx, effect.ID, map[string]any{"status": domainpayment.EffectStatusReverted, "refund_id": refund.ID, "refund_no": refund.RefundNo, "reverted_at": now}); err != nil {
			return err
		}
	}
	if err := s.payments.UpdateRefund(ctx, tx, refund.ID, map[string]any{"status": domainpayment.RefundStatusSucceeded, "channel_confirmed_at": now, "completed_at": now}); err != nil {
		return err
	}
	if err := s.payments.UpdatePayment(ctx, tx, payment.ID, map[string]any{"status": domainpayment.StatusRefunded}); err != nil {
		return err
	}
	return s.orders.Update(ctx, tx, order.ID, map[string]any{"status": domainorder.StatusClosed, "payment_status": domainorder.PaymentStatusRefunded, "closed_at": now})
}

func paymentItem(row mysqlpayment.PaymentRow) admindto.PaymentItem {
	return admindto.PaymentItem{PaymentNo: row.PaymentNo, OrderNo: row.OrderNo, User: admindto.OrderUserSummary{ID: row.UserID, Username: row.Username, Email: row.Email, DisplayName: row.DisplayName}, Provider: row.Provider, Method: row.Method, Status: row.Status, AmountCents: row.AmountCents, Currency: row.Currency, ExpiresAt: row.ExpiresAt, PaidAt: row.PaidAt, CreatedAt: row.CreatedAt, OrderStatus: row.OrderStatus, OrderType: row.OrderType}
}

func refundItem(row mysqlpayment.RefundRow) admindto.RefundItem {
	return admindto.RefundItem{RefundNo: row.RefundNo, PaymentNo: row.PaymentNo, OrderNo: row.OrderNo, User: admindto.OrderUserSummary{ID: row.UserID, Username: row.Username, Email: row.Email, DisplayName: row.DisplayName}, Provider: row.Provider, Status: row.Status, AmountCents: row.AmountCents, Currency: row.Currency, Reason: row.Reason, CreatedAt: row.CreatedAt, CompletedAt: row.CompletedAt, FailedAt: row.FailedAt}
}

func (s *Service) createChannelRefund(ctx context.Context, refund mysqlpayment.RefundTransaction) (integrationpayment.RefundResult, error) {
	adapter, err := s.adapters.Adapter(refund.Provider)
	if err != nil {
		return integrationpayment.RefundResult{}, err
	}
	cfg, err := s.paymentConfig(ctx, refund.Provider)
	if err != nil {
		return integrationpayment.RefundResult{}, err
	}
	return adapter.CreateRefund(ctx, cfg, integrationpayment.CreateRefundRequest{RefundNo: refund.RefundNo, PaymentNo: refund.PaymentNo, UpstreamTradeNo: valueOf(refund.UpstreamTradeNo), AmountCents: refund.AmountCents, Currency: refund.Currency, Reason: refund.Reason})
}

func (s *Service) paymentConfig(ctx context.Context, provider string) (integrationpayment.Config, error) {
	var rows []struct {
		ConfigKey   string  `gorm:"column:config_key"`
		ConfigValue *string `gorm:"column:config_value"`
	}
	if err := s.db.WithContext(ctx).Table("system_configs").Select("config_key, config_value").Where("config_key LIKE ?", "payment.%").Find(&rows).Error; err != nil {
		return integrationpayment.Config{}, err
	}
	values := map[string]string{}
	for _, row := range rows {
		values[row.ConfigKey] = valueOf(row.ConfigValue)
	}
	cfg := integrationpayment.Config{Provider: provider, Values: values}
	if err := integrationpayment.ValidateProductionConfig(cfg, ""); err != nil {
		return integrationpayment.Config{}, err
	}
	return cfg, nil
}

func valueOf(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func nullableString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(value)
	return &trimmed
}

func truncateString(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	return value[:max]
}

func (s *Service) recordAlert(ctx context.Context, event paymentalert.Event) {
	if s.alerts != nil {
		s.alerts.Record(ctx, event)
	}
}
