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
- 菜单范围：当前只返回已完成功能菜单，第一阶段只包含 `/dashboard`。
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
- 菜单范围：当前只返回已完成功能菜单，第一阶段只包含 `/dashboard`。
- 指标字段：`key`、`title`、`value`、`unit`。
- 失败响应：未登录、token 无效、token 已吊销、token 过期或缺少权限。
