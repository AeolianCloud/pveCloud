# PVE Cloud Backend Realization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在当前 skeleton 后端基础上，按 `docs/superpowers/specs/2026-04-22-pvecloud-backend-realization-design.md` 的范围完成“真实后端闭环”，让 MariaDB 成为业务事实源，支付回调、任务执行、实例开通和 public/admin API 形成可验证闭环。

**Architecture:** 继续保持 `public-api`、`admin-api`、`worker` 三入口的 Go 模块化单体。优先补齐 YAML 配置分层、`database/sql` 事务基础设施、各领域真实 repository 和事务边界，再把支付回调、worker、mock resource provider、API 装配和测试矩阵串成真实闭环。

**Tech Stack:** Go, `database/sql`, MariaDB, Redis, `net/http` `ServeMux`, YAML config, Go test

---

## 当前设计与仓库现状冲突

- 当前配置仍是扁平结构：`[server/config/config.yaml](/abs/path/D:/UGit/pveCloud/server/config/config.yaml)` 与 `[server/internal/bootstrap/config/config.go](/abs/path/D:/UGit/pveCloud/server/internal/bootstrap/config/config.go)` 只有 `mysql_dsn`、`redis_addr`、`jwt_*` 等扁平字段，尚未实现设计要求的 `payment/resource/worker` 嵌套配置。
- 当前数据库基础设施只有 `Open`：`[server/internal/common/database/mysql.go](/abs/path/D:/UGit/pveCloud/server/internal/common/database/mysql.go)` 没有 `WithTx`、`BeginTx` 封装，也没有统一事务注入方式，和设计要求的强事务边界不一致。
- 领域服务大多仍是接口驱动 skeleton：`[server/internal/order/service.go](/abs/path/D:/UGit/pveCloud/server/internal/order/service.go)`、`[server/internal/payment/service.go](/abs/path/D:/UGit/pveCloud/server/internal/payment/service.go)`、`[server/internal/task/service.go](/abs/path/D:/UGit/pveCloud/server/internal/task/service.go)`、`[server/internal/instance/service.go](/abs/path/D:/UGit/pveCloud/server/internal/instance/service.go)` 已有服务壳，但缺少 MariaDB-backed repository 落地。
- `bootstrap` 只注册了最少路由：`[server/internal/bootstrap/app.go](/abs/path/D:/UGit/pveCloud/server/internal/bootstrap/app.go)` 目前只有 `/healthz`、用户登录注册、管理员登录，未覆盖设计文档要求的 public/admin 闭环 API。
- 支付回调链路仍是简化版：`[server/internal/payment/handler/callback_handler.go](/abs/path/D:/UGit/pveCloud/server/internal/payment/handler/callback_handler.go)` 直接从 query 读取 `payment_order_no`，没有 provider verifier、配置驱动验签和回调日志上下文。
- worker 还不是可运行闭环：`[server/internal/task/worker.go](/abs/path/D:/UGit/pveCloud/server/internal/task/worker.go)` 只有 claim 包装，没有轮询、执行、重试、日志追加和实例开通编排。
- e2e 仍是硬编码假实现：`[server/internal/e2e/provisioning_flow.go](/abs/path/D:/UGit/pveCloud/server/internal/e2e/provisioning_flow.go)` 直接返回成功结果，不是 DB-backed flow。
- 仓库里没有此前计划中提到的 `server/internal/common/testutil/testdb.go`，说明测试基础设施需要按现状重建，不能直接沿用旧计划假设。

以上冲突不阻止规划，但决定了这一轮必须先补“配置 + 事务 + repository 基础设施”，不能直接从 handler 或 e2e 表层推进。

## 这一轮必须完成

- `payment/resource/worker` 嵌套配置模型与 YAML 读取校验。
- `internal/common/database` 事务 helper，以及各模块 repository 的 tx 注入约定。
- `catalog/billing/order/payment` 的真实 MariaDB repository 与三条关键事务边界中的前两条。
- 配置驱动的真实支付回调链路，含回调日志、幂等、任务创建。
- `task/worker` 的 claim/execute/retry/log 闭环。
- `instance/resource` 的 mock provider + provisioning 持久化闭环。
- 设计文档限定的 public/admin API 路由与鉴权装配。
- repository integration tests 与真实 DB-backed e2e。

