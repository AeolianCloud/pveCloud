# 平台与管理基础 API

本文档维护系统检查、管理端认证会话、Dashboard、管理员、角色权限、管理员会话和系统配置相关接口。跨接口通用约定见 `docs/server/api/conventions.md`，API 目录入口见 `docs/server/api/README.md`。

## 系统检查

### `GET /healthz`

- 鉴权：无
- 作用：检查 API 进程、MariaDB 和 Redis 是否可用

### `GET /admin-api/ping`

- 鉴权：无
- 作用：管理端 API 入口连通性检查

## 管理端认证与会话

### `GET /admin-api/auth/captcha`

- 鉴权：无
- 作用：获取管理端登录验证码
- 成功数据包含：`captcha_id`、验证码图片、有效期

### `POST /admin-api/auth/login`

- 鉴权：无
- 作用：管理员账号密码登录
- 请求字段：`username`、`password`、`captcha_id`、`captcha_code`
- 成功数据包含：
  - `access_token`
  - `token_type`
  - `expires_in`
  - `admin`
  - `role_ids`
  - `permission_codes`
  - `menus`
  - `session`

### `GET /admin-api/auth/me`

- 鉴权：管理端 Bearer Token
- 作用：恢复当前管理员、权限快照、后端菜单树与会话状态
- 成功数据包含：
  - `admin`
  - `role_ids`
  - `permission_codes`
  - `menus`
  - `session`
- `menus` 由 `admin_permissions` 中 `type=menu` 且当前管理员拥有的权限节点生成，前端侧栏按该树渲染。

### `POST /admin-api/auth/logout`

- 鉴权：管理端 Bearer Token
- 作用：注销当前会话

### `POST /admin-api/auth/refresh`

- 鉴权：管理端 Bearer Token
- 作用：轮换新 token 和新会话
- 成功响应结构与登录成功响应保持一致

## 管理端 Dashboard

### `GET /admin-api/dashboard`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.dashboard`
- 作用：获取当前管理端 Dashboard 首页数据
- 成功数据包含：
  - `admin`
  - `role_ids`
  - `permission_codes`
  - `menus`
  - `session`
  - `metrics`
  - `business_metrics`

`metrics` 表示基础后台指标。每项包含：

- `key`：指标唯一键
- `title`：展示标题
- `value`：整数值
- `unit`：单位，可为空

当前基础后台指标包括：

- `active_admins`：启用管理员数量，来源 `admin_users.status = active AND deleted_at IS NULL`
- `active_roles`：启用角色数量，来源 `admin_roles.status = active`
- `active_sessions`：活跃会话数量，来源 `admin_sessions.status = active AND expires_at > NOW(3)`
- `audit_logs_today`：今日操作日志数量，来源 `admin_audit_logs.created_at >= 当日 00:00:00`

`business_metrics` 表示运营待办和异常指标。每项包含：

- `key`：指标唯一键
- `title`：展示标题
- `value`：整数值
- `unit`：单位，可为空
- `description`：指标说明
- `target_path`：目标管理端页面路径，可为空
- `target_permission`：目标页面权限，可为空
- `severity`：展示严重度，允许 `info`、`warning`、`error`

当前业务指标包括：

- `pending_orders`：待处理订单数量，来源 `orders.status = pending`，目标 `/orders`，权限 `page.orders`
- `order_errors`：交付异常订单数量，来源 `orders.status = error`，目标 `/orders`，权限 `page.orders`
- `instance_errors`：异常实例数量，来源 `instances.status = error`，目标 `/instances`，权限 `page.instances`
- `failed_async_tasks`：失败异步任务数量，来源 `async_tasks.status = failed`，目标 `/async-tasks`，权限 `page.async-tasks`
- `pending_tickets`：待后台处理工单数量，来源 `tickets.status = waiting_admin`，目标 `/tickets`，权限 `page.tickets`
- `invoice_todo`：待处理发票数量，来源 `invoice_applications.status IN (pending, processing)`，目标 `/invoices`，权限 `page.invoices`
- `payment_exceptions`：支付异常数量，来源 `payment_transactions.status = failed` 与 `refund_transactions.status IN (pending, failed)` 的合计，目标 `/payments`，权限 `page.payments`

约束：

