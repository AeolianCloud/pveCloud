# 部署与运维说明

## 部署组成

```text
api
worker
admin
web（未来）
MariaDB
Redis
```

## 代理边界

- 代理 `/api/*` 到用户端 API
- 代理 `/admin-api/*` 到管理端 API
- `admin` 和未来的 `web` 可使用独立域名或同域不同路径
- `worker` 不暴露公开业务 HTTP 端点

## 配置与依赖

- 配置示例契约：`server/config.example.yaml`
- 真实部署配置不得提交到仓库
- Redis 是运行时基础依赖，不提供生产降级模式
- 管理端会话最终状态、RBAC、订单、支付、实例、任务和审计最终状态仍以 MariaDB 为准

## 本地开发脚本与生产的区别

仓库根目录的 `scripts/dev.mjs` 仅面向开发环境。
它不是生产进程管理方案。

## 运维关注点

- API 与 Worker 启动时都必须检查 Redis 可用性
- `/healthz` 应能反映核心依赖健康状态
- PVE 操作需要通过本地恢复锚点和任务状态追踪
- 高危管理操作必须进入审计域

## 备份与恢复

第一阶段至少覆盖：

- MariaDB 备份
- 配置安全备份
- PVE 节点与模板配置记录

恢复演练需要覆盖订单、支付、实例、异步任务和审计日志。
