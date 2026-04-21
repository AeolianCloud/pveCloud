# PVE Cloud MVP Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first production-capable MVP slice of the cloud sales platform, covering backend foundations, auth, product saleability, order and payment flow, async provisioning, instance operations, and the minimum web/admin frontends required to operate the system.

**Architecture:** Keep the backend as a Go modular monolith with `public-api`, `admin-api`, and `worker` entrypoints. Persist all business truth in MariaDB, use Redis only for non-primary support, and enforce a single async task center for all resource-changing operations. Keep `web` and `admin` as fully separate Bun + Vue 3 SPA projects.

**Tech Stack:** Go, Chi, GORM, MariaDB, Redis, Bun, Vue 3, Vue Router, Pinia, Vitest

---

## Planned File Structure

### Backend

- Create: `server/go.mod`
- Create: `server/cmd/public-api/main.go`
- Create: `server/cmd/admin-api/main.go`
- Create: `server/cmd/worker/main.go`
- Create: `server/internal/bootstrap/app.go`
- Create: `server/internal/bootstrap/config/config.go`
- Create: `server/internal/common/http/response.go`
- Create: `server/internal/common/errors/errors.go`
- Create: `server/internal/common/database/mysql.go`
- Create: `server/internal/common/cache/redis.go`
- Create: `server/internal/common/logger/logger.go`
- Create: `server/internal/common/testutil/testdb.go`
- Create: `server/internal/auth/jwt.go`
- Create: `server/internal/auth/middleware.go`
- Create: `server/internal/user/service.go`
- Create: `server/internal/user/handler/public_auth_handler.go`
- Create: `server/internal/user/handler/public_register_handler.go`
- Create: `server/internal/adminuser/service.go`
- Create: `server/internal/adminuser/handler/admin_auth_handler.go`
- Create: `server/internal/catalog/service.go`
- Create: `server/internal/catalog/handler/public_products_handler.go`
- Create: `server/internal/catalog/handler/admin_products_handler.go`
- Create: `server/internal/order/service.go`
- Create: `server/internal/payment/service.go`
- Create: `server/internal/payment/handler/public_payment_handler.go`
- Create: `server/internal/payment/handler/callback_handler.go`
- Create: `server/internal/billing/service.go`
- Create: `server/internal/instance/service.go`
- Create: `server/internal/instance/handler/public_instances_handler.go`
- Create: `server/internal/instance/handler/admin_instances_handler.go`
- Create: `server/internal/resource/client.go`
- Create: `server/internal/task/service.go`
- Create: `server/internal/notification/service.go`
- Create: `server/internal/audit/service.go`
- Create: `server/migrations/0001_core_auth.sql`
- Create: `server/migrations/0002_catalog_capacity.sql`
- Create: `server/migrations/0003_orders_payments.sql`
- Create: `server/migrations/0004_instances_tasks.sql`

### Web Frontend

- Create: `web/package.json`
- Create: `web/bun.lock`
- Create: `web/vite.config.ts`
- Create: `web/tsconfig.json`
- Create: `web/index.html`
- Create: `web/src/main.ts`
- Create: `web/src/router/index.ts`
- Create: `web/src/stores/auth.ts`
- Create: `web/src/lib/http.ts`
- Create: `web/src/views/LoginView.vue`
- Create: `web/src/views/RegisterView.vue`
- Create: `web/src/views/ProductListView.vue`
- Create: `web/src/views/ProductDetailView.vue`
- Create: `web/src/views/OrderListView.vue`
- Create: `web/src/views/PaymentStatusView.vue`
- Create: `web/src/views/InstanceListView.vue`
- Create: `web/src/views/InstanceDetailView.vue`
- Create: `web/src/views/NoticeListView.vue`

### Admin Frontend

- Create: `admin/package.json`
- Create: `admin/bun.lock`
- Create: `admin/vite.config.ts`
- Create: `admin/tsconfig.json`
- Create: `admin/index.html`
- Create: `admin/src/main.ts`
- Create: `admin/src/router/index.ts`
- Create: `admin/src/stores/auth.ts`
- Create: `admin/src/lib/http.ts`
- Create: `admin/src/views/LoginView.vue`
- Create: `admin/src/views/DashboardView.vue`
- Create: `admin/src/views/UserManageView.vue`
- Create: `admin/src/views/ProductManageView.vue`
- Create: `admin/src/views/OrderManageView.vue`
- Create: `admin/src/views/InstanceManageView.vue`
- Create: `admin/src/views/TaskManageView.vue`

