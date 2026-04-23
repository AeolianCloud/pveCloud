package order_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/billing"
	"github.com/AeolianCloud/pveCloud/server/internal/catalog"
	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
	"github.com/AeolianCloud/pveCloud/server/internal/order"
	"github.com/AeolianCloud/pveCloud/server/internal/payment"
)

type fakeOrderRepo struct {
	order             order.Order
	boundReservation  uint64
	createOrderCalled bool
	createOrderQ      database.Querier
	bindReservationQ  database.Querier
}

func (f *fakeOrderRepo) CreateOrder(ctx context.Context, q database.Querier, in order.CreateOrderParams) (order.Order, error) {
	f.createOrderCalled = true
	f.createOrderQ = q
	f.order = order.Order{
		ID:             5001,
		OrderNo:        "O5001",
		UserID:         in.UserID,
		SKUID:          in.SKUID,
		RegionID:       in.RegionID,
		Status:         "pending_payment",
		Cycle:          in.Cycle,
		OriginalAmount: in.OriginalAmount,
		DiscountAmount: in.DiscountAmount,
		PayableAmount:  in.PayableAmount,
	}
	return f.order, nil
}

func (f *fakeOrderRepo) BindReservation(ctx context.Context, q database.Querier, orderID, reservationID uint64) error {
	f.bindReservationQ = q
	f.boundReservation = reservationID
	return nil
}

func (f *fakeOrderRepo) GetOrderByID(ctx context.Context, orderID uint64) (order.Order, error) {
	return f.order, nil
}

func (f *fakeOrderRepo) ListOrdersByUser(ctx context.Context, userID uint64) ([]order.Order, error) {
	return []order.Order{f.order}, nil
}

func (f *fakeOrderRepo) ListAllOrders(ctx context.Context) ([]order.Order, error) {
	return []order.Order{f.order}, nil
}

type fakeBillingService struct {
	createRecordCount int
	createRecordQ     database.Querier
}

func (f *fakeBillingService) Quote(ctx context.Context, skuID uint64, cycle string) (order.BillingQuote, error) {
	return order.BillingQuote{
		Cycle:          cycle,
		OriginalAmount: 10000,
		DiscountAmount: 0,
		PayableAmount:  10000,
	}, nil
}

func (f *fakeBillingService) CreateRecord(ctx context.Context, q database.Querier, in billing.CreateRecordInput) (billing.Record, error) {
	f.createRecordCount++
	f.createRecordQ = q
	return billing.Record{
		ID:             8001,
		OrderID:        in.OrderID,
		BillingType:    in.BillingType,
		Cycle:          in.Cycle,
		OriginalAmount: in.OriginalAmount,
		DiscountAmount: in.DiscountAmount,
		PayableAmount:  in.PayableAmount,
	}, nil
}

type fakePaymentService struct {
	err               error
	createPaymentCall int
	createPaymentQ    database.Querier
}

func (f *fakePaymentService) CreatePendingPayment(ctx context.Context, q database.Querier, orderID uint64, payableAmount int64) (payment.PaymentOrder, error) {
	f.createPaymentCall++
	f.createPaymentQ = q
	if f.err != nil {
		return payment.PaymentOrder{}, f.err
	}
	return payment.PaymentOrder{
		ID:             7001,
		PaymentOrderNo: "P7001",
		OrderID:        orderID,
		PayStatus:      "pending",
		PayableAmount:  payableAmount,
	}, nil
}

type fakeCatalogService struct {
	reserveQ database.Querier
}

func (f *fakeCatalogService) ReserveCapacityWithQuerier(ctx context.Context, q database.Querier, in catalog.ReserveInput) (catalog.Reservation, error) {
	f.reserveQ = q
	return catalog.Reservation{
		ID:        6001,
		UserID:    in.UserID,
		SKUID:     in.SKUID,
		RegionID:  in.RegionID,
		NodeID:    4001,
		Status:    "reserved",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}, nil
}

