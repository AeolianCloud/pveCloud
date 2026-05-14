# Admin 路由与权限契约

## 路由原则

- 受保护页面统一挂在后台壳下。
- 后端 `admin_permissions` 是菜单和操作权限的唯一目录来源。
- 服务端 `menus` 是当前管理员可见侧栏菜单树，管理端侧栏必须按该树渲染。
- 前端路由只作为本地页面组件白名单和兜底守卫，不再作为侧栏菜单事实来源。
- 前端不接受后端任意组件路径；只能按服务端菜单 `path` 匹配本地已注册受保护路由。

## 当前页面权限

| 页面 | 路径 | 菜单权限 |
| --- | --- | --- |
| Dashboard | `/dashboard` | `page.dashboard` |
| 系统配置 | `/system/settings` | `page.system-settings.config` |
| 管理员设置 | `/system/admin-users` | `page.system-settings.admin-users` 或 `page.system-settings.admin-roles` 或 `page.system-settings.admin-sessions` |
| 日志管理 | `/system/audit-logs` | `page.system-settings.audit-logs` |
| 文件管理 | `/files` | `page.file-management` |
| Web 用户管理 | `/web/users` | `page.web-users` 或 `page.web-user-sessions` |
| 实名管理 | `/web/real-names` | `page.real-name-management` |
| 产品管理 | `/products` | `page.products` |
| 订单管理 | `/orders` | `page.orders` |
| 工单管理 | `/tickets` | `page.tickets` |

管理员设置页面使用 `permissionMode: any` 时，只要具备管理员账号、管理组权限或管理员会话入口之一即可进入页面；页面内部能力继续按按钮或功能块权限控制。

Web 用户管理页面使用 `permissionMode: any` 时，只要具备 Web 用户账号或用户状态入口之一即可进入页面；用户状态作为第二个 tab 展示，不作为独立侧栏菜单。

## 权限码命名

- 菜单权限统一采用 `page.xxx`，也是页面访问和侧栏可见的授权节点。
- 操作权限统一采用 `resource:action`，例如 `admin-user:create`、`web-user:create`、`real-name:sync`、`product:update`。
- 支持 `resource:*` 风格的模块全权限，例如 `admin-user:*`，并在前后端权限判断中覆盖同资源细权限。
- 操作权限在权限目录中必须挂到明确菜单父节点；授权时选中操作权限会自动保留父级菜单权限。

## 权限判断

- 页面权限声明在路由 `meta.permission`，用于直接访问路由时兜底校验。
- 侧栏菜单只使用后端 `menus` 渲染。
- 按钮权限通过统一权限指令或工具函数判断。
- 页面模板中不散写 `permissionCodes.includes(...)`。
- 新页面若需要独立控制，必须先补数据库权限码。
- 新按钮、标签页或页面内功能块若需要独立显隐，必须先补对应权限码，再挂接 `meta.permission` 或 `v-permission`。
- 工单管理页面内操作权限包括 `ticket:reply`、`ticket:close`、`ticket:assign`、`ticket:collaborate`、`ticket:note`、`ticket:priority`、`ticket:tag`、`ticket:tag-manage`，均由 `ticket:*` 覆盖。

## 新增页面流程

新增 admin 页面前必须先确认：

- 页面是否属于当前管理端开放范围。
- 是否需要新增路由。
- 是否需要新增菜单权限。
- 是否需要新增操作权限。
- 是否需要新增或调整 API 契约。
- 是否会改变侧栏菜单结构。

涉及以上任意一项时，先更新对应文档并等待维护者确认，再进入实现。
