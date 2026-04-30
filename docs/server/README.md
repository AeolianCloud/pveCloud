# Server 文档总览

本目录维护 `server/` 的项目事实、契约和架构说明。

## 文档职责

- `architecture.md`
  后端整体架构、边界、鉴权和当前阶段范围
- `go-technical.md`
  技术栈、目录结构、命令和运行依赖
- `api/`
  当前开放的管理端 API 契约与跨接口约定
- `database/design.md`
  当前数据库契约与表分组
- `jobs.md`
  已下线 Worker / 异步任务的历史说明入口
- `integrations/`
  外部集成历史说明入口

## 建议阅读顺序

按后端任务定位时，建议先按下面顺序建立上下文：
1. `architecture.md`
2. `go-technical.md`
3. 再按任务进入对应子域：
   - API：`api/`
   - 数据库：`database/design.md`
   - 历史下线能力：`jobs.md`

## 当前实现范围

- 后端当前只把 `/admin-api/*` 和 `/healthz` 作为现行契约
- `cmd/api` 负责当前唯一的服务入口
- MariaDB 是基础后台事实来源
- Redis 是运行时基础依赖，用于管理端会话、验证码、短 TTL 状态、限流和缓存

当前不再把用户端 `/api/*`、用户端账号、产品、订单、支付、实例、工单和异步任务视为现行业务契约。
