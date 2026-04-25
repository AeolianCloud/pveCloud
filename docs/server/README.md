# Server 文档索引

后端文档维护 pveCloud API 进程、Worker、数据库、接口约定和外部集成设计。AI 实现守则在 `.codex/skills/pvecloud-document-first/references/backend.md`，不要把接口字段或表结构只写进 skill。

| 路径 | 内容 |
| --- | --- |
| `docs/server/architecture.md` | 后端架构、业务规则、鉴权权限、状态机 |
| `docs/server/go-technical.md` | Go 技术栈、目录结构、配置、命令、验收基线 |
| `docs/server/api/conventions.md` | API 响应、错误码、鉴权、幂等约定 |
| `docs/server/api/openapi-src/` | OpenAPI 源文件，按路径和 schema 拆分维护 |
| `docs/server/api/openapi.yaml` | 自动生成的 OpenAPI 机器可读接口契约，不手动编辑 |
| `docs/server/database/design.md` | 数据库设计、表分组、约束、事务边界 |
| `docs/server/jobs.md` | 异步任务和 Worker 规则 |
| `docs/server/integrations/README.md` | PVE、支付、通知等外部系统边界 |
| `server/migrations/` | MariaDB 可执行迁移契约 |
| `server/config.example.yaml` | 后端配置示例契约 |
