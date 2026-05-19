# Worker、实例生命周期和续费阶段代码开工说明

本文档用于下一轮进入代码实现前快速恢复上下文。
它不是最终接口、数据库、页面或配置契约；如有冲突，以 `docs/server/api/`、`docs/server/`、`docs/admin/`、`docs/web/`、`server/migrations/` 和 `server/config.example.yaml` 为准。

## 当前结论

文档和机器契约已先完成，下一轮可以进入代码实现。

本阶段只基于当前 MCP PVE client API 已提供能力推进：

- 查询节点、节点详情、节点 VM、存储
- 创建 VM
- 查询 VM
- 启动 VM
- 停止 VM
- 删除 VM
- 查询 operation

不得实现 MCP 未提供的能力：

- 重启
- 重装
- 重置密码
- 控制台
- 快照
- 备份
- 迁移
- 监控
- 防火墙
- 资源池或通用 PVE 运维

真实支付网关不在本阶段。续费订单只做订单和实例服务期延长，支付字段为占位；后台人工确认续费，未来支付回调复用同一确认逻辑。

## 开工前必读

按顺序读取：

1. `AGENTS.md`
2. `.codex/skills/pvecloud-document-first/SKILL.md`
3. `docs/server/README.md`
4. `docs/server/architecture.md`
5. `docs/server/jobs.md`
6. `docs/server/api/endpoints.md`
7. `docs/server/database/design.md`
8. `server/migrations/037_worker_instance_lifecycle.sql`
9. `server/config.example.yaml`
10. `docs/admin/pages/async-tasks.md`
11. `docs/admin/pages/instance-management.md`
12. `docs/admin/pages/order-management.md`
13. `docs/web/pages/instances.md`
14. `docs/web/pages/orders.md`
15. `docs/web/pages/order-detail.md`

同时参考上一阶段交接：

- `docs/progress/mcp-pve-instance-handoff.md`

## 本阶段实现范围

后端：

- 新增 Worker 进程入口 `cmd/worker`。
- 补齐配置加载结构：`worker`、`instance_lifecycle`、`notification`。
- 新增异步任务领域、仓储和执行器。
- 支持任务领取、锁定、执行、重试、失败落库和人工重试。
- 支持任务类型：
  - `instance_operation_sync`
  - `instance_expiry_notice`
  - `instance_expiry_release`
  - `notification_email_send`
  - `notification_sms_placeholder`
- 新增管理端异步任务查询和失败重试接口。
- 新增用户端创建续费订单接口。
- 新增管理端确认续费订单接口。
- 新增管理端调整实例到期时间接口。
- 实例交付完成时写入 `service_started_at` 和 `expires_at`。
- Worker 自动同步未完成 operation。
- Worker 到期前投递提醒任务。
- Worker 到期后按配置投递自动释放任务。
- 自动释放只调用现有 MCP 删除 VM 能力，且受 `instance_lifecycle.auto_release_enabled` 控制。
- 邮件通知复用现有 SMTP；短信只生成占位任务和通知记录。

管理端：

- 新增 Async Tasks 页面和路由 `/async-tasks`。
- 权限使用 `page.async-tasks`、`async-task:retry`、`async-task:*`。
- 实例页面展示服务期、到期提醒、自动释放计划、因到期释放时间和续费摘要。
- 订单页面展示订单类型、支付占位状态，并支持续费订单人工确认。
- 实例页面支持后台调整到期时间。

用户端：

- 实例列表和详情展示服务开始时间、到期时间、到期状态、释放倒计时和最近续费订单摘要。
- 用户可为自己的未释放实例创建续费订单。
- 订单列表和详情展示 `purchase`、`renewal` 两类订单，以及支付占位状态。
- 不展示真实支付入口。

## 推荐实现顺序

1. 迁移和配置模型
   - 确认 `037_worker_instance_lifecycle.sql` 可执行。
   - 更新 Go config struct、默认值和校验。

2. 后端领域常量和 model
   - 补订单类型、支付状态、任务类型、任务状态、通知状态、实例生命周期字段。
   - 更新 GORM model 和 repository。

3. 续费订单闭环
   - 用户端创建续费订单。
   - 管理端确认续费。
   - 管理端调整实例到期时间。
   - 同事务更新订单、实例和审计。

