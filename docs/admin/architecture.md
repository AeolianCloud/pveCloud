# Admin 前端架构

本文档描述当前管理端前端的整体契约、职责边界和结构原则。
页面级行为、路由权限和具体功能范围拆分到 `docs/admin/pages/` 与 `docs/admin/routing-permissions.md`。

## 文档入口

- Admin 文档索引：`docs/admin/README.md`
- 页面索引：`docs/admin/pages/README.md`
- 路由与权限：`docs/admin/routing-permissions.md`
- 具体 API 契约：`docs/server/api/`

## 定位

管理端面向平台运营、客服和管理员。
接口边界固定为 `/admin-api/*`。
最终接口契约以 `docs/server/api/` 为准。

## 当前实现状态

- `admin/` 是当前已存在的管理端前端实现。
- 当前开放页面以 `docs/admin/pages/README.md` 为准。
- 当前路由和权限口径以 `docs/admin/routing-permissions.md` 为准。
- 管理员会话能力当前收敛在 `System Settings -> 管理员设置` 的第三个 tab 内，不提供独立页面、菜单入口或受保护路由。
- 审计日志和高危操作日志的后端接口仍然存在，但当前前端不提供这些独立页面、菜单入口和受保护路由。

## 技术栈

| 领域 | 选择 |
| --- | --- |
| Package/script runner | Bun |
| Build | Vite |
| UI framework | Vue 3 |
| Language | TypeScript |
| Router | Vue Router |
| State | Pinia |
| HTTP | Axios |
| Base UI | Element Plus |

## 独立边界

- `admin/` 只调用 `/admin-api/*`
- 不调用 `/api/*`
- 不导入 `web/` 的运行时代码
- 不创建跨前端共享运行时代码包

## 前后端边界配套

管理端前端的 API 消费边界，对应后端的管理端实现边界：

- 前端调用边界：`/admin-api/*`
- 后端实现边界：`server/internal/admin/*`

因此新增管理端页面、菜单、权限码、请求封装或页面行为时，应同步检查：

- `docs/admin/*`
- `docs/server/api/*`
- `docs/server/architecture.md`

不要把管理端页面契约建立在 `server/internal/web/*` 或用户端 `/api/*` 之上。

## 架构原则

- route-meta driven：标题、图标、权限、隐藏与 affix 由路由元信息承载
- layout-first：受保护页面统一挂在后台壳下
- api-by-domain：请求封装按业务域组织
- store-by-concern：认证、权限、布局状态分离
- centralized permission：路由准入、侧栏菜单和按钮权限统一围绕权限模块实现

## 目录口径

当前目标结构如下：

```text
admin/src/
  api/
  components/
  directives/
  layouts/
  plugins/
  router/
    modules/
  store/
    modules/
  styles/
  utils/
  views/
```

说明：

- `api/`：按业务域组织请求包装
- `router/modules/`：按业务域组织受保护路由
- `store/modules/`：全局 concern
- `views/`：页面入口及页面私有组件
- `layouts/`：后台壳
- `components/`：项目级复用 UI 组合

## 样式原则

- 管理端基础控件优先使用 Element Plus
- 全局样式只承载 tokens、reset、shell 和少量共享样式
- 页面私有样式放在页面或组件内
- 不建议替代 Element Plus 的本地工具类体系

## 本地开发

```powershell
cd admin
bun install
bun dev
```

构建验证：

```powershell
cd admin
bun run build
```
