# MCP PVE 实例交付阶段交接

本文档用于下个会话复盘和审查本轮实现是否正确。
它不是最终接口、数据库或页面契约；如有冲突，以 `docs/server/api/`、`docs/admin/`、`docs/web/`、`server/migrations/` 和 `server/config.example.yaml` 为准。

> 续接说明：本文记录的是 MCP PVE 实例交付阶段实现复盘。后续“Worker、实例生命周期、续费订单和通知”已进入新的文档契约确认阶段；审查当前范围时应以 owner docs 和最新迁移为准，不再把本文中的“未实现自动轮询/自动重试”理解为长期产品边界。

## 背景

本轮目标是基于维护者提供的 MCP PVE client API，先把当前 API 文档中已经存在的能力接入 pveCloud。

确认后的边界：

- MCP PVE client API 是内部上游接口，不直接暴露给用户端。
- pveCloud 自己暴露管理端 `/admin-api/*` 和用户端 `/api/*` 业务接口。
- 只接 MCP 已有能力：节点只读、存储只读、VM 创建、VM 查询、VM 启动、VM 停止、VM 删除、operation 查询。
- 不实现 MCP 未提供的能力：重启、重装、重置密码、控制台、快照、备份、迁移、监控、防火墙、资源池、通用 PVE 运维。
- 外部平台或内部平台负责的能力不写入本系统流程。

## 契约和文档变更

新增或更新的主要契约：

- `docs/server/api/endpoints.md`
  - 新增管理端交付映射、MCP 只读资源、实例、订单交付接口。
  - 新增用户端实例列表、详情、启动、停止接口。
- `docs/server/database/design.md`
  - 新增实例交付映射、实例、实例操作记录设计说明。
  - 订单状态加入 `provisioning`、`fulfilled`。
- `server/migrations/036_instance_mcp_pve.sql`
  - 修改订单状态列注释。
  - 新增 `instance_provision_mappings`、`instances`、`instance_operations`。
  - 新增实例管理权限目录。
- `server/config.example.yaml`
  - 新增 `mcp_pve` 配置。
- `docs/admin/pages/instance-management.md`
  - 定义管理端实例管理页面范围和验收。
- `docs/web/pages/instances.md`
  - 定义用户端实例页面范围和验收。
- `docs/admin/architecture.md`、`docs/web/architecture.md`
  - 将实例管理纳入当前开放页面范围。
- `docs/admin/routing-permissions.md`
  - 新增 `page.instances` 和 `instance:*` 相关权限。

## 后端实现

配置与外部适配：

- `server/internal/platform/config/config.go`
  - 新增 `MCPPVEConfig`。
  - `mcp_pve.enabled=true` 时校验 `base_url` 和 `timeout_seconds`。
- `server/internal/integration/mcppve/client.go`
  - 封装上游 MCP PVE client API。
  - 支持 Bearer token、超时、202 响应头中的 `Location` 和 `Operation-Location`。
  - 对上游错误做适配，面向前端的错误文案已改为中文。

领域与存储：

- `server/internal/domain/instance/policy.go`
  - 定义实例状态：`creating`、`running`、`stopped`、`error`、`releasing`、`released`。
  - 定义操作：`provision`、`start`、`stop`、`release`、`sync`。
  - 定义可启动、可停止、可释放状态判断。
- `server/internal/domain/order/policy.go`
  - 新增订单状态 `provisioning`、`fulfilled`。
  - `pending` 可交付。
  - `fulfilled` 可关闭。
- `server/internal/repository/mysql/instance/`
  - 新增交付映射、实例、实例操作记录 GORM model 和 repository。

管理端用例与接口：

- `server/internal/usecase/admin/instance/service.go`
  - 交付映射列表、新增、编辑。
  - MCP 节点、节点详情、节点 VM、存储只读读取。
  - 从订单触发实例交付。
  - 实例列表、详情、启动、停止、释放、同步。
  - 操作审计：交付、映射维护、实例操作、同步。
  - MCP 禁用时在产生本地副作用前返回不可用。
