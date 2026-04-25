# AI 协作入口

本文件是 AI 助手进入 pveCloud 仓库后的第一入口。它只做总导航、硬规则和文档优先流程说明；具体设计细节放在 `docs/` 对应目录里。

## 工作原则

先写文档，再写代码。

任何实现、迁移、接口、页面或业务流程变更，都必须先找到对应文档，补齐设计和验收点。文档确认没有问题后，再进入代码实现。不要直接凭感觉改代码结构、接口字段或数据库表。

## 必读顺序

1. `AGENTS.md`：AI 协作入口。
2. `docs/README.md`：文档总索引。
3. `docs/process/document-first.md`：文档先行流程。
4. `docs/ai/context.md`：当前项目上下文。
5. 按任务进入对应工程域文档。

## 工程域索引

文档目录要镜像未来代码目录。前端 `web` 和 `admin` 完全独立管理，不设置公共 `shared/` 前端包。

```text
server/  -> docs/server/
web/     -> docs/web/
admin/   -> docs/admin/
```

| 任务类型 | 先读文档 |
| --- | --- |
| 后端架构、业务流程、状态机 | `docs/server/architecture.md` |
| 后端 Go 技术栈、依赖、启动入口、公共包 | `docs/server/go-technical.md` |
| API 响应、错误码、分页、鉴权 | `docs/server/api/conventions.md` |
| 数据库表、索引、事务、迁移 | `docs/server/database/design.md` 和 `server/migrations/001_init.sql` |
| PVE、支付、通知等外部系统 | `docs/server/integrations/README.md` |
| 异步任务和 worker | `docs/server/jobs.md` |
| 官网和用户中心 | `docs/web/architecture.md` |
| 管理后台 | `docs/admin/architecture.md` |
| 本地开发环境 | `docs/development/local-setup.md` |
| 部署和运维 | `docs/operations/deployment.md` |

PVE、支付、通知等外部系统的协议适配、client 代码结构、回调处理和错误映射，归 `docs/server/integrations/`。部署拓扑、运行配置、凭据保存、备份恢复和运维演练，归 `docs/operations/`。

## 已确认技术口径

- 后端：Go 单体应用，不做微服务，不做复杂 DDD。
- Go 版本：1.26.2。
- 后端配置：YAML 配置文件，不使用 `.env` 或环境变量作为主配置来源。
- 前端：Bun + Vue 3 + Vite + TypeScript。
- 数据库：MariaDB 11.4.9 + InnoDB。
- 缓存：Redis 前期可只预留。
- 用户端 API：`/api/*`。
- 管理端 API：`/admin-api/*`。
- 异步任务：独立 `cmd/worker` 进程执行，API 进程只创建任务。
- 后台专属表：统一使用 `admin_` 前缀。
- 金额字段：统一使用整数分，字段名使用 `_cents` 后缀。
- 状态字段：使用字符串常量控制，不使用数据库 enum。

## 文档先行门禁

开始写代码前，必须满足：

- 已阅读 `docs/process/document-first.md`。
- 已找到本次任务对应的工程域文档。
- 若设计缺失，先补文档，不先补代码。
- 若涉及数据库，已同步更新 `docs/server/database/design.md` 和 `server/migrations/`。
- 若涉及接口，已同步更新 `docs/server/api/conventions.md` 或后续模块 API 文档。
- 若涉及前端页面，已同步更新 `docs/web/` 或 `docs/admin/`。
- 若涉及前端请求封装、类型、状态枚举或工具函数，必须分别更新 `docs/web/` 和 `docs/admin/`，不能抽到公共前端包。
- 若涉及 PVE、支付、通知等外部系统代码适配，已同步更新 `docs/server/integrations/`。
- 若涉及部署、运行配置、凭据、备份或恢复演练，已同步更新 `docs/operations/`。

## 关键业务规则

- 支付成功必须先区分 `payment_scene` 和 `order_type`：
  - 新购订单：更新订单为 `paid`，创建唯一实例开通任务。
  - 续费订单：更新订单为 `paid`，延长到期时间或创建续费同步任务。
  - 余额充值：写钱包流水并增加余额，不创建实例任务。
- 实例开通必须先持久化本地恢复锚点：
  - `instances.vmid`
  - `instances.provisioning_key`
  - `instances.pve_task_upid`
- 新购订单状态流转：
  - `pending -> paid -> provisioning -> active`
  - 失败时进入 `failed` 或保留 `provisioning` 等待补偿。
- 支付回调、人工入账、退款、实例开通、实例删除都必须幂等。
- PVE 是外部资源系统，本地 MariaDB 是业务事实源。
- 管理端权限使用 RBAC，后台高风险操作必须写 `admin_audit_logs`。

## 协作规则

- 修改前先看 `git status`，不要覆盖用户未提交改动。