## 后续真实外部接口接入时再做

- 第三方支付厂商的真实下单、验签、退款、账单对账。
- 真实 PVE 或其他 VM 平台 provider。
- 前端 `web/`、`admin/` 对真实 API 的接入改造。
- 高级实例操作，如开机、关机、重启、重装、续费、升降配。
- Redis 驱动的任务加速分发、缓存、限流等非主事实源优化。
- 非闭环必需的 admin 运营 API 与通知实际发送通道。

## 实施任务

### Task 1: 补齐配置结构与事务基础设施

**目标：** 先把配置模型、应用装配入口、数据库事务 helper 调整到设计文档要求的形态，为后续 repository 和 worker 实现提供稳定基础。

**涉及文件：**
- Modify: `server/config/config.yaml`
- Modify: `server/internal/bootstrap/config/config.go`
- Modify: `server/internal/bootstrap/app.go`
- Modify: `server/internal/common/database/mysql.go`
- Create: `server/internal/common/database/tx.go`
- Add or Modify Test: `server/internal/bootstrap/config/config_test.go`
- Add or Modify Test: `server/internal/common/database/tx_test.go`

**实现步骤：**
- [x] 在 `config.go` 中新增 `PaymentConfig`、`ResourceConfig`、`WorkerConfig`，保留现有字段兼容入口，但以嵌套字段作为后续业务依赖。
- [x] 更新 `config.yaml`，补齐 `payment.provider`、`payment.callback_base_url`、`payment.notify_path`、`payment.merchant_id`、`payment.merchant_secret`、`resource.provider`、`resource.api_endpoint`、`resource.api_token`、`worker.poll_interval`、`worker.batch_size`。
- [x] 在 `internal/common/database` 中新增 `WithTx(ctx, db, fn)`，约定 repository 在需要事务时接收 `*sql.Tx` 或抽象出的 querier。
- [x] 在 `bootstrap/app.go` 中把共享依赖装配拆清楚，形成“配置 -> DB/Redis -> repository -> service -> handler”的固定装配顺序，为后续 public/admin/worker 共用。
- [x] 为配置加载和事务 helper 增加测试，覆盖必填项校验、事务提交、事务回滚。

**验证命令：**
- `go -C server test ./internal/bootstrap/config ./internal/common/database -v`
- `go -C server build ./cmd/public-api ./cmd/admin-api ./cmd/worker`

**预期结果：**
- 配置文件可直接表达支付、资源、worker 策略。
- 事务 helper 可被各模块统一复用。
- 三个入口仍可编译通过。

**提交边界：**
- 单独提交“配置结构 + 通用事务基础设施”，不混入业务 repository 或 handler 改动。
- 建议提交信息：`feat: add nested backend config and tx helpers`

### Task 2: 落地 catalog/billing/order/payment 真实 repository 与下单事务

**目标：** 用 MariaDB-backed repository 替换当前 skeleton 接口实现，先打通“创建订单事务”，确保 `orders`、`billing_records`、`payment_orders`、`resource_reservations` 真正入库。

**涉及文件：**
- Modify: `server/internal/catalog/repository.go`
- Create: `server/internal/catalog/mysql_repository.go`
- Modify: `server/internal/billing/service.go`
- Create: `server/internal/billing/repository.go`
- Create: `server/internal/billing/mysql_repository.go`
- Modify: `server/internal/order/service.go`
- Create: `server/internal/order/repository.go`
- Create: `server/internal/order/mysql_repository.go`
- Modify: `server/internal/payment/service.go`
- Create: `server/internal/payment/repository.go`
- Create: `server/internal/payment/mysql_repository.go`
- Add or Modify Test: `server/internal/catalog/repository_integration_test.go`
- Add or Modify Test: `server/internal/order/repository_integration_test.go`
- Add or Modify Test: `server/internal/payment/repository_integration_test.go`

