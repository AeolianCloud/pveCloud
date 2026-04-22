package catalog

import "time"

type Product struct {
	ID          uint64    `json:"id"`
	ProductNo   string    `json:"product_no"`
	ProductName string    `json:"product_name"`
	ProductType string    `json:"product_type"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SKU struct {
	ID            uint64    `json:"id"`
	SKUNo         string    `json:"sku_no"`
	ProductID     uint64    `json:"product_id"`
	SKUName       string    `json:"sku_name"`
	CPUCores      int       `json:"cpu_cores"`
	MemoryMB      int       `json:"memory_mb"`
	DiskGB        int       `json:"disk_gb"`
	BandwidthMbps int       `json:"bandwidth_mbps"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SaleableProduct struct {
	Product Product `json:"product"`
	SKUs    []SKU   `json:"skus"`
}

type ResourceNode struct {
	ID                uint64    `json:"id"`
	NodeNo            string    `json:"node_no"`
	RegionID          uint64    `json:"region_id"`
	NodeName          string    `json:"node_name"`
	TotalInstances    int       `json:"total_instances"`
	UsedInstances     int       `json:"used_instances"`
	ReservedInstances int       `json:"reserved_instances"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Reservation struct {
	ID            uint64    `json:"id"`
	ReservationNo string    `json:"reservation_no"`
	UserID        uint64    `json:"user_id"`
	SKUID         uint64    `json:"sku_id"`
	RegionID      uint64    `json:"region_id"`
	NodeID        uint64    `json:"node_id"`
	Status        string    `json:"status"`
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ReserveInput struct {
	UserID   uint64 `json:"user_id"`
	SKUID    uint64 `json:"sku_id"`
	RegionID uint64 `json:"region_id"`
}

type CreateSKUInput struct {
	SKUName       string `json:"sku_name"`
	CPUCores      int    `json:"cpu_cores"`
	MemoryMB      int    `json:"memory_mb"`
	DiskGB        int    `json:"disk_gb"`
	BandwidthMbps int    `json:"bandwidth_mbps"`
	Status        string `json:"status"`
}
