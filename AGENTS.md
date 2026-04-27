# AI 协作入口

本文件只定义 AI 进入 `pveCloud` 仓库后的协作入口、读取顺序和职责边界。
它不是接口文档、数据库设计文档、前端页面契约或产品说明书。

## 双轨协作原则

`pveCloud` 采用两条并行但职责分离的协作轨道：

1. AI 轨道：`AGENTS.md` 和 `.codex/skills/`
   作用是约束 AI 的读取顺序、工作流、实现守则和验收方式。
2. 项目轨道：`docs/`、`server/migrations/`、`server/config.example.yaml`
   作用是承载给人和代码共同使用的事实、契约、架构、计划和进度。

不要把这两条轨道混写：

- 不要把接口字段、响应结构、表结构、配置项说明写进 skill。
- 不要把 AI 的提示词、执行门禁、工具偏好写进业务文档。

## 权威来源

- 接口契约：`docs/server/api/`
- 后端架构与业务规则：`docs/server/`
- 管理端前端契约：`docs/admin/`
- 用户端前端契约与规划：`docs/web/`
- 数据库可执行契约：`server/migrations/`
- 配置示例契约：`server/config.example.yaml`
- AI 工作流与守则：`.codex/skills/`

如果 skill 和项目文档冲突，以 `docs/`、迁移 SQL 和配置示例为准，然后修正 skill。

## 必读顺序

1. `AGENTS.md`
2. `.codex/skills/pvecloud-document-first/SKILL.md`
3. 对应领域的 skill reference
4. 对应项目文档或机器契约

按任务类型继续读取：

- 后端与 API：`docs/server/`、`docs/server/api/`
- 管理端前端：`docs/admin/`
- 用户端前端：`docs/web/`
- 数据库：`docs/server/database/`、`server/migrations/`
- 配置与部署：`docs/development/`、`docs/operations/`、`server/config.example.yaml`
- 计划与进度：`docs/analysis/`、`docs/plan/`、`docs/progress/`

## 文档先行门禁

以下变更都属于契约或行为变更，必须先更新文档或机器契约，再暂停等待维护者确认：

- API、错误码、鉴权、权限码、菜单来源、路由语义
- 数据库结构、迁移、事务边界、幂等规则
- 页面行为、页面范围、状态语义、请求包装、权限判断
- 配置项、部署方式、运行依赖、运维流程
- 业务流程、阶段边界、产品开放范围

以下变更属于纯 UI/UX 视觉优化，可以直接实现，不触发确认门禁：

- 布局、间距、颜色、字体、图标、响应式样式
- 不影响流程和契约的静态文案调整

## 仓库级硬规则

- 不要覆盖用户未提交改动。
- `admin/` 和未来的 `web/` 独立管理，不创建跨前端共享运行时代码包。
- 管理端只调用 `/admin-api/*`。
- 用户端只调用 `/api/*`。
- 如果仓库里尚不存在 `web/` 前端目录，`docs/web/` 视为规划与契约草案，不能假定用户端实现已经存在。

## 目标

这套入口的目标不是替代项目文档，而是让 AI 始终先读对地方、按对流程工作、在该停的时候停下来。