**实现步骤：**
- [ ] 为 `catalog` 增加真实 SQL repository，覆盖 saleable product 查询、saleable node 选择、reservation 创建与按需锁定查询。
- [ ] 为 `billing` 增加快照落库 repository，并明确 `billing` 只负责金额与周期快照，不承担支付或订单状态流转。
- [ ] 重构 `order.Service.CreateOrder`，把“预占资源 + 创建订单 + 创建 billing record + 创建 payment order + 绑定 reservation”纳入一个明确事务边界。
- [ ] 为 `payment` 增加真实 pending payment 创建查询能力，替换当前只返回内存对象的假实现。
- [ ] 为上述 repository 写 integration tests，验证唯一键、字段回写、事务回滚和 reservation 绑定正确性。

**验证命令：**
- `go -C server test ./internal/catalog ./internal/billing ./internal/order ./internal/payment -v`
- `go -C server test ./internal/catalog -run Repository ./internal/order -run Repository ./internal/payment -run Repository -v`

**预期结果：**
- 创建订单后，MariaDB 中真实生成订单、账单快照、支付单、预占关系。
- 下单失败时不会留下半成品业务记录。

**提交边界：**
- 单独提交“catalog/billing/order/payment repository + order create tx”，不引入 callback/worker 逻辑。
- 建议提交信息：`feat: implement order creation repositories and transaction`

### Task 3: 实现真实支付回调链路与支付成功事务

**目标：** 把支付回调从简化 handler 升级为配置驱动、可幂等、可追踪、可入任务中心的真实业务链路。

**涉及文件：**
- Modify: `server/internal/payment/service.go`
- Modify: `server/internal/payment/handler/callback_handler.go`
- Modify: `server/internal/payment/handler/public_payment_handler.go`
- Create: `server/internal/payment/provider.go`
- Create: `server/internal/payment/mock_provider.go`
- Modify: `server/internal/payment/repository.go`
- Modify: `server/internal/payment/mysql_repository.go`
- Modify: `server/internal/order/mysql_repository.go`
- Create: `server/internal/audit/repository.go`
- Create: `server/internal/audit/mysql_repository.go`
- Add or Modify Test: `server/internal/payment/service_test.go`
- Add or Modify Test: `server/internal/payment/repository_integration_test.go`

**实现步骤：**
- [ ] 在 `payment` 模块中拆出“业务服务”和“provider verifier”，让回调处理从 YAML 读取 provider/secret，而不是硬编码或直接信任 query 参数。
- [ ] 改造 callback handler，请求体解析后调用 provider verifier，再把标准化后的结果交给 payment service。
- [ ] 在支付成功事务中按设计要求完成：写 callback log、`payment_orders.pending -> success`、`orders.pending_payment -> paid`、创建唯一 `create_instance` 任务。
- [ ] 约束重复回调行为：同一 `payment_order_no` 重放只能新增可追踪日志，不得重复推进订单或重复建任务。
- [ ] 补充支付状态查询接口所需 repository 能力，给后续 public API 装配做准备。

**验证命令：**
- `go -C server test ./internal/payment ./internal/order ./internal/task -v`
- `go -C server test ./internal/payment -run Callback ./internal/payment -run Repository -v`

**预期结果：**
- 支付回调真正由配置驱动处理。
- 重复回调不会创建重复任务。
- 支付成功后数据库状态稳定推进到 `orders.paid`。

**提交边界：**
- 单独提交“支付 provider/verifier + callback tx + payment status query”，不包含 worker 执行。
- 建议提交信息：`feat: implement real payment callback transaction`

### Task 4: 落地 task repository、worker 执行循环与任务日志

**目标：** 从“只能 claim”推进到“能 claim、能执行、能记录日志、能重试”的真实任务中心。

**涉及文件：**
- Modify: `server/internal/task/repository.go`
- Create: `server/internal/task/mysql_repository.go`
- Modify: `server/internal/task/service.go`
- Modify: `server/internal/task/worker.go`
- Create: `server/internal/task/executor.go`
- Create: `server/internal/task/log_repository.go`
- Modify: `server/cmd/worker/main.go`
- Modify: `server/internal/bootstrap/app.go`
- Add or Modify Test: `server/internal/task/service_test.go`
- Add or Modify Test: `server/internal/task/worker_test.go`
- Add or Modify Test: `server/internal/task/repository_integration_test.go`

