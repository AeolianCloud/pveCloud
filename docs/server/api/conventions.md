# API 约定

接口最终契约维护在 `docs/server/api/openapi.yaml`。本文件记录跨接口通用约定：响应包裹、错误码、鉴权、幂等和 OpenAPI 暴露方式。

## OpenAPI

- API 契约源文件是 `docs/server/api/openapi-src/`。
- 机器可执行的最终 API 契约是生成文件 `docs/server/api/openapi.yaml`。
- 不要手动编辑生成后的 `docs/server/api/openapi.yaml`；接口变更后运行 `node ./scripts/generate-openapi.mjs`。
- 可以运行 `node ./scripts/generate-openapi.mjs --check` 检查生成文件是否最新。
- API 进程在启用 OpenAPI 时暴露 `GET /openapi.yaml`。
- 启动阶段启用 OpenAPI 时必须校验规范文件。
- 初始化检查接口：
  - `GET /healthz`
  - `GET /api/ping`
  - `GET /admin-api/ping`
- `/healthz` 是轻量健康检查；数据库或 Redis ping 失败时返回非 2xx。

## 响应格式

所有业务响应使用统一包裹：

```json
{"code":0,"message":"成功","data":{}}
```

错误响应：

```json
{"code":40001,"message":"参数错误","data":null}
```

## 错误码范围

- `0`：成功
- `400xx`：参数或校验错误
- `401xx`：未登录、token 无效、token 过期
- `403xx`：无权限
- `404xx`：资源不存在
- `429xx`：请求过于频繁、登录失败限流
- `409xx`：状态冲突、重复提交
- `500xx`：服务端内部错误
- `600xx`：支付错误
- `700xx`：PVE 或实例错误
- `800xx`：管理端操作错误

## 鉴权

- 用户端 JWT 使用用户端 secret 和 issuer。
- 管理端 JWT 使用管理端 secret 和 issuer。
- 管理端接口通过 OpenAPI `security` 声明保护状态。
- 管理端权限码采用 `domain:action` 格式，例如 `dashboard:view`、`payment:manual_credit`。
- 缺少、错误或过期 token 返回 `40101 未登录或登录已过期`。
- 权限不足返回 `40301 无权限`。
- 管理端 JWT 必须包含 `jti`，`jti` 对应 `admin_sessions.session_id`；受保护管理端接口除校验签名、issuer、token type 和过期时间外，还必须校验会话状态。
- 管理端退出登录使用 `POST /admin-api/auth/logout` 吊销当前会话；刷新 token 使用 `POST /admin-api/auth/refresh` 轮换新会话并吊销旧会话。
- 管理端会话自检使用 `GET /admin-api/auth/me`，返回当前管理员、角色 ID、权限码、可见菜单和会话摘要，前端启动和刷新页面时应优先用它确认 localStorage 中 token 是否仍有效。
- 管理端权限以数据库 RBAC 为准；JWT 中的权限快照不能替代服务端对当前角色和权限关系的校验。
- 管理端登录密码长度允许 6 到 72 个字符；本地开发可使用短密码，生产环境仍应使用高强度随机密码。
- 管理端登录必须先调用 `GET /admin-api/auth/captcha` 获取验证码图片和 `captcha_id`；`POST /admin-api/auth/login` 必须提交 `captcha_id` 和 `captcha_code`，后端校验通过后立即删除验证码，校验失败或过期返回 `400xx` 参数/校验错误。
- 管理端登录验证码使用 Redis 保存短 TTL 临时状态，推荐有效期 120 秒；验证码 Redis key 使用 `redis.key_prefix` 前缀并按 `admin:login_captcha:<captcha_id>` 组织，不能把验证码答案返回给前端。
- 管理端登录失败次数使用 Redis 按 `IP + 账号标识哈希` 做 15 分钟短窗口计数，计数达到 5 次时返回 HTTP `429` 和业务错误码 `42901`，错误响应仍使用统一 `code/message/data` 包裹。
- 管理端登录失败、成功、退出和刷新仍写入 `admin_audit_logs`，Redis 只承担短窗口限流计数，不替代审计日志。

## 幂等

- 支付回调、人工入账、退款、实例开通、实例删除和异步任务执行必须具备幂等保护。
- 幂等保护优先使用业务唯一键、任务幂等键、外部交易号或本地状态机重入检查。
- 涉及外部系统的接口不得仅依赖前端防重复提交。
