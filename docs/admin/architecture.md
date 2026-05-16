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
- 日志管理中心以独立一级菜单开放；第一阶段提供操作审计和登录安全两个页面，复用 `admin_audit_logs` 和 `GET /admin-api/audit-logs`。
- 附件管理以 `File Management` 独立页面开放，用于管理上传图片和附件。
- 产品管理以独立一级菜单开放，用于维护服务器产品、固定套餐、周期价格、销售地域和服务器系统模板；不承载订单、支付或实例操作。
- 实名管理以独立一级菜单开放，用于查看用户端个人实名申请并同步支付宝/微信侧核验结果；不承载企业实名或后台人工审核。支付宝/微信侧供应商密钥配置在系统设置的“实名设置”分组维护。
- 订单管理以独立一级菜单开放，用于查看和处理用户端订单；不支持后台创建订单，不承载支付、实例或 PVE 操作。
- 工单管理以独立一级菜单开放，用于查看和处理用户端工单；支持后台内部指派、转派、协作者、内部备注、内部 SLA 时限、标签和优先级升级；不支持后台创建用户工单，不承载支付、实例或 PVE 操作。

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
| Base UI | Naive UI |
| Icon set | @vicons/ionicons5 |

## 独立边界

- `admin/` 只调用 `/admin-api/*`
- 不调用 `/api/*`
- 不导入 `web/` 的运行时代码
- 不创建跨前端共享运行时代码包

## 安全边界

通用安全基线见 `docs/security.md`。

- 管理端前端权限判断只用于菜单、路由、按钮和区块可见性；后端 RBAC 是最终裁决。
- 管理端不得读取、保存或展示 secret、token、验证码、密码、敏感配置明文和未脱敏身份信息。
- 管理端页面不得通过前端传入的权限、角色、管理员 ID、用户 ID 或状态绕过后端校验。
- `401xx` 和 HTTP 401 进入管理端登录态恢复或退出流程；`403xx` 和 HTTP 403 展示无权限反馈，不作为未登录处理。

## 前后端边界配套

管理端前端的 API 消费边界，对应后端的管理端实现边界：

- 前端调用边界：`/admin-api/*`
- 后端实现边界：`server/internal/delivery/http/admin/*` 聚合路由，业务编排落在 `server/internal/usecase/admin/*`，GORM model 和可复用查询对象落在 `server/internal/repository/mysql/*`

因此新增管理端页面、菜单、权限码、请求封装或页面行为时，应同步检查：

- `docs/admin/*`
- `docs/server/api/*`
- `docs/server/architecture.md`

不要把管理端页面契约建立在用户端 `/api/*` 或 `server/internal/delivery/http/web/*` 之上。

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

## 页面目录结构

管理端页面按复杂度选择目录结构。

简单页面可以使用单文件入口：

```text
views/<page>/index.vue
```

适用场景：

- 单一业务实体
- 无页面内 tab，或只有轻量展示切换
- 无多个编辑弹窗
- 页面状态和表单状态较少

复杂管理页必须使用页面容器结构：

```text
views/<page>/
  index.vue
  types.ts
  components/
    <FeatureTab>.vue
    <EditorDialog>.vue
```

触发复杂管理页结构的条件：

- 页面内包含多个 tab，并且 tab 对应不同业务实体或资源
- 页面同时维护多个实体，例如账号、角色、会话，或产品、套餐、地域、系统模板
- 页面包含多个编辑、重置、关联或详情弹窗
- 页面需要多组分页、筛选、表单状态或权限判断
- 单个 `index.vue` 会同时承载请求编排、表格、弹窗和表单校验

复杂管理页职责划分：

- `index.vue`：页面容器，只负责状态、请求、权限、事件编排和组件组合
- `types.ts`：页面私有状态、表单、tab key、快照和枚举类型
- `components/*Tab.vue`：对应 tab 的查询区、表格、分页和操作按钮
- `components/*Dialog.vue`：对应新增、编辑、重置、关联、详情等弹窗表单
- 页面私有样式优先留在对应 tab 或 dialog 组件内

现有样板：

- `views/admin-users/` 是复杂管理页样板
- 新增同类页面时，优先对齐该目录结构，不应把多实体、多弹窗逻辑堆在单个 `index.vue`

## 样式原则

- 管理端基础控件统一使用 Naive UI（`naive-ui`），图标统一使用 `@vicons/ionicons5`
- 不再使用 Element Plus 与 `@element-plus/icons-vue`，新增页面/组件不得引入 `el-*` 组件或相关样式
- 不并行引入第二套 UI 框架；通过 Naive UI `n-config-provider` 注入主题与本地化
- 全局样式只承载 tokens、reset、shell 和少量共享样式
- 页面私有样式放在页面或组件内
- 不建设替代 Naive UI 的本地工具类体系

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
