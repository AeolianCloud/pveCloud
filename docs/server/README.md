# Server 文档索引

本目录对应未来代码目录 `server/`。

## 当前文档

| 文档 | 对应代码区域 | 说明 |
| --- | --- | --- |
| `architecture.md` | `server/cmd/`、`server/internal/` | 后端整体架构、目录、流程、状态 |
| `go-technical.md` | `server/go.mod`、`server/cmd/`、`server/internal/` | Go 技术栈、依赖、启动、工程约定 |
| `api/conventions.md` | `server/internal/api/`、`server/internal/pkg/response` | API 响应、错误码、分页、鉴权 |
| `database/design.md` | `server/migrations/`、`server/internal/models/` | MariaDB 表设计、索引、事务边界 |
| `integrations/README.md` | `server/internal/integrations/` | PVE、支付、通知等外部系统边界 |
| `jobs.md` | `server/cmd/worker/`、`server/internal/jobs/` | 异步任务和 worker 行为 |

## 维护规则

- 新增后端模块时，先补本文档索引。
- 新增 Go 依赖、启动入口、公共包或工程约定时，同步 `go-technical.md`。
- 新增表结构时，同步 `database/design.md` 和 `server/migrations/`。
- 新增外部系统时，同步 `integrations/README.md`。
- 新增异步任务时，同步 `jobs.md`。