**实现步骤：**
- [ ] 为 `async_tasks`、`async_task_logs` 增加真实 repository，覆盖 create、claim、mark success、mark retry、mark failed、append log、list tasks。
- [ ] 在 worker 中实现基于 `worker.poll_interval`、`worker.batch_size` 的轮询执行模型，至少支持单次 loop 和持续 loop 两种入口，便于测试。
- [ ] 引入 task executor，把 `create_instance` 任务分派给 instance provisioning 服务，避免在 worker 中直接堆业务逻辑。
- [ ] 明确 retry 语义：可重试错误写 `retrying` 和 `next_run_at`，终态错误写 `failed`，每次执行都追加 task log。
- [ ] 为 claim 竞争、重复业务键、重试次数和日志追加写 integration tests。

**验证命令：**
- `go -C server test ./internal/task/... -v`
- `go -C server test ./internal/task -run Worker ./internal/task -run Repository -v`

**预期结果：**
- worker 可从 MariaDB 抢占任务并执行。
- 任务失败会留下日志和重试元数据。
- 管理端后续可查询到真实任务状态。

**提交边界：**
- 单独提交“task repository + worker loop + retry/log”，不混入 instance 持久化细节。
- 建议提交信息：`feat: implement task worker execution flow`

### Task 5: 实现 instance repository 与 mock resource provider 开通闭环

**目标：** 在不接真实外部云平台的前提下，用稳定 mock provider 完成“支付后创建实例、写服务周期事实、消费预占、推进订单 active”的第三条事务边界。

**涉及文件：**
- Modify: `server/internal/resource/client.go`
- Modify: `server/internal/resource/service.go`
- Create: `server/internal/resource/mock_client.go`
- Modify: `server/internal/instance/service.go`
- Create: `server/internal/instance/repository.go`
- Create: `server/internal/instance/mysql_repository.go`
- Modify: `server/internal/catalog/mysql_repository.go`
- Modify: `server/internal/order/mysql_repository.go`
- Modify: `server/internal/notification/service.go`
- Modify: `server/internal/audit/service.go`
- Add or Modify Test: `server/internal/instance/service_test.go`
- Add or Modify Test: `server/internal/resource/service_test.go`
- Add or Modify Test: `server/internal/instance/repository_integration_test.go`

**实现步骤：**
- [ ] 固化 `resource.VMClient` 最终契约，并提供 `mock` provider 实现，返回稳定、可预测的 VM 创建结果。
- [ ] 为 `instance` 增加真实 repository，覆盖 paid order 加锁读取、instance/instance_service 落库、reservation consume、order `paid -> provisioning -> active` 推进。
- [ ] 让 task executor 调用 `instance.Service.HandleCreateInstanceTask`，在外部 provider 调用成功后，通过事务统一落实例事实并更新任务状态。
- [ ] 对 provider 失败场景定义 retryable/terminal 边界，并把失败原因写入 task log 与 audit/notification 入口。
- [ ] 为实例创建事务和幂等保护写 integration tests，确保 worker retry 不会创建重复实例。

**验证命令：**
- `go -C server test ./internal/resource ./internal/instance -v`
- `go -C server test ./internal/instance -run Repository ./internal/instance -run Service -v`

**预期结果：**
- 支付成功后的任务能真正生成实例记录与服务周期记录。
- 订单最终可推进到 `active`。
- retry 不会重复创建实例。

**提交边界：**
- 单独提交“instance repository + mock provider + provision tx”，不混入 API surface 扩充。
- 建议提交信息：`feat: implement provisioning closure with mock provider`

### Task 6: 补齐 public/admin API 闭环装配

**目标：** 仅实现设计文档要求的闭环 API，不做前端真实接入，但让 public/admin 两套门面都能查询真实后端数据。

