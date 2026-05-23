# 部署与运维说明

## 部署组成

```text
api
worker
admin
web
MariaDB
Redis
```

## 代理边界

- 代理 `/admin-api/*` 到管理端 API
- 代理 `/api/*` 到用户端 API 边界
- `admin` 可使用独立域名或同域不同路径
- `web` 可使用独立域名或站点根路径
- `web` 不代理 `/admin-api/*`，`admin` 不代理 `/api/*`
- `worker` 不暴露 HTTP 入口，不配置反向代理路径

## 配置与依赖

- 配置示例契约：`server/config.example.yaml`
- 通用安全基线：`docs/security.md`
- 真实部署配置默认不提交到仓库；维护者明确要求提交时，应避免在日志、回复和示例中展示密钥内容
- API 运行时区由 `app.timezone` 指定，必须使用有效 IANA 时区名；服务启动后应按该配置设置进程默认时区，不能依赖宿主机系统时区
- Redis 是运行时基础依赖，不提供生产降级模式
- 管理端会话最终状态、RBAC 和审计最终状态仍以 MariaDB 为准
- 用户端密码找回依赖 SMTP 邮件发送配置；生产环境开放密码找回前必须配置 `mail` 并确认重置链接域名指向 Web 端
- 实例到期邮件提醒复用 SMTP 邮件发送配置；生产环境启用 `notification.email_enabled=true` 前必须确认 `mail` 可用
- 用户端实名依赖后台系统设置中的实名供应商配置；生产环境开放支付宝实名前必须配置供应商密钥、回调基础地址或异步通知地址、返回地址和证件摘要密钥；开放微信侧实名前必须配置腾讯云密钥、规则、返回地址和证件摘要密钥
- Web 当前开放用户账号自助、支付宝/微信侧用户实名、服务器产品目录、订单、续费订单、实例和工单，不代表真实支付 API 已开放
- Worker 生产进程必须与 API 使用同一份 `server/config.yaml`，并能访问 MariaDB、Redis、SMTP 和 MCP PVE client API
- 多 Worker 部署时 `worker.id` 必须唯一；`worker.lock_ttl_seconds` 应大于单轮任务的正常执行耗时

## 本地开发脚本与生产的区别

仓库根目录的 `scripts/dev.mjs` 仅面向开发环境。它不是生产进程管理方案。

## Worker 启动口径

- API 进程入口为 `server/cmd/api`，Worker 进程入口为 `server/cmd/worker`；两者应作为独立进程由进程管理器分别守护。
- Worker 使用 `-config` 指定同一份后端 YAML 配置；不注册 HTTP 路由，也不需要反向代理健康入口。
- `worker.enabled=false` 时，Worker 进程会保持空闲等待退出信号，不领取任务；生产启用后台任务前必须设为 `true`。
- 多 Worker 并行时依赖 MariaDB 中的任务锁字段避免重复执行，仍必须保证每个进程的 `worker.id` 唯一，便于排查锁持有者。

## 运维关注点

- API 启动时必须检查 Redis 可用性
- `/healthz` 应能反映核心依赖健康状态
- Worker 启动前必须确认 MariaDB、Redis 和配置可用；Worker 失败不应通过反向代理对外暴露
- 只有在确认实例生命周期策略后才启用 `instance_lifecycle.auto_release_enabled=true`；到期自动释放只允许调用 MCP 当前已有的删除 VM 能力
- 高危管理操作必须进入审计域
- 支付宝实名供应商回调路径必须能被外部供应商访问，并在反向代理层保留原始请求方法、请求体和必要签名字段；当前微信/腾讯云不开放异步回调，结果通过服务端同步查询确认
- MCP PVE client API 只由后端服务端访问，不应由反向代理作为用户端或管理端公开路径暴露；真实 `mcp_pve.bearer_token` 只写入 `server/config.yaml`
- 实名供应商密钥、SecretKey 和证件摘要密钥保存在后台敏感配置中，不得出现在部署日志、反向代理日志、备份明文或前端构建产物中
- `admin` 和 `web` 的静态资源、域名和代理边界必须分开配置
- 若未来新增支付、钱包或其它 `/api/*` 业务能力，需要同步更新 API 契约、后端实现边界和代理规则
- 日志、备份、部署输出和示例配置不得包含真实 token、password、secret、验证码、SMTP 凭据、数据库密码、Redis 密码或对象存储密钥
- 反向代理必须明确区分 `/admin-api/*`、`/api/*`、管理端静态资源和用户端静态资源

## 备份与恢复

第一阶段至少覆盖：

- MariaDB 备份
- 配置安全备份
- Worker 无独立持久化状态；异步任务、通知、实例生命周期事实均随 MariaDB 备份恢复

恢复演练至少覆盖管理员账号、角色权限、系统配置、审计日志、异步任务、通知记录和实例生命周期字段。
