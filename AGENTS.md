# AI 协作入口

本文件只定义 AI 进入 pveCloud 仓库后的协作入口、阅读顺序和边界。不要把接口字段、接口示例、业务返回结构、数据库表结构等契约内容写进本文件。

## 核心边界

- 接口文档和接口契约写在 `docs/`，其中机器可执行的最终 API 契约是自动生成的 `docs/server/api/openapi.yaml`。
- `docs/server/api/openapi.yaml` 不要手动编辑；接口变更先改 `docs/server/api/openapi-src/` 下的小源文件，再运行 `node ./scripts/generate-openapi.mjs` 生成。
- 数据库表结构的最终契约是 `server/migrations/` 下的 SQL。
- 示例配置的最终契约是 `server/config.example.yaml`。
- 项目开发技能、协作流程、文档先行门禁、工程规则、代码书写规则和验收流程写在 `.codex/skills/pvecloud-document-first/`。
- 技能可以索引和提醒应更新哪些文档或契约，但不要替代 `docs/` 成为接口文档。

## 必读顺序

1. `AGENTS.md`：本入口，只确认协作边界和阅读顺序。
2. `.codex/skills/pvecloud-document-first/SKILL.md`：强制工作流。
3. `.codex/skills/pvecloud-document-first/references/`：按任务读取对应领域开发规则。
4. 按任务读取对应契约或文档：
   - API：`docs/server/api/openapi-src/`、生成后的 `docs/server/api/openapi.yaml` 和 `docs/server/api/`
   - 后端设计：`docs/server/`
   - 管理端前端：`docs/admin/`
   - 用户端前端：`docs/web/`
   - 数据库：`server/migrations/`
   - 配置：`server/config.example.yaml`

## 不可混淆

- 新增或修改接口时，先更新 `docs/server/api/openapi-src/`，再运行 `node ./scripts/generate-openapi.mjs` 生成 `docs/server/api/openapi.yaml`；不要直接手写生成后的 `openapi.yaml`，也不要只改技能 reference。
- 新增或修改页面行为、流程、权限、配置、部署、数据结构时，先更新对应 `docs/` 或机器契约；纯 UI/UX 视觉美化（布局、间距、颜色、字体、图标、响应式样式、非契约文案）不需要为了确认而写入 `docs/`；技能只维护开发规则和工作流。
- `.codex/skills/pvecloud-document-first/references/` 可以记录技术口径、目录约定、权限检查原则、前后端边界和验收命令，但不承载具体接口字段契约。
- 如果发现技能内容和 `docs/` 或机器契约冲突，以 `docs/` 和机器契约为准，并先修正文档或契约。

## 不可绕过的规则

- 任何实现、迁移、接口、页面行为、配置、部署或业务流程变更，都必须按技能执行文档先行门禁，并在文档/契约更新后暂停等待维护者确认。
- 纯 UI/UX 视觉美化不触发文档先行确认门禁；未经维护者明确确认，不得进入触发文档/契约变更的代码实现。
- 不要覆盖用户未提交改动。
- `admin/` 和 `web/` 独立管理，不创建公共前端 `shared/` 包。
- OpenAPI 源文件生成的 `openapi.yaml` 是接口最终契约；数据库迁移 SQL 是表结构最终契约。
