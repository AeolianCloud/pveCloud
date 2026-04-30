# 异步任务与 Worker

当前仓库契约已从基础后台阶段收口，`async_tasks` 与 Worker 不再属于现阶段数据库和 API 契约。

## 当前状态

- 本文仅作为已下线能力的历史说明入口
- 现阶段不再定义任务类型、状态机、领取规则和重试规则
- 现阶段不再保留 `cmd/worker`、`internal/job` 或用户端异步业务流实现

## 重新开放前的前置条件

- 补回 `async_tasks` 及相关业务域表结构迁移
- 更新 `docs/server/architecture.md`
- 更新 `docs/server/database/design.md`
- 更新相关 API 文档和验收口径
