# pveCloud 文档总览

`docs/` 是 `pveCloud` 的项目事实层。
这里存放给人和代码共同使用的契约、架构、流程、计划和进度说明。

它不承担 AI 提示词职责。
AI 的工作方法、读取顺序和执行门禁请看 `AGENTS.md` 与 `.codex/skills/`。

## 文档分层

### 1. 契约层

用于描述系统必须遵守的对外和对内约束：

- API 契约：`docs/server/api/`
- 数据库设计口径：`docs/server/database/`
- 可执行数据库契约：`server/migrations/`
- 配置示例契约：`server/config.example.yaml`

### 2. 架构层

用于描述系统当前结构、职责边界和实现口径：

- 后端：`docs/server/`
- 管理端入口：`docs/admin/README.md`
- 管理端总体架构：`docs/admin/architecture.md`
- 管理端页面契约：`docs/admin/pages/`
- 管理端路由权限：`docs/admin/routing-permissions.md`
- 用户端：`docs/web/`

### 3. 流程层

用于描述项目协作和交付流程，而不是 AI 提示：

- 文档先行说明：`docs/process/document-first.md`
- 本地开发：`docs/development/local-setup.md`
- 部署与运维：`docs/operations/deployment.md`

### 4. 计划层

用于记录阶段目标、缺口和实施路线：

- 缺口分析：`docs/analysis/`
- 实施计划：`docs/plan/`

### 5. 进度层

用于记录已完成内容、当前状态和下一步：

- 进度说明：`docs/progress/README.md`
- 进度总览：`docs/progress/MASTER.md`
- 分阶段进度：`docs/progress/phase-*.md`

进度文档是阶段账本和历史记录，不是最终契约。
如果它和 API、前端架构、数据库迁移或配置示例冲突，应先以权威契约为准，再同步修正或归档进度文档。

## 协作分工

### 技能和 AGENTS 的职责

- 告诉 AI 先读什么
- 告诉 AI 什么时候必须停下来确认
- 告诉 AI 不要把什么写错地方
- 约束 AI 的实现习惯、目录边界和验证方式

### 文档的职责

- 说明系统当前是什么
- 说明接口、页面、流程和状态应该怎样工作
- 说明阶段范围、设计约束、计划和进度
- 给维护者、开发者、AI 和未来代码变更提供统一参考

## 当前项目状态

- 后端 `server/` 是当前主实现。
- 管理端 `admin/` 已存在，是当前实际前端实现。
- `docs/web/` 当前主要承担用户端规划与契约准备；如果仓库里尚无 `web/` 目录，不应把它误读成已有实现说明。

## 维护原则

- 改接口，先改 `docs/server/api/`。
- 改数据库结构，先改迁移和数据库设计文档。
- 改页面行为、路由、权限、状态、菜单，先改对应前端架构文档。
- 改配置和部署，先改 `server/config.example.yaml` 与运维文档。
- 改 AI 工作流，改 `AGENTS.md` 或 `.codex/skills/`，不要改业务文档来表达提示词。
- 阶段完成后，把长期有效事实沉淀到权威文档；进度文档只保留状态、背景和验收记录。
