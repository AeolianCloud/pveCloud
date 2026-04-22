package payment_test

import (
	"context"
	"testing"

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
	txRepo *fakeTxRepo
}

func (f *fakePaymentRepo) CreatePendingPayment(ctx context.Context, orderID uint64, payableAmount int64) (payment.PaymentOrder, error) {
	return payment.PaymentOrder{
		ID:             7001,
		PaymentOrderNo: "P7001",
		OrderID:        orderID,
		PayStatus:      "pending",
		PayableAmount:  payableAmount,
	}, nil
}

func (f *fakePaymentRepo) WithTx(ctx context.Context, fn func(payment.TxRepo) error) error {
	return fn(f.txRepo)
}

func TestMarkPaymentSuccessIsIdempotent(t *testing.T) {
	repo := &fakePaymentRepo{txRepo: &fakeTxRepo{}}
	svc := payment.NewService(repo)

	if err := svc.MarkPaymentSuccess(context.Background(), "P7001", []byte(`{"status":"success"}`)); err != nil {
		t.Fatalf("mark payment success: %v", err)
	}
	if repo.txRepo.insertCallbackCount != 1 {
		t.Fatalf("expected callback log insert once, got %d", repo.txRepo.insertCallbackCount)
	}
	if repo.txRepo.markSuccessCount != 1 {
		t.Fatalf("expected mark success once, got %d", repo.txRepo.markSuccessCount)
	}
	if repo.txRepo.insertTaskCount != 1 {
		t.Fatalf("expected task insert once, got %d", repo.txRepo.insertTaskCount)
	}

	repo.txRepo.hasSuccessfulCallback = true
	if err := svc.MarkPaymentSuccess(context.Background(), "P7001", []byte(`{"status":"success"}`)); err != nil {
		t.Fatalf("mark payment success duplicate: %v", err)
	}
	if repo.txRepo.insertCallbackCount != 1 {
		t.Fatalf("expected duplicate callback to skip insert, got %d", repo.txRepo.insertCallbackCount)
	}
	if repo.txRepo.markSuccessCount != 1 {
		t.Fatalf("expected duplicate callback to skip mark success, got %d", repo.txRepo.markSuccessCount)
	}
	if repo.txRepo.insertTaskCount != 1 {
		t.Fatalf("expected duplicate callback to skip task insert, got %d", repo.txRepo.insertTaskCount)
	}
}
