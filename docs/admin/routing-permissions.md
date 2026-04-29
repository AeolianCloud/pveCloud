# Admin 路由与权限契约

## 路由原则

- 受保护页面统一挂在后台壳下。
- 路由元信息承担标题、图标、权限、隐藏与 affix 等页面契约。
- 侧栏菜单由可访问路由推导。
- 不直接以服务端 `menus` 渲染完整菜单树。
- 服务端 `menus` 当前仅作为兼容快照和未来扩展信息。

## 当前页面权限

| 页面 | 路径 | 权限 |
| --- | --- | --- |
| Dashboard | `/dashboard` | `page.dashboard` |
| 系统配置 | `/system/settings` | `page.system-settings.config` |
| 管理员设置 | `/system/admin-users` | `page.system-settings.admin-users` 或 `page.system-settings.admin-roles` 或 `page.system-settings.admin-sessions` |

管理员设置页面使用 `permissionMode: any` 时，只要具备管理员账号、管理组权限或管理员会话入口之一即可进入页面；页面内部能力继续按按钮或功能块权限控制。

## 权限码命名

- 页面入口权限建议统一采用 `page.xxx`。
- 资源权限建议统一采用 `resource:action`，例如 `admin-user:view`。
- 建议支持 `resource:*` 作为资源全权限，并在前后端权限判断中覆盖同资源细权限。

## 权限判断

- 页面权限声明在路由 `meta.permission`。
- 按钮权限通过统一权限指令或工具函数判断。
- 页面模板中不散写 `permissionCodes.includes(...)`。
- 新页面若需要独立控制，必须先补数据库权限码。
- 新按钮、标签页或页面内功能块若需要独立显隐，必须先补对应权限码，再挂接 `meta.permission` 或 `v-permission`。

## 新增页面流程

新增 admin 页面前必须先确认：

- 页面是否属于当前管理端开放范围。
- 是否需要新增路由。
- 是否需要新增页面入口权限。
- 是否需要新增资源操作权限。
- 是否需要新增或调整 API 契约。
- 是否会改变侧栏菜单结构。

涉及以上任意一项时，先更新对应文档并等待维护者确认，再进入实现。