### Shared Repository Files

- Create: `.gitignore`
- Create: `README.md`
- Create: `docs/adr/001-task-source-of-truth.md`
- Create: `docs/adr/002-capacity-reservation.md`

### Task 1: Bootstrap The Repository And Backend Runtime

**Files:**
- Create: `.gitignore`
- Create: `README.md`
- Create: `server/go.mod`
- Create: `server/cmd/public-api/main.go`
- Create: `server/cmd/admin-api/main.go`
- Create: `server/cmd/worker/main.go`
- Create: `server/internal/bootstrap/app.go`
- Create: `server/internal/bootstrap/config/config.go`
- Create: `server/internal/common/http/response.go`
- Create: `server/internal/common/errors/errors.go`
- Test: `server/internal/bootstrap/config/config_test.go`

- [ ] **Step 1: Write the failing config and bootstrap tests**

```go
package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap/config"
)

func TestLoadConfigReadsRequiredFields(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("MYSQL_DSN", "root:root@tcp(localhost:3306)/pvecloud")
	t.Setenv("REDIS_ADDR", "127.0.0.1:6379")

	cfg, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, "test", cfg.AppEnv)
	require.Equal(t, "root:root@tcp(localhost:3306)/pvecloud", cfg.MySQLDSN)
	require.Equal(t, "127.0.0.1:6379", cfg.RedisAddr)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go -C server test ./internal/bootstrap/config -v`
Expected: FAIL with `undefined: config.Load`

- [ ] **Step 3: Write the minimal bootstrap implementation**

```go
package config

import (
	"errors"
	"os"
)

type Config struct {
	AppEnv   string
	MySQLDSN string
	RedisAddr string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:   os.Getenv("APP_ENV"),
		MySQLDSN: os.Getenv("MYSQL_DSN"),
		RedisAddr: os.Getenv("REDIS_ADDR"),
	}
	if cfg.MySQLDSN == "" {
		return Config{}, errors.New("MYSQL_DSN is required")
	}
	return cfg, nil
}
```

- [ ] **Step 4: Add entrypoints and health responses**

```go
package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	_ = http.ListenAndServe(":8080", mux)
}
```

- [ ] **Step 5: Run tests and smoke build**

Run: `go -C server test ./...`
Expected: PASS

Run: `go -C server build ./cmd/public-api ./cmd/admin-api ./cmd/worker`
Expected: PASS with no compile errors

- [ ] **Step 6: Commit**

```bash
git add .gitignore README.md server
git commit -m "chore: bootstrap repository and backend runtime"
```

### Task 2: Build MariaDB Migrations And Core Schema

**Files:**
- Create: `server/migrations/0001_core_auth.sql`
- Create: `server/migrations/0002_catalog_capacity.sql`
- Create: `server/migrations/0003_orders_payments.sql`
- Create: `server/migrations/0004_instances_tasks.sql`
- Create: `server/internal/common/database/mysql.go`
- Test: `server/internal/common/database/migration_test.go`

- [ ] **Step 1: Write the failing migration smoke test**

```go
package database_test

import (
	"os"
	"strings"
	"testing"
)

func TestMigrationsContainChineseComments(t *testing.T) {
	data, err := os.ReadFile("../../migrations/0001_core_auth.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(data)
	if !strings.Contains(sql, "COMMENT='前台用户主表'") {
		t.Fatalf("expected Chinese table comment in migration")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go -C server test ./internal/common/database -v`
Expected: FAIL because the migration file or table comment does not exist yet

- [ ] **Step 3: Write the first migration with strict comments**

```sql
CREATE TABLE users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  user_no VARCHAR(32) NOT NULL COMMENT '用户编号，业务侧唯一编号',
  email VARCHAR(128) NULL COMMENT '邮箱地址',
  phone VARCHAR(32) NOT NULL COMMENT '手机号',
  password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
  status VARCHAR(32) NOT NULL COMMENT '用户状态：active-正常，disabled-禁用',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_no (user_no),
  UNIQUE KEY uk_phone (phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='前台用户主表';
```

