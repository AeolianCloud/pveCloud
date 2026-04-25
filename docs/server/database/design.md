# 数据库设计

## 设计目标

本设计面向 pveCloud 第一期后端闭环：用户注册登录、产品套餐、下单支付、余额流水、实例开通、工单、后台权限、异步任务和审计日志。

数据库使用 MariaDB 11.4.9 + InnoDB。业务事实以数据库为准，Redis 只作为缓存、会话或幂等辅助。

目标数据库名使用 `pvecloud`，与部署连接配置保持一致。

## 基础约定

- 主键统一使用 `BIGINT UNSIGNED AUTO_INCREMENT`。
- 金额统一使用整数分，字段后缀为 `_cents`，禁止使用浮点数。
- 状态字段使用 `VARCHAR` 保存，由 Go 常量控制，不使用数据库 enum，避免后续状态扩展困难。
- 所有业务表和字段都需要带 MariaDB `COMMENT` 注释，便于后台、迁移工具和数据库客户端直接查看结构含义。
- 时间字段统一使用 `DATETIME(3)`，保留毫秒。
- 常用表包含 `created_at`、`updated_at`，需要软删除的表包含 `deleted_at`。
- 对外展示使用业务编号，如 `order_no`、`payment_no`、`instance_no`，不要直接暴露自增 ID。
- JSON 字段只保存扩展快照、第三方回包、配置片段，不作为高频查询条件。

## 状态枚举

订单状态：

```text
pending        待支付
paid           已支付，等待按订单类型处理
provisioning   新购实例开通中
active         已完成，实例已交付
cancelled      已取消
expired        超时关闭
failed         处理失败，需要人工介入或补偿
refunded       已退款
```

支付状态：

```text
created        支付单已创建
pending        等待第三方支付结果
success        支付成功
failed         支付失败
closed         支付关闭
refunding      退款中
refunded       已退款
```

实例状态：

```text
creating       创建中
running        运行中
stopped        已关机
suspended      已暂停
expired        已到期
deleting       删除中
deleted        已删除
error          异常
```

异步任务状态：

```text
pending        待执行
running        执行中
success        执行成功
failed         执行失败，可重试或人工处理
cancelled      已取消
```

工单状态：

```text
open           已提交
pending_admin  等待客服/管理员处理
pending_user   等待用户回复
closed         已关闭
```

## 表分组

### 账号和权限

| 表名 | 说明 |
| --- | --- |
| `users` | 用户账号 |
| `admin_users` | 管理员账号 |
| `admin_roles` | 管理端角色 |
| `admin_permissions` | 管理端权限码 |
| `admin_user_roles` | 管理员和角色关联 |
| `admin_role_permissions` | 角色和权限关联 |

后台专属表统一使用 `admin_` 前缀，方便从表名区分管理端数据域。权限码以 `domain:action` 命名，例如 `order:view`、`payment:manual_credit`。管理员 JWT 可以缓存权限码，但最终权限来源以数据库角色权限为准。

### 产品目录

| 表名 | 说明 |
| --- | --- |
| `products` | 产品系列，例如云服务器 |
| `product_plans` | 套餐规格，CPU、内存、磁盘、带宽 |
| `regions` | 地域 |
| `pve_nodes` | PVE 节点配置 |
| `images` | 系统镜像逻辑定义 |
| `region_images` | 地域和节点可用镜像，以及对应 PVE 模板映射 |
| `plan_prices` | 套餐在地域和周期下的价格 |

价格查询以 `plan_id + region_id + billing_period` 为唯一维度。下单时把套餐、镜像、地域、价格快照写入订单扩展字段，避免后续产品改价影响历史订单。

镜像选择必须按地域过滤。`images` 只保存系统镜像的逻辑信息，真实 PVE 模板 ID 放在 `region_images.pve_template_id`。用户选择地域后，只展示该地域下 `region_images.status=active` 的镜像；worker 选定具体 `node_id` 后，再读取该节点对应的 `pve_template_id` 开通实例。

### 订单、支付和余额

| 表名 | 说明 |
| --- | --- |
| `orders` | 新购、续费订单 |
| `payment_orders` | 第三方支付、余额支付和人工入账支付单 |
| `payment_notify_logs` | 第三方支付回调原始记录 |
| `wallet_accounts` | 用户余额账户 |
| `wallet_transactions` | 钱包流水 |

第一期订单模型明确为一单一实例，不支持单个订单批量购买多个实例。因此 `orders` 不设计 `quantity` 字段，`instances.order_id` 保持唯一。后续如果要支持批量购买，应新增订单明细表或批量父订单，而不是复用当前一单一实例结构。

支付成功处理必须按 `payment_scene` 和 `order_type` 分支：

