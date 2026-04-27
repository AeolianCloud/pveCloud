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

管理端权限采用双层模型：

1. 页面入口权限：控制菜单、路由、tab 是否显示
2. 资源操作权限：控制进入页面后能执行哪些读写操作

#### 页面入口权限

页面入口权限格式统一为：

```text
page.<menu>.<feature>
```

例如：

- `page.dashboard`
- `page.system-settings.config`
- `page.system-settings.admin-users`
- `page.system-settings.admin-roles`

约定：

- 页面入口权限只负责“能不能看到这个入口”
- 页面入口权限不负责接口写操作授权
- 页面入口权限不引入 `*` 通配

#### 资源操作权限

资源操作权限格式统一为：

```text
resource:action
```

例如：

- `dashboard:view`
- `system-config:view`
- `system-config:update`
- `admin-user:*`
- `admin-user:password-reset`

约定：

- `resource:view`：查看列表、详情、页面主数据
- `resource:create`：创建
- `resource:update`：编辑、状态切换、权限分配等更新类操作
- `resource:delete`：删除
- 特殊动作单独拆分，例如 `password-reset`、`revoke`、`operate`
- `resource:*`：该资源全权限，语义上覆盖同资源下所有子权限

实现要求：

- 前端菜单、路由和 tab 显示判断使用页面入口权限
- 前端按钮、区块、提交动作和后端接口鉴权使用资源操作权限
- 如果持有 `resource:*`，则应视为同时拥有该资源全部细粒度权限
- 同一资源权限可被多个页面复用，不因页面增多而新增重复资源权限
- 当前仓库若仍存在旧权限码（如 `admin:manage`、`system:update`），在迁移完成前可兼容保留，但新设计和新增页面不再继续扩展旧命名

新增功能权限 checklist：

1. 新后台页面或新 tab：至少补一个 `page.*` 页面入口权限
2. 新页面若读取独立资源数据：至少补一个对应的 `resource:view`
3. 新按钮或新写操作：补对应的 `create`、`update`、`delete`、`revoke`、`password-reset`、`operate` 等资源权限
4. 新资源若需要“资源管理员”快捷授权：再补一个 `resource:*`
5. 以上权限码必须写入文档，并同步进入数据库 `admin_permissions`
6. 前端菜单/路由/tab 与后端接口不要共用同一个权限码层次，避免页面入口和资源操作混在一起

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
