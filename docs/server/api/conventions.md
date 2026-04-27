# API 通用约定

本文件只描述跨接口共用的约定。
具体接口字段、请求参数和响应数据请看 `docs/server/api/endpoints.md` 及对应业务文档。

## 文档职责

这里描述：

- 统一响应包裹
- 错误码分段
- 鉴权与权限约定
- 幂等与限流原则

不在这里重复每个接口的业务细节。

## 路由边界

- 用户端：`/api/*`
- 管理端：`/admin-api/*`
- 健康检查：`/healthz`

## 统一响应格式

成功响应：

```json
{"code":0,"message":"成功","data":{}}
```

错误响应：

```json
{"code":40001,"message":"参数错误","data":null}
```

## 错误码分段

- `0`：成功
- `400xx`：参数或校验错误
- `401xx`：未登录、token 无效、token 过期、会话失效
- `403xx`：无权限
- `404xx`：资源不存在
- `409xx`：状态冲突、重复提交
- `429xx`：请求过于频繁、登录失败限流
- `500xx`：服务端内部错误
- `600xx`：支付错误
- `700xx`：PVE 或实例错误
- `800xx`：管理端业务操作错误

## 鉴权约定

### 用户端

- 使用用户端 JWT secret 和 issuer
- 作用范围为 `/api/*`

### 管理端

- 使用管理端 JWT secret 和 issuer
- 作用范围为 `/admin-api/*`
- JWT 必须带 `jti`
- `jti` 对应 `admin_sessions.session_id`
- 受保护管理端接口不仅校验 token，还要校验当前会话状态和当前数据库 RBAC

### 管理端权限码

权限码格式统一为：

```text
domain:action
```

例如：

- `dashboard:view`
- `admin:manage`
- `audit:view`
- `system:update`

## 管理端登录与会话约定

- 登录前先调用 `GET /admin-api/auth/captcha`
- 登录请求必须带 `captcha_id` 和 `captcha_code`
- 验证码使用 Redis 存储短 TTL 临时状态
- 登录失败限流按 `IP + 账号标识哈希` 计数
- `GET /admin-api/auth/me` 是前端恢复登录态的首选接口
- `POST /admin-api/auth/logout` 吊销当前会话
- `POST /admin-api/auth/refresh` 轮换新会话并吊销旧会话

## 幂等原则

以下能力必须具备幂等保护：

- 支付回调
- 人工入账
- 退款
- 实例开通
- 实例删除
- 异步任务执行

幂等优先依赖：

- 业务唯一键
- 外部交易号
- 任务幂等键
- 本地状态机重入检查

不能只依赖前端防重复点击。