- `server/internal/delivery/http/admin/instance/handler.go`
  - 管理端实例 handler。
- `server/internal/delivery/http/admin/routes/routes.go`
  - 新增管理端实例相关路由。
- `server/internal/app/api/app.go`、`server/internal/app/api/routes.go`
  - 初始化 MCP PVE client 并注入 route set。

用户端用例与接口：

- `server/internal/usecase/web/instance/service.go`
  - 当前用户自己的实例列表、详情、启动、停止。
  - 只通过当前登录用户 ID 做资源归属校验。
  - 不返回上游节点、存储、磁盘来源、VMID、operation ID 或管理端错误详情。
- `server/internal/delivery/http/web/instance/handler.go`
  - 用户端实例 handler。
- `server/internal/delivery/http/web/routes/routes.go`
  - 新增用户端实例路由。

通用错误：

- `server/internal/shared/errors/errors.go`
  - 新增 `ErrExternalUnavailable = 70002`。

## 前端实现

管理端：

- `admin/src/api/instance.ts`
  - 新增实例、交付映射、MCP 只读资源、订单交付请求封装。
- `admin/src/router/constants.ts`
  - 新增 `/instances` 路由常量。
- `admin/src/router/view-routes.ts`
  - 新增实例管理页面路由，权限 `page.instances`。
- `admin/src/layouts/components/AppSidebar.vue`
  - 新增 `Server` 图标映射。
- `admin/src/views/instances/`
  - 新增复杂页面结构：
    - `index.vue`
    - `types.ts`
    - `components/InstancesTab.vue`
    - `components/ProvisionMappingsTab.vue`
    - `components/McpResourcesTab.vue`
  - 页面提供实例列表/详情、开机、关机、释放、同步、交付映射维护、MCP 只读资源查看。
- `admin/src/views/orders/index.vue`
  - 新增 `instance:provision` 下的“触发交付”按钮。
  - 订单状态展示新增 `provisioning` 和 `fulfilled`。
- `admin/src/api/order.ts`
  - 订单状态类型新增 `provisioning` 和 `fulfilled`。

用户端：

- `web/src/api/instance.ts`
  - 新增用户端实例请求封装。
- `web/src/router/routes.ts`
  - 新增 `/user/instances` 和 `/user/instances/:instanceNo`。
- `web/src/views/instances/`
  - 新增实例列表和实例详情。
  - 用户端只展示订单、状态、产品/套餐/地域/系统模板、规格、操作摘要。
  - 不展示 PVE 节点、存储、磁盘来源、VMID 或 operation ID。
- `web/src/views/user-center/index.vue`
  - 用户中心新增实例管理入口。
- `web/src/views/orders/index.vue`、`web/src/views/orders/detail.vue`
  - 订单状态展示新增 `provisioning` 和 `fulfilled`。
- `web/src/components/AppConfirmDialog.vue`
  - 将确认弹窗可见标签改为中文。
- `web/src/api/order.ts`
  - 订单状态类型新增 `provisioning` 和 `fulfilled`。

## 当前流程

交付流程：

1. 用户创建订单后，订单为 `pending`。
2. 管理员在订单页点击“触发交付”，或直接调用 `POST /admin-api/orders/{order_no}/provision`。
3. 后端在本地事务内：
   - 锁定订单。
   - 匹配 active 交付映射。
   - 分配 `next_vmid`。
   - 创建 `instances` 记录，状态 `creating`。
   - 创建 `instance_operations` 记录，操作 `provision`，状态 `running`。
   - 将订单置为 `provisioning`。
   - 写入管理端审计。
