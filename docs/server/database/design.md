# 数据库设计

可执行表结构最终以 `server/migrations/` 为准。
本文件记录数据库设计口径、表分组、关键约束和事务边界。

## 基础环境

```text
database: pvecloud
engine: MariaDB 11.4.x / InnoDB
charset: utf8mb4
collation: utf8mb4_unicode_ci
```

## 设计原则

- 主键使用 `BIGINT UNSIGNED AUTO_INCREMENT`
- 金额字段使用整数分，字段名以 `_cents` 结尾
- 状态字段使用字符串，不使用数据库 enum
- 表和字段写明 `COMMENT`
- 常规时间字段使用 `DATETIME(3)`
- 对外展示优先使用业务编号，不直接暴露自增 ID
- JSON 字段只用于快照、第三方 payload 或低频配置片段

## 表分组

### 账号与权限

```text
users
admin_users
admin_roles
admin_permissions
admin_user_roles
admin_role_permissions
admin_sessions
```

### 产品目录

```text
products
product_plans
regions
pve_nodes
images
region_images
plan_prices
```

### 订单、支付、钱包

```text
orders
payment_orders
payment_notify_logs
wallet_accounts
wallet_transactions
```

### 实例与任务

```text
instances
async_tasks
```

### 工单、配置、审计

```text
tickets
ticket_messages
system_configs
admin_audit_logs
admin_risk_logs
```

## 管理端关键规则

- 管理端专用表使用 `admin_` 前缀
- 权限码分为页面入口权限和资源操作权限：
  - 页面入口权限使用 `page.<menu>.<feature>`
  - 资源操作权限使用 `resource:action`
- 管理端会话最终状态以 `admin_sessions` 为准
- `super_admin` 角色应始终拥有当前 `admin_permissions` 中定义的全部权限
- JWT 中的角色和权限快照只用于登录响应与前端体验，不替代服务端当前 RBAC 校验
- `system_configs.is_secret=1` 的配置不得通过接口返回明文
- 高危操作同时写入 `admin_audit_logs` 和 `admin_risk_logs`
- 风险日志属于审计域

## 当前阶段说明

当前基础后台阶段，后端仍保留认证、RBAC、会话、系统配置、审计和高危日志等管理域数据结构。
这不意味着当前管理端前端必须保留这些独立页面。

当前开放的管理端权限码以 `server/migrations/004_admin_permission_refactor.sql`、`005_admin_permission_cleanup.sql`、`006_admin_page_permissions.sql`、`007_admin_session_permissions.sql` 和 `008_super_admin_full_permissions.sql` 的最终结果为准。
`admin-session:*`、`admin-session:view`、`admin-session:revoke` 与 `page.system-settings.admin-sessions` 当前已通过 `007_admin_session_permissions.sql` 重新纳入开放范围；`008_super_admin_full_permissions.sql` 负责把 `super_admin` 回填并同步到全部现存权限。
审计日志查询和高危日志查询相关权限码当前仍不开放；重新开放时必须新增迁移并同步 API 文档和路由。

## 关键唯一约束示例

- `orders.order_no`
- `payment_orders.payment_no`
- `instances.order_id`
- `instances.provisioning_key`
- `async_tasks.idempotency_key`
- `wallet_accounts.user_id`
- `admin_sessions.session_id`

## 一致性原则

- MariaDB 是业务事实来源
- Redis 只做缓存、限流、短 TTL 状态和辅助幂等
- 外部系统调用通过本地恢复锚点、补偿或重试来保持一致性
- 不在长事务中调用 PVE、支付或通知系统
