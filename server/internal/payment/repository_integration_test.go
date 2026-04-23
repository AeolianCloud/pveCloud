package payment_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/billing"
	"github.com/AeolianCloud/pveCloud/server/internal/catalog"
	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
	"github.com/AeolianCloud/pveCloud/server/internal/order"
	"github.com/AeolianCloud/pveCloud/server/internal/payment"
	"github.com/AeolianCloud/pveCloud/server/internal/testutil"
)

func TestMarkPaymentSuccessTxMovesOrderAndCreatesSingleTask(t *testing.T) {
	db := testutil.OpenMariaDB(t)
	testutil.SeedUser(t, db, 21001)
	seed := testutil.SeedSaleableCatalogWithBase(t, db, 22000)

	catalogSvc := catalog.NewService(catalog.NewMySQLRepository(db), 15*time.Minute)
	billingSvc := &billingAdapter{svc: billing.NewService(billing.NewMySQLRepository(db))}
	paymentRepo := payment.NewMySQLRepository(db)
	paymentSvc := payment.NewServiceWithCallbackStore(paymentRepo, payment.NewMySQLCallbackStore(db))
	orderSvc := order.NewService(db, order.NewMySQLRepository(db), billingSvc, paymentSvc, catalogSvc)

	result, err := orderSvc.CreateOrder(context.Background(), order.CreateInput{
		UserID:   21001,
		SKUID:    seed.SKUID,
		RegionID: seed.RegionID,
		Cycle:    "month",
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	if err := paymentSvc.MarkPaymentSuccess(context.Background(), result.PaymentOrder.PaymentOrderNo, []byte(`{"status":"success"}`)); err != nil {
		t.Fatalf("mark payment success: %v", err)
	}
	if err := paymentSvc.MarkPaymentSuccess(context.Background(), result.PaymentOrder.PaymentOrderNo, []byte(`{"status":"success"}`)); err != nil {
		t.Fatalf("mark payment success duplicate: %v", err)
	}

	var orderStatus string
	if err := db.QueryRow(`SELECT order_status FROM orders WHERE id = ?`, result.Order.ID).Scan(&orderStatus); err != nil {
		t.Fatalf("load order status: %v", err)
	}
	if orderStatus != "paid" {
		t.Fatalf("expected paid order status, got %s", orderStatus)
	}

	assertTableCount(t, db, "async_tasks", 1)
}

func assertTableCount(t *testing.T, db *sql.DB, table string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&got); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if got != want {
		t.Fatalf("expected %d rows in %s, got %d", want, table, got)
	}
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
