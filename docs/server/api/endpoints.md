# API 接口清单

本文件维护当前已确认的基础接口契约。跨接口响应、错误码、鉴权和幂等规则见 `docs/server/api/conventions.md`。

## 系统检查

### `GET /healthz`

- 鉴权：无。
- 用途：检查 API 进程、MariaDB 和 Redis 是否可用。
- 成功响应：`data.app`、`data.env`、`data.database`、`data.redis`、`data.time`。
- 失败响应：服务或依赖不可用时返回服务端错误。

### `GET /api/ping`

- 鉴权：无。
- 用途：用户端 API 入口检查。
- 成功响应：`data.scope="api"`、`data.pong=true`。

### `GET /admin-api/ping`

- 鉴权：无。
- 用途：管理端 API 入口检查。
- 成功响应：`data.scope="admin-api"`、`data.pong=true`。

## 管理端登录和会话

### `GET /admin-api/auth/captcha`

- 鉴权：无。
- 用途：获取管理员登录验证码。
- 成功响应：`data.captcha_id`、`data.image`、`data.expires_in`。
- 失败响应：验证码获取过于频繁时返回限流错误。

### `POST /admin-api/auth/login`

- 鉴权：无。
- 用途：管理员账号密码和验证码登录。
- 请求参数：`username`、`password`、`captcha_id`、`captcha_code`。
- 成功响应：`data.access_token`、`data.token_type`、`data.expires_in`、`data.admin`、`data.role_ids`、`data.permission_codes`、`data.session`。
- 失败响应：参数或验证码错误、账号密码错误、登录失败次数过多。

### `GET /admin-api/auth/me`

- 鉴权：管理端 Bearer Token。
- 用途：恢复当前管理员、角色、权限、菜单和会话状态。
- 成功响应：`data.admin`、`data.role_ids`、`data.permission_codes`、`data.menus`、`data.session`。
- 菜单范围：当前阶段只返回基础后台管理菜单，不返回产品、订单、支付、实例、工单等业务菜单。
- 失败响应：未登录、token 无效、token 已吊销、token 过期或管理员账号已禁用。

### `POST /admin-api/auth/logout`

- 鉴权：管理端 Bearer Token。
- 用途：吊销当前管理员会话。
- 成功响应：`data=null`。
- 失败响应：未登录、token 无效、token 已吊销或 token 过期。

### `POST /admin-api/auth/refresh`

- 鉴权：管理端 Bearer Token。
- 用途：轮换管理端访问令牌，并吊销旧会话。
- 成功响应：同 `POST /admin-api/auth/login`。
- 失败响应：未登录、token 无效、token 已吊销、token 过期或管理员账号已禁用。

## 管理端首页

### `GET /admin-api/dashboard`

- 鉴权：管理端 Bearer Token。
- 权限：`dashboard:view`。
- 用途：获取管理端首页初始数据。
- 成功响应：同 `GET /admin-api/auth/me`，并额外返回 `data.metrics`。
- 菜单范围：当前阶段只返回基础后台管理菜单。
- 指标字段：`key`、`title`、`value`、`unit`。
- 当前阶段指标：`active_admins`、`active_roles`、`active_sessions`、`risk_logs_today`。
- 失败响应：未登录、token 无效、token 已吊销、token 过期或缺少权限。

## 基础后台管理员

### `GET /admin-api/admin-users`

- 鉴权：管理端 Bearer Token。
- 权限：`admin:manage`。
- 用途：分页查询管理员账号。
- 查询参数：`page`、`per_page`、`keyword`、`status`、`role_id`。
- 成功响应：`data.list`、`data.total`、`data.page`、`data.per_page`、`data.last_page`。列表项包含 `id`、`username`、`email`、`display_name`、`status`、`role_ids`、`roles`、`last_login_at`、`last_login_ip`、`created_at`、`updated_at`。
- 失败响应：未登录、无权限、参数错误。

### `POST /admin-api/admin-users`

- 鉴权：管理端 Bearer Token。
- 权限：`admin:manage`。
- 用途：创建管理员账号。
- 请求参数：`username`、`email`、`display_name`、`password`、`status`、`role_ids`。
- 成功响应：新建管理员摘要。
- 失败响应：未登录、无权限、参数错误、用户名或邮箱重复。

### `GET /admin-api/admin-users/{id}`

- 鉴权：管理端 Bearer Token。
- 权限：`admin:manage`。
- 用途：查看管理员详情。
- 成功响应：管理员摘要、角色、权限和会话摘要。
- 失败响应：未登录、无权限、管理员不存在。

### `PATCH /admin-api/admin-users/{id}`

- 鉴权：管理端 Bearer Token。
- 权限：`admin:manage`。
- 用途：更新管理员资料、状态和角色。
- 请求参数：`email`、`display_name`、`status`、`role_ids`，均为可选字段。
- 成功响应：更新后的管理员摘要。
- 失败响应：未登录、无权限、参数错误、管理员不存在、禁止禁用当前账号。

### `POST /admin-api/admin-users/{id}/password`

- 鉴权：管理端 Bearer Token。
- 权限：`admin:manage`。
- 用途：重置管理员密码。
- 请求参数：`password`。
- 成功响应：`data=null`。
- 失败响应：未登录、无权限、参数错误、管理员不存在。

## 基础后台角色和权限

### `GET /admin-api/admin-roles`

