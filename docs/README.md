# 文档索引

项目流程和主要设计口径已经迁移到项目内技能：

```text
.codex/skills/pvecloud-document-first/
```

## 当前保留内容

| 路径 | 角色 |
| --- | --- |
| `.codex/skills/pvecloud-document-first/SKILL.md` | AI 强制工作流入口 |
| `.codex/skills/pvecloud-document-first/references/workflow.md` | 协作流程、文档先行门禁、项目硬规则 |
| `.codex/skills/pvecloud-document-first/references/backend.md` | 后端架构、Go 技术栈、API、任务、集成边界 |
| `.codex/skills/pvecloud-document-first/references/database.md` | 数据库设计、约束、事务规则 |
| `.codex/skills/pvecloud-document-first/references/frontend.md` | `admin/` 和 `web/` 前端规则 |
| `.codex/skills/pvecloud-document-first/references/operations.md` | 本地开发、部署和运维边界 |
| `docs/server/api/openapi.yaml` | OpenAPI 3.x 机器可读接口契约 |
| `server/migrations/` | MariaDB 可执行迁移契约 |

## 维护方式

- AI 执行任务时先读技能，再读对应 reference。
- 接口变更继续更新 `docs/server/api/openapi.yaml`。
- 数据库结构变更继续更新 `server/migrations/`，并同步 `references/database.md`。
- 其他规则变更优先更新 `.codex/skills/pvecloud-document-first/references/`。