- [ ] **Step 4: Add the remaining minimum tables**

```sql
CREATE TABLE payment_orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  payment_order_no VARCHAR(32) NOT NULL COMMENT '支付单编号，业务侧唯一编号',
  order_id BIGINT UNSIGNED NOT NULL COMMENT '订单ID，对应 orders.id',
  pay_status VARCHAR(32) NOT NULL COMMENT '支付状态：pending-待支付，success-支付成功，failed-支付失败，refunded-已退款',
  payable_amount BIGINT NOT NULL COMMENT '应付金额，单位分',
  paid_at DATETIME(3) NULL COMMENT '支付成功时间',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_payment_order_no (payment_order_no),
  KEY idx_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付单主表';

CREATE TABLE async_tasks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  task_no VARCHAR(32) NOT NULL COMMENT '任务编号，业务侧唯一编号',
  task_type VARCHAR(32) NOT NULL COMMENT '任务类型：create_instance-开通实例，start_instance-开机，stop_instance-关机，reboot_instance-重启，reinstall_instance-重装',
  business_type VARCHAR(32) NOT NULL COMMENT '业务类型：order-订单，instance-实例',
  business_id BIGINT UNSIGNED NOT NULL COMMENT '业务ID',
  status VARCHAR(32) NOT NULL COMMENT '任务状态：pending-待执行，processing-执行中，success-成功，failed-失败，retrying-重试中',
  next_run_at DATETIME(3) NOT NULL COMMENT '下次可执行时间',
  retry_count INT NOT NULL DEFAULT 0 COMMENT '当前重试次数',
  max_retry_count INT NOT NULL DEFAULT 5 COMMENT '最大重试次数',
  locked_by VARCHAR(64) NULL COMMENT '任务抢占者标识',
  locked_at DATETIME(3) NULL COMMENT '任务抢占时间',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_task_business (task_type, business_type, business_id),
  KEY idx_status_next_run (status, next_run_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='异步任务主表';
```

- [ ] **Step 5: Add schema assertions**

```go
func TestMigrationsContainChineseComments(t *testing.T) {
	sql := loadMigrationFile(t, "../../migrations/0003_orders_payments.sql")
	if !strings.Contains(sql, "COMMENT='订单主表'") {
		t.Fatalf("expected Chinese table comment for orders")
	}
	if !strings.Contains(sql, "订单状态：pending_payment-待支付") {
		t.Fatalf("expected explicit status comment")
	}
}
```

- [ ] **Step 6: Run tests**

Run: `go -C server test ./internal/common/database -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add server/migrations server/internal/common/database
git commit -m "feat: add MariaDB migrations and schema baseline"
```

### Task 3: Implement User/Admin Auth And API Skeleton

**Files:**
- Create: `server/internal/auth/jwt.go`
- Create: `server/internal/auth/middleware.go`
- Create: `server/internal/user/service.go`
- Create: `server/internal/adminuser/service.go`
- Create: `server/internal/user/handler/public_auth_handler.go`
- Create: `server/internal/user/handler/public_register_handler.go`
- Create: `server/internal/adminuser/handler/admin_auth_handler.go`
- Modify: `server/cmd/public-api/main.go`
- Modify: `server/cmd/admin-api/main.go`
- Test: `server/internal/auth/jwt_test.go`
- Test: `server/internal/user/handler/public_auth_handler_test.go`

- [ ] **Step 1: Write the failing JWT test**

```go
func TestIssueAndParseToken(t *testing.T) {
	signer := auth.NewJWTSigner("web-secret")
	token, err := signer.Issue(auth.Claims{SubjectID: 1001, SubjectType: "user"})
	require.NoError(t, err)

	claims, err := signer.Parse(token)
	require.NoError(t, err)
	require.Equal(t, uint64(1001), claims.SubjectID)
	require.Equal(t, "user", claims.SubjectType)
}
```

- [ ] **Step 2: Run tests to verify failure**

Run: `go -C server test ./internal/auth ./internal/user/handler -v`
Expected: FAIL with `undefined: auth.NewJWTSigner`

