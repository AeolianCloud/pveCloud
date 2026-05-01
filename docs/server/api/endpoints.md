# API 接口总览

本文档维护当前已确认的接口清单与主要契约口径。跨接口通用约定见 `docs/server/api/conventions.md`。

## 实现边界提示

接口契约按访问边界区分：

- `/admin-api/*`：对应管理端后端实现边界 `server/internal/admin/*`

这里描述的是 API 契约，不直接替代具体代码结构；但当接口重新开放、迁移或新增时，路由注册、权限校验和实现目录应与上述边界保持一致。

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
- 菜单权限：`page.system-settings.admin-users`
- 作用：分页查询管理员账号

### `POST /admin-api/admin-users`

- 鉴权：管理端 Bearer Token
- 操作权限：`admin-user:create` 或 `admin-user:*`
- 作用：创建管理员账号

### `GET /admin-api/admin-users/{id}`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.admin-users`
- 作用：查看管理员详情

### `PATCH /admin-api/admin-users/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`admin-user:update` 或 `admin-user:*`
- 作用：更新管理员信息、状态和角色

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

### `GET /admin-api/admin-roles/{id}`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.admin-roles`
- 作用：查看角色详情

### `PATCH /admin-api/admin-roles/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`admin-role:update` 或 `admin-role:*`
- 作用：更新角色信息、状态和权限

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

### `PATCH /admin-api/system-configs/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`system-config:update` 或 `system-config:*`
- 作用：更新系统配置

## 日志管理域

### `GET /admin-api/audit-logs`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.audit-logs`
- 敏感详情权限：`audit-log:sensitive-view` 或 `audit-log:*`
- 作用：分页查询普通后台审计日志，可用于日志管理页面的操作日志 tab 和登录日志 tab
- 查询参数支持：`page`、`per_page`、`admin_id`、`action`、`object_type`、`object_id`、`date_from`、`date_to`
- 成功数据包含：
  - `list`
  - `total`
  - `page`
  - `per_page`
  - `last_page`

列表项包含操作者摘要、会话 ID、请求 ID、请求方法、请求路径、操作动作、对象类型、对象 ID、IP、备注和创建时间。
未具备敏感详情权限时，`before_data`、`after_data` 和 `user_agent` 不返回。

登录日志 tab 不新增独立接口或表，使用本接口并固定 `object_type=admin_auth` 查询认证相关日志；如需按动作类型筛选，继续使用单个 `action` 查询参数。

## 暂未开放的管理域

密码、token、secret、验证码和敏感配置明文不得出现在任何接口响应中。

## 当前不在契约内的业务域

以下业务域已经从当前 API 契约中移除：

- 用户端 API
- 用户端账号
- 产品
- 订单
- 支付
- 实例
- 工单
- 异步任务
