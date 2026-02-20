package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"pvecloud/backend/internal/model"
	"pvecloud/backend/internal/pveclient"
	"pvecloud/backend/internal/repository"

	"gorm.io/gorm"
)

var (
	errInsufficientBalance = errors.New("余额不足，请充值")
	errProductUnavailable  = errors.New("商品已下架")
	errOrderNotFound       = errors.New("订单不存在")
	errInvalidBillingCycle = errors.New("无效计费周期")
)

// OrderService 封装下单、续费、任务异步处理等流程。
type OrderService struct {
	db           *gorm.DB
	productRepo  *repository.ProductRepository
	walletRepo   *repository.WalletRepository
	orderRepo    *repository.OrderRepository
	taskRepo     *repository.TaskRepository
	instanceRepo *repository.InstanceRepository
	pve          pveclient.PVEClient
}

// NewOrderService 创建订单服务。
func NewOrderService(
	db *gorm.DB,
	productRepo *repository.ProductRepository,
	walletRepo *repository.WalletRepository,
	orderRepo *repository.OrderRepository,
	taskRepo *repository.TaskRepository,
	instanceRepo *repository.InstanceRepository,
	pve pveclient.PVEClient,
) *OrderService {
	return &OrderService{db: db, productRepo: productRepo, walletRepo: walletRepo, orderRepo: orderRepo, taskRepo: taskRepo, instanceRepo: instanceRepo, pve: pve}
}

// CreateOrderRequest 描述下单请求体。
type CreateOrderRequest struct {
	ProductID    uint   `json:"product_id"`
	BillingCycle string `json:"billing_cycle"`
	OS           string `json:"os"`
	CPU          int    `json:"cpu"`
	MemoryGB     int    `json:"memory_gb"`
	DiskGB       int    `json:"disk_gb"`
}

// CreateOrder 下单事务：检查商品/余额，扣费，创建订单和任务，并异步触发 PVE。
// CreateOrder 创建订单并执行相关业务逻辑
// ctx: 上下文信息，用于传递请求范围的数据和控制取消信号
// userID: 用户ID，标识创建订单的用户
// req: 创建订单请求，包含订单所需的各种信息
// 返回值: 创建的订单、任务对象和可能的错误
func (s *OrderService) CreateOrder(ctx context.Context, userID uint, req CreateOrderRequest) (*model.Order, *model.Task, error) {
	var createdOrder model.Order // 创建的订单对象
	var createdTask model.Task   // 创建的任务对象

	// 使用数据库事务确保订单创建的原子性
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var product model.Product
		if err := tx.WithContext(ctx).First(&product, req.ProductID).Error; err != nil {
			return err
		}
		// 检查产品是否可用
		if product.Status != "published" {
			return errProductUnavailable
		}

		cpu, memoryGB, diskGB, err := resolveOrderSpec(&product, req)
		if err != nil {
			return err
		}
		req.CPU = cpu
		req.MemoryGB = memoryGB
		req.DiskGB = diskGB

		var prices []model.ProductPrice
		if err := tx.WithContext(ctx).Where("product_id = ?", req.ProductID).Find(&prices).Error; err != nil {
			return err
		}

		var unitPrice float64 // 单价
		found := false        // 是否找到匹配的计费周期
		// 查找匹配的计费周期价格
		for _, price := range prices {
			if price.BillingCycle == req.BillingCycle {
				unitPrice = price.UnitPrice
				found = true
				break
			}
		}
		if !found {
			return errInvalidBillingCycle
		}

		// 创建配置快照
		snapshotMap := map[string]interface{}{
			"cpu":           cpu,
			"memory":        memoryGB,
			"disk":          diskGB,
			"bandwidth":     product.BandwidthMbps,
			"os":            req.OS,
			"unit_price":    unitPrice,
			"billing_cycle": req.BillingCycle,
			"total_amount":  unitPrice,
		}
		snapshotBytes, _ := json.Marshal(snapshotMap)

		// 创建订单记录
		createdOrder = model.Order{UserID: userID, ProductID: req.ProductID, Amount: unitPrice, BillingCycle: req.BillingCycle, Status: "pending", ConfigSnapshot: string(snapshotBytes)}
		if err := tx.Create(&createdOrder).Error; err != nil {
			return err
		}

		// 扣除用户余额
		orderID := createdOrder.ID
		if err := s.walletRepo.ChangeBalanceWithTx(ctx, tx, userID, -unitPrice, "consume", &orderID, "下单扣费"); err != nil {
			if strings.Contains(err.Error(), "insufficient balance") {
				return errInsufficientBalance
			}
			return err
		}

		createdTask = model.Task{UserID: userID, OrderID: &orderID, Type: "create_instance", Status: "pending", Progress: 0, Message: "订单已创建，等待下发"}
		return tx.Create(&createdTask).Error
	})
	if err != nil {
		return nil, nil, err
	}

	// 异步调用 PVE，不阻塞下单接口响应。
	go s.provisionInstance(context.Background(), createdOrder, createdTask, req)
	return &createdOrder, &createdTask, nil
}

