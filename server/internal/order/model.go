package order

import "github.com/AeolianCloud/pveCloud/server/internal/payment"

type Order struct {
	ID             uint64 `json:"id"`
	OrderNo        string `json:"order_no"`
	UserID         uint64 `json:"user_id"`
	SKUID          uint64 `json:"sku_id"`
	RegionID       uint64 `json:"region_id"`
	ReservationID  uint64 `json:"reservation_id"`
	Status         string `json:"status"`
	Cycle          string `json:"cycle"`
	OriginalAmount int64  `json:"original_amount"`
	DiscountAmount int64  `json:"discount_amount"`
	PayableAmount  int64  `json:"payable_amount"`
}

type BillingQuote struct {
	Cycle          string `json:"cycle"`
	OriginalAmount int64  `json:"original_amount"`
	DiscountAmount int64  `json:"discount_amount"`
	PayableAmount  int64  `json:"payable_amount"`
}

type CreateInput struct {
	UserID   uint64 `json:"user_id"`
	SKUID    uint64 `json:"sku_id"`
	RegionID uint64 `json:"region_id"`
	Cycle    string `json:"cycle"`
}

type CreateOrderParams struct {
	UserID         uint64 `json:"user_id"`
	SKUID          uint64 `json:"sku_id"`
	RegionID       uint64 `json:"region_id"`
	Cycle          string `json:"cycle"`
	OriginalAmount int64  `json:"original_amount"`
	DiscountAmount int64  `json:"discount_amount"`
	PayableAmount  int64  `json:"payable_amount"`
}

type CreateResult struct {
	Order        Order                `json:"order"`
	PaymentOrder payment.PaymentOrder `json:"payment_order"`
}
