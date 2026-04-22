package billing

import (
	"context"

	"github.com/AeolianCloud/pveCloud/server/internal/common/database"
	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
)

type QuoteResult struct {
	Cycle          string `json:"cycle"`
	OriginalAmount int64  `json:"original_amount"`
	DiscountAmount int64  `json:"discount_amount"`
	PayableAmount  int64  `json:"payable_amount"`
}

type Record struct {
	ID             uint64 `json:"id"`
	OrderID        uint64 `json:"order_id"`
	BillingType    string `json:"billing_type"`
	Cycle          string `json:"cycle"`
	OriginalAmount int64  `json:"original_amount"`
	DiscountAmount int64  `json:"discount_amount"`
	PayableAmount  int64  `json:"payable_amount"`
}

type CreateRecordInput struct {
	OrderID        uint64 `json:"order_id"`
	BillingType    string `json:"billing_type"`
	Cycle          string `json:"cycle"`
	OriginalAmount int64  `json:"original_amount"`
	DiscountAmount int64  `json:"discount_amount"`
	PayableAmount  int64  `json:"payable_amount"`
}

type Repository interface {
	CreateRecord(ctx context.Context, q database.Querier, in CreateRecordInput) (Record, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Quote(ctx context.Context, skuID uint64, cycle string) (QuoteResult, error) {
	if skuID == 0 || cycle == "" {
		return QuoteResult{}, errorsx.ErrBadRequest
	}

	priceByCycle := map[string]int64{
		"month":   10000,
		"quarter": 27000,
		"year":    100000,
	}
	originalAmount, ok := priceByCycle[cycle]
	if !ok {
		return QuoteResult{}, errorsx.ErrBadRequest
	}

	return QuoteResult{
		Cycle:          cycle,
		OriginalAmount: originalAmount,
		DiscountAmount: 0,
		PayableAmount:  originalAmount,
	}, nil
}

func (s *Service) CreateRecord(ctx context.Context, q database.Querier, in CreateRecordInput) (Record, error) {
	if in.OrderID == 0 || in.Cycle == "" || in.BillingType == "" || in.PayableAmount <= 0 {
		return Record{}, errorsx.ErrBadRequest
	}
	if s.repo == nil {
		return Record{}, errorsx.ErrInternal
	}

	return s.repo.CreateRecord(ctx, q, in)
}