func TestCreateOrderBuildsBillingSnapshotAndPaymentOrder(t *testing.T) {
	db, state := openTestDB(t)
	orderRepo := &fakeOrderRepo{}
	billingSvc := &fakeBillingService{}
	paymentSvc := &fakePaymentService{}
	catalogSvc := &fakeCatalogService{}
	svc := order.NewService(db, orderRepo, billingSvc, paymentSvc, catalogSvc)

	result, err := svc.CreateOrder(context.Background(), order.CreateInput{
		UserID:   1001,
		SKUID:    2001,
		RegionID: 3001,
		Cycle:    "month",
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if result.Order.Status != "pending_payment" {
		t.Fatalf("expected order status pending_payment, got %s", result.Order.Status)
	}
	if result.Order.DiscountAmount != 0 {
		t.Fatalf("expected discount amount 0, got %d", result.Order.DiscountAmount)
	}
	if result.PaymentOrder.PaymentOrderNo == "" {
		t.Fatalf("expected payment order no to be set")
	}
	if orderRepo.boundReservation != 6001 {
		t.Fatalf("expected reservation 6001 to be bound, got %d", orderRepo.boundReservation)
	}
	if billingSvc.createRecordCount != 1 {
		t.Fatalf("expected billing record to be created once, got %d", billingSvc.createRecordCount)
	}
	if paymentSvc.createPaymentCall != 1 {
		t.Fatalf("expected payment order to be created once, got %d", paymentSvc.createPaymentCall)
	}
	if catalogSvc.reserveQ == nil {
		t.Fatalf("expected catalog reservation to receive transaction querier")
	}
	if orderRepo.createOrderQ == nil {
		t.Fatalf("expected order create to receive transaction querier")
	}
	if billingSvc.createRecordQ == nil {
		t.Fatalf("expected billing record create to receive transaction querier")
	}
	if orderRepo.bindReservationQ == nil {
		t.Fatalf("expected reservation binding to receive transaction querier")
	}
	if paymentSvc.createPaymentQ == nil {
		t.Fatalf("expected payment create to receive transaction querier")
	}
	if orderRepo.createOrderQ != catalogSvc.reserveQ ||
		billingSvc.createRecordQ != catalogSvc.reserveQ ||
		orderRepo.bindReservationQ != catalogSvc.reserveQ ||
		paymentSvc.createPaymentQ != catalogSvc.reserveQ {
		t.Fatalf("expected all collaborators to share the same transaction querier")
	}
	if state.commitCount != 1 || state.rollbackCount != 0 {
		t.Fatalf("expected committed transaction, got commits=%d rollbacks=%d", state.commitCount, state.rollbackCount)
	}
}

func TestCreateOrderRollsBackWhenPaymentCreationFails(t *testing.T) {
	db, state := openTestDB(t)
	paymentErr := errors.New("payment failed")
	svc := order.NewService(db, &fakeOrderRepo{}, &fakeBillingService{}, &fakePaymentService{err: paymentErr}, &fakeCatalogService{})

	_, err := svc.CreateOrder(context.Background(), order.CreateInput{
		UserID:   1001,
		SKUID:    2001,
		RegionID: 3001,
		Cycle:    "month",
	})
	if !errors.Is(err, paymentErr) {
		t.Fatalf("expected payment error, got %v", err)
	}
	if state.commitCount != 0 || state.rollbackCount != 1 {
		t.Fatalf("expected rolled back transaction, got commits=%d rollbacks=%d", state.commitCount, state.rollbackCount)
	}
}

type txState struct {
	mu            sync.Mutex
	beginCount    int
	commitCount   int
	rollbackCount int
}

type txDriver struct {
	state *txState
}

func (d *txDriver) Open(name string) (driver.Conn, error) {
	return &txConn{state: d.state}, nil
}

type txConn struct {
	state *txState
}

func (c *txConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported")
}

func (c *txConn) Close() error {
	return nil
}

func (c *txConn) Begin() (driver.Tx, error) {
	c.state.mu.Lock()
	c.state.beginCount++
	c.state.mu.Unlock()
	return &txHandle{state: c.state}, nil
}

func (c *txConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return c.Begin()
}

type txHandle struct {
	state *txState
}

func (t *txHandle) Commit() error {
	t.state.mu.Lock()
	t.state.commitCount++
	t.state.mu.Unlock()
	return nil
}

func (t *txHandle) Rollback() error {
	t.state.mu.Lock()
	t.state.rollbackCount++
	t.state.mu.Unlock()
	return nil
}

func openTestDB(t *testing.T) (*sql.DB, *txState) {
	t.Helper()

	state := &txState{}
	name := fmt.Sprintf("order-tx-test-%s", strings.NewReplacer("/", "-", " ", "-").Replace(t.Name()))
	sql.Register(name, &txDriver{state: state})

	db, err := sql.Open(name, "test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	return db, state
}
