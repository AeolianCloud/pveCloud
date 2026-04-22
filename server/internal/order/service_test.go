package order_test

import (
	"context"
	"testing"
	"time"

	"github.com/AeolianCloud/pveCloud/server/internal/catalog"
	"github.com/AeolianCloud/pveCloud/server/internal/order"
	"github.com/AeolianCloud/pveCloud/server/internal/payment"
)

type fakeOrderRepo struct {
	order             order.Order
	boundReservation  uint64
	createOrderCalled bool
}

func (f *fakeOrderRepo) CreateOrder(ctx context.Context, in order.CreateOrderParams) (order.Order, error) {
	f.createOrderCalled = true
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

func (f *fakeOrderRepo) BindReservation(ctx context.Context, orderID, reservationID uint64) error {
	f.boundReservation = reservationID
	return nil
}

type fakeBillingService struct{}

func (f *fakeBillingService) Quote(ctx context.Context, skuID uint64, cycle string) (order.BillingQuote, error) {
	return order.BillingQuote{
		Cycle:          cycle,
		OriginalAmount: 10000,
		DiscountAmount: 0,
		PayableAmount:  10000,
	}, nil
}

type fakePaymentService struct{}

func (f *fakePaymentService) CreatePendingPayment(ctx context.Context, orderID uint64, payableAmount int64) (payment.PaymentOrder, error) {
	return payment.PaymentOrder{
		ID:             7001,
		PaymentOrderNo: "P7001",
		OrderID:        orderID,
		PayStatus:      "pending",
		PayableAmount:  payableAmount,
	}, nil
}

type fakeCatalogService struct{}

func (f *fakeCatalogService) ReserveCapacity(ctx context.Context, in catalog.ReserveInput) (catalog.Reservation, error) {
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
	svc := order.NewService(&fakeOrderRepo{}, &fakeBillingService{}, &fakePaymentService{}, &fakeCatalogService{})
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
}