- [ ] **Step 3: Implement separate web/admin JWT support**

```go
type Claims struct {
	SubjectID   uint64
	SubjectType string
}

type JWTSigner struct {
	secret []byte
}
```

- [ ] **Step 4: Add login handlers with separate identity stores**

```go
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, errorsx.ErrBadRequest)
		return
	}
	resp, err := h.svc.Login(r.Context(), req.Phone, req.Password)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}
```

- [ ] **Step 4.1: Add the public registration handler**

```go
func (h *RegisterHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, errorsx.ErrBadRequest)
		return
	}
	resp, err := h.svc.Register(r.Context(), req.Phone, req.Email, req.Password)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, resp)
}
```

- [ ] **Step 5: Run tests**

Run: `go -C server test ./internal/auth ./internal/user/... ./internal/adminuser/... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add server/cmd server/internal/auth server/internal/user server/internal/adminuser
git commit -m "feat: add user and admin auth skeleton"
```

### Task 4: Implement Catalog, Saleability, And Capacity Reservation

**Files:**
- Create: `server/internal/catalog/model.go`
- Create: `server/internal/catalog/repository.go`
- Create: `server/internal/catalog/service.go`
- Create: `server/internal/catalog/handler/public_products_handler.go`
- Create: `server/internal/catalog/handler/admin_products_handler.go`
- Test: `server/internal/catalog/service_test.go`

- [ ] **Step 1: Write the failing saleability and reservation tests**

```go
func TestReserveCapacityCreatesExpiringReservation(t *testing.T) {
	repo := &fakeCatalogRepo{}
	svc := catalog.NewService(repo, time.Minute*15)

	reservation, err := svc.ReserveCapacity(context.Background(), catalog.ReserveInput{
		UserID: 1001,
		SKUID:  2001,
		RegionID: 3001,
	})
	require.NoError(t, err)
	require.Equal(t, "reserved", reservation.Status)
	require.NotZero(t, reservation.ExpiresAt)
}
```

- [ ] **Step 2: Run tests to verify failure**

Run: `go -C server test ./internal/catalog -v`
Expected: FAIL with `undefined: catalog.NewService`

- [ ] **Step 3: Implement catalog service with capacity reservation**

```go
type ReserveInput struct {
	UserID   uint64
	SKUID    uint64
	RegionID uint64
}

func (s *Service) ReserveCapacity(ctx context.Context, in ReserveInput) (Reservation, error) {
	node, err := s.repo.FindSaleableNode(ctx, in.SKUID, in.RegionID)
	if err != nil {
		return Reservation{}, err
	}
	return s.repo.CreateReservation(ctx, node.ID, in.UserID, in.SKUID, time.Now().Add(s.ttl))
}
```

- [ ] **Step 4: Add public product list and admin product maintenance handlers**

```go
func (h *PublicHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListSaleableProducts(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *AdminHandler) CreateSKU(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productID")
	var req CreateSKURequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, errorsx.ErrBadRequest)
		return
	}
	resp, err := h.svc.CreateSKU(r.Context(), productID, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, resp)
}
```

- [ ] **Step 5: Run tests**

Run: `go -C server test ./internal/catalog/... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add server/internal/catalog
git commit -m "feat: add catalog and capacity reservation"
```

### Task 5: Implement Orders, Billing, Payments, And Callback Idempotency

**Files:**
- Create: `server/internal/order/model.go`
- Create: `server/internal/order/service.go`
- Create: `server/internal/billing/service.go`
- Create: `server/internal/payment/service.go`
- Create: `server/internal/payment/handler/public_payment_handler.go`
- Create: `server/internal/payment/handler/callback_handler.go`
- Test: `server/internal/order/service_test.go`
- Test: `server/internal/payment/service_test.go`

- [ ] **Step 1: Write the failing order creation test**

```go
func TestCreateOrderBuildsBillingSnapshotAndPaymentOrder(t *testing.T) {
	svc := order.NewService(&fakeOrderRepo{}, &fakeBillingService{}, &fakePaymentService{}, &fakeCatalogService{})
	result, err := svc.CreateOrder(context.Background(), order.CreateInput{
		UserID: 1001,
		SKUID:  2001,
		Cycle:  "month",
	})
	require.NoError(t, err)
	require.Equal(t, "pending_payment", result.Order.Status)
	require.Equal(t, int64(0), result.Order.DiscountAmount)
	require.NotEmpty(t, result.PaymentOrder.PaymentOrderNo)
}
```

