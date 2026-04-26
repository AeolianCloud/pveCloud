# pveCloud 文档索引

`docs/` 是项目事实、设计说明和接口文档的维护位置。AI 协作流程和开发技能只放在 `.codex/skills/pvecloud-document-first/`，不要把这两类内容混在一起。

## 文档边界

| 内容 | 维护位置 |
| --- | --- |
| API 最终契约 | `docs/server/api/` 和对应业务文档 |
| API 响应、错误码、鉴权约定 | `docs/server/api/conventions.md` |
| 后端架构和业务规则 | `docs/server/architecture.md` |
| 后端技术栈、目录、配置和命令 | `docs/server/go-technical.md` |
| 数据库设计说明 | `docs/server/database/design.md` |
| 数据库可执行结构 | `server/migrations/` |
| 异步任务和 Worker | `docs/server/jobs.md` |
| 外部系统集成 | `docs/server/integrations/README.md` |
| 管理端前端设计 | `docs/admin/architecture.md` |
| 用户端前端设计 | `docs/web/architecture.md` |
| 本地开发 | `docs/development/local-setup.md` |
| 部署和运维 | `docs/operations/deployment.md` |
| AI 文档先行工作流 | `.codex/skills/pvecloud-document-first/` |

## 维护规则

- 新增或修改接口：先更新 `docs/server/api/` 下的接口文档，再按需更新对应业务文档和 `docs/server/api/conventions.md`。
- 新增或修改表结构：先更新 `server/migrations/`，设计口径同步写入 `docs/server/database/design.md`。
- 新增或修改页面、状态、菜单、路由、权限展示：更新对应前端文档。
- 新增或修改开发流程、AI 门禁、工程实现守则：更新 `.codex/skills/pvecloud-document-first/`。
- 如果 skill 和 `docs/` 冲突，以 `docs/` 和机器契约为准，然后修正 skill。