**涉及文件：**
- Modify: `server/internal/bootstrap/app.go`
- Modify: `server/internal/auth/middleware.go`
- Modify: `server/internal/catalog/handler/public_products_handler.go`
- Modify: `server/internal/catalog/handler/admin_products_handler.go`
- Create or Modify: `server/internal/order/handler/public_orders_handler.go`
- Create or Modify: `server/internal/order/handler/admin_orders_handler.go`
- Modify: `server/internal/payment/handler/public_payment_handler.go`
- Modify: `server/internal/payment/handler/callback_handler.go`
- Modify: `server/internal/instance/handler/public_instances_handler.go`
- Create or Modify: `server/internal/instance/handler/public_instance_detail_handler.go`
- Modify: `server/internal/instance/handler/admin_instances_handler.go`
- Create or Modify: `server/internal/task/handler/admin_tasks_handler.go`
- Add or Modify Test: `server/internal/order/handler/public_orders_handler_test.go`
- Add or Modify Test: `server/internal/payment/handler/callback_handler_test.go`
- Add or Modify Test: `server/internal/instance/handler/public_instances_handler_test.go`

**实现步骤：**
- [ ] 在 `bootstrap/app.go` 中按模块清晰注册 public/admin 路由，并保持 `http.ServeMux`，不引入新 router。
- [ ] public API 实现并保护：`POST /auth/register`、`POST /auth/login`、`GET /products`、`GET /products/:id`、`POST /orders`、`GET /orders`、`GET /payments/:paymentOrderNo`、`POST /payments/callback`、`GET /instances`、`GET /instances/:id`。
- [ ] admin API 实现并保护：`POST /auth/login`、`GET /products`、`GET /orders`、`GET /instances`、`GET /tasks`。
- [ ] 使用现有 JWT middleware 区分 user/admin 身份，确保 public/admin 路由不共 URL、不共鉴权链。
- [ ] 对输入错误、权限错误、状态冲突、内部错误统一返回结构化应用错误。

**验证命令：**
- `go -C server test ./internal/auth ./internal/catalog/... ./internal/order/... ./internal/payment/... ./internal/instance/... ./internal/task/... -v`
- `go -C server build ./cmd/public-api ./cmd/admin-api`

**预期结果：**
- public/admin API 均可查询或触发真实后端链路。
- 仍然保持“门面只做接入，不承载核心业务规则”的边界。

**提交边界：**
- 单独提交“public/admin API 闭环路由与 handler 装配”，不掺杂 e2e 测试重构。
- 建议提交信息：`feat: expose backend closure through public and admin apis`

### Task 7: 建立 repository integration tests 基础设施

**目标：** 为真实 DB-backed repository 测试补齐当前仓库缺失的测试基础设施，覆盖设计文档要求的高风险事务与锁语义。

**涉及文件：**
- Create: `server/internal/testutil/mariadb.go`
- Create: `server/internal/testutil/migrations.go`
- Modify: `server/internal/catalog/repository_integration_test.go`
- Modify: `server/internal/order/repository_integration_test.go`
- Modify: `server/internal/payment/repository_integration_test.go`
- Modify: `server/internal/task/repository_integration_test.go`
- Modify: `server/internal/instance/repository_integration_test.go`
- Optionally Modify: `docker-compose.yml`
- Optionally Modify: `README.md`

**实现步骤：**
- [ ] 在现有目录结构下补建 `server/internal/testutil`，不要继续引用不存在的 `server/internal/common/testutil/testdb.go`。
- [ ] 实现测试 DB 打开、迁移执行、清表和测试数据构造 helper，供 repository integration tests 复用。
- [ ] 覆盖设计文档要求的四个重点：`CreateOrderTx`、`MarkPaymentSuccessTx`、`ClaimPendingTask`、`CreateInstanceTx`。
- [ ] 在测试中显式验证唯一键、状态流转、重复回调幂等、worker claim 锁和实例幂等。
- [ ] 如果需要，补充 README 或 compose 说明，明确本地如何准备 MariaDB 后跑 integration tests。

