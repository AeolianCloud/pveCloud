# 部署和运维说明

本文件记录部署和运维边界。详细部署脚本后续补充。

PVE、支付、通知等外部系统的协议适配、client 代码结构、回调验签和错误映射，写在 `docs/server/integrations/`。本文件只记录这些外部系统在部署和运维层面的要求，例如凭据保存、网络访问、备份恢复和故障演练。

## 服务组成

```text
api      用户端和管理端 HTTP API
worker   异步任务处理进程
web      官网和用户中心前端
admin    管理后台前端
MariaDB  业务事实源
Redis    缓存、会话或队列增强，前期可选
```

## 部署边界

- `/api/*` 代理到 Go API 用户端路由。
- `/admin-api/*` 代理到 Go API 管理端路由。
- `web` 和 `admin` 可以独立域名部署，也可以同域不同路径部署。
- `worker` 不对外暴露 HTTP 业务接口。

## PVE 运维边界

PVE 是外部资源系统，本地 MariaDB 是业务事实源。

- PVE 操作必须通过后端 service 和 worker 编排。
- PVE 请求成功不等于实例交付成功，必须查询 task/UPID 结果。
- 远端成功、本地失败时，通过 `vmid`、`provisioning_key`、`pve_task_upid` 补偿。

## 后台操作要求

- 后台高风险操作必须记录 `admin_audit_logs`。

## 备份建议

第一阶段至少需要备份：

- MariaDB 全量和增量备份。
- 后端 `config.yaml` 等部署配置的安全备份。
- PVE 节点和模板配置记录。

备份恢复演练应覆盖订单、支付、余额、实例、异步任务和审计日志。
