package order_test

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

func TestCreateOrderTxPersistsOrderBillingPaymentAndReservation(t *testing.T) {
	db := testutil.OpenMariaDB(t)
	testutil.SeedUser(t, db, 11001)
	seed := testutil.SeedSaleableCatalogWithBase(t, db, 12000)

	catalogSvc := catalog.NewService(catalog.NewMySQLRepository(db), 15*time.Minute)
	billingSvc := &billingAdapter{svc: billing.NewService(billing.NewMySQLRepository(db))}
	paymentSvc := payment.NewService(payment.NewMySQLRepository(db))
	orderSvc := order.NewService(db, order.NewMySQLRepository(db), billingSvc, paymentSvc, catalogSvc)

	result, err := orderSvc.CreateOrder(context.Background(), order.CreateInput{
		UserID:   11001,
		SKUID:    seed.SKUID,
		RegionID: seed.RegionID,
		Cycle:    "month",
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if result.Order.ID == 0 || result.PaymentOrder.ID == 0 || result.Order.ReservationID == 0 {
		t.Fatalf("unexpected create result: %+v", result)
	}

	assertCount(t, db, "orders", 1)
	assertCount(t, db, "billing_records", 1)
	assertCount(t, db, "payment_orders", 1)
	assertCount(t, db, "resource_reservations", 1)
}

func assertCount(t *testing.T, db *sql.DB, table string, want int) {
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
