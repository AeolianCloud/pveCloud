package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gorm.io/gorm"

	domaininstance "github.com/AeolianCloud/pveCloud/server/internal/domain/instance"
	domainorder "github.com/AeolianCloud/pveCloud/server/internal/domain/order"
	"github.com/AeolianCloud/pveCloud/server/internal/integration/mail"
	"github.com/AeolianCloud/pveCloud/server/internal/integration/mcppve"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/config"
	mysqlinstance "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/instance"
	mysqlorder "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/order"
	mysqltx "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/tx"
	admininstance "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/instance"
)

type Runner struct {
	db           *gorm.DB
	log          *slog.Logger
	tasks        *mysqlinstance.Repository
	orders       *mysqlorder.Repository
	instanceSvc  *admininstance.Service
	mail         *mail.Sender
	workerCfg    config.WorkerConfig
	lifecycleCfg config.InstanceLifecycleConfig
	notifyCfg    config.NotificationConfig
}

type taskPayload struct {
	InstanceNo     string `json:"instance_no,omitempty"`
	ExpiresAt      string `json:"expires_at,omitempty"`
	NotificationNo string `json:"notification_no,omitempty"`
}

var errPaymentProvisionSkipped = errors.New("payment provision task skipped")

func NewRunner(db *gorm.DB, log *slog.Logger, mcp *mcppve.Client, mailSender *mail.Sender, workerCfg config.WorkerConfig, lifecycleCfg config.InstanceLifecycleConfig, notifyCfg config.NotificationConfig) *Runner {
	return &Runner{
		db:           db,
		log:          log,
		tasks:        mysqlinstance.NewRepository(db),
		orders:       mysqlorder.NewRepository(db),
		instanceSvc:  admininstance.NewService(db, mcp, nil, lifecycleCfg),
		mail:         mailSender,
		workerCfg:    workerCfg,
		lifecycleCfg: lifecycleCfg,
		notifyCfg:    notifyCfg,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	if !r.workerCfg.Enabled {
		r.log.Info("Worker 未启用，进程空闲等待退出")
		<-ctx.Done()
		return nil
	}
	ticker := time.NewTicker(r.pollInterval())
	defer ticker.Stop()
	for {
		if err := r.PollOnce(ctx); err != nil {
			r.log.Error("Worker 轮询失败", "error", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (r *Runner) PollOnce(ctx context.Context) error {
	tasks, err := r.claim(ctx)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if err := r.execute(ctx, task); err != nil {
			if errors.Is(err, admininstance.ErrOperationPending) {
				if updateErr := r.markDeferred(ctx, task); updateErr != nil {
					r.log.Error("异步任务延后状态落库失败", "task_no", task.TaskNo, "error", updateErr)
				}
				continue
			}
			r.log.Error("异步任务执行失败", "task_no", task.TaskNo, "task_type", task.TaskType, "error", err)
			if updateErr := r.markFailedOrRetry(ctx, task, err); updateErr != nil {
				r.log.Error("异步任务失败状态落库失败", "task_no", task.TaskNo, "error", updateErr)
			}
			continue
		}
		if err := r.markSucceeded(ctx, task); err != nil {
			r.log.Error("异步任务成功状态落库失败", "task_no", task.TaskNo, "error", err)
		}
	}
	return nil
}

func (r *Runner) claim(ctx context.Context) ([]mysqlinstance.Task, error) {
	lockUntil := time.Now().Add(time.Duration(r.workerCfg.LockTTLSeconds) * time.Second)
	var rows []mysqlinstance.Task
	err := mysqltx.NewManager(r.db).WithinContext(ctx, func(tx *gorm.DB) error {
		var err error
		rows, err = r.tasks.ClaimTasks(ctx, tx, strings.TrimSpace(r.workerCfg.ID), r.workerCfg.BatchSize, lockUntil)
		return err
	})
	return rows, err
}

func (r *Runner) execute(ctx context.Context, task mysqlinstance.Task) error {
	switch task.TaskType {
	case domaininstance.TaskTypeOperationSync:
		payload := parsePayload(task.Payload)
		if strings.TrimSpace(payload.InstanceNo) == "" && task.ObjectNo != nil {
			payload.InstanceNo = *task.ObjectNo
		}
		_, err := r.instanceSvc.SyncByWorker(ctx, payload.InstanceNo)
		return err
	case domaininstance.TaskTypeExpiryNotice:
		return r.expiryNotice(ctx, task)
	case domaininstance.TaskTypeExpiryRelease:
		return r.expiryRelease(ctx, task)
	case domaininstance.TaskTypePaymentProvision:
		return r.paymentOrderProvision(ctx, task)
	case domaininstance.TaskTypeEmailSend:
		return r.notificationEmailSend(ctx, task)
	case domaininstance.TaskTypeSMSPlaceholder:
		return r.notificationPlaceholder(ctx, task)
	default:
		return fmt.Errorf("不支持的任务类型：%s", task.TaskType)
	}
}

func (r *Runner) paymentOrderProvision(ctx context.Context, task mysqlinstance.Task) error {
	orderNo := strings.TrimSpace(pointerValue(task.ObjectNo))
	if orderNo == "" {
		return errors.New("支付自动交付任务缺少订单编号")
	}
	if err := r.preparePaymentProvisionOrder(ctx, orderNo); err != nil {
		if errors.Is(err, errPaymentProvisionSkipped) {
			return nil
		}
		return err
	}
	// 自动交付复用管理端交付规则；worker 不持有管理员身份，因此审计中的 admin_id 使用 0 表示系统触发。
	_, err := r.instanceSvc.Provision(ctx, 0, orderNo)
	return err
}

func (r *Runner) preparePaymentProvisionOrder(ctx context.Context, orderNo string) error {
	orders := r.orders
	if orders == nil {
		orders = mysqlorder.NewRepository(r.db)
	}
	tasks := r.tasks
	if tasks == nil {
		tasks = mysqlinstance.NewRepository(r.db)
	}
	return mysqltx.NewManager(r.db).WithinContext(ctx, func(tx *gorm.DB) error {
		order, err := orders.OrderForUpdate(ctx, tx, orderNo)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errPaymentProvisionSkipped
		}
		if err != nil {
			return err
		}
		if order.OrderType != domainorder.TypePurchase || order.PaymentStatus != domainorder.PaymentStatusPaid {
			return errPaymentProvisionSkipped
		}
		if order.Status != domainorder.StatusPending && order.Status != domainorder.StatusError {
			return errPaymentProvisionSkipped
		}
		if _, err := tasks.InstanceByOrderID(ctx, order.ID); err == nil {
			return errPaymentProvisionSkipped
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if order.Status == domainorder.StatusError {
			return orders.Update(ctx, tx, order.ID, map[string]any{"status": domainorder.StatusPending})
		}
		return nil
	})
}

func (r *Runner) expiryNotice(ctx context.Context, task mysqlinstance.Task) error {
	payload := parsePayload(task.Payload)
	instanceNo := firstNonEmpty(payload.InstanceNo, pointerValue(task.ObjectNo))
	instance, err := r.tasks.InstanceForUpdate(ctx, nil, instanceNo)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if !sameExpiresAt(instance.ExpiresAt, payload.ExpiresAt) {
		return nil
	}
	row, err := r.tasks.Detail(ctx, instanceNo)
	if err != nil {
		return err
	}
	return mysqltx.NewManager(r.db).WithinContext(ctx, func(tx *gorm.DB) error {
		if err := r.enqueueNotificationTask(ctx, tx, task.TaskNo, row, domaininstance.NotificationChannelEmail); err != nil {
			return err
		}
		if r.notifyCfg.SMSEnabled {
			if err := r.enqueueNotificationTask(ctx, tx, task.TaskNo, row, domaininstance.NotificationChannelSMS); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Runner) expiryRelease(ctx context.Context, task mysqlinstance.Task) error {
	if !r.lifecycleCfg.AutoReleaseEnabled {
		return nil
	}
	payload := parsePayload(task.Payload)
	instanceNo := firstNonEmpty(payload.InstanceNo, pointerValue(task.ObjectNo))
	instance, err := r.tasks.InstanceForUpdate(ctx, nil, instanceNo)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if !sameExpiresAt(instance.ExpiresAt, payload.ExpiresAt) || instance.Status == domaininstance.StatusReleased || instance.Status == domaininstance.StatusReleasing {
		return nil
	}
	if instance.ExpiresAt == nil || instance.ExpiresAt.After(time.Now()) {
		return nil
	}
	expectedExpiresAt, ok := parseExpiresAt(payload.ExpiresAt)
	if !ok {
		return nil
	}
	_, err = r.instanceSvc.ReleaseExpiredByWorker(ctx, instance.InstanceNo, expectedExpiresAt)
	return err
}

func (r *Runner) notificationEmailSend(ctx context.Context, task mysqlinstance.Task) error {
	payload := parsePayload(task.Payload)
	notificationNo := firstNonEmpty(payload.NotificationNo, pointerValue(task.ObjectNo))
	notification, err := r.tasks.NotificationByNo(ctx, notificationNo)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if notification.Status == domaininstance.NotificationStatusSent || notification.Status == domaininstance.NotificationStatusSkipped {
		return nil
	}
	now := time.Now()
	if !r.notifyCfg.EmailEnabled || r.mail == nil || !r.mail.Enabled() {
		if err := r.tasks.UpdateNotification(ctx, nil, notification.ID, map[string]any{"status": domaininstance.NotificationStatusSkipped, "sent_at": now}); err != nil {
			return err
		}
		return r.markInstanceNoticeSent(ctx, notification, now)
	}
	subject := "实例即将到期"
	if notification.Subject != nil && strings.TrimSpace(*notification.Subject) != "" {
		subject = strings.TrimSpace(*notification.Subject)
	}
	body := "实例即将到期，请及时处理续费。"
	if notification.ContentSummary != nil && strings.TrimSpace(*notification.ContentSummary) != "" {
		body = strings.TrimSpace(*notification.ContentSummary)
	}
	if err := r.mail.SendPlain(notification.Target, subject, body); err != nil {
		_ = r.tasks.UpdateNotification(ctx, nil, notification.ID, map[string]any{"status": domaininstance.NotificationStatusFailed, "error_code": "mail_send_failed", "error_message": "邮件发送失败"})
		return err
	}
	if err := r.tasks.UpdateNotification(ctx, nil, notification.ID, map[string]any{"status": domaininstance.NotificationStatusSent, "sent_at": now, "error_code": nil, "error_message": nil}); err != nil {
		return err
	}
	return r.markInstanceNoticeSent(ctx, notification, now)
}

func (r *Runner) notificationPlaceholder(ctx context.Context, task mysqlinstance.Task) error {
	now := time.Now()
	payload := parsePayload(task.Payload)
	notificationNo := firstNonEmpty(payload.NotificationNo, pointerValue(task.ObjectNo))
	if notificationNo != "" {
		notification, err := r.tasks.NotificationByNo(ctx, notificationNo)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			if updateErr := r.tasks.UpdateNotification(ctx, nil, notification.ID, map[string]any{"status": domaininstance.NotificationStatusSkipped, "sent_at": now, "error_code": nil, "error_message": nil}); updateErr != nil {
				return updateErr
			}
		}
	}
	result := `{"status":"skipped","reason":"placeholder"}`
	return r.tasks.UpdateTask(ctx, nil, task.ID, map[string]any{"result": result, "completed_at": now})
}

func (r *Runner) enqueueNotificationTask(ctx context.Context, tx *gorm.DB, sourceTaskNo string, row mysqlinstance.InstanceRow, channel string) error {
	suffix := "EMAIL"
	taskType := domaininstance.TaskTypeEmailSend
	target := row.Email
	status := domaininstance.NotificationStatusPending
	if channel == domaininstance.NotificationChannelSMS {
		suffix = "SMS"
		taskType = domaininstance.TaskTypeSMSPlaceholder
		target = "sms_placeholder"
	}
	notificationNo := fmt.Sprintf("NTF-%s-%s", sourceTaskNo, suffix)
	taskNo := fmt.Sprintf("TASK-%s", notificationNo)
	payload := map[string]string{"notification_no": notificationNo}
	data, _ := json.Marshal(payload)
	objectType := "notification"
	notification := mysqlinstance.Notification{
		NotificationNo:    notificationNo,
		UserID:            row.UserID,
		Channel:           channel,
		Scene:             "instance_expiry_notice",
		Target:            target,
		Status:            status,
		Subject:           stringPtr("实例即将到期"),
		ContentSummary:    stringPtr("实例 " + row.InstanceNo + " 即将到期。当前阶段续费订单需后台人工确认。"),
		RelatedObjectType: stringPtr("instance"),
		RelatedObjectNo:   stringPtr(row.InstanceNo),
		TaskNo:            stringPtr(taskNo),
	}
	if err := r.tasks.CreateNotificationIgnoreDuplicate(ctx, tx, &notification); err != nil {
		return err
	}
	idempotencyKey := "notification_send:" + notificationNo
	task := mysqlinstance.Task{TaskNo: taskNo, TaskType: taskType, IdempotencyKey: &idempotencyKey, Status: domaininstance.TaskStatusPending, ObjectType: &objectType, ObjectNo: &notificationNo, Payload: stringPtr(string(data)), MaxAttempts: 10, ScheduledAt: time.Now()}
	return r.tasks.CreateTaskIgnoreDuplicate(ctx, tx, &task)
}

func (r *Runner) markInstanceNoticeSent(ctx context.Context, notification mysqlinstance.Notification, sentAt time.Time) error {
	if notification.RelatedObjectType == nil || *notification.RelatedObjectType != "instance" || notification.RelatedObjectNo == nil {
		return nil
	}
	instance, err := r.tasks.InstanceForUpdate(ctx, nil, *notification.RelatedObjectNo)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return r.tasks.UpdateInstance(ctx, nil, instance.ID, map[string]any{"expire_notice_sent_at": sentAt})
}

func (r *Runner) markSucceeded(ctx context.Context, task mysqlinstance.Task) error {
	now := time.Now()
	return r.tasks.UpdateTask(ctx, nil, task.ID, map[string]any{"status": domaininstance.TaskStatusSucceeded, "locked_by": nil, "locked_until": nil, "last_error_code": nil, "last_error_message": nil, "completed_at": now})
}

func (r *Runner) markDeferred(ctx context.Context, task mysqlinstance.Task) error {
	return r.tasks.UpdateTask(ctx, nil, task.ID, map[string]any{"status": domaininstance.TaskStatusPending, "locked_by": nil, "locked_until": nil, "last_error_code": nil, "last_error_message": nil, "scheduled_at": time.Now().Add(retryDelay(task.Attempts))})
}

func (r *Runner) markFailedOrRetry(ctx context.Context, task mysqlinstance.Task, err error) error {
	message := "任务执行失败"
	if err != nil && strings.TrimSpace(err.Error()) != "" {
		message = err.Error()
	}
	if len(message) > 500 {
		message = message[:500]
	}
	updates := map[string]any{"locked_by": nil, "locked_until": nil, "last_error_code": "task_failed", "last_error_message": message}
	if task.Attempts >= task.MaxAttempts {
		now := time.Now()
		updates["status"] = domaininstance.TaskStatusFailed
		updates["completed_at"] = now
		if task.TaskType == domaininstance.TaskTypePaymentProvision {
			if updateErr := r.markPaymentProvisionError(ctx, task); updateErr != nil {
				return updateErr
			}
		}
	} else {
		updates["status"] = domaininstance.TaskStatusPending
		updates["scheduled_at"] = time.Now().Add(retryDelay(task.Attempts))
	}
	return r.tasks.UpdateTask(ctx, nil, task.ID, updates)
}

func (r *Runner) markPaymentProvisionError(ctx context.Context, task mysqlinstance.Task) error {
	orderNo := strings.TrimSpace(pointerValue(task.ObjectNo))
	if orderNo == "" {
		return nil
	}
	orders := r.orders
	if orders == nil {
		orders = mysqlorder.NewRepository(r.db)
	}
	return mysqltx.NewManager(r.db).WithinContext(ctx, func(tx *gorm.DB) error {
		order, err := orders.OrderForUpdate(ctx, tx, orderNo)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if order.OrderType != domainorder.TypePurchase || order.PaymentStatus != domainorder.PaymentStatusPaid {
			return nil
		}
		if order.Status != domainorder.StatusPending && order.Status != domainorder.StatusProvisioning {
			return nil
		}
		return orders.Update(ctx, tx, order.ID, map[string]any{"status": domainorder.StatusError})
	})
}

func (r *Runner) pollInterval() time.Duration {
	if r.workerCfg.PollIntervalSeconds <= 0 {
		return 5 * time.Second
	}
	return time.Duration(r.workerCfg.PollIntervalSeconds) * time.Second
}

func retryDelay(attempts int) time.Duration {
	if attempts < 1 {
		attempts = 1
	}
	if attempts > 6 {
		attempts = 6
	}
	return time.Duration(attempts*attempts) * time.Minute
}

func parsePayload(raw *string) taskPayload {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return taskPayload{}
	}
	var payload taskPayload
	_ = json.Unmarshal([]byte(*raw), &payload)
	return payload
}

func sameExpiresAt(value *time.Time, encoded string) bool {
	if strings.TrimSpace(encoded) == "" {
		return true
	}
	if value == nil {
		return false
	}
	parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(encoded))
	if err != nil {
		return false
	}
	return value.Truncate(time.Millisecond).Equal(parsed.Truncate(time.Millisecond))
}

func parseExpiresAt(encoded string) (time.Time, bool) {
	if strings.TrimSpace(encoded) == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(encoded))
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func pointerValue(value *string) string {
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

func stringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
