# AI 协作入口

pveCloud 的详细协作流程和项目口径已经沉淀到项目内技能：

```text
.codex/skills/pvecloud-document-first/
```

AI 进入本仓库后，必须优先使用 `pvecloud-document-first` 技能。该技能承载文档先行门禁、工程域索引、技术口径、关键业务规则、代码书写规则和验收流程。

## 必读顺序

1. `AGENTS.md`：本入口。
2. `.codex/skills/pvecloud-document-first/SKILL.md`：强制工作流。
3. `.codex/skills/pvecloud-document-first/references/`：按任务读取对应领域规则。
4. 机器可执行契约：
   - `docs/server/api/openapi.yaml`
   - `server/migrations/`
   - `server/config.example.yaml`

## 不可绕过的规则

- 任何实现、迁移、接口、页面、配置、部署或业务流程变更，都必须先按技能完成文档/契约更新，并暂停等待维护者确认。
- 未经维护者明确确认，不得进入代码实现。
- 不要覆盖用户未提交改动。
- `admin/` 和 `web/` 独立管理，不创建公共前端 `shared/` 包。
- OpenAPI 仍是接口最终契约；数据库迁移 SQL 仍是表结构最终契约。
