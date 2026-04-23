package e2e_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/audit"
	"github.com/AeolianCloud/pveCloud/server/internal/billing"
	"github.com/AeolianCloud/pveCloud/server/internal/catalog"
	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
	"github.com/AeolianCloud/pveCloud/server/internal/instance"
	"github.com/AeolianCloud/pveCloud/server/internal/notification"
	"github.com/AeolianCloud/pveCloud/server/internal/order"
	"github.com/AeolianCloud/pveCloud/server/internal/payment"
	"github.com/AeolianCloud/pveCloud/server/internal/resource"
	"github.com/AeolianCloud/pveCloud/server/internal/task"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil"
)

type ProvisioningHarness struct {
	t      *testing.T
	db     *sql.DB
	userID uint64
	seed   testutil.CatalogSeed
}

type FlowResult struct {
	OrderStatus    string
	TaskStatus     string
	InstanceStatus string
	InstanceNo     string
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	return testutil.OpenMariaDB(t)
}

func (h *ProvisioningHarness) RunPaidProvisioningFlow() (FlowResult, error) {
	h.userID = 41001
	testutil.SeedUser(h.t, h.db, h.userID)
	h.seed = testutil.SeedSaleableCatalogWithBase(h.t, h.db, 42000)

	catalogSvc := catalog.NewService(catalog.NewMySQLRepository(h.db), 15*time.Minute)
	billingSvc := &billingAdapter{svc: billing.NewService(billing.NewMySQLRepository(h.db))}
	paymentRepo := payment.NewMySQLRepository(h.db)
	paymentSvc := payment.NewServiceWithCallbackStore(paymentRepo, payment.NewMySQLCallbackStore(h.db))
	orderSvc := order.NewService(h.db, order.NewMySQLRepository(h.db), billingSvc, paymentSvc, catalogSvc)

	createResult, err := orderSvc.CreateOrder(context.Background(), order.CreateInput{
		UserID:   h.userID,
		SKUID:    h.seed.SKUID,
		RegionID: h.seed.RegionID,
		Cycle:    "month",
	})
	if err != nil {
		return FlowResult{}, err
	}

	if err := paymentSvc.MarkPaymentSuccess(context.Background(), createResult.PaymentOrder.PaymentOrderNo, []byte(`{"status":"success"}`)); err != nil {
		return FlowResult{}, err
	}
	if err := paymentSvc.MarkPaymentSuccess(context.Background(), createResult.PaymentOrder.PaymentOrderNo, []byte(`{"status":"success"}`)); err != nil {
		return FlowResult{}, err
	}

	instanceSvc := instance.NewService(
		instance.NewMySQLRepository(h.db),
		resource.NewMockClient(),
		audit.NewService(audit.NewMySQLRepository(h.db)),
		notification.NewService(),
	)
	taskRepo := task.NewMySQLRepository(h.db)
	worker := task.NewWorker(
		taskRepo,
		task.NewMySQLLogRepository(h.db),
		task.NewDispatchingExecutor(func(ctx context.Context, orderID uint64) error {
			_, err := instanceSvc.HandleCreateInstanceTask(ctx, orderID)
			return err
		}),
		"e2e-worker",
	)

	if err := worker.RunOnce(context.Background()); err != nil {
		return FlowResult{}, err
	}

	if err := h.assertSingleProvisioningArtifacts(createResult.Order.ID); err != nil {
		return FlowResult{}, err
	}

	return h.loadFlowResult(createResult.Order.ID)
}

func (h *ProvisioningHarness) assertSingleProvisioningArtifacts(orderID uint64) error {
	var taskCount int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM async_tasks WHERE business_type = 'order' AND business_id = ?`, orderID).Scan(&taskCount); err != nil {
		return err
	}
	if taskCount != 1 {
		return fmt.Errorf("expected 1 async task, got %d", taskCount)
	}

	var instanceCount int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM instances WHERE order_id = ?`, orderID).Scan(&instanceCount); err != nil {
		return err
	}
	if instanceCount != 1 {
		return fmt.Errorf("expected 1 instance, got %d", instanceCount)
	}
	return nil
}

func (h *ProvisioningHarness) loadFlowResult(orderID uint64) (FlowResult, error) {
	var result FlowResult
	if err := h.db.QueryRow(`SELECT order_status FROM orders WHERE id = ?`, orderID).Scan(&result.OrderStatus); err != nil {
		return FlowResult{}, err
	}
	if err := h.db.QueryRow(`SELECT status FROM async_tasks WHERE business_type = 'order' AND business_id = ?`, orderID).Scan(&result.TaskStatus); err != nil {
		return FlowResult{}, err
	}
	if err := h.db.QueryRow(`SELECT instance_status, instance_no FROM instances WHERE order_id = ?`, orderID).Scan(&result.InstanceStatus, &result.InstanceNo); err != nil {
		return FlowResult{}, err
	}
	return result, nil
}

type billingAdapter struct {
	svc *billing.Service
}

func (a *billingAdapter) Quote(ctx context.Context, skuID uint64, cycle string) (order.BillingQuote, error) {
	row, err := a.svc.Quote(ctx, skuID, cycle)
	if err != nil {
		return order.BillingQuote{}, err
	}
	return order.BillingQuote{
		Cycle:          row.Cycle,
		OriginalAmount: row.OriginalAmount,
		DiscountAmount: row.DiscountAmount,
		PayableAmount:  row.PayableAmount,
	}, nil
}

func (a *billingAdapter) CreateRecord(ctx context.Context, q database.Querier, in billing.CreateRecordInput) (billing.Record, error) {
	return a.svc.CreateRecord(ctx, q, in)
}
