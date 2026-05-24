# 异步任务与 Worker

通用异步任务平台重新开放，用于执行实例生命周期和通知等不能放入 API 请求事务内的后台工作。

## 运行边界

- API 进程负责创建业务事实和投递任务。
- Worker 进程负责领取 `async_tasks`、执行任务、记录结果和按规则重试。
- Worker 与 API 共用 MariaDB、Redis、配置文件和日志配置；支付商户配置来自 `system_configs`，不来自前端环境变量。
- Worker 不对外暴露 HTTP API；管理端只通过 `/admin-api/async-tasks/*` 查询和重试任务。

## 任务类型

首批开放任务类型：

- `instance_operation_sync`：同步 MCP operation 和 VM 状态。
- `instance_expiry_notice`：生成实例到期提醒。
- `instance_expiry_release`：到期宽限期后释放实例。
- `notification_email_send`：发送邮件通知。
- `notification_sms_placeholder`：短信通知占位记录；本阶段不接真实短信供应商。
- `payment_order_provision`：真实支付成功后为新购订单触发实例交付。
- `payment_refund_sync`：退款状态不可确认或渠道异步确认延迟时，同步渠道退款状态并完成本地回滚。

## 状态机

任务状态只允许：

- `pending`：待执行。
- `running`：已被 Worker 领取，正在执行。
- `succeeded`：执行成功。
- `failed`：达到最大重试次数后失败。
- `cancelled`：被业务状态覆盖或人工取消后不再执行。

Worker 领取任务必须使用数据库短事务和锁定字段，不能让多个 Worker 同时执行同一个任务。

## 重试与幂等

- 每个任务必须有 `task_no`。
- 可重复投递的任务必须提供 `idempotency_key`，同一类型和幂等键只能存在一条未取消任务；取消任务时必须释放内部幂等投影，重试失败任务时复用原任务行。
- `max_attempts`、`attempts`、`scheduled_at`、`locked_by`、`locked_until` 用于控制重试和领取。
- 任务执行失败但未达到最大次数时，应更新 `scheduled_at` 延后重试。
- 实例操作同步任务在 MCP operation 未完成时不计为业务失败，应延后重试。
- 实例生命周期任务的幂等键必须包含能区分业务版本的信息，例如实例编号和目标到期时间；实例续费或后台调整到期时间后，旧释放任务应被取消或在执行时因状态不匹配而跳过。
- `payment_order_provision` 的幂等键必须使用支付编号或订单编号，执行时必须重新锁定订单并确认 `status=error|pending`、`payment_status=paid`、`order_type=purchase` 且未存在实例；状态已变化时跳过，不重复创建实例。
- `payment_refund_sync` 的幂等键必须使用退款编号，只有渠道退款成功或查询确认成功后才能回滚本地支付生效记录；渠道失败或不可确认时保持退款可排查状态，不扣回用户服务期。

## 审计与安全

- 管理端人工重试任务必须写入后台操作审计。
- `payload` 和 `result` 不得保存 token、密码、SMTP 凭据、MCP Bearer Token、商户私钥、微信 API v3 key、签名串、用户敏感明文、完整支付回调 payload 或完整上游响应。
- 用户端不得看到任务内部错误、上游 operation ID、PVE 节点、VMID 或 Worker 标识。
