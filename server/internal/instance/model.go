package instance

import "time"

type Instance struct {
	ID          uint64    `json:"id"`
	InstanceNo  string    `json:"instance_no"`
	UserID      uint64    `json:"user_id"`
	OrderID     uint64    `json:"order_id"`
	NodeID      uint64    `json:"node_id"`
	Status      string    `json:"status"`
	InstanceRef string    `json:"instance_ref"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ServiceFact struct {
	ID                   uint64    `json:"id"`
	InstanceID           uint64    `json:"instance_id"`
	CurrentPeriodStartAt time.Time `json:"current_period_start_at"`
	CurrentPeriodEndAt   time.Time `json:"current_period_end_at"`
	BillingStatus        string    `json:"billing_status"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type PaidOrder struct {
	ID            uint64 `json:"id"`
	OrderNo       string `json:"order_no"`
	UserID        uint64 `json:"user_id"`
	SKUID         uint64 `json:"sku_id"`
	RegionID      uint64 `json:"region_id"`
	Cycle         string `json:"cycle"`
	PayableAmount int64  `json:"payable_amount"`
}

type ProvisionResult struct {
	Instance Instance    `json:"instance"`
	Service  ServiceFact `json:"service"`
}
