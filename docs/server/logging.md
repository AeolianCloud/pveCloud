# 日志系统

本文档说明后端当前日志体系的边界和链路。

## 日志分类

pveCloud 当前有两类容易混用的日志：

- 运行时日志：API 进程输出到 stdout 的结构化 JSON 日志，用于运维排障和服务运行观察。
- 后台操作审计：写入 `admin_audit_logs` 的管理端操作事实，用于还原后台操作历史；管理端日志管理中心的操作审计页面查询的就是这类日志。

管理端“登录安全”不是独立表，也不是独立接口。它是 `admin_audit_logs` 中 `object_type=admin_auth` 的认证相关记录视图。

## 运行时日志

运行时日志使用 Go 标准库 `log/slog`，当前由 `internal/platform/logger` 初始化为 JSON handler，并输出到 stdout。

当前运行配置只有：

- `log.level`：支持 `debug`、`info`、`warn`/`warning`、`error`；未知值按 `info` 处理。

当前不提供日志文件输出、日志轮转、采样、慢请求阈值、日志导出或运行日志查询页面。生产环境的采集、留存和轮转应由部署环境、反向代理或日志采集系统负责。

## 请求链路 ID 与访问日志

全局 HTTP 中间件会读取请求头 `X-Request-ID`。客户端未传时由服务端生成，并回写响应头。

每个 HTTP 请求结束后都会写一条访问日志，当前字段包括：

- `request_id`
- `method`
- `path`
- `status`
- `latency_ms`
- `client_ip`

panic 会被 recovery 中间件捕获，写入 error 级别日志，并返回统一内部错误响应。访问日志、panic 日志和后台操作审计都应带同一个请求链路 ID，方便串联排查。

## 后台操作审计

普通后台操作写入 `admin_audit_logs`。该表是管理端操作日志和登录日志的数据库事实来源。

审计记录保存：

- 操作者快照：`admin_id`、`admin_username`、`admin_display_name`
- 请求上下文：`session_id`、`request_id`、`request_method`、`request_path`、`ip`、`user_agent`
- 业务动作：`action`、`object_type`、`object_id`
- 业务快照：`before_data`、`after_data`
- 补充说明：`remark`、`created_at`

管理端中间件统一采集请求上下文并写入 request context。业务用例写审计时只传动作、对象、前后快照和备注；不应在每个业务模块重复拼装 IP、会话、请求路径等通用上下文。

本地数据库状态变更和对应后台审计写入应尽量在同一事务中完成。高危管理操作不得在审计完全不可用时静默成功，除非对应 owner docs 明确允许降级。

后台审计动作值和当前管理端写接口覆盖情况见 `audit-actions.md`。

## 日志管理中心

管理端日志管理中心是独立一级菜单。当前提供操作审计、登录安全、用户安全日志、用户业务日志、前端错误日志和后端运行日志页面。

- 操作审计：分页查询 `admin_audit_logs` 中非认证类普通后台操作日志。
- 登录安全：复用同一接口和表，固定查询 `object_type=admin_auth` 的认证相关日志。

关联接口：

- `GET /admin-api/audit-logs`

页面入口权限：

- 日志管理中心父级：`page.logs`
- 操作审计：`page.logs.admin-operations`
- 登录安全：`page.logs.admin-security`

操作权限：

- 操作审计列表与敏感详情：`audit-log:*`、`audit-log:sensitive-view`
- 登录安全列表：`admin-security-log:*`、`admin-security-log:view`

敏感详情权限：

- `audit-log:sensitive-view` 或 `audit-log:*`

未具备敏感详情权限时，接口不返回 `before_data`、`after_data` 和 `user_agent`。

## 用户安全日志与用户业务日志

`user_security_logs` 保存用户端安全事件，`user_business_logs` 保存实名、订单和工单等关键业务事件。两者由用户端关键流程写入，管理端分别通过独立查询接口查看。

## 前端错误日志

`frontend_error_logs` 保存 admin/web 前端上报的错误摘要，供管理端查看脱敏后的错误路径、错误类型、API 关联和请求链路。

## 后端运行日志

`backend_runtime_logs` 保存结构化运行日志摘要，供管理端查看访问、panic 和关键错误的可查询记录。它不替代 stdout 运行日志输出。

## 支付告警事件日志

真实支付一期的告警先复用运行时日志和 `backend_runtime_logs`，不新增独立告警通道、Webhook、短信或邮件配置。生产环境由日志采集系统或人工巡检基于固定字段触发外部告警。

支付告警事件必须同时写 stdout 结构化日志和 `backend_runtime_logs`：

- `level` 固定为 `error`。
- `category` 固定为 `runtime`。
- `message` 使用 `payment_alert`。
- `module` 固定为 `payment`。
- `event` 只允许 `payment_create_failed`、`payment_callback_signature_failed`、`refund_pending`、`refund_failed`。
- 必须包含可排查业务锚点：`payment_no`、`refund_no`、`order_no`、`provider`、`method`、`status` 中能够确定的字段。
- 错误详情只保存本地错误码或 500 字以内的脱敏摘要，不保存商户密钥、签名串、完整回调 payload、完整上游响应或用户敏感明文。

事件触发口径：

- 支付创建调用渠道失败并将本地支付交易更新为 `failed` 时，写 `payment_create_failed`。
- 支付回调供应商验签失败时，写 `payment_callback_signature_failed`；该事件可能缺少 `payment_no`，但必须包含 `provider` 和请求链路 ID。
- 退款调用渠道失败并将本地退款更新为 `failed` 时，写 `refund_failed`。
- 退款创建后渠道未同步确认成功、仍保持 `pending` 时，写 `refund_pending`；后续若补充 Worker 周期扫描，只能对超过运维文档确认阈值的 `pending` 退款重复写同类事件，且不得改变退款本地状态。

## 日志导出与清理

`log_export_records` 保存日志导出锚点。导出、清理和留存策略应作为单独受控能力处理，并写入管理端审计。

## 脱敏与安全

日志和审计不得成为敏感明文泄露渠道。密码、token、secret、验证码、SMTP 凭据、数据库密码、Redis 密码、对象存储密钥、证件号码明文、供应商完整响应和真实配置不得进入运行日志、审计详情、测试输出或示例文档。

审计写入会对常见敏感 key 做统一脱敏，例如 `password`、`token`、`secret`、`captcha`、`config_value` 等。业务模块仍应在生成审计快照前避免放入不应记录的敏感原文。

邮箱、手机号、姓名、证件号、地址、密钥片段、User-Agent 等敏感或半敏感字段的脱敏口径以 `docs/security.md` 和对应 owner docs 为准。

## 当前非目标

当前日志系统不包含以下能力：

- OpenTelemetry 或跨服务追踪
- 日志文件轮转、远程采集配置或外部告警发送通道

新增这些能力前必须先更新对应 owner docs 或机器契约，并按文档先行流程确认。
