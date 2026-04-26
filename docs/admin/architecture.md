# Admin 前端

管理后台面向平台运营、客服和管理员工作流。接口边界是 `/admin-api/*`，最终接口契约以 `docs/server/api/` 和对应后端业务文档为准。

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
| Base UI components | PrimeVue + admin-owned wrapper components when needed |
| Component theme | PrimeVue Styled Mode with official preset theme |
| Utility CSS | Optional PrimeFlex for layout utilities |

## 独立边界

- `admin/` 只调用 `/admin-api/*`。
- `admin/` 不调用 `/api/*`。
- `admin/` 不导入 `web/` 的页面、组件、请求、状态、类型、常量或工具。
- 不创建公共前端 `shared/` 包。

## 样式组织

- `src/style.css` 只承载全局设计变量、基础 reset、应用外壳布局和跨页面复用工具类。
- 页面或组件私有样式写在对应 Vue SFC 的 `<style scoped>` 中，避免把页面级 class 长期堆进全局 CSS。
- 需要在管理端多个页面复用的样式，只能在 `admin/` 内部抽取，不与 `web/` 共用样式包。
- 主题相关颜色、边框、阴影和交互态使用语义化 CSS 变量；页面局部变量可定义在页面根 class 上。
- 管理端基础组件采用 PrimeVue。优先直接使用 PrimeVue 的稳定组件能力，只有当 pveCloud 需要统一业务语义、固定组合结构或屏蔽重复配置时，才在 `admin/src/components/` 内部进行二次封装。
- 管理端控件视觉采用 PrimeVue Styled Mode 和官方 preset 主题作为基础，不以项目自写 CSS 变量系统作为主要控件样式来源。项目样式只负责 pveCloud 特有的布局修正、品牌细节、页面级组合和必要覆盖。
- 布局辅助可以使用 PrimeFlex。PrimeFlex 只负责 flex、grid、间距、对齐和响应式等布局工具，不替代 PrimeVue 主题，也不用于重写 PrimeVue 控件状态。
- 成熟第三方库可以用于标准化、易出错或已有稳定实现的前端能力，例如弹窗、选择器、日期、表格交互、虚拟滚动、图表和复杂可访问性组件。不要手写这类底层行为来重复造轮子。

## 页面范围

当前阶段只完善基础后台管理能力，不开放产品套餐、订单、支付、实例、工单等业务模块。

基础后台页面范围包含 Login、Dashboard、Admin users、Roles、Permissions、Admin sessions、System settings、Audit logs、Risk logs 和 403 无权访问页。业务模块在对应功能完成前不出现在侧边栏菜单和受保护业务路由中。

除 Login 外，管理后台受保护页面统一按 PrimeVue 套件逐步改造。Dashboard、管理员、角色权限、登录会话、系统设置、审计日志、高危操作日志和 403 页面应优先使用 PrimeVue 组件、PrimeVue Styled Mode 主题和必要的管理端封装组件承载 UI，不继续扩散手写控件样式。

基础后台菜单按后端 `data.menus` 渲染。第一阶段可开放的菜单为 Dashboard、管理员、角色权限、登录会话、系统设置、审计日志、高危操作日志。菜单可见性仍由后端 RBAC 返回结果决定，前端 fallback 菜单只作为恢复登录态之前的兜底。

## 状态设计

- `auth`：管理员 token、管理员资料、角色 ID、权限码、会话摘要。
- `permission`：权限码集合和菜单可见性。
- `layout`：侧边栏、主题、折叠状态。
- `tabs`：可选的管理端多标签。

## 登录

```text
GET /admin-api/auth/captcha
POST /admin-api/auth/login
```

登录页加载时先调用 `GET /admin-api/auth/captcha` 获取验证码图片、验证码标识和有效期；提交登录时把管理员账号、密码、`captcha_id` 和 `captcha_code` 一起发送到 `POST /admin-api/auth/login`。验证码错误、过期或登录失败时，登录页展示后端返回的错误消息并重新获取验证码。

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
- 登录表单必须提交有效验证码；验证码不作为前端本地校验逻辑，最终以后端校验结果为准。
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

## 基础后台管理页

管理员页面调用 `/admin-api/admin-users`，支持列表、创建、编辑、禁用、启用、重置密码和分配角色。超级管理员账号可以查看，但不能被自己禁用或删除。

角色权限页面调用 `/admin-api/admin-roles` 和 `/admin-api/admin-permissions`，支持角色列表、创建、编辑、启用、禁用和分配权限。权限码由系统维护，前端只读展示，不提供新增或删除权限码入口。

登录会话页面调用 `/admin-api/admin-sessions`，支持查看会话列表和吊销指定活跃会话。当前会话不能通过列表操作吊销，退出登录继续使用 `/admin-api/auth/logout`。

系统设置页面调用 `/admin-api/system-configs`，支持按分组查看和更新配置。敏感配置不返回明文，只显示是否已设置；更新敏感配置时只提交新值。

审计日志页面调用 `/admin-api/audit-logs`，只读展示后台登录、退出、刷新、管理员、角色、系统设置和会话吊销等普通操作事实。审计日志不提供编辑和删除入口。

高危操作日志页面调用 `/admin-api/risk-logs`，只读展示重置管理员密码、禁用管理员、修改管理员角色、修改角色权限、吊销他人会话、修改敏感配置、登录失败触发限制等高危行为。高危操作日志不提供编辑和删除入口。高危行为应同时写入普通审计日志和高危操作日志，普通行为只写入普通审计日志。

所有基础后台列表默认使用 `page` 和 `per_page` 分页参数，筛选条件写入 URL query。页面必须提供加载态、空状态、错误重试、提交中禁用、删除或吊销类操作二次确认和后端错误消息展示。

## 本地开发

```powershell
cd admin
bun install
bun dev
```

Vite dev server 使用端口 `5174`，并把 `/admin-api/*` 代理到 `http://127.0.0.1:8080`。
