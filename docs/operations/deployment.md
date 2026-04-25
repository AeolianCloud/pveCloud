# 部署和运维说明

## 部署组成

```text
api      用户端和管理端 HTTP API
worker   异步任务进程
web      官网和用户中心
admin    管理后台
MariaDB  业务事实来源
Redis    缓存、会话或队列增强，第一阶段可选
```

## 代理边界

- 代理 `/api/*` 到 Go API 用户端路由。
- 代理 `/admin-api/*` 到 Go API 管理端路由。
- `web` 和 `admin` 可使用独立域名，也可使用同域不同路径。
- `worker` 不暴露公开业务 HTTP 端点。

## 配置

后端配置示例维护在 `server/config.example.yaml`。真实部署配置不得提交到仓库。

`server/config.example.yaml` 同时作为配置键说明入口，所有示例配置组和配置项都应保留中文注释。注释需要说明用途、单位、默认值语义和安全注意事项，避免运维人员只能通过代码理解配置含义。

建议配置组：

```text
app
database
redis
jwt
worker
openapi
pve
payment
mail
sms
log
```

## 本地启动辅助脚本

本地开发可使用仓库根目录 `scripts/dev.mjs` 同时启动后端 API 和前端开发服务。该脚本只面向开发环境，不作为生产进程管理方案。

脚本边界：

- 使用 Node.js 编写，保持跨平台路径处理，不依赖 PowerShell 专有语法。
- 使用 `server/.air.toml` 启动 API 热重载进程。
- 使用 `server/.air.worker.toml` 启动 Worker 热重载进程。
- 默认启动 API、Worker、`admin` 和已存在的 `web`。
- 后续存在 `web/` 前端时，可自动启动 `web` Vite dev server。
- 脚本不提供运行参数，避免不同开发者启动组合不一致。
- 脚本不管理 MariaDB、Redis、反向代理或生产守护进程。
- 脚本不得生成、输出或提交真实生产密钥。

## PVE 运维

- MariaDB 是业务事实来源，PVE 是外部资源系统。
- PVE 操作由后端服务和 Worker 编排。
- PVE HTTP 请求成功不等于实例交付成功，需要查询 task/UPID。
- 远端成功、本地失败时，必须能通过 `vmid`、`provisioning_key`、`pve_task_upid` 恢复。

## 管理端操作和审计

- 高风险管理端操作必须写入 `admin_audit_logs`。
- 审计应覆盖操作者、对象、操作前后状态、IP、User-Agent 和备注。

## 备份

第一阶段备份应包含：

- MariaDB 全量和增量备份。
- `config.yaml` 和部署配置的安全备份。
- PVE 节点和模板配置记录。

恢复演练应覆盖订单、支付、钱包余额、实例、异步任务和审计日志。