func (s *OrderService) provisionInstance(ctx context.Context, order model.Order, task model.Task, req CreateOrderRequest) {
	result, err := s.pve.CreateInstance(ctx, pveclient.CreateInstanceReq{
		Name:          fmt.Sprintf("order-%d", order.ID),
		CPU:           req.CPU,
		MemoryMB:      req.MemoryGB * 1024,
		DiskGB:        req.DiskGB,
		BandwidthMbps: 100,
		Template:      req.OS,
		Password:      "ChangeMe123!",
		RegionCode:    "default",
	})
	if err != nil {
		s.markOrderFailedAndRefund(ctx, order, task, "调用 PVE 失败")
		return
	}

	task.PveTaskID = result.PveTaskID
	task.Status = "running"
	task.Message = "实例创建中"
	task.Progress = 50
	_ = s.taskRepo.Update(ctx, &task)

	status, err := s.pve.GetTaskStatus(ctx, result.TaskID)
	if err != nil || status.Status == "failed" {
		s.markOrderFailedAndRefund(ctx, order, task, "PVE 返回创建失败")
		return
	}

	now := time.Now()
	expire := now.Add(30 * 24 * time.Hour)
	instance := &model.Instance{
		UserID:        order.UserID,
		OrderID:       order.ID,
		PVEInstanceID: result.InstanceID,
		Name:          fmt.Sprintf("vm-%d", order.ID),
		IP:            "10.0.0.8",
		Status:        "active",
		CPU:           req.CPU,
		MemoryGB:      req.MemoryGB,
		DiskGB:        req.DiskGB,
		ExpireAt:      &expire,
	}
	_ = s.instanceRepo.Create(ctx, instance)

	order.Status = "active"
	_ = s.orderRepo.Update(ctx, &order)

	task.Status = "success"
	task.Message = "实例创建完成"
	task.Progress = 100
	task.InstanceID = &instance.ID
	_ = s.taskRepo.Update(ctx, &task)
}

func (s *OrderService) markOrderFailedAndRefund(ctx context.Context, order model.Order, task model.Task, message string) {
	order.Status = "failed"
	_ = s.orderRepo.Update(ctx, &order)
	orderID := order.ID
	_ = s.walletRepo.ChangeBalance(ctx, order.UserID, order.Amount, "refund", &orderID, "实例创建失败自动退款")

	task.Status = "failed"
	task.Message = message
	task.Progress = 100
	_ = s.taskRepo.Update(ctx, &task)
}

// ListOrders 查询用户订单。
func (s *OrderService) ListOrders(ctx context.Context, userID uint, status string) ([]model.Order, error) {
	return s.orderRepo.ListByUser(ctx, userID, status)
}

// GetOrderDetail 查询用户订单详情并做归属校验。
func (s *OrderService) GetOrderDetail(ctx context.Context, userID uint, orderID uint) (*model.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, errOrderNotFound
	}
	if order.UserID != userID {
		return nil, WrapForbidden("无权限访问该订单")
	}
	return order, nil
}

// RenewOrder 续费：从钱包扣费并延长实例到期时间。
func (s *OrderService) RenewOrder(ctx context.Context, userID uint, orderID uint, amount float64) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return errOrderNotFound
	}
	if order.UserID != userID {
		return WrapForbidden("无权限续费该订单")
	}
	if order.Status != "active" {
		return errors.New("仅 active 订单可续费")
	}

	if err := s.walletRepo.ChangeBalance(ctx, userID, -amount, "consume", &orderID, "订单续费扣费"); err != nil {
		return err
	}

	instanceList, err := s.instanceRepo.ListByUser(ctx, userID)
	if err != nil {
		return err
	}
	for i := range instanceList {
		if instanceList[i].OrderID == orderID && instanceList[i].ExpireAt != nil {
			next := instanceList[i].ExpireAt.Add(30 * 24 * time.Hour)
			instanceList[i].ExpireAt = &next
			_ = s.instanceRepo.Update(ctx, &instanceList[i])
			break
		}
	}
	return nil
}

// GetTaskStatus 查询任务状态。
func (s *OrderService) GetTaskStatus(ctx context.Context, userID uint, taskID uint) (*model.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.UserID != userID {
		return nil, WrapForbidden("无权限查看该任务")
	}
	return task, nil
}

// ListAdminOrders 查询后台订单列表。
func (s *OrderService) ListAdminOrders(ctx context.Context, userID uint, status string, dateRange string) ([]model.Order, error) {
	return s.orderRepo.ListForAdmin(ctx, userID, status, dateRange)
}

func resolveOrderSpec(product *model.Product, req CreateOrderRequest) (int, int, int, error) {
	// 非自定义商品强制使用商品默认规格，避免客户端伪造参数。
	if !product.IsCustomizable {
		return product.CPU, product.MemoryGB, product.DiskGB, nil
	}

	cpu := req.CPU
	memoryGB := req.MemoryGB
	diskGB := req.DiskGB
	if cpu == 0 {
		cpu = product.CPU
	}
	if memoryGB == 0 {
		memoryGB = product.MemoryGB
	}
	if diskGB == 0 {
		diskGB = product.DiskGB
	}

	if err := validateRange("CPU", cpu, product.MinCPU, product.MaxCPU); err != nil {
		return 0, 0, 0, err
	}
	if err := validateRange("内存", memoryGB, product.MinMemoryGB, product.MaxMemoryGB); err != nil {
		return 0, 0, 0, err
	}
	if err := validateRange("磁盘", diskGB, product.MinDiskGB, product.MaxDiskGB); err != nil {
		return 0, 0, 0, err
	}
	return cpu, memoryGB, diskGB, nil
}

func validateRange(label string, value int, min int, max int) error {
	if min <= 0 && max <= 0 {
		if value <= 0 {
			return fmt.Errorf("%s 规格无效", label)
		}
		return nil
	}
	if value < min || value > max {
		return fmt.Errorf("%s 超出可选范围(%d-%d)", label, min, max)
	}
	return nil
}