**验证命令：**
- `go -C server test ./internal/catalog ./internal/order ./internal/payment ./internal/task ./internal/instance -v`
- `go -C server test ./internal/... -run Repository -v`

**预期结果：**
- repository 层关键事务和锁行为有可重复执行的集成测试保障。
- 测试基础设施与当前仓库结构一致，不再依赖缺失文件。

**提交边界：**
- 单独提交“repository integration test harness + high-risk repo tests”。
- 建议提交信息：`test: add repository integration test harness`

### Task 8: 用真实 DB-backed 流程替换 e2e 假实现

**目标：** 把 `internal/e2e` 从硬编码成功返回替换成真实闭环验证，证明从下单到实例可查的主链路已经可运行。

**涉及文件：**
- Modify: `server/internal/e2e/provisioning_flow.go`
- Modify: `server/internal/e2e/provisioning_flow_test.go`
- Modify: `server/internal/bootstrap/app.go`
- Modify: `README.md`

**实现步骤：**
- [ ] 重写 e2e harness：准备测试库、种子用户/商品/节点、调用真实 service 或 HTTP handler 完成“查商品 -> 创建订单 -> 模拟支付回调 -> 跑一轮 worker -> 查实例”。
- [ ] 明确只验证本轮闭环所需路径，不扩展到真实外部支付或真实 VM provider。
- [ ] 在 e2e 中校验最终事实：`orders.active`、`async_tasks.success`、`instances.running`、实例对用户可见、重复 payment callback 不会重复开通。
- [ ] 确保 harness 可在本地 MariaDB 条件下重复运行，必要时补充测试数据清理。
- [ ] 更新 README 的 backend test matrix，写明 unit / repository integration / e2e 的执行方式。

**验证命令：**
- `go -C server test ./internal/e2e -v`
- `go -C server test ./...`

**预期结果：**
- `internal/e2e` 不再是假数据返回，而是验证真实闭环。
- 全部后端测试通过时，可证明“后端真实闭环”目标达成。

**提交边界：**
- 单独提交“real e2e harness + backend test matrix docs”。
- 建议提交信息：`test: replace fake e2e with real backend closure flow`

## 推荐执行顺序与里程碑

- [x] Milestone 1: 完成 Task 1，冻结配置结构和事务 helper。
- [ ] Milestone 2: 完成 Task 2，创建订单事务真实入库。
- [ ] Milestone 3: 完成 Task 3，支付回调成功事务真实落地。
- [ ] Milestone 4: 完成 Task 4，worker 可 claim/execute/retry。
- [ ] Milestone 5: 完成 Task 5，实例开通与 mock provider 闭环达成。
- [ ] Milestone 6: 完成 Task 6，public/admin API 面向真实数据开放。
- [ ] Milestone 7: 完成 Task 7 与 Task 8，测试矩阵与 e2e 收口。

## 完成定义

- [ ] YAML 配置已支持 `payment/resource/worker` 嵌套结构。
- [ ] 订单创建、支付成功、开通成功三条事务边界全部落地。
- [ ] `catalog/billing/order/payment/task/instance` 均有真实 MariaDB repository。
- [ ] 支付回调能做到幂等、可追踪、可建唯一任务。
- [ ] worker 能 claim 任务、执行开通、记录日志并处理 retry。
- [ ] mock resource provider 可替换且不泄漏业务逻辑。
- [ ] public/admin API 覆盖设计文档要求的闭环最小面。
- [ ] repository integration tests 与 e2e 均基于真实 DB 路径运行通过。

## 自检

- 设计文档顺序覆盖：
  - 配置结构和事务基础设施：Task 1
  - `catalog/billing/order/payment` 真实 repository 和事务：Task 2
  - 真实支付回调链路：Task 3
  - `task/worker`：Task 4
  - `instance/resource` mock provider：Task 5
  - public/admin API：Task 6
  - e2e 和 repository integration tests：Task 7-8
- “这一轮必须完成”和“后续再做”已显式区分。
- 所有任务都包含目标、涉及文件、实现步骤、验证命令、预期结果、提交边界。

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-04-22-pvecloud-backend-realization-implementation-plan.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**