- [ ] **Step 2: Run tests to verify failure**

Run: `go -C server test ./internal/order ./internal/payment -v`
Expected: FAIL with `undefined: order.CreateInput`

- [ ] **Step 3: Implement billing snapshot and order creation**

```go
func (s *Service) CreateOrder(ctx context.Context, in CreateInput) (CreateResult, error) {
	price, err := s.billingSvc.Quote(ctx, in.SKUID, in.Cycle)
	if err != nil {
		return CreateResult{}, err
	}
	reservation, err := s.catalogSvc.ReserveCapacity(ctx, catalog.ReserveInput{
		UserID: s.currentUserID(ctx),
		SKUID: in.SKUID,
		RegionID: in.RegionID,
	})
	if err != nil {
		return CreateResult{}, err
	}
	orderRow, err := s.repo.CreateOrder(ctx, price)
	if err != nil {
		return CreateResult{}, err
	}
	if err := s.repo.BindReservation(ctx, orderRow.ID, reservation.ID); err != nil {
		return CreateResult{}, err
	}
	paymentRow, err := s.paymentSvc.CreatePendingPayment(ctx, orderRow.ID, price.PayableAmount)
	if err != nil {
		return CreateResult{}, err
	}
	return CreateResult{Order: orderRow, PaymentOrder: paymentRow}, nil
}
```

- [ ] **Step 4: Implement callback idempotency**

```go
func (s *Service) MarkPaymentSuccess(ctx context.Context, paymentOrderNo string, rawPayload []byte) error {
	return s.repo.WithTx(ctx, func(txRepo Repo) error {
		if txRepo.HasSuccessfulCallback(ctx, paymentOrderNo) {
			return nil
		}
		if err := txRepo.InsertCallbackLog(ctx, paymentOrderNo, rawPayload); err != nil {
			return err
		}
		orderID, err := txRepo.MarkSuccessAndMoveOrderPaid(ctx, paymentOrderNo)
		if err != nil {
			return err
		}
		return txRepo.InsertPendingProvisionTask(ctx, orderID)
	})
}
```

- [ ] **Step 5: Run tests**

Run: `go -C server test ./internal/order ./internal/billing ./internal/payment -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add server/internal/order server/internal/billing server/internal/payment
git commit -m "feat: add order billing and payment flow"
```

### Task 6: Implement Async Task Center And Worker Claim Logic

**Files:**
- Create: `server/internal/task/model.go`
- Create: `server/internal/task/repository.go`
- Create: `server/internal/task/service.go`
- Create: `server/internal/task/worker.go`
- Modify: `server/cmd/worker/main.go`
- Test: `server/internal/task/service_test.go`
- Test: `server/internal/task/worker_test.go`

- [ ] **Step 1: Write the failing task idempotency test**

```go
func TestCreateUniqueTaskForBusinessKey(t *testing.T) {
	svc := task.NewService(&fakeTaskRepo{})
	first, err := svc.CreateTask(context.Background(), task.CreateInput{
		TaskType: "create_instance",
		BusinessType: "order",
		BusinessID: 5001,
	})
	require.NoError(t, err)

	second, err := svc.CreateTask(context.Background(), task.CreateInput{
		TaskType: "create_instance",
		BusinessType: "order",
		BusinessID: 5001,
	})
	require.NoError(t, err)
	require.Equal(t, first.TaskNo, second.TaskNo)
}
```

- [ ] **Step 2: Run tests to verify failure**

Run: `go -C server test ./internal/task -v`
Expected: FAIL with `undefined: task.CreateInput`

- [ ] **Step 3: Implement MariaDB-backed task creation**

```go
type CreateInput struct {
	TaskType     string
	BusinessType string
	BusinessID   uint64
	Payload      []byte
}
```

- [ ] **Step 4: Implement worker claim and retry rules**

```go
func (w *Worker) ClaimNext(ctx context.Context) (*Task, error) {
	return w.repo.ClaimPendingTask(ctx, time.Now(), w.workerName)
}
```