- 鉴权：管理端 Bearer Token。
- 权限：`admin:manage`。
- 用途：分页查询管理端角色。
- 查询参数：`page`、`per_page`、`keyword`、`status`。
- 成功响应：分页角色列表。列表项包含 `id`、`code`、`name`、`description`、`status`、`permission_codes`、`created_at`、`updated_at`。
- 失败响应：未登录、无权限、参数错误。

### `POST /admin-api/admin-roles`

- 鉴权：管理端 Bearer Token。
- 权限：`admin:manage`。
- 用途：创建管理端角色。
- 请求参数：`code`、`name`、`description`、`status`、`permission_codes`。
- 成功响应：新建角色摘要。
- 失败响应：未登录、无权限、参数错误、角色编码重复。

### `GET /admin-api/admin-roles/{id}`

- 鉴权：管理端 Bearer Token。
- 权限：`admin:manage`。
- 用途：查看角色详情。
- 成功响应：角色摘要和权限码列表。
- 失败响应：未登录、无权限、角色不存在。

### `PATCH /admin-api/admin-roles/{id}`

- 鉴权：管理端 Bearer Token。
- 权限：`admin:manage`。
- 用途：更新角色资料、状态和权限。
- 请求参数：`name`、`description`、`status`、`permission_codes`，均为可选字段。
- 成功响应：更新后的角色摘要。
- 失败响应：未登录、无权限、参数错误、角色不存在、禁止禁用内置超级管理员角色。

### `GET /admin-api/admin-permissions`

- 鉴权：管理端 Bearer Token。
- 权限：`admin:manage`。
- 用途：查询系统权限码清单。
- 查询参数：`group_name`。
- 成功响应：权限分组列表。权限项包含 `id`、`code`、`name`、`group_name`、`description`。
- 失败响应：未登录、无权限。

## 基础后台登录会话

### `GET /admin-api/admin-sessions`

- 鉴权：管理端 Bearer Token。
- 权限：`admin:manage`。
- 用途：分页查询管理端登录会话。
- 查询参数：`page`、`per_page`、`admin_id`、`status`、`keyword`。
- 成功响应：分页会话列表。列表项包含 `id`、`session_id`、`admin`、`status`、`issued_at`、`expires_at`、`last_seen_at`、`last_seen_ip`、`user_agent`、`revoked_at`、`revoke_reason`。
- 失败响应：未登录、无权限、参数错误。

### `POST /admin-api/admin-sessions/{id}/revoke`

- 鉴权：管理端 Bearer Token。
- 权限：`admin:manage`。
- 用途：吊销指定活跃会话。
- 成功响应：`data=null`。
- 失败响应：未登录、无权限、会话不存在、会话已失效、禁止通过该接口吊销当前会话。

## 基础后台系统配置

### `GET /admin-api/system-configs`

- 鉴权：管理端 Bearer Token。
- 权限：`system:update`。
- 用途：按分组查询系统配置。
- 查询参数：`group_name`。
- 成功响应：配置分组列表。配置项包含 `id`、`config_key`、`config_value`、`value_type`、`group_name`、`is_secret`、`has_value`、`description`、`updated_at`。`is_secret=true` 时不返回明文 `config_value`。
- 失败响应：未登录、无权限。

### `PATCH /admin-api/system-configs/{id}`

- 鉴权：管理端 Bearer Token。
- 权限：`system:update`。
- 用途：更新系统配置值。
- 请求参数：`config_value`。
- 成功响应：更新后的配置摘要，敏感配置仍不返回明文。
- 失败响应：未登录、无权限、参数错误、配置不存在。

## 基础后台审计日志

### `GET /admin-api/audit-logs`

- 鉴权：管理端 Bearer Token。
- 权限：`audit:view`。
- 用途：分页查询后台操作审计日志。
- 查询参数：`page`、`per_page`、`admin_id`、`action`、`object_type`、`object_id`、`date_from`、`date_to`。
- 成功响应：分页审计日志列表。列表项包含 `id`、`admin`、`action`、`object_type`、`object_id`、`before_data`、`after_data`、`ip`、`user_agent`、`remark`、`created_at`。
- 敏感字段权限：只有同时具备 `audit:sensitive_view` 时才返回 `before_data`、`after_data` 和 `user_agent`；仅具备 `audit:view` 时这些字段返回 `null`。密码、token、secret、验证码和敏感配置明文永不返回。
- 失败响应：未登录、无权限、参数错误。

## 基础后台高危操作日志

### `GET /admin-api/risk-logs`

- 鉴权：管理端 Bearer Token。
- 权限：`audit:view`。
- 用途：分页查询后台高危操作日志。
- 查询参数：`page`、`per_page`、`admin_id`、`risk_level`、`action`、`object_type`、`object_id`、`date_from`、`date_to`。
- 成功响应：分页高危操作日志列表。列表项包含 `id`、`audit_log_id`、`admin`、`risk_level`、`action`、`object_type`、`object_id`、`risk_reason`、`before_data`、`after_data`、`ip`、`user_agent`、`remark`、`created_at`。
- 敏感字段权限：只有同时具备 `audit:sensitive_view` 时才返回 `before_data`、`after_data` 和 `user_agent`；仅具备 `audit:view` 时这些字段返回 `null`。密码、token、secret、验证码和敏感配置明文永不返回。
- 失败响应：未登录、无权限、参数错误。
