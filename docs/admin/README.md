# Admin 文档入口

`docs/admin/` 描述管理端前端的当前契约、页面范围、路由权限和实现边界。

这里是管理端文档索引。新增或修改 admin 功能时，先读本文件，再按任务读取对应细分文档。

## 必读入口

- 总体架构与边界：`docs/admin/architecture.md`
- 路由与权限：`docs/admin/routing-permissions.md`
- 页面索引：`docs/admin/pages/README.md`

## 按任务读取

新增或修改页面：

- 先读 `docs/admin/architecture.md`
- 再读 `docs/admin/routing-permissions.md`
- 再读对应 `docs/admin/pages/*.md`
- 涉及请求时读对应 `docs/server/api/`

修改登录、恢复、刷新、退出：

- 读 `docs/admin/pages/login.md`
- 读 `docs/admin/routing-permissions.md`
- 读 `docs/server/api/` 中认证相关契约

修改 Dashboard：

- 读 `docs/admin/pages/dashboard.md`
- 读 `docs/server/api/` 中 Dashboard 相关契约

修改系统设置：

- 读 `docs/admin/pages/system-settings.md`
- 读系统配置、管理员、角色、权限、管理员会话相关 API 契约

修改 403 或权限反馈：

- 读 `docs/admin/pages/403.md`
- 读 `docs/admin/routing-permissions.md`

## 维护规则

- 本目录写管理端当前事实，不写 AI 提示词。
- 页面范围、路由语义、权限判断、菜单来源、请求包装和状态语义变化，必须先更新本目录对应文档。
- 具体 API 字段和响应结构以 `docs/server/api/` 为准。
- 具体数据库结构以 `server/migrations/` 和 `docs/server/database/` 为准。
- 如果本目录文档与 `docs/progress/` 冲突，以本目录当前契约为准，再同步或归档进度文档。
