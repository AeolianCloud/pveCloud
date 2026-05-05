# API 通用约定

本文档只描述跨接口共用的约定。具体接口字段、请求参数和响应数据请看 `docs/server/api/endpoints.md`。

## 文档职责

这里描述：

- 统一响应包裹
- 错误码分段
- 鉴权与权限约定
- 幂等与限流原则

## 路由边界

- 管理端：`/admin-api/*`
- 用户端：`/api/*`
- 健康检查：`/healthz`

## 统一响应格式

成功响应：

```json
{"code":0,"message":"成功","data":{}}
```

错误响应：

```json
{"code":40001,"message":"参数错误","data":null}
```

## 错误码分段

- `0`：成功
- `400xx`：参数或校验错误
- `401xx`：未登录、token 无效、token 过期、会话失效
- `403xx`：无权限
- `404xx`：资源不存在
- `409xx`：状态冲突、重复提交
- `429xx`：请求过于频繁、登录失败限流
- `500xx`：服务端内部错误
- `600xx`：预留的外部业务错误段
- `700xx`：预留的外部系统错误段
- `800xx`：管理端业务操作错误

## 鉴权约定

### 管理端

- 使用管理端 JWT secret 和 issuer
- 作用范围为 `/admin-api/*`
- JWT 必须带 `jti`
- `jti` 对应 `admin_sessions.session_id`
- 受保护管理端接口不仅校验 token，还要校验当前会话状态和当前数据库 RBAC

### 用户端

- 使用用户端 JWT secret 和 issuer
- 作用范围为 `/api/*`
- JWT 必须带 `jti`
- `jti` 对应 `user_sessions.session_id`
- 受保护用户端接口需要同时校验 token 和当前用户会话状态
- 当前阶段用户端不引入权限码；登录后即可访问 `/user` 控制台入口
- `GET /api/server-catalog` 是公开产品目录接口，不要求用户登录，但不得返回订单、支付、实例或 PVE 节点信息
- 用户实名接口属于用户端受保护接口，实名状态只作为业务状态，不引入用户端权限码
- 当 `real_name.required_for_order=true` 时，任何机器购买或订单创建接口都必须要求当前用户实名状态为 `approved`

## 管理端权限目录

管理端权限以 `admin_permissions` 作为唯一目录来源，采用菜单节点和操作节点一体化模型：

1. 菜单权限：控制服务端下发菜单、页面路由访问和页面主数据读取
2. 操作权限：控制进入页面后能执行哪些按钮、写操作、危险操作或敏感详情读取

### 菜单权限

菜单权限格式统一为：

```text
page.<menu>.<feature>
```

例如：

- `page.dashboard`
- `page.system-settings.config`
- `page.system-settings.admin-users`
- `page.system-settings.admin-roles`
- `page.system-settings.admin-sessions`
- `page.system-settings.audit-logs`
- `page.file-management`
- `page.web-users`
- `page.web-user-sessions`
- `page.real-name-management`
- `page.products`

菜单权限在权限目录中使用 `type=menu`。`/admin-api/auth/me`、登录恢复和 Dashboard 响应中的 `menus` 必须按当前管理员拥有的菜单权限生成。

`menus` 节点结构：

```json
{
  "key": "page.dashboard",
  "title": "控制台",
  "path": "/dashboard",
  "icon": "Odometer",
  "permission_code": "page.dashboard",
  "children": []
}
```

### 操作权限

操作权限格式统一为：

```text
resource:action
```

例如：

- `dashboard:view`
- `system-config:view`
- `system-config:update`
- `admin-user:*`
- `admin-user:password-reset`
- `admin-session:view`
- `admin-session:revoke`
- `audit-log:view`
- `audit-log:sensitive-view`
- `file:upload`
- `file:delete`
- `file:*`
- `web-user:view`
- `web-user:create`
- `web-user:update`
- `web-user:password-reset`
- `web-user-session:view`
- `web-user-session:revoke`
- `real-name:view`
- `real-name:review`
- `product:view`
- `product:create`
- `product:update`
- `product:publish`

实现要求：

- 前端侧栏菜单使用服务端 `menus`
- 前端路由、tab 显示和页面主数据读取使用菜单权限
- 前端按钮、区块、提交动作和后端写接口鉴权使用操作权限
- 如果持有 `resource:*`，则应视为同时拥有该资源全部细粒度权限
- 操作权限在权限目录中必须挂到明确父级菜单；创建或更新角色时，后端需要归一化权限集合，保留被选中操作权限的父级菜单权限

### 权限分配边界

管理端角色和管理员账号编辑属于二次授权入口，后端不能只校验“是否能编辑角色/账号”，还必须校验“是否能分配目标权限或角色”。

角色权限分配规则：

- `POST /admin-api/admin-roles` 和 `PATCH /admin-api/admin-roles/{id}` 中提交的 `permission_codes` 必须是当前操作者实时数据库权限集合的子集。
- 如果操作者拥有 `resource:*`，可以分配该资源下的细粒度权限；如果操作者只拥有细粒度权限，不能分配对应 `resource:*`。
- 菜单权限可随操作者已拥有的操作权限父级归一化保留，但不能借角色编辑新增操作者未拥有的菜单或操作权限。
- 该规则用于防止持有 `admin-role:create` 或 `admin-role:update` 的管理员把自己没有的高危权限授予角色。

管理员角色分配规则：

- `PATCH /admin-api/admin-users/{id}` 中提交的 `role_ids` 对应的全部权限，必须是当前操作者实时数据库权限集合的子集。
- 管理员不能通过编辑自己的 `role_ids` 修改自身角色集合，避免自我提权或自锁。
- 分配角色时只校验 `active` 角色还不够，后端必须展开目标角色权限后再执行上述子集校验。

内置 `super_admin` 角色仍应始终拥有当前权限目录中的全部权限；但普通角色编辑和账号编辑接口不能被用来绕过上述分配边界。

## 幂等原则

以下能力必须具备幂等保护：

- 管理端会话刷新
- 管理员密码重置
- 用户端 access token refresh
- 用户端密码重置申请和确认

幂等优先依赖：

- 业务唯一键
- 会话唯一标识
- 本地状态检查

不能只依赖前端防重复点击。

## 当前不在契约内的业务域

以下业务域仍不在当前 API 契约内：

- 用户端业务 API（公开站点配置、用户账号自助、用户实名和服务器产品目录接口除外）
- 订单
- 支付
- 实例
- 工单
- 异步任务
