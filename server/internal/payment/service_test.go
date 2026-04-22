package payment_test

import (
	"context"
	"testing"

	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
	"github.com/AeolianCloud/pveCloud/server/internal/payment"
)

type fakeTxRepo struct {
	hasSuccessfulCallback bool
	insertCallbackCount   int
	markSuccessCount      int
	insertTaskCount       int
}

func (f *fakeTxRepo) HasSuccessfulCallback(ctx context.Context, paymentOrderNo string) bool {
	return f.hasSuccessfulCallback
}

func (f *fakeTxRepo) InsertCallbackLog(ctx context.Context, paymentOrderNo string, rawPayload []byte) error {
	f.insertCallbackCount++
	return nil
}

func (f *fakeTxRepo) MarkSuccessAndMoveOrderPaid(ctx context.Context, paymentOrderNo string) (uint64, error) {
	f.markSuccessCount++
	return 5001, nil
}

func (f *fakeTxRepo) InsertPendingProvisionTask(ctx context.Context, orderID uint64) error {
	f.insertTaskCount++
	return nil
}

type fakePaymentRepo struct {
}

func (f *fakePaymentRepo) CreatePendingPayment(ctx context.Context, q database.Querier, orderID uint64, payableAmount int64) (payment.PaymentOrder, error) {
	return payment.PaymentOrder{
		ID:             7001,
		PaymentOrderNo: "P7001",
		OrderID:        orderID,
		PayStatus:      "pending",
		PayableAmount:  payableAmount,
	}, nil
}

func (f *fakePaymentRepo) GetByPaymentOrderNo(ctx context.Context, paymentOrderNo string) (payment.PaymentOrder, error) {
	return payment.PaymentOrder{PaymentOrderNo: paymentOrderNo}, nil
}

type fakeCallbackStore struct {
	txRepo *fakeTxRepo
}

func (f *fakeCallbackStore) WithTx(ctx context.Context, fn func(payment.CallbackTxRepository) error) error {
	return fn(f.txRepo)
}

func TestCreatePendingPaymentDelegatesToRepository(t *testing.T) {
	repo := &fakePaymentRepo{}
	svc := payment.NewService(repo)

	got, err := svc.CreatePendingPayment(context.Background(), nil, 5001, 10000)
	if err != nil {
		t.Fatalf("create pending payment: %v", err)
	}
	if got.PaymentOrderNo != "P7001" {
		t.Fatalf("expected payment order no P7001, got %s", got.PaymentOrderNo)
	}
}

func TestMarkPaymentSuccessIsIdempotent(t *testing.T) {
	store := &fakeCallbackStore{txRepo: &fakeTxRepo{}}
	svc := payment.NewServiceWithCallbackStore(nil, store)

	if err := svc.MarkPaymentSuccess(context.Background(), "P7001", []byte(`{"status":"success"}`)); err != nil {
		t.Fatalf("mark payment success: %v", err)
	}
	if store.txRepo.insertCallbackCount != 1 {
		t.Fatalf("expected callback log insert once, got %d", store.txRepo.insertCallbackCount)
	}
	if store.txRepo.markSuccessCount != 1 {
		t.Fatalf("expected mark success once, got %d", store.txRepo.markSuccessCount)
	}
	if store.txRepo.insertTaskCount != 1 {
		t.Fatalf("expected task insert once, got %d", store.txRepo.insertTaskCount)
	}

	store.txRepo.hasSuccessfulCallback = true
	if err := svc.MarkPaymentSuccess(context.Background(), "P7001", []byte(`{"status":"success"}`)); err != nil {
		t.Fatalf("mark payment success duplicate: %v", err)
	}
	if store.txRepo.insertCallbackCount != 1 {
		t.Fatalf("expected duplicate callback to skip insert, got %d", store.txRepo.insertCallbackCount)
	}
	if store.txRepo.markSuccessCount != 1 {
		t.Fatalf("expected duplicate callback to skip mark success, got %d", store.txRepo.markSuccessCount)
	}
	if store.txRepo.insertTaskCount != 1 {
		t.Fatalf("expected duplicate callback to skip task insert, got %d", store.txRepo.insertTaskCount)
	}
}
