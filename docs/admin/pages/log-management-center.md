# Log Management Center 页面

## 页面范围

日志管理中心是管理端独立一级菜单，当前覆盖管理端操作审计、管理端登录安全、用户安全日志、用户业务日志、前端错误日志和后端运行日志。

当前开放页面：

- 操作审计：`/logs/admin-operations`
- 登录安全：`/logs/admin-security`
- 用户安全日志：`/logs/user-security`
- 用户业务日志：`/logs/user-business`
- 前端错误日志：`/logs/frontend-errors`
- 后端运行日志：`/logs/backend-runtime`

当前不开放：

- 用户中心自查登录记录
- 外部日志平台深度检索
- 日志归档和备份管理

## 操作审计

- 路由：`/logs/admin-operations`
- 菜单权限：`page.logs.admin-operations`
- 数据来源：`admin_audit_logs` 中非认证类后台操作审计记录
- 接口：`GET /admin-api/audit-logs`

页面能力：

- 支持按管理员、操作动作、对象类型、对象 ID 和时间范围筛选。
- 查询接口传 `log_type=admin_operation`，排除 `object_type=admin_auth` 的登录安全记录。
- 展示操作者、操作动作、对象、请求方法、请求路径、请求 ID、IP、备注和创建时间。
- `before_data`、`after_data`、`user_agent` 等敏感详情默认不展示，只有具备敏感详情权限时展示。

权限建议：

- 页面入口：`page.logs.admin-operations`
- 日志列表资源：`audit-log:view` 或 `audit-log:*`
- 敏感详情：`audit-log:sensitive-view` 或 `audit-log:*`

## 登录安全

- 路由：`/logs/admin-security`
- 菜单权限：`page.logs.admin-security`
- 数据来源：`admin_audit_logs` 中 `object_type=admin_auth` 的认证相关审计记录
- 接口：`GET /admin-api/audit-logs`

页面能力：

- 支持按管理员、认证动作类型和时间范围筛选。
- 查询接口传 `log_type=admin_security`，固定查询 `object_type=admin_auth` 的认证相关记录。
- 认证动作类型包含登录成功、登录失败、登录限流、验证码限流、退出登录和会话刷新。
- 展示字段按认证场景展示操作者、认证动作、请求方法、请求路径、请求 ID、IP、备注和创建时间。

权限建议：

- 页面入口：`page.logs.admin-security`
- 日志列表资源：`admin-security-log:view` 或 `admin-security-log:*`

## 用户安全日志

- 路由：`/logs/user-security`
- 菜单权限：`page.logs.user-security`
- 数据来源：`user_security_logs`
- 接口：`GET /admin-api/logs/user-security`

页面能力：

- 支持按用户、动作、结果、请求 ID、IP 和时间范围筛选。
- 展示用户摘要、会话 ID、动作、结果、请求、请求 ID、IP、备注和创建时间。

权限建议：

- 页面入口：`page.logs.user-security`
- 日志列表资源：`user-security-log:view` 或 `user-security-log:*`

## 用户业务日志

- 路由：`/logs/user-business`
- 菜单权限：`page.logs.user-business`
- 数据来源：`user_business_logs`
- 接口：`GET /admin-api/logs/user-business`

页面能力：

- 支持按用户、模块、动作、对象类型、对象 ID、请求 ID 和时间范围筛选。
- 展示用户摘要、模块、动作、对象、请求、请求 ID、IP、摘要和创建时间。

权限建议：

- 页面入口：`page.logs.user-business`
- 日志列表资源：`user-business-log:view` 或 `user-business-log:*`

## 前端错误日志

- 路由：`/logs/frontend-errors`
- 菜单权限：`page.logs.frontend-errors`
- 数据来源：`frontend_error_logs`
- 接口：`GET /admin-api/logs/frontend-errors`

页面能力：

- 支持按来源应用、错误类型、API 路径、HTTP 状态、请求 ID 和时间范围筛选。
- 展示来源、页面路径、错误类型、消息摘要、关联 API、浏览器摘要、请求 ID 和创建时间。

权限建议：

- 页面入口：`page.logs.frontend-errors`
- 日志列表资源：`frontend-error-log:view` 或 `frontend-error-log:*`

## 后端运行日志

- 路由：`/logs/backend-runtime`
- 菜单权限：`page.logs.backend-runtime`
- 数据来源：`backend_runtime_logs`
- 接口：`GET /admin-api/logs/backend-runtime`

页面能力：

- 支持按级别、分类、状态码、路径、请求 ID 和时间范围筛选。
- 展示级别、分类、请求、状态码、耗时、请求 ID、IP、消息和创建时间。

权限建议：

- 页面入口：`page.logs.backend-runtime`
- 日志列表资源：`backend-runtime-log:view` 或 `backend-runtime-log:*`

## 路由兼容

旧 `/system/audit-logs` 不再作为菜单入口，仅在前端保留兼容重定向到 `/logs/admin-operations`。

## 边界

- 登录安全继续复用 `admin_audit_logs` 和 `GET /admin-api/audit-logs`。
- 后端运行日志不写入 `admin_audit_logs`。
- 运行时日志、访问日志、后台操作审计和管理端日志管理边界以 `docs/server/logging.md` 为准。
