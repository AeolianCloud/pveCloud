package model

import "time"

// User 用户账户信息，角色字段用于 RBAC 权限判断。
type User struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	Email            string     `gorm:"size:120;uniqueIndex;not null" json:"email"`
	PasswordHash     string     `gorm:"size:255;not null" json:"-"`
	Role             string     `gorm:"size:20;default:user" json:"role"`
	Status           string     `gorm:"size:20;default:active" json:"status"`
	LoginFailedCount int        `gorm:"default:0" json:"login_failed_count"`
	LockedUntil      *time.Time `json:"locked_until"`
	EmailVerified    bool       `gorm:"default:false" json:"email_verified"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// Wallet 用户钱包，balance 表示可用余额，frozen_balance 用于预留冻结金额场景。
type Wallet struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	Balance       float64   `gorm:"type:decimal(12,2);default:0" json:"balance"`
	FrozenBalance float64   `gorm:"type:decimal(12,2);default:0" json:"frozen_balance"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// WalletLog 钱包流水，记录充值、消费、退款等资金变化。
type WalletLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Type      string    `gorm:"size:20;index;not null" json:"type"`
	Amount    float64   `gorm:"type:decimal(12,2);not null" json:"amount"`
	OrderID   *uint     `gorm:"index" json:"order_id"`
	Remark    string    `gorm:"size:255" json:"remark"`
	CreatedAt time.Time `json:"created_at"`
}

// Product 套餐商品定义，status 控制前台可见性。
type Product struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Name           string    `gorm:"size:120;not null" json:"name"`
	Description    string    `gorm:"type:text" json:"description"`
	RegionID       uint      `gorm:"index" json:"region_id"`
	CPU            int       `json:"cpu"`
	MemoryGB       int       `json:"memory_gb"`
	DiskGB         int       `json:"disk_gb"`
	BandwidthMbps  int       `json:"bandwidth_mbps"`
	DiskType       string    `gorm:"size:20" json:"disk_type"`
	OSOptions      string    `gorm:"type:text" json:"os_options"`
	IsCustomizable bool      `gorm:"default:false" json:"is_customizable"`
	MinCPU         int       `json:"min_cpu"`
	MaxCPU         int       `json:"max_cpu"`
	MinMemoryGB    int       `json:"min_memory_gb"`
	MaxMemoryGB    int       `json:"max_memory_gb"`
	MinDiskGB      int       `json:"min_disk_gb"`
	MaxDiskGB      int       `json:"max_disk_gb"`
	Status         string    `gorm:"size:20;default:draft;index" json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ProductPrice 存储商品在不同计费周期下的单价。
type ProductPrice struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ProductID    uint      `gorm:"index;not null" json:"product_id"`
	BillingCycle string    `gorm:"size:20;index;not null" json:"billing_cycle"`
	UnitPrice    float64   `gorm:"type:decimal(12,2);not null" json:"unit_price"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Order 订单主体，config_snapshot 锁定下单时规格和价格。
type Order struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"index;not null" json:"user_id"`
	ProductID      uint      `gorm:"index;not null" json:"product_id"`
	Amount         float64   `gorm:"type:decimal(12,2);not null" json:"amount"`
	BillingCycle   string    `gorm:"size:20;not null" json:"billing_cycle"`
	Status         string    `gorm:"size:20;default:pending;index" json:"status"`
	ConfigSnapshot string    `gorm:"type:json" json:"config_snapshot"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Instance 云主机实例记录，和订单一对一或一对多都可扩展。
type Instance struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	UserID        uint       `gorm:"index;not null" json:"user_id"`
	OrderID       uint       `gorm:"index;not null" json:"order_id"`
	PVEInstanceID string     `gorm:"size:100;index" json:"pve_instance_id"`
	Name          string     `gorm:"size:120" json:"name"`
	IP            string     `gorm:"size:64" json:"ip"`
	Status        string     `gorm:"size:20;default:pending;index" json:"status"`
	CPU           int        `json:"cpu"`
	MemoryGB      int        `json:"memory_gb"`
	DiskGB        int        `json:"disk_gb"`
	ExpireAt      *time.Time `gorm:"index" json:"expire_at"`
	DeletedAt     *time.Time `gorm:"index" json:"deleted_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// InstanceSnapshot 实例快照记录，便于查询和审计。
type InstanceSnapshot struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	InstanceID uint      `gorm:"index;not null" json:"instance_id"`
	Name       string    `gorm:"size:120;not null" json:"name"`
	Status     string    `gorm:"size:20;default:available" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Task 统一追踪异步任务状态。
type Task struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	OrderID    *uint     `gorm:"index" json:"order_id"`
	InstanceID *uint     `gorm:"index" json:"instance_id"`
	Type       string    `gorm:"size:50;not null" json:"type"`
	Status     string    `gorm:"size:20;default:pending;index" json:"status"`
	PveTaskID  string    `gorm:"size:100;index" json:"pve_task_id"`
	Message    string    `gorm:"size:255" json:"message"`
	Progress   int       `gorm:"default:0" json:"progress"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Ticket 工单主表。
type Ticket struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	InstanceID *uint     `gorm:"index" json:"instance_id"`
	Title      string    `gorm:"size:120;not null" json:"title"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	Priority   string    `gorm:"size:20;default:medium" json:"priority"`
	Status     string    `gorm:"size:20;default:open;index" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TicketReply 工单对话内容。
type TicketReply struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TicketID  uint      `gorm:"index;not null" json:"ticket_id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	IsAdmin   bool      `gorm:"default:false" json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
}

// Region 地域定义，供商品归属和筛选使用。
type Region struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Code      string    `gorm:"size:32;uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"size:80;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SystemConfig 系统级配置键值对。
type SystemConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ConfigKey string    `gorm:"size:80;uniqueIndex;not null" json:"config_key"`
	ConfigVal string    `gorm:"type:text;not null" json:"config_val"`
	Remark    string    `gorm:"size:255" json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AllModels 返回需要自动迁移的模型列表，便于 main 中统一调用。
func AllModels() []interface{} {
	return []interface{}{
		&User{},
		&Wallet{},
		&WalletLog{},
		&Product{},
		&ProductPrice{},
		&Order{},
		&Instance{},
		&InstanceSnapshot{},
		&Task{},
		&Ticket{},
		&TicketReply{},
		&Region{},
		&SystemConfig{},
	}
}
