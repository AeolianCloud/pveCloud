# API 约定

本文件记录用户端 `/api/*` 和管理端 `/admin-api/*` 的通用接口约定。具体接口清单后续按模块补充。

## 路由边界

```text
/api/*        官网和用户中心 API
/admin-api/*  管理后台 API
```

两个入口共享统一响应格式，但鉴权 token、中间件和权限规则分开。

## 初始化检查接口

后端初始化阶段提供以下无需鉴权的检查接口，用于本地开发、反向代理和进程探活：

```text
GET /healthz          服务健康检查，返回应用、环境和数据库连接状态
GET /openapi.yaml     OpenAPI 规范文件
GET /api/ping         用户端 API 入口检查
GET /admin-api/ping   管理端 API 入口检查
```

`/healthz` 必须只做轻量检查，不执行业务写入；数据库不可用时返回非 2xx 状态，便于部署平台或运维脚本判断服务状态。

## OpenAPI

OpenAPI 规范文件位于：

```text
docs/server/api/openapi.yaml
```

接口新增或变更时，必须先更新 OpenAPI 3.x 规范，再进入 handler、DTO 和 service 实现。OpenAPI 描述请求路径、方法、参数、响应结构、错误响应和鉴权要求；业务流程、幂等、事务和补偿策略仍放在对应后端设计文档。

API 进程启动时校验 OpenAPI 文件；校验失败应阻止服务启动，避免文档和实现继续分叉。

Go handler 可以写块注释标签用于就近说明接口行为，例如 `@route`、`@request`、`@response`、`@auth`。这些注释必须与 OpenAPI 3.x 规范文件一致；接口契约仍以 `docs/server/api/openapi.yaml` 为准。

## 统一响应

成功响应：

```json
{
  "code": 0,
  "message": "成功",
  "data": {}
}
```

失败响应：

```json
{
  "code": 40001,
  "message": "参数错误",
  "data": null
}
```

分页响应：

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "items": [],
    "page": 1,
    "page_size": 20,
    "total": 0
  }
}
```

## 错误码分段

```text
0      成功
400xx  请求参数、校验错误
401xx  未登录、token 无效、token 过期
403xx  无权限
404xx  资源不存在
409xx  状态冲突、重复提交
500xx  系统内部错误
600xx  支付相关错误
700xx  PVE/实例相关错误
800xx  后台操作相关错误
```

handler 不直接拼错误响应，统一使用 `pkg/errors` 和 `pkg/response`。

API `message` 必须使用中文。第三方支付或外部协议要求固定英文响应时，允许按协议返回，但必须在实现处用中文注释说明原因。

## 鉴权

用户端使用用户 JWT：

```text
user_id
token_type=user
issued_at
expires_at
```

管理端使用管理员 JWT：

```text
admin_id
token_type=admin
role_ids
permission_codes
issued_at
expires_at
```

管理端接口必须声明权限码，并由 `middleware/admin_permission.go` 统一校验。

## 幂等

支付回调、人工入账、退款、实例开通、实例删除都必须幂等。涉及支付和异步任务时，优先使用本地支付单号、第三方交易号、业务编号或 `idempotency_key` 做重复处理保护。

## 支付渠道

支付渠道字段统一使用 `channel`，有效值如下：

```text
alipay    支付宝
wechat    微信支付
balance   余额支付
manual    后台人工入账或人工处理
```

支付场景字段统一使用 `payment_scene`：

```text
order     订单支付
topup     余额充值
```
