# Admin UI 框架迁移：Element Plus → Naive UI

本文档定义管理端从 Element Plus 全量替换为 Naive UI 的范围、策略与验收口径。
属于一次性 UI 框架变更，不变更任何 `/admin-api/*` 契约、权限码、菜单来源、数据库结构或业务流程。

## 背景

- 原契约：`Base UI = Element Plus`（见 `docs/admin/architecture.md` 历史版本与 CLAUDE.md 旧守则）。
- 现契约：`Base UI = Naive UI`，图标统一为 `@vicons/ionicons5`。
- 仅影响管理端 `admin/`；用户端 `web/` 不在本次范围内。

## 决策记录

维护者于 2026-05-10 确认：

1. 未提交改动：先由维护者 stash/提交后再开始重构（方案 A）。
2. 替换范围：借机重做 admin 视觉体系（布局、主题、图标、表单范式）（方案 B）。
3. 图标体系：统一使用 `@vicons/ionicons5`（方案 A）。
4. 过渡策略：一次性全量替换，不允许双 UI 框架共存（方案 A）。
5. 规则改写：CLAUDE.md / AGENTS.md 直接改为"管理端使用 Naive UI"（方案 A）。

## 范围（In-scope）

- `admin/package.json`：移除 `element-plus`、`@element-plus/icons-vue`，新增 `naive-ui`、`@vicons/ionicons5`。
- `admin/src/main.ts`、`plugins/`、`styles/`：移除 Element Plus 注册、主题与样式入口；新增 Naive UI `n-config-provider` 与全局 message/dialog/notification provider。
- `admin/src/layouts/`：后台壳（侧栏、顶栏、面包屑、tab、用户菜单）使用 Naive UI 重构。
- `admin/src/views/**`：所有页面 `el-*` 组件 → `n-*` 组件，包括但不限于：
  - `login/`
  - `dashboard/`
  - `403/`
  - `admin-users/`（含 `AdminUsersTab` / `AdminRolesTab` / `AdminSessionsTab` / `RoleEditorDialog`）
  - `system-settings/`
  - `file-management/`
  - `product-management/`
  - `real-name-management/`
  - `orders/`（注意当前未提交，待维护者先提交后纳入）
- `admin/src/components/`、`directives/`、`utils/`：所有依赖 Element Plus 的复用组件、消息封装、表单工具一并替换。
- 路由 meta 中的图标 key：迁移到 `@vicons/ionicons5` 对应名称。
- `docs/admin/architecture.md`、`docs/admin/pages/*`、`docs/admin/routing-permissions.md`：同步描述、示例和组件名。

## 不在范围（Out-of-scope）

- `/admin-api/*` 任何接口、权限码、菜单来源、错误码语义。
- 数据库迁移、配置示例、后端业务流程。
- 用户端 `web/` 与 `docs/web/`。
- Phase Basic Admin 当前开放页面范围（仍以 `docs/admin/pages/README.md` 为准）。

## 替换映射（核心约定）

| 旧 (Element Plus) | 新 (Naive UI) |
| --- | --- |
| `el-button` | `n-button` |
| `el-input` / `el-input-number` | `n-input` / `n-input-number` |
| `el-select` / `el-option` | `n-select`（`options` prop） |
| `el-form` / `el-form-item` | `n-form` / `n-form-item` |
| `el-table` / `el-table-column` | `n-data-table`（`columns` 配置式） |
| `el-pagination` | `n-pagination` |
| `el-dialog` | `n-modal`（preset=card 或 dialog） |
| `el-drawer` | `n-drawer` + `n-drawer-content` |
| `el-tabs` / `el-tab-pane` | `n-tabs` / `n-tab-pane` |
| `el-message` / `ElMessage` | `useMessage()`（来自 `n-message-provider`） |
| `el-message-box` | `useDialog()` |
| `el-notification` | `useNotification()` |
| `el-menu` | `n-menu`（`options` 配置式） |
| `el-tag` | `n-tag` |
| `el-switch` | `n-switch` |
| `el-tooltip` | `n-tooltip` |
| `el-icon` + `@element-plus/icons-vue` | `n-icon` + `@vicons/ionicons5` |

`n-data-table` 与 `n-menu` 改为 columns/options 配置式，所有原 slot/template-style 表格列、菜单项需要重写为渲染函数或 `render` 选项，详见各页面 PR。

## 主题与本地化

- 全局通过 `n-config-provider` 注入主题、语言（`zhCN` + `dateZhCN`）、组件级主题覆盖。
- 全局 provider 顺序：`n-config-provider > n-loading-bar-provider > n-dialog-provider > n-notification-provider > n-message-provider > App`。
- 业务代码统一通过 `useMessage()` / `useDialog()` / `useNotification()` 获取，不再直接 import 静态实例。

## 风险与回滚

- 风险：`n-data-table` 配置式与原 `el-table` slot 写法差异较大，分页/筛选/选择/展开行需要逐页核对。
- 风险：表单校验规则签名差异（Element 的 `Arrayable<FormItemRule>` vs Naive 的 `FormItemRule`）；统一在迁移时将 `validator` 改写为 Naive 形式。
- 风险：`useMessage` 等 hook 必须在 provider 子树内调用，登录页等独立页要确保挂在 provider 之下。
- 回滚：本次为一次性切换，不保留 Element 入口；如需回滚，回退到本次 PR 之前的 commit。

## 验收

- `cd admin && bun run build` 通过。
- 当前开放页面（Login、Dashboard、403、System Settings、File Management、Product Management、Real-name Management、Admin Users、Orders）功能不回退，可登录、可看菜单、可进入受保护页面。
- 仓库内不再存在 `element-plus`、`@element-plus/icons-vue`、`el-*` 组件、`ElMessage`/`ElMessageBox`/`ElNotification` 调用。
- `docs/admin/architecture.md` 与 `CLAUDE.md`、`AGENTS.md` 中关于 UI 框架的描述一致指向 Naive UI。

## 执行步骤（实现阶段）

> 文档先行已完成，下面是确认后才执行的实现顺序，写在这里仅作备忘。

1. 维护者先提交/stash 当前未提交的 admin/web/orders 改动。
2. 改 `admin/package.json` 与依赖锁，跑 `bun install`。
3. 改入口 `main.ts` 与 provider 链，提供全局主题与 message/dialog hook。
4. 改后台壳 `layouts/` 与路由 meta 图标。
5. 按页面逐个替换：Login → Dashboard → 403 → Admin Users → System Settings → File Management → Product Management → Real-name Management → Orders。
6. 删除遗留的 Element Plus 样式、图标、消息封装。
7. `bun run build` 与人工冒烟。
8. 同步更新 `docs/admin/pages/*` 中残留的 Element 词汇与示例。