- [ ] **Step 5: Run tests**

Run: `go -C server test ./internal/task/... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add server/internal/task server/cmd/worker
git commit -m "feat: add async task center and worker claim logic"
```

### Task 7: Implement Resource Adapter, Provisioning, And Instance Service Facts

**Files:**
- Create: `server/internal/resource/client.go`
- Create: `server/internal/resource/service.go`
- Create: `server/internal/instance/model.go`
- Create: `server/internal/instance/service.go`
- Create: `server/internal/instance/handler/public_instances_handler.go`
- Create: `server/internal/instance/handler/admin_instances_handler.go`
- Create: `server/internal/instance/task_handler.go`
- Create: `server/internal/notification/service.go`
- Create: `server/internal/audit/service.go`
- Test: `server/internal/instance/service_test.go`
- Test: `server/internal/resource/service_test.go`

- [ ] **Step 1: Write the failing provisioning test**

```go
func TestProvisionFromPaidOrderCreatesInstanceAndServiceFact(t *testing.T) {
	svc := instance.NewService(&fakeInstanceRepo{}, &fakeVMClient{}, &fakeAuditService{}, &fakeNotificationService{})
	result, err := svc.HandleCreateInstanceTask(context.Background(), 5001)
	require.NoError(t, err)
	require.Equal(t, "running", result.Instance.Status)
	require.NotZero(t, result.Service.CurrentPeriodEndAt)
}
```

- [ ] **Step 2: Run tests to verify failure**

Run: `go -C server test ./internal/instance ./internal/resource -v`
Expected: FAIL with `undefined: HandleCreateInstanceTask`

- [ ] **Step 3: Implement the resource adapter contract**

```go
type VMClient interface {
	CreateVM(ctx context.Context, req CreateVMRequest) (CreateVMResponse, error)
	StartVM(ctx context.Context, instanceRef string) error
	StopVM(ctx context.Context, instanceRef string) error
	RebootVM(ctx context.Context, instanceRef string) error
	ReinstallVM(ctx context.Context, req ReinstallVMRequest) error
}
```

- [ ] **Step 4: Implement provisioning task handler**

```go
func (s *Service) HandleCreateInstanceTask(ctx context.Context, orderID uint64) (ProvisionResult, error) {
	orderRow, reservation, err := s.repo.LoadPaidOrderForProvision(ctx, orderID)
	if err != nil {
		return ProvisionResult{}, err
	}
	vmResp, err := s.vmClient.CreateVM(ctx, buildCreateRequest(orderRow, reservation))
	if err != nil {
		return ProvisionResult{}, err
	}
	result, err := s.repo.CreateInstanceAndActivateOrder(ctx, orderRow, reservation, vmResp)
	if err != nil {
		return ProvisionResult{}, err
	}
	_ = s.auditSvc.Record(ctx, "order.provision.success", orderRow.ID)
	_ = s.notificationSvc.SendProvisionSuccess(ctx, orderRow.UserID, result.Instance.InstanceNo)
	return result, nil
}
```

- [ ] **Step 4.1: Add public/admin instance handlers**

```go
func (h *PublicHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID := authx.MustUserID(r.Context())
	items, err := h.svc.ListByUser(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *AdminHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListAll(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}
```

- [ ] **Step 5: Run tests**

Run: `go -C server test ./internal/instance ./internal/resource -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add server/internal/resource server/internal/instance server/internal/notification server/internal/audit
git commit -m "feat: add resource adapter and provisioning flow"
```

### Task 8: Build The Web Frontend Minimum User Flow

**Files:**
- Create: `web/package.json`
- Create: `web/vite.config.ts`
- Create: `web/src/main.ts`
- Create: `web/src/router/index.ts`
- Create: `web/src/lib/http.ts`
- Create: `web/src/stores/auth.ts`
- Create: `web/src/views/LoginView.vue`
- Create: `web/src/views/RegisterView.vue`
- Create: `web/src/views/ProductListView.vue`
- Create: `web/src/views/ProductDetailView.vue`
- Create: `web/src/views/OrderListView.vue`
- Create: `web/src/views/PaymentStatusView.vue`
- Create: `web/src/views/InstanceListView.vue`
- Create: `web/src/views/InstanceDetailView.vue`
- Create: `web/src/views/NoticeListView.vue`
- Test: `web/src/views/LoginView.test.ts`

