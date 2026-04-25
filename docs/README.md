# 文档索引

pveCloud 采用文档先行。代码目录和文档目录保持镜像关系，先把设计写清楚、评审通过，再生成或修改代码。

## 顶层结构

```text
pveCloud/
├─ AGENTS.md
├─ docs/
│  ├─ README.md
│  ├─ ai/
│  ├─ process/
│  ├─ server/
│  ├─ web/
│  ├─ admin/
│  ├─ development/
│  └─ operations/
├─ server/
├─ web/
└─ admin/
```

## 文档和代码映射

| 代码目录 | 文档目录 | 说明 |
| --- | --- | --- |
| `server/` | `docs/server/` | Go 后端、API、数据库、任务、集成 |
| `web/` | `docs/web/` | 官网和用户中心 |
| `admin/` | `docs/admin/` | 管理后台 |
| 根目录 | `docs/process/` | 工作流和文档先行流程 |
| 部署环境 | `docs/operations/` | 部署拓扑、运行配置、凭据管理、备份、安全 |

## 集成和运维边界

PVE、支付、通知等外部系统的代码适配、协议封装、回调处理、错误映射和业务调用边界，统一写在 `docs/server/integrations/`。

部署、运行环境、反向代理、进程管理、凭据保存、备份恢复和运维演练，统一写在 `docs/operations/`。运维文档可以引用 PVE、支付等外部系统，但不定义业务编排和 client 代码结构。

## 推荐阅读顺序

1. `AGENTS.md`
2. `docs/README.md`（本文件）
3. `docs/process/document-first.md`
4. `docs/ai/context.md`
5. `docs/server/architecture.md`
6. `docs/server/database/design.md`
7. `server/migrations/001_init.sql`
8. `docs/server/api/conventions.md`
9. `docs/web/architecture.md`
10. `docs/admin/architecture.md`

## 当前文档清单

| 路径 | 说明 |
| --- | --- |
| `ai/context.md` | AI 项目上下文 |
| `process/document-first.md` | 文档先行流程 |
| `server/README.md` | 后端文档索引 |
| `server/architecture.md` | 后端 Go 架构设计 |
| `server/go-technical.md` | 后端 Go 技术栈、依赖、启动和工程约定 |
| `server/api/conventions.md` | API 响应、错误码、分页、鉴权约定 |
| `server/api/openapi.yaml` | OpenAPI 3.x 接口规范 |
| `server/database/design.md` | MariaDB 数据库设计 |
| `server/integrations/README.md` | 外部系统集成边界 |
| `server/jobs.md` | 异步任务和 worker 约定 |
| `web/architecture.md` | 官网和用户中心文档 |
| `admin/architecture.md` | 管理后台文档 |
| `development/local-setup.md` | 本地开发环境说明 |
| `operations/deployment.md` | 部署和运维说明 |

## 维护规则

- 文档移动后必须同步更新 `AGENTS.md` 和 `docs/README.md`。
- 后端边界、状态、流程变更时，更新 `docs/server/architecture.md`。
- 后端 Go 版本、依赖、启动入口、公共包和工程约定变更时，更新 `docs/server/go-technical.md`。
- HTTP 接口新增或变更时，更新 `docs/server/api/openapi.yaml`。
- 表结构、索引、字段注释变更时，更新 `docs/server/database/design.md` 和 `server/migrations/`。
- API 响应格式、错误码、分页结构变更时，更新 `docs/server/api/conventions.md`。
- 官网和用户中心变更时，更新 `docs/web/`。
- 管理后台变更时，更新 `docs/admin/`。
- 前端请求封装、类型、状态枚举、工具函数变更时，分别更新 `docs/web/` 和 `docs/admin/`。两个前端不建立公共 `shared/` 包。
