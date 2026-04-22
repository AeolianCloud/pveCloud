package billing_test

import (
	"context"
	"testing"

	"github.com/AeolianCloud/pveCloud/server/internal/billing"
	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
)

type fakeBillingRepo struct {
	lastInput billing.CreateRecordInput
}

func (f *fakeBillingRepo) CreateRecord(ctx context.Context, q database.Querier, in billing.CreateRecordInput) (billing.Record, error) {
	f.lastInput = in
	return billing.Record{
		ID:             9001,
		OrderID:        in.OrderID,
		BillingType:    in.BillingType,
		Cycle:          in.Cycle,
		OriginalAmount: in.OriginalAmount,
		DiscountAmount: in.DiscountAmount,
		PayableAmount:  in.PayableAmount,
	}, nil
}

func TestCreateRecordDelegatesToRepository(t *testing.T) {
	repo := &fakeBillingRepo{}
	svc := billing.NewService(repo)

	record, err := svc.CreateRecord(context.Background(), nil, billing.CreateRecordInput{
		OrderID:        5001,
		BillingType:    "create",
		Cycle:          "month",
		OriginalAmount: 10000,
		DiscountAmount: 0,
		PayableAmount:  10000,
	})
	if err != nil {
		t.Fatalf("create record: %v", err)
	}
	if record.ID != 9001 {
		t.Fatalf("expected record id 9001, got %d", record.ID)
	}
	if repo.lastInput.OrderID != 5001 {
		t.Fatalf("expected order id 5001, got %d", repo.lastInput.OrderID)
	}
}