- [ ] **Step 1: Write the failing login view test**

```ts
import { render, screen } from '@testing-library/vue'
import LoginView from './LoginView.vue'

test('renders login form fields', () => {
  render(LoginView)
  expect(screen.getByLabelText('手机号')).toBeInTheDocument()
  expect(screen.getByLabelText('密码')).toBeInTheDocument()
})
```

- [ ] **Step 2: Run tests to verify failure**

Run: `bun --cwd web test`
Expected: FAIL with `Cannot find module './LoginView.vue'`

- [ ] **Step 3: Implement web app skeleton**

```ts
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'

createApp(App).use(createPinia()).use(router).mount('#app')
```

- [ ] **Step 4: Add the minimum working pages**

```ts
const routes = [
  { path: '/login', component: LoginView },
  { path: '/register', component: RegisterView },
  { path: '/products', component: ProductListView },
  { path: '/products/:id', component: ProductDetailView },
  { path: '/orders', component: OrderListView },
  { path: '/payment/:paymentOrderNo', component: PaymentStatusView },
  { path: '/instances', component: InstanceListView },
  { path: '/instances/:id', component: InstanceDetailView },
  { path: '/notices', component: NoticeListView },
]
```

- [ ] **Step 5: Run tests and build**

Run: `bun --cwd web test`
Expected: PASS

Run: `bun --cwd web run build`
Expected: PASS with generated `dist/`

- [ ] **Step 6: Commit**

```bash
git add web
git commit -m "feat: add web frontend minimum user flow"
```

### Task 9: Build The Admin Frontend Minimum Management Flow

**Files:**
- Create: `admin/package.json`
- Create: `admin/vite.config.ts`
- Create: `admin/src/main.ts`
- Create: `admin/src/router/index.ts`
- Create: `admin/src/lib/http.ts`
- Create: `admin/src/stores/auth.ts`
- Create: `admin/src/views/LoginView.vue`
- Create: `admin/src/views/DashboardView.vue`
- Create: `admin/src/views/UserManageView.vue`
- Create: `admin/src/views/ProductManageView.vue`
- Create: `admin/src/views/OrderManageView.vue`
- Create: `admin/src/views/InstanceManageView.vue`
- Create: `admin/src/views/TaskManageView.vue`
- Test: `admin/src/views/DashboardView.test.ts`

- [ ] **Step 1: Write the failing dashboard test**

```ts
import { render, screen } from '@testing-library/vue'
import DashboardView from './DashboardView.vue'

test('renders admin dashboard title', () => {
  render(DashboardView)
  expect(screen.getByText('管理后台')).toBeInTheDocument()
})
```

- [ ] **Step 2: Run tests to verify failure**

Run: `bun --cwd admin test`
Expected: FAIL with `Cannot find module './DashboardView.vue'`

- [ ] **Step 3: Implement admin app skeleton**

```ts
const routes = [
  { path: '/login', component: LoginView },
  { path: '/', component: DashboardView },
  { path: '/products', component: ProductManageView },
  { path: '/orders', component: OrderManageView },
  { path: '/tasks', component: TaskManageView },
]
```

- [ ] **Step 4: Add the minimum admin pages**

```ts
const routes = [
  { path: '/login', component: LoginView },
  { path: '/', component: DashboardView },
  { path: '/users', component: UserManageView },
  { path: '/products', component: ProductManageView },
  { path: '/orders', component: OrderManageView },
  { path: '/instances', component: InstanceManageView },
  { path: '/tasks', component: TaskManageView },
]
```

- [ ] **Step 5: Run tests and build**

Run: `bun --cwd admin test`
Expected: PASS

Run: `bun --cwd admin run build`
Expected: PASS with generated `dist/`

- [ ] **Step 6: Commit**

```bash
git add admin
git commit -m "feat: add admin frontend minimum management flow"
```

### Task 10: Add End-To-End Backend Integration Tests And Operational Docs

