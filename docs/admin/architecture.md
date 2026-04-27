# Admin 前端架构

本文件描述当前管理端前端的实际契约、页面范围和结构约束。

它既不是 AI skill，也不是纯视觉规范。

## 定位

管理端面向平台运营、客服和管理员。
接口边界固定为 `/admin-api/*`。
最终接口契约以 `docs/server/api/` 为准。

## 当前实现状态

- `admin/` 是当前已存在的管理端前端实现。
- 当前前端范围收缩为三个页面：
  - `Login`
  - `Dashboard`
  - `403`
- 管理员、角色权限、登录会话、系统设置、审计日志和高危操作日志的后端接口仍然存在，但当前前端不再提供这些独立页面、菜单入口和受保护路由。

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

- 是当前唯一受保护业务页
- 需要 `dashboard:view`
- 只展示当前已开放的基础后台指标
- 不展示订单、支付、实例、工单或其它未开放业务模块指标

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
