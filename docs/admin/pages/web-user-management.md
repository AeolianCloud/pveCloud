# Web 用户管理页面契约

## 页面定位

Web 用户管理用于在管理端维护用户端账号，并查看用户端登录状态。

当前在同一个后台页面内承载两个 tab：

- Web 用户
- 用户状态

不承载：

- 用户端注册开放开关
- 实名申请查看与供应商结果同步、钱包、订单、实例或工单业务资料；实名申请由独立实名管理页面承载
- 用户端权限码体系

## 路由结构

页面：

- Web 用户管理：`/web/users`

`Web 用户管理` 作为管理端侧栏中的独立菜单项展示，不挂在 `System Settings` 下。
页面内使用 tab 区分“Web 用户”和“用户状态”。
`用户状态` 不做独立侧栏菜单，也不新增独立受保护路由。

本页属于复杂管理页，应采用 `index.vue + types.ts + components/` 页面容器结构：

- `index.vue`：状态、请求、权限和事件编排
- `types.ts`：页面私有状态和表单类型
- `components/WebUsersTab.vue`：Web 用户 tab
- `components/WebUserSessionsTab.vue`：用户状态 tab
- `components/WebUserEditorDialog.vue`：用户新增/编辑弹窗
- `components/WebUserPasswordDialog.vue`：重置密码弹窗

## Web 用户

页面职责：

- 分页展示 `users` 表中的用户端账号。
- 支持按关键词和状态筛选。
- 支持创建用户端账号。
- 支持编辑用户邮箱、显示名称和状态。
- 支持重置用户密码。
- 不支持删除用户。
- 不展示 `password_hash`。

权限建议：

- 页面入口：`page.web-users`
- 用户列表资源：`web-user:view` 或 `web-user:*`
- 新建用户：`web-user:create` 或 `web-user:*`
- 编辑用户和状态切换：`web-user:update` 或 `web-user:*`
- 重置用户密码：`web-user:password-reset` 或 `web-user:*`

关联接口：

- `GET /admin-api/users`
- `POST /admin-api/users`
- `GET /admin-api/users/{id}`
- `PATCH /admin-api/users/{id}`
- `POST /admin-api/users/{id}/password`

具体字段、响应和错误码以 `docs/server/api/` 为准。

## 用户状态

页面职责：

- 分页展示 `user_sessions` 对应的用户端登录会话。
- 展示用户摘要、会话状态、签发时间、过期时间、最近访问时间、最近访问 IP 和 User-Agent。
- 支持按用户、状态和时间范围筛选。
- 支持吊销 active 状态用户会话。
- 不支持从该页面直接修改用户账号资料。

权限建议：

- tab 入口：`page.web-user-sessions`
- 会话列表资源：`web-user-session:view` 或 `web-user-session:*`
- 吊销用户会话：`web-user-session:revoke` 或 `web-user-session:*`

关联接口：

- `GET /admin-api/user-sessions`
- `PATCH /admin-api/user-sessions/{session_id}`

具体字段、响应和错误码以 `docs/server/api/` 为准。

## 验收重点

- 管理端页面只调用 `/admin-api/*`。
- Web 用户账号与用户状态在同一页面内通过 tab 展示，tab 入口使用独立权限控制。
- 页面按钮显隐使用统一权限能力判断。
- 密码哈希不得出现在任何管理端响应中。
- 用户状态页吊销会话后，Web 端对应 token 后续访问必须失效。
