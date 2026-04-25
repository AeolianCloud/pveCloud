# Server 文档索引

后端设计规则已经迁移到项目内技能：

```text
.codex/skills/pvecloud-document-first/references/backend.md
.codex/skills/pvecloud-document-first/references/database.md
```

## 仍需直接维护的契约

| 路径 | 角色 |
| --- | --- |
| `docs/server/api/openapi.yaml` | OpenAPI 3.x 接口契约 |
| `server/migrations/` | MariaDB 迁移契约 |
| `server/config.example.yaml` | 后端配置示例 |

后端架构、Go 技术栈、任务和外部集成规则变更时，优先更新技能 references。
