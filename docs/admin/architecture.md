# Admin 前端架构

本文件描述当前管理端前端的实际契约、页面范围和结构约束。

它既不是 AI skill，也不是纯视觉规范。

## 定位

管理端面向平台运营、客服和管理员。
接口边界固定为 `/admin-api/*`。
最终接口契约以 `docs/server/api/` 为准。

## 当前实现状态

- `admin/` 是当前已存在的管理端前端实现。
- 当前前端范围包括以下页面：
  - `Login`
  - `Dashboard`
  - `System Settings`（系统设置，含系统配置、管理员设置子页面）
  - `403`
- 登录会话、审计日志和高危操作日志的后端接口仍然存在，但当前前端不再提供这些独立页面、菜单入口和受保护路由。

这条边界是当前管理端前端的有效契约。

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

## 页面范围

### Login

- 登录页先获取验证码
- 登录成功后写入 token、本地登录态和最小管理员摘要
- 登录错误留在登录页处理

### Dashboard

- 是受保护业务页
- 需要 `dashboard:view`
- 只展示当前已开放的基础后台指标
- 不展示订单、支付、实例、工单或其它未开放业务模块指标

### System Settings

- 系统设置父级菜单，包含以下子页面：

#### 系统配置

- 按分组展示和编辑 `system_configs` 表中的配置项
- 页面入口权限建议为 `page.system-settings.config`
- 页面可见资源权限建议为 `system-config:view` 或 `system-config:*`
- 更新权限建议为 `system-config:update` 或 `system-config:*`
- 接口：`GET /admin-api/system-configs`（按分组查询）、`PATCH /admin-api/system-configs/{id}`（更新）
- `is_secret=1` 的配置值不得展示明文

#### 管理员设置

- 在同一页面内承载两块能力：
  - 管理员账号列表、创建、编辑、状态切换、密码重置
  - 管理组列表、创建、编辑、状态切换、权限码分配
- 页面内可使用标签页、分区或其它明确的信息架构，但不单独新增系统设置侧栏子菜单
- 权限建议拆分如下：
  - 管理员账号 tab 入口：`page.system-settings.admin-users`
  - 页面与管理员列表资源：`admin-user:view` 或 `admin-user:*`
  - 新建管理员：`admin-user:create` 或 `admin-user:*`
  - 编辑管理员与状态切换：`admin-user:update` 或 `admin-user:*`
  - 重置管理员密码：`admin-user:password-reset` 或 `admin-user:*`
  - 管理组权限 tab 入口：`page.system-settings.admin-roles`
  - 管理组列表资源：`admin-role:view` 或 `admin-role:*`
  - 新建管理组：`admin-role:create` 或 `admin-role:*`
  - 编辑管理组、状态切换、权限分配：`admin-role:update` 或 `admin-role:*`
- 接口：
  - 管理员：`GET/POST /admin-api/admin-users`、`GET/PATCH /admin-api/admin-users/{id}`、`POST /admin-api/admin-users/{id}/password`
  - 管理组与权限：`GET/POST /admin-api/admin-roles`、`GET/PATCH /admin-api/admin-roles/{id}`、`GET /admin-api/admin-permissions`

### 403

- 作为受保护错误页保留
- 用于权限不足时的明确反馈

## 登录与会话恢复

- 首次进入或刷新后，如果本地存在 token，前端优先调用 `GET /admin-api/auth/me` 恢复登录态
- `GET /admin-api/dashboard` 只负责首页数据，不承担登录态恢复职责
- `POST /admin-api/auth/logout` 无论成功失败都要清理本地状态
- `POST /admin-api/auth/refresh` 失败按登录过期处理

## 权限系统

- 页面权限声明在路由 `meta.permission`
- 侧栏菜单由可访问路由推导，不直接以服务端 `menus` 渲染完整菜单树
- 服务端 `menus` 当前仅作为兼容快照和未来扩展信息
- 按钮权限通过统一权限指令或工具函数判断
- 页面模板中不散写 `permissionCodes.includes(...)`
- 页面入口权限建议统一采用 `page.xxx`
- 资源权限建议统一采用 `resource:action`，例如 `admin-user:view`
- 建议支持 `resource:*` 作为资源全权限，并在前后端权限判断中覆盖同资源细权限
- 新页面若需要独立控制，必须先补数据库权限码，否则页面无法进入角色权限分配列表，也无法做独立授权
- 新按钮、标签页或页面内功能块若需要独立显隐，必须先补对应权限码，再挂接 `meta.permission` 或 `v-permission`

## 样式原则

- 管理端基础控件优先使用 Element Plus
- 全局样式只承载 tokens、reset、shell 和少量共享样式
- 页面私有样式放在页面或组件内
- 不再建设替代 Element Plus 的本地工具类体系

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
