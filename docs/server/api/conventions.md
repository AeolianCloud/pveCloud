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

## 管理端权限码

管理端权限采用双层模型：

1. 页面入口权限：控制菜单、路由、tab 是否显示
2. 资源操作权限：控制进入页面后能执行哪些读写操作

### 页面入口权限

页面入口权限格式统一为：

```text
page.<menu>.<feature>
```

例如：

- `page.dashboard`
- `page.system-settings.config`
- `page.system-settings.admin-users`
- `page.system-settings.admin-roles`
- `page.system-settings.admin-sessions`

### 资源操作权限

资源操作权限格式统一为：

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

实现要求：

- 前端菜单、路由和 tab 显示判断使用页面入口权限
- 前端按钮、区块、提交动作和后端接口鉴权使用资源操作权限
- 如果持有 `resource:*`，则应视为同时拥有该资源全部细粒度权限

## 幂等原则

以下能力必须具备幂等保护：

- 管理端会话刷新
- 管理员密码重置
- 高危操作审计写入

幂等优先依赖：

- 业务唯一键
- 会话唯一标识
- 本地状态检查

不能只依赖前端防重复点击。

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
