# Server 文档总览

本目录维护 `server/` 的项目事实、契约和架构说明。它不承担 AI 提示词职责。

## 文档职责

- `architecture.md`
  后端整体架构、边界、鉴权、状态机、核心业务规则
- `go-technical.md`
  技术栈、目录结构、命令、运行依赖、验收基线
- `api/`
  接口契约与跨接口约定
- `database/design.md`
  数据库设计口径、表分组、事务边界、关键约束
- `jobs.md`
  异步任务与 Worker 规则
- `integrations/`
  外部系统集成边界，以及 `server/internal/platform/integrations/` 的职责约束

## 建议阅读顺序

按后端任务定位时，建议先按下面顺序建立上下文：

1. `architecture.md`
2. `go-technical.md`
3. 再按任务进入对应子域：
   - API：`api/`
   - 数据库：`database/design.md`
   - Worker：`jobs.md`
   - 外部集成：`integrations/`

如果只是查看某一个接口、某一个表或某一个任务类型，不要把整个 `docs/server/` 全部读完后再开始。

## 与 skill 的关系

- skill 负责告诉 AI 先读这里、什么时候必须停下来确认。
- 本目录负责告诉所有协作者，后端系统当前是什么、应该怎样工作。

## 当前实现范围

- 后端提供 `/api/*` 用户端接口和 `/admin-api/*` 管理端接口。
- `cmd/api` 负责 HTTP API。
- `cmd/worker` 负责异步任务。
- MariaDB 是业务事实来源。
- Redis 是运行时基础依赖，用于短 TTL 状态、限流、缓存和辅助幂等，但不替代最终业务事实。
