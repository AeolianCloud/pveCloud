# API 接口总览

本文件维护当前已确认的接口清单与主要契约口径。
跨接口通用约定见 `docs/server/api/conventions.md`。

## 系统检查

### `GET /healthz`

- 鉴权：无
- 作用：检查 API 进程、MariaDB 与 Redis 是否可用

### `GET /api/ping`

- 鉴权：无
- 作用：用户端 API 入口连通性检查

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
  - `session`

### `GET /admin-api/auth/me`

- 鉴权：管理端 Bearer Token
- 作用：恢复当前管理员、权限快照、菜单快照与会话状态
- 成功数据包含：
  - `admin`
  - `role_ids`
  - `permission_codes`
  - `menus`
  - `session`

### `POST /admin-api/auth/logout`

- 鉴权：管理端 Bearer Token
- 作用：吊销当前会话

### `POST /admin-api/auth/refresh`

- 鉴权：管理端 Bearer Token
- 作用：轮换新 token 和新会话
- 成功响应结构与登录成功响应保持一致

## 管理端 Dashboard

### `GET /admin-api/dashboard`

- 鉴权：管理端 Bearer Token
- 页面入口权限建议：`page.dashboard`
- 资源权限：`dashboard:view` 或 `dashboard:*`
- 作用：获取当前基础后台首页数据
- 成功数据包含：
  - `admin`
  - `role_ids`
  - `permission_codes`
  - `menus`
  - `session`
  - `metrics`

当前阶段 Dashboard 只展示基础后台相关指标，不展示未开放业务模块数据。

## 管理员账号域

### `GET /admin-api/admin-users`

- 鉴权：管理端 Bearer Token
- 页面入口权限建议：`page.system-settings.admin-users`
- 资源权限：`admin-user:view` 或 `admin-user:*`
- 作用：分页查询管理员账号

### `POST /admin-api/admin-users`

- 鉴权：管理端 Bearer Token
- 页面入口权限建议：`page.system-settings.admin-users`
- 资源权限：`admin-user:create` 或 `admin-user:*`
- 作用：创建管理员账号

### `GET /admin-api/admin-users/{id}`

- 鉴权：管理端 Bearer Token
- 页面入口权限建议：`page.system-settings.admin-users`
- 资源权限：`admin-user:view` 或 `admin-user:*`
- 作用：查看管理员详情

### `PATCH /admin-api/admin-users/{id}`

- 鉴权：管理端 Bearer Token
- 页面入口权限建议：`page.system-settings.admin-users`
- 资源权限：`admin-user:update` 或 `admin-user:*`
- 作用：更新管理员信息、状态和角色

### `POST /admin-api/admin-users/{id}/password`

- 鉴权：管理端 Bearer Token
- 页面入口权限建议：`page.system-settings.admin-users`
- 资源权限：`admin-user:password-reset` 或 `admin-user:*`
- 作用：重置管理员密码

## 角色与权限域

### `GET /admin-api/admin-roles`

- 鉴权：管理端 Bearer Token
- 页面入口权限建议：`page.system-settings.admin-roles`
- 资源权限：`admin-role:view` 或 `admin-role:*`
- 作用：查询角色列表

### `POST /admin-api/admin-roles`

- 鉴权：管理端 Bearer Token
- 页面入口权限建议：`page.system-settings.admin-roles`
- 资源权限：`admin-role:create` 或 `admin-role:*`
- 作用：创建角色

### `GET /admin-api/admin-roles/{id}`

- 鉴权：管理端 Bearer Token
- 页面入口权限建议：`page.system-settings.admin-roles`
- 资源权限：`admin-role:view` 或 `admin-role:*`
- 作用：查看角色详情

### `PATCH /admin-api/admin-roles/{id}`

- 鉴权：管理端 Bearer Token
- 页面入口权限建议：`page.system-settings.admin-roles`
- 资源权限：`admin-role:update` 或 `admin-role:*`
- 作用：更新角色信息、状态和权限

### `GET /admin-api/admin-permissions`

- 鉴权：管理端 Bearer Token
- 页面入口权限建议：`page.system-settings.admin-roles`
- 资源权限：`admin-role:view` 或 `admin-role:*`
- 作用：只读查询权限码分组

## 会话与系统配置域

### `GET /admin-api/admin-sessions`

- 鉴权：管理端 Bearer Token
- 页面入口权限建议：`page.admin-sessions`
- 资源权限：`admin-session:view` 或 `admin-session:*`
- 作用：查询管理端会话列表

### `POST /admin-api/admin-sessions/{id}/revoke`

- 鉴权：管理端 Bearer Token
- 页面入口权限建议：`page.admin-sessions`
- 资源权限：`admin-session:revoke` 或 `admin-session:*`
- 作用：吊销他人会话

### `GET /admin-api/system-configs`

- 鉴权：管理端 Bearer Token
- 页面入口权限建议：`page.system-settings.config`
- 资源权限：`system-config:view` 或 `system-config:*`
- 作用：按配置分组查询系统配置

### `PATCH /admin-api/system-configs/{id}`

- 鉴权：管理端 Bearer Token
- 页面入口权限建议：`page.system-settings.config`
- 资源权限：`system-config:update` 或 `system-config:*`
- 作用：更新系统配置

## 审计域

### `GET /admin-api/audit-logs`

- 鉴权：管理端 Bearer Token
- 页面入口权限建议：`page.audit-logs`
- 资源权限：`audit-log:view` 或 `audit-log:*`
- 作用：查询后台审计日志

### `GET /admin-api/risk-logs`

- 鉴权：管理端 Bearer Token
- 页面入口权限建议：`page.risk-logs`
- 资源权限：`risk-log:view` 或 `risk-log:*`
- 作用：查询高危操作日志

### 敏感字段查看

- `audit-log:view`：仅查看日志主信息
- `audit-log:sensitive-view` 或 `audit-log:*`：查看已脱敏的 `before_data`、`after_data` 和 `user_agent`

密码、token、secret、验证码和敏感配置明文不得出现在接口响应中。

## 建议权限清单

当前后台建议采用以下细粒度权限清单：

- `page.dashboard`
- `page.system-settings.config`
- `page.system-settings.admin-users`
- `page.system-settings.admin-roles`
- `page.admin-sessions`
- `page.audit-logs`
- `page.risk-logs`

- `dashboard:*`
- `dashboard:view`

- `system-config:*`
- `system-config:view`
- `system-config:update`

- `admin-user:*`
- `admin-user:view`
- `admin-user:create`
- `admin-user:update`
- `admin-user:password-reset`

- `admin-role:*`
- `admin-role:view`
- `admin-role:create`
- `admin-role:update`

- `admin-session:*`
- `admin-session:view`
- `admin-session:revoke`

- `audit-log:*`
- `audit-log:view`
- `audit-log:sensitive-view`

- `risk-log:*`
- `risk-log:view`

扩展原则：

- 新页面或新 tab 若未来可能独立授权，至少预留一个 `page.*`
- 新页面存在独立资源读写时，再补对应 `resource:view`、`create`、`update`、`delete` 或其它特殊动作
- `resource:*` 只作为资源快捷授权入口，不替代页面入口权限

## 前端范围说明

当前这些接口里，管理端前端实际消费以下范围：

- 登录、会话恢复、退出、刷新
- Dashboard
- 系统设置下的系统配置
- 系统设置下的管理员设置

会话管理、审计日志和高危日志接口虽然仍然存在，但当前前端仍未提供对应独立页面。