- `payment_scene=order` 且 `order_type=new`：订单进入 `paid`，创建唯一实例开通任务。
- `payment_scene=order` 且 `order_type=renew`：订单进入 `paid`，延长实例到期时间或创建续费同步任务。
- `payment_scene=topup`：只写钱包流水并增加余额，不创建实例任务。

支付渠道 `payment_orders.channel` 有效值：

```text
alipay    支付宝
wechat    微信支付
balance   余额支付
manual    后台人工入账或人工处理
```

`payment_orders.third_trade_no` 可为空。MariaDB 的唯一索引允许多个 `NULL`，所以 `channel + NULL` 不会冲突；第三方支付成功后，同一渠道的同一 `third_trade_no` 必须唯一。

钱包流水必须记录 `balance_after_cents`，用于后台追账和对账。余额变动使用 `wallet_accounts.version` 做乐观锁或事务内行锁。

`wallet_accounts.frozen_cents` 第一阶段只作为预留汇总字段。当前设计不启用冻结和解冻明细流程；后续如果需要预授权、争议冻结或退款冻结，应新增钱包冻结明细表，记录冻结来源、金额、状态和释放时间。

退款第一阶段只在 `payment_orders` 上预留 `refunding`、`refunded` 状态和 `refunded_at`。后续如果需要部分退款、退款审核、退款失败重试或多次退款，应新增独立退款记录表。

### 实例和异步任务

| 表名 | 说明 |
| --- | --- |
| `instances` | 云服务器实例 |
| `async_tasks` | 数据库任务队列 |

实例开通必须先创建或复用本地 `instances` 占位记录，再分配并持久化 `vmid`、`provisioning_key`、`pve_task_upid` 等恢复锚点。这样即使 PVE 创建成功后本地流程中断，也能通过补偿任务找回远端 VM。

`async_tasks.idempotency_key` 用于限制同一业务对象的同类任务只能存在一条有效任务。例如新购订单的开通任务可以使用 `instance_create:{order_no}`。

### 工单、配置和审计

| 表名 | 说明 |
| --- | --- |
| `tickets` | 工单主表 |
| `ticket_messages` | 工单回复 |
| `system_configs` | 系统配置 |
| `admin_audit_logs` | 后台操作审计日志 |

审计日志只记录后台关键操作，敏感字段需要脱敏后再写入 `before_data` 和 `after_data`。

## 关键约束

- `orders.order_no` 唯一。
- `payment_orders.payment_no` 唯一。
- `payment_orders.third_trade_no` 可为空，但第三方支付成功后应保证同渠道唯一。
- `region_images.node_id + region_images.image_id` 唯一，防止同一节点重复配置同一镜像模板。
- `instances.order_id` 唯一，第一期一单一实例。
- `instances.node_id + instances.vmid` 唯一，防止 PVE VMID 冲突。
- `instances.provisioning_key` 唯一，防止重复开通。
- `async_tasks.idempotency_key` 唯一，防止重复任务。
- `wallet_accounts.user_id` 唯一，一个用户只有一个余额账户。

## 主要索引

- 用户查询：`users.email`、`users.phone`、`users.username`。
- 订单列表：`orders.user_id + created_at`、`orders.status + expired_at`。
- 支付补偿：`payment_orders.status + created_at`。
- 地域镜像：`region_images.region_id + status + sort_order`。
- 实例列表：`instances.user_id + status`、`instances.expire_at`。
- 任务拉取：`async_tasks.status + run_at`、`async_tasks.locked_until`。
- 工单后台列表：`tickets.status + updated_at`。
- 审计查询：`admin_audit_logs.admin_id + created_at`、`admin_audit_logs.object_type + object_id`。

## 事务边界

### 下单事务

在同一事务内完成：

- 校验套餐、地域、镜像和价格。
- 创建 `orders`。
- 创建 `payment_orders`，或为余额支付预留扣款逻辑。

### 支付成功事务

在同一事务内完成：

- 锁定并更新 `payment_orders`。
- 按支付场景更新 `orders` 或 `wallet_accounts`。
- 写入 `wallet_transactions`。
- 创建唯一 `async_tasks`。

事务内不能执行长耗时外部调用。

### 实例开通事务

至少拆成两个本地事务：

1. claim 任务后，锁定订单并创建或复用 `instances` 占位记录，持久化 VMID、幂等键和订单 `provisioning` 状态。
2. PVE 任务成功后，写入实例最终信息，更新实例、订单和任务状态。

PVE HTTP 调用不放在数据库事务内。

## 初始化 SQL

初始化 SQL 位于：

```text
server/migrations/001_init.sql
```

该文件包含建库、建表、基础权限码、默认超级管理员角色和系统配置占位数据。默认管理员账号后续通过命令行脚本或后台初始化流程创建。
