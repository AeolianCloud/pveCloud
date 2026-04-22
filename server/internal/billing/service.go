package billing

import (
	"context"

	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
)

type QuoteResult struct {
	Cycle          string `json:"cycle"`
	OriginalAmount int64  `json:"original_amount"`
	DiscountAmount int64  `json:"discount_amount"`
	PayableAmount  int64  `json:"payable_amount"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
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