4. 本地事务提交后调用 MCP 创建 VM。
5. MCP 调用失败时，本地实例进入 `error`，operation 进入 `failed`。
6. MCP 调用成功后保存上游资源位置和 operation 位置。
7. 管理员点击实例同步时，优先查询未完成的上游 operation；成功后查询 VM 状态。
8. 如果本地实例从 `creating` 同步到 `running` 或 `stopped`，订单推进到 `fulfilled`。

实例操作流程：

- 管理端：
  - `stopped` 可开机。
  - `running` 可关机。
  - 非 `releasing/released` 可释放。
  - 任意实例可同步。
- 用户端：
  - 只能操作当前登录用户自己的实例。
  - `stopped` 可启动。
  - `running` 可停止。

## 已知取舍和重点审查点

需要重点审查：

- 交付幂等：当前同一订单已有实例时返回冲突，不重复创建上游 VM。
- 外部副作用边界：创建 VM、启动、停止、删除都在本地事务提交后调用。
- 失败恢复：MCP 调用失败会写入本地失败状态，但没有补偿删除上游资源。
- 同步语义：订单从 `provisioning` 到 `fulfilled` 依赖管理端手动同步。
- 释放语义：释放提交后本地先进入 `releasing`，同步时置为 `released`。
- MCP 只读资源：管理端只返回经过服务端筛选包装的节点、存储和 VM 基础字段，当前仅用于后台配置和排障，用户端不开放。
- 用户端信息隐藏：用户端 DTO 和页面不展示节点、VMID、operation ID 或上游错误详情。
- 错误文案：返回给前端的错误 message 已改为中文；机器字段、状态枚举、权限码、error_code 仍为英文契约。
- 测试覆盖：本轮未新增长期 `*_test.go`；仅跑现有全量构建和测试。

当前没有实现：

- 自动后台轮询同步；后续由 Worker 异步任务契约接管。
- 自动重试；后续由 Worker 异步任务契约接管。
- 上游成功、本地后续更新失败的补偿。
- 初始密码、重置密码、控制台、重装、重启、快照、备份、迁移、监控。
- 支付、钱包、发票。

## 验证记录

本轮已执行并通过：

```bash
cd server && go test ./...
cd admin && bun run build
cd web && bun run build
git diff --check
```

构建备注：

- `admin` 构建有 Vite chunk size warning，属于现有大包警告，不是本轮编译失败。
- 构建产物 `dist/` 未出现在 `git status --short` 中。

## 下个会话建议审查顺序

1. 先读 `AGENTS.md` 和 `.codex/skills/pvecloud-document-first/SKILL.md`。
2. 读本文件恢复上下文。
3. 对照权威契约：
   - `docs/server/api/endpoints.md`
   - `docs/server/database/design.md`
   - `server/migrations/036_instance_mcp_pve.sql`
   - `server/config.example.yaml`
   - `docs/admin/pages/instance-management.md`
   - `docs/web/pages/instances.md`
4. 审查后端：
   - `server/internal/integration/mcppve/client.go`
   - `server/internal/usecase/admin/instance/service.go`
   - `server/internal/usecase/web/instance/service.go`
   - `server/internal/repository/mysql/instance/repository.go`
   - `server/internal/delivery/http/admin/routes/routes.go`
   - `server/internal/delivery/http/web/routes/routes.go`
5. 审查前端：
   - `admin/src/api/instance.ts`
   - `admin/src/views/instances/`
   - `admin/src/views/orders/index.vue`
   - `web/src/api/instance.ts`
   - `web/src/views/instances/`
   - `web/src/router/routes.ts`
6. 复跑验证命令。

## 工作区注意事项

本轮交接时工作区仍有未提交改动。

已知非本轮业务实现重点：

- `.vscode/settings.json` 是已有本地改动，不应被误纳入审查结论。
- `server/config.yaml` 是本地运行配置文件，当前显示为已修改；审查时不要把真实本地配置内容当作公共契约，公共配置契约看 `server/config.example.yaml`。

不要在未确认前执行 `git add`、`git commit`、`git reset` 或回滚这些本地改动。