- Dashboard 只读聚合当前已开放业务事实，不直接修改订单、支付、退款、实例、异步任务、工单或发票状态。
- Dashboard 业务指标不得展示商户密钥、完整回调 payload、完整上游响应、PVE/MCP Bearer Token、PVE 节点、VMID、Worker payload/result 或其它敏感详情。
- 前端可根据 `target_permission` 控制跳转入口可见性；目标业务页面和目标业务接口仍必须按各自菜单权限、操作权限和状态规则最终裁决。
- Dashboard 不开放未纳入当前契约的专票、红冲、部分退款、提现、人工调账、余额转账、自动对账、实例控制台、重装、快照、备份或通用 PVE 运维能力。

## 管理员账号域

### `GET /admin-api/admin-users`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.admin-users`
- 作用：分页查询管理员账号

### `POST /admin-api/admin-users`

- 鉴权：管理端 Bearer Token
- 操作权限：`admin-user:create` 或 `admin-user:*`
- 作用：创建管理员账号
- 约束：若创建时分配角色，目标角色展开后的全部权限必须是当前操作者实时数据库权限集合的子集

### `GET /admin-api/admin-users/{id}`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.admin-users`
- 作用：查看管理员详情

### `PATCH /admin-api/admin-users/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`admin-user:update` 或 `admin-user:*`
- 作用：更新管理员信息、状态和角色
- 约束：管理员不能通过该接口修改自己的 `role_ids`；给其它管理员分配角色时，目标角色展开后的全部权限必须是当前操作者实时数据库权限集合的子集

### `POST /admin-api/admin-users/{id}/password`

- 鉴权：管理端 Bearer Token
- 操作权限：`admin-user:password-reset` 或 `admin-user:*`
- 作用：重置管理员密码

## 角色与权限域

### `GET /admin-api/admin-roles`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.admin-roles`
- 作用：查询角色列表

### `POST /admin-api/admin-roles`

- 鉴权：管理端 Bearer Token
- 操作权限：`admin-role:create` 或 `admin-role:*`
- 作用：创建角色
- 约束：提交的 `permission_codes` 必须是当前操作者实时数据库权限集合的子集，禁止通过角色创建授予操作者未拥有的权限

### `GET /admin-api/admin-roles/{id}`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.admin-roles`
- 作用：查看角色详情

### `PATCH /admin-api/admin-roles/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`admin-role:update` 或 `admin-role:*`
- 作用：更新角色信息、状态和权限
- 约束：提交的 `permission_codes` 必须是当前操作者实时数据库权限集合的子集，禁止通过角色编辑授予操作者未拥有的权限

### `GET /admin-api/admin-permissions`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.admin-roles`
- 作用：只读查询菜单和操作权限目录树
- 成功数据为树形节点数组，每个节点包含：`code`、`name`、`type`、`parent_code`、`path`、`icon`、`sort_order`、`description`、`children`

## 管理员会话域

### `GET /admin-api/admin-sessions`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.admin-sessions`
- 作用：分页查询管理员会话列表
- 查询参数支持：`page`、`per_page`、`keyword`、`status`

### `PATCH /admin-api/admin-sessions/{session_id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`admin-session:revoke` 或 `admin-session:*`
- 作用：吊销指定管理员会话
- 请求字段：`status`，当前固定为 `revoked`
- 约束：不得通过该接口吊销当前会话自身

## 系统配置域

### `GET /admin-api/system-configs`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.config`
- 作用：按配置分组查询系统配置
- 成功数据包含配置分组和配置项；配置项包含 `id`、`config_key`、`config_value`、`value_type`、`group_name`、`is_secret`、`has_value`、`description`
- 约束：
  - `is_secret=1` 的配置不得返回明文，`config_value` 必须为空或固定掩码，只通过 `has_value` 表示是否已配置
  - 支付宝、微信侧实名供应商密钥和证件摘要密钥都作为后台敏感配置管理

### `PATCH /admin-api/system-configs/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`system-config:update` 或 `system-config:*`
- 作用：更新系统配置
- 约束：
  - 更新 `is_secret=1` 配置时，仅非空新值会覆盖旧值；空值表示保留旧敏感值
  - `real_name.identity_digest_secret` 是外部供应商实名申请和证件摘要重复校验的敏感配置；缺少时外部供应商不可用，但不影响人工审核实名入口；已有当前 HMAC 版本实名申请后不得通过普通系统设置直接修改
  - 更新实名供应商启用开关时，服务端必须校验对应供应商必要后台配置是否完整
