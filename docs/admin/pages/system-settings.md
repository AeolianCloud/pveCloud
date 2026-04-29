# System Settings 页面契约

## 页面定位

`System Settings` 是当前管理端开放的系统设置父级菜单。

当前承载：

- 系统配置
- 管理员设置

不承载：

- 审计日志
- 高危操作日志

## 路由结构

父级菜单：

- 路径：`/system`
- 标题：系统设置
- 作为侧栏父级菜单展示

子页面：

- 系统配置：`/system/settings`
- 管理员设置：`/system/admin-users`

当前不为系统设置继续拆更多侧栏层级。

## 系统配置

页面职责：

- 按分组展示 `system_configs` 表中的配置项。
- 支持编辑允许更新的配置项。
- `is_secret=1` 的配置值不得展示明文。

权限建议：

- 页面入口：`page.system-settings.config`
- 页面可见资源：`system-config:view` 或 `system-config:*`
- 更新：`system-config:update` 或 `system-config:*`

关联接口：

- `GET /admin-api/system-configs`
- `PATCH /admin-api/system-configs/{id}`

具体字段、响应和错误码以 `docs/server/api/` 为准。

## 管理员设置

页面职责：

- 在同一页面内承载管理员账号、管理组权限和管理员会话三块能力。
- 可以用标签页、分区或其它明确的信息架构切分三块能力。
- 不恢复 `/admin-users` 和 `/admin-roles` 独立侧栏菜单。
- 不新增管理员会话独立侧栏菜单或受保护路由。

管理员账号能力：

- 管理员列表
- 创建管理员
- 编辑管理员
- 状态切换
- 密码重置

管理组权限能力：

- 管理组列表
- 创建管理组
- 编辑管理组
- 状态切换
- 权限码分配

管理员会话能力：

- 按管理员设置页第三个 tab 展示 `admin_sessions` 对应的会话列表
- 展示会话状态、签发时间、过期时间、最近访问时间、最近访问 IP、User-Agent 等核心信息
- 区分当前会话与其它会话
- 支持吊销其它管理员会话
- 不支持从该 tab 吊销当前会话自身

权限建议：

- 管理员账号 tab 入口：`page.system-settings.admin-users`
- 管理员列表资源：`admin-user:view` 或 `admin-user:*`
- 新建管理员：`admin-user:create` 或 `admin-user:*`
- 编辑管理员与状态切换：`admin-user:update` 或 `admin-user:*`
- 重置管理员密码：`admin-user:password-reset` 或 `admin-user:*`
- 管理组权限 tab 入口：`page.system-settings.admin-roles`
- 管理组列表资源：`admin-role:view` 或 `admin-role:*`
- 新建管理组：`admin-role:create` 或 `admin-role:*`
- 编辑管理组、状态切换、权限分配：`admin-role:update` 或 `admin-role:*`
- 管理员会话 tab 入口：`page.system-settings.admin-sessions`
- 管理员会话列表资源：`admin-session:view` 或 `admin-session:*`
- 吊销管理员会话：`admin-session:revoke` 或 `admin-session:*`

关联接口：

- `GET /admin-api/admin-users`
- `POST /admin-api/admin-users`
- `GET /admin-api/admin-users/{id}`
- `PATCH /admin-api/admin-users/{id}`
- `POST /admin-api/admin-users/{id}/password`
- `GET /admin-api/admin-roles`
- `POST /admin-api/admin-roles`
- `GET /admin-api/admin-roles/{id}`
- `PATCH /admin-api/admin-roles/{id}`
- `GET /admin-api/admin-permissions`
- `GET /admin-api/admin-sessions`
- `PATCH /admin-api/admin-sessions/{session_id}`

具体字段、响应和错误码以 `docs/server/api/` 为准。

## 验收重点

- 系统设置只调用 `/admin-api/*`。
- 页面入口、标签页和按钮显隐都通过统一权限能力判断。
- 页面模板中不散写 `permissionCodes.includes(...)`。
- 敏感配置不展示明文。
- 管理员、管理组和管理员会话能力都不恢复独立侧栏菜单。
- 管理员会话 tab 需要对当前会话提供明确标识，并阻止自吊销误操作。
