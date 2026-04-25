# API 约定

接口最终契约维护在 `docs/server/api/openapi.yaml`。本文件记录跨接口通用约定：响应包裹、错误码、鉴权、幂等和 OpenAPI 暴露方式。

## OpenAPI

- API 契约源是 `docs/server/api/openapi.yaml`。
- API 进程在启用 OpenAPI 时暴露 `GET /openapi.yaml`。
- 启动阶段启用 OpenAPI 时必须校验规范文件。
- 初始化检查接口：
  - `GET /healthz`
  - `GET /api/ping`
  - `GET /admin-api/ping`
- `/healthz` 是轻量健康检查；数据库 ping 失败时返回非 2xx。

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

## 幂等

- 支付回调、人工入账、退款、实例开通、实例删除和异步任务执行必须具备幂等保护。
- 幂等保护优先使用业务唯一键、任务幂等键、外部交易号或本地状态机重入检查。
- 涉及外部系统的接口不得仅依赖前端防重复提交。
