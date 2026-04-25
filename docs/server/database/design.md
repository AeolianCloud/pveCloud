# 数据库设计

数据库可执行结构以 `server/migrations/` 为准。本文件记录设计口径、表分组、关键约束和事务边界。

## 目标环境

```text
database: pvecloud
engine: MariaDB 11.4.9 / InnoDB
charset: utf8mb4
collation: utf8mb4_unicode_ci
```

MariaDB 是业务事实来源。PVE 是外部资源系统。

## 基础约定

- 主键使用 `BIGINT UNSIGNED AUTO_INCREMENT`。
- 金额字段使用整数分，字段名以 `_cents` 结尾。
- 状态字段使用 `VARCHAR` 和 Go 常量，不使用数据库 enum。
- 表和字段必须写 MariaDB `COMMENT`。
- 时间字段使用 `DATETIME(3)`。
- 常规表包含 `created_at` 和 `updated_at`；软删除表包含 `deleted_at`。
- 对外展示使用业务编号，例如 `order_no`、`payment_no`、`instance_no`，不直接暴露自增 ID。
- JSON 字段用于快照、第三方 payload、配置片段，不用于高频查询条件。

## 表分组

账号和权限：

```text
users
admin_users
admin_roles
admin_permissions
admin_user_roles
admin_role_permissions
admin_sessions
```

产品目录：

```text
products
product_plans
regions
pve_nodes
images
region_images
plan_prices
```

订单、支付、钱包：

```text
orders
payment_orders
payment_notify_logs
wallet_accounts
wallet_transactions
```

实例和任务：

```text
instances
async_tasks
```

工单、配置、审计：

```text
tickets
ticket_messages
system_configs
admin_audit_logs
```

## 关键业务规则

- 管理端专用表使用 `admin_` 前缀。
- 管理端权限码使用 `domain:action`，例如 `order:view` 和 `payment:manual_credit`。
- 管理端登录会话使用 `admin_sessions` 持久化，`session_id` 对应 JWT `jti`，用于退出登录、刷新轮换、账号禁用后的会话吊销和会话自检。
- `admin_sessions.status` 使用 `active`、`revoked`、`expired`；退出登录将当前会话标记为 `revoked/logout`，刷新 token 将旧会话标记为 `revoked/refresh` 并创建新会话。
- 管理端权限以数据库 RBAC 为最终边界；JWT 中的角色和权限只作为登录响应快照，不作为受保护接口的唯一授权来源。
- 产品价格唯一性是 `plan_id + region_id + billing_period`。
- 第一阶段一个订单对应一个实例，不添加 `quantity`。
- 订单保存产品和价格快照，避免后续改价影响历史订单。
- `images` 是逻辑系统镜像，实际 PVE 模板存储在 `region_images.pve_template_id`。
- 用户只能看到所选地域启用的镜像。
- 钱包流水必须写入 `balance_after_cents`。
- 钱包余额变更使用乐观锁或 `wallet_accounts.version` 行锁策略。
- `wallet_accounts.frozen_cents` 仅预留；实现真实冻结流程前需增加冻结明细表。
- `payment_orders` 上的退款状态当前为预留；支持部分退款、多次退款或退款审核前需增加退款表。

## 关键约束

- `orders.order_no` 唯一。
- `payment_orders.payment_no` 唯一。
- `payment_orders.channel + third_trade_no` 在 `third_trade_no` 存在时唯一。
- `region_images.node_id + image_id` 唯一。
- `instances.order_id` 唯一。
- `instances.node_id + vmid` 唯一。
- `instances.provisioning_key` 唯一。
- `async_tasks.idempotency_key` 唯一。
- `wallet_accounts.user_id` 唯一。
- `admin_sessions.session_id` 唯一。

## 重要索引

- `users.email`、`users.phone`、`users.username`。
- `orders.user_id + created_at`、`orders.status + expired_at`。
- `payment_orders.status + created_at`。
- `region_images.region_id + status + sort_order`。
- `instances.user_id + status`、`instances.expire_at`。
- `async_tasks.status + run_at`、`async_tasks.locked_until`。
- `tickets.status + updated_at`。
- `admin_audit_logs.admin_id + created_at`、`admin_audit_logs.object_type + object_id`。
- `admin_sessions.admin_id + status`、`admin_sessions.status + expires_at`。

管理端认证事务：

- 登录成功时校验账号、密码和状态，创建 `admin_sessions`，签发带 `jti` 的管理端 JWT，更新 `admin_users.last_login_at` 和 `last_login_ip`，并写入 `admin_audit_logs`。
- 登录失败时不创建会话，但按账号标识哈希和 IP 做短窗口失败次数限制，并写入 `admin_audit_logs`，`admin_id` 可为空。
- 登录失败短窗口计数放在 Redis，MariaDB 的 `admin_audit_logs` 只作为审计记录，不承担登录限流计数兜底。
- Redis 可用于幂等短锁和防重复提交标记，但订单、支付、钱包流水、实例、任务和审计的最终状态必须写入 MariaDB。
- 受保护接口校验 JWT 后必须确认 `admin_sessions.status=active` 且 `expires_at` 未过期，再从数据库读取当前管理员状态、角色和权限写入请求上下文。
- 刷新 token 时在同一事务内吊销旧会话并创建新会话，避免同一个旧 token 被重复刷新。
- 退出登录只吊销当前会话，不影响同一管理员的其它设备会话。

## 事务

订单创建事务：

- 校验套餐、地域、镜像和价格。
- 创建 `orders`。
- 创建 `payment_orders`，或预留余额支付扣减逻辑。

支付成功事务：

- 锁定并更新 `payment_orders`。
- 按 `payment_scene` 和 `order_type` 分支。
- 更新 `orders` 或 `wallet_accounts`。
- 写入 `wallet_transactions`。
- 创建唯一 `async_tasks`。

实例开通至少拆为两个本地事务：

1. 领取任务、锁定订单、创建或复用实例占位、持久化 VMID/幂等锚点，并标记订单开通中。
2. PVE 任务成功后，更新实例、订单和任务最终状态。

不要在长数据库事务中调用 PVE HTTP。