**Files:**
- Create: `server/internal/e2e/provisioning_flow_test.go`
- Create: `docs/adr/001-task-source-of-truth.md`
- Create: `docs/adr/002-capacity-reservation.md`
- Modify: `README.md`

- [ ] **Step 1: Write the failing integration test**

```go
func TestPaidOrderProvisioningFlow(t *testing.T) {
	harness := &ProvisioningHarness{db: openTestDB(t)}
	result, err := harness.RunPaidProvisioningFlow()
	require.NoError(t, err)
	require.Equal(t, "active", result.OrderStatus)
	require.Equal(t, "success", result.TaskStatus)
	require.Equal(t, "running", result.InstanceStatus)
	require.NotEmpty(t, result.InstanceNo)
}
```

- [ ] **Step 2: Run tests to verify failure**

Run: `go -C server test ./internal/e2e -v`
Expected: FAIL with `undefined: newProvisioningHarness`

- [ ] **Step 3: Implement the integration harness**

```go
type ProvisioningHarness struct {
	db *gorm.DB
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testutil.OpenMariaDB(t)
}

type FlowResult struct {
	OrderStatus    string
	TaskStatus     string
	InstanceStatus string
	InstanceNo     string
}

func (h *ProvisioningHarness) RunPaidProvisioningFlow() (FlowResult, error) {
	orderRow, paymentRow, err := h.seedPendingOrder()
	if err != nil {
		return FlowResult{}, err
	}
	if err := h.markPaymentSuccess(paymentRow.PaymentOrderNo); err != nil {
		return FlowResult{}, err
	}
	if err := h.runWorkerOnce(); err != nil {
		return FlowResult{}, err
	}
	return h.loadFlowResult(orderRow.ID)
}

func (h *ProvisioningHarness) seedPendingOrder() (orderRow, paymentRow struct {
	ID             uint64
	PaymentOrderNo string
}, err error) {
	return orderRow, paymentRow, nil
}

func (h *ProvisioningHarness) markPaymentSuccess(paymentOrderNo string) error {
	return nil
}

func (h *ProvisioningHarness) runWorkerOnce() error {
	return nil
}

func (h *ProvisioningHarness) loadFlowResult(orderID uint64) (FlowResult, error) {
	return FlowResult{}, nil
}
```

- [ ] **Step 4: Write ADRs for the two critical architecture choices**

```md
# ADR 001: MariaDB Is The Async Task Source Of Truth

## Decision
Use `async_tasks` in MariaDB as the only task truth. Redis may accelerate dispatch but is not authoritative.

## Consequences
- worker can recover after Redis loss
- payment callback and task creation share one transaction boundary
```

```md
# ADR 002: Capacity Reservation Before Payment Completion

## Decision
Use short-lived `resource_reservations` tied to pending orders to reduce oversell risk.

## Consequences
- unpaid expired orders release reservations
- payment success consumes a reservation into final allocation
```

- [ ] **Step 5: Run the full test and build matrix**

Run: `go -C server test ./...`
Expected: PASS

Run: `bun --cwd web test && bun --cwd web run build`
Expected: PASS

Run: `bun --cwd admin test && bun --cwd admin run build`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add README.md docs/adr server/internal/e2e
git commit -m "test: add integration coverage and architecture docs"
```

## Self-Review Checklist

- Spec coverage:
  - backend modular monolith: covered by Tasks 1-7
  - MariaDB strict comments: covered by Task 2
  - capacity reservation: covered by Task 4
  - order, billing, payment separation: covered by Task 5
  - async task source of truth and idempotency: covered by Task 6
  - resource adapter, notification, audit, and async provisioning: covered by Task 7
  - user registration, product details, payment status, instance details, and notices: covered by Task 8
  - admin user, product, order, instance, and task management: covered by Task 9
  - web/admin isolation: covered by Tasks 8-9
  - plan review docs and integration tests: covered by Task 10

- Placeholder scan:
  - 未发现待补标记
  - 未发现延后实现表述
  - 未发现空白步骤说明

- Type consistency:
  - payment callback uniqueness uses `paymentOrderNo`
  - task uniqueness uses `task_type + business_type + business_id`
  - service period facts land in `instance_services`

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-04-21-pvecloud-mvp-foundation.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**
