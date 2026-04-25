# Admin 前端

管理后台面向平台运营、客服和管理员工作流。接口边界是 `/admin-api/*`，最终接口契约以 `docs/server/api/openapi.yaml` 为准。

## 技术栈

| 领域 | 选择 |
| --- | --- |
| Package/script runner | Bun |
| Build | Vite |
| UI framework | Vue 3 composition API |
| Language | TypeScript |
| Router | Vue Router |
| State | Pinia |
| HTTP | Axios |
| Icons | lucide-vue-next |

## 独立边界

- `admin/` 只调用 `/admin-api/*`。
- `admin/` 不调用 `/api/*`。
- `admin/` 不导入 `web/` 的页面、组件、请求、状态、类型、常量或工具。
- 不创建公共前端 `shared/` 包。

## 页面范围

- Login
- Dashboard
- Users
- Products / plans / regions / nodes / images / prices
- Orders
- Payments / wallet flows / manual credit
- Instances
- Tickets
- Admin users / roles / permissions
- System settings and audit logs

## 状态设计

- `auth`：管理员 token、管理员资料、角色 ID、权限码、会话摘要。
- `permission`：权限码集合和菜单可见性。
- `layout`：侧边栏、主题、折叠状态。
- `tabs`：可选的管理端多标签。

## 登录

```text
POST /admin-api/auth/login
```

登录成功后返回管理端 JWT、管理员摘要、角色 ID、权限码和会话摘要。前端把 `access_token` 写入 auth store 和 `localStorage`，后续请求发送：

```text
Authorization: Bearer <access_token>
```

前端可用 `permission_codes` 控制页面、菜单和按钮显示，但后端 RBAC 仍是最终权限边界。

登录成功响应还会返回 `session` 摘要，包含 `session_id`、`issued_at` 和 `expires_at`。前端不直接操作 `session_id`；服务端通过 JWT `jti` 和 `admin_sessions` 判断当前 token 是否仍有效。

## 登录会话

```text
GET /admin-api/auth/me
POST /admin-api/auth/logout
POST /admin-api/auth/refresh
```

- 应用启动或页面刷新后，如果本地存在 `access_token`，先调用 `GET /admin-api/auth/me` 恢复管理员资料、角色、权限和菜单；接口返回 `401` 或 `401xx` 时清理 auth store 和 `localStorage`。
- 已登录判断以 auth store 中 token 和 `GET /admin-api/auth/me` 成功结果为准，不只信任 localStorage 快照。
- 退出登录时先调用 `POST /admin-api/auth/logout`，无论接口是否成功都清理 auth store 和 `localStorage`，再跳转 `/login`。
- token 临近过期时可以调用 `POST /admin-api/auth/refresh` 获取新 token；刷新成功后替换本地 token、管理员资料、权限和会话摘要。
- `POST /admin-api/auth/refresh` 失败按登录过期处理。
- 登录表单允许提交 6 到 72 位密码；登录失败次数过多时，登录页展示后端返回的限流错误消息，不自动重定向。
- 同一管理员在其它设备的会话不受当前退出登录影响。

## 路由守卫

- `/login` 公开访问，已登录管理员访问时跳转 `/dashboard`。
- 管理端业务路由使用 `meta.requiresAuth=true`。
- 页面可声明 `meta.permissionCode`。
- 未登录管理员跳转 `/login?redirect=<target>`。
- 权限码缺失时阻止进入页面并显示无权限状态。
- 第一阶段受保护首页是 `/dashboard`，需要 `dashboard:view`。

## Axios 处理

- 请求拦截器为 `/admin-api/*` 请求附加 `Authorization: Bearer <access_token>`。
- 受保护请求遇到 HTTP `401` 或响应包裹 `401xx` 时，清理 auth store 和 `localStorage`，并跳转 `/login?redirect=<current>`。
- 登录请求自身的账号密码错误展示在登录页面，不触发自动退出重定向。
- HTTP `403` 或响应包裹 `403xx` 展示为权限错误。
- `GET /admin-api/auth/me` 是恢复登录态的首选接口；Dashboard 只负责首页指标，不再承担会话恢复职责。
- `POST /admin-api/auth/logout` 的失败不阻塞本地退出，避免用户被坏 token 困在后台。
- `POST /admin-api/auth/refresh` 失败按登录过期处理。

## Dashboard

```text
GET /admin-api/dashboard
```

Dashboard 在路由守卫通过后调用，后端仍校验 JWT、会话状态和 `dashboard:view`。页面使用返回的管理员摘要、角色 ID、权限码、可见菜单和概览指标渲染初始管理工作台。

## 本地开发

```powershell
cd admin
bun install
bun dev
```

Vite dev server 使用端口 `5174`，并把 `/admin-api/*` 代理到 `http://127.0.0.1:8080`。