4. 异步任务平台
   - 任务创建、幂等投递、领取、锁定、执行、重试。
   - 管理端任务列表和失败重试。
   - payload/result 只保存摘要，不保存敏感原文。

5. Worker 进程
   - `cmd/worker` 和 `internal/app/worker` 装配。
   - Worker 不注册 HTTP 路由。
   - 支持轮询、批量领取、锁 TTL 和优雅退出。

6. 实例 operation 自动同步
   - 交付、启动、停止、释放后投递 `instance_operation_sync`。
   - operation 未完成时延后重试，不算业务失败。

7. 实例生命周期
   - 交付完成写入服务期。
   - 到期提醒任务。
   - 到期释放任务。
   - 续费或手动调整到期后，旧释放任务必须取消或执行时跳过。

8. 通知
   - 邮件发送任务。
   - 短信占位任务。
   - 通知记录写入和失败记录。

9. 管理端页面
   - Async Tasks 页面。
   - 订单续费确认入口。
   - 实例服务期和到期调整入口。

10. 用户端页面
    - 实例服务期展示和续费入口。
    - 订单类型和支付占位状态展示。

## 重点约束

- API 进程只负责本地业务事实和任务投递，Worker 负责后台执行。
- Worker 与 API 共用 MariaDB、Redis、配置和日志配置。
- Worker 不对外暴露 HTTP API。
- 用户端不得看到任务内部错误、Worker ID、PVE 节点、VMID、operation ID 或上游原始错误。
- `async_tasks.payload`、`result`、`notifications` 不得保存 token、password、SMTP 凭据、MCP Bearer Token、支付回调敏感原文或完整上游响应。
- `async_tasks(task_type, idempotency_active_key)` 只约束未取消任务；取消任务后允许重新投递同类型同幂等键。
- 实例释放、续费确认、到期调整、任务人工重试必须写后台审计。
- 外部 MCP 调用、邮件发送不得放入长事务。
- Redis 只能做短锁、限流、缓存或短 TTL 辅助状态；最终事实以 MariaDB 为准。
- 管理端只调用 `/admin-api/*`。
- 用户端只调用 `/api/*`。
- 管理端使用 Naive UI 和 `@vicons/ionicons5`，不要新增 Element Plus。

## 重点审查点

- 续费确认的事务：订单状态、支付占位状态、`paid_at`、实例 `expires_at` 和审计必须一致。
- 续费顺延规则：未到期从原 `expires_at` 顺延；已到期从当前时间顺延。
- 到期释放规则：实例已释放、已续费或到期时间已变化时，旧释放任务不能删除 VM。
- operation 同步规则：operation 未完成、缺 operation ID 或无法确认成功时，不提前推进本地状态。
- 任务领取并发：多 Worker 不能执行同一个任务。
- 任务重试：未达到最大次数延后重试；达到最大次数进入 `failed`。
- 人工重试：只允许 `failed`，重置锁定字段并回到 `pending`。
- 邮件失败：不得泄露 SMTP 凭据，失败应可重试或可排查。
- 短信：只占位，不调用真实供应商。
- 前端展示：不出现真实支付入口，不出现 MCP 未提供的运维入口。

## 验证建议

后端：

```bash
cd server
gofmt -w .
go test ./...
```

管理端：

```bash
cd admin
bun run build
```

用户端：

```bash
cd web
bun run build
```

通用：

```bash
git diff --check
```

建议补充的手工流程：

- 新购订单交付后自动同步到实例完成。
- 用户为未释放实例创建续费订单。
- 管理端确认续费后实例到期时间顺延。
- 管理端手动调整实例到期时间。
- Worker 发送到期提醒邮件任务。
- `auto_release_enabled=false` 时到期不删除 VM。
- `auto_release_enabled=true` 且宽限期到达后释放 VM。
- 失败任务在 Async Tasks 页面可见并可重试。

## 当前不做

- 真实支付发起、支付页面、支付回调和退款。
- 钱包、余额、发票。
- 短信真实供应商。
- 用户端任务查询。
- 通用 PVE 运维。
- MCP 未提供的重启、重装、重置密码、控制台、快照、备份、迁移、监控、防火墙。

