# 后端架构

`pveCloud` 后端当前以 Go 基础后台为主，现行契约范围聚焦管理端 API。

## 进程职责

- API：提供管理端 HTTP 接口和健康检查

当前契约不再把 Worker、用户端 API、订单、支付、实例、工单和其他用户端业务流纳入现阶段交付范围。

## 路由边界

- 健康检查：`/healthz`
- 管理端 API：`/admin-api/*`

当前仓库不再把 `/api/*` 作为现行后端契约。

## 事实来源

- MariaDB：基础后台事实来源
- Redis：运行时辅助依赖

Redis 可以保存缓存、限流、验证码、一次性 token、短锁和短期状态。它不能替代管理端会话最终状态、RBAC 关系或审计事实。

## 目标目录原则

后端目录当前按“基础后台可用”口径维护：

- 先按边界分：`admin`
- 再按模块分：一个模块一个目录
- 基础设施进入 `platform`
- 只有稳定且无业务语义的能力才能进入 `shared`

目标结构如下：

```text
server/
  cmd/
    api/
    setup-admin/
  internal/
    admin/
      routes/
      middleware/
      modules/
        auth/
        dashboard/
        admin_user/
        admin_role/
        system_config/
        audit/
    platform/
      bootstrap/
      database/
      cache/
      logger/
    shared/
      errors/
      jwt/
      password/
      response/
      validator/
  migrations/
```

## 目录职责

### `internal/admin`

管理端专属边界，承载 `/admin-api/*` 的路由、中间件和模块实现。

- `routes/`：管理端路由注册
- `middleware/`：管理端鉴权、权限校验等中间件
- `modules/*`：按模块组织的管理端 handler、service、repository、dto、test

### `internal/platform`

基础设施层。

- `bootstrap/`：应用启动与依赖装配
- `database/`：数据库初始化和通用持久化基础设施
- `cache/`：缓存客户端与基础封装
- `logger/`：日志初始化与封装

### `internal/shared`

仅存放稳定、无明确业务语义、被多边界长期复用的基础能力。

允许进入 `shared` 的典型内容：

- `errors`
- `jwt`
- `password`
- `response`
- `validator`

## 模块组织原则

当前示例：

- 管理端：`auth`、`dashboard`、`admin_user`、`admin_role`、`system_config`

## 领域边界

当前只保留 `admin` 管理域：

- 管理端认证
- 会话
- RBAC
- 系统配置
- 审计写入
- 高危日志写入

用户端账号、产品、订单、支付、实例、工单、异步任务和用户端业务流已经从当前阶段契约中收口。后续如需恢复，必须先更新文档与迁移。

## 鉴权与权限

### 管理端

- 使用管理端 JWT secret 和 issuer
- JWT 必须带 `jti`
- `jti` 对应 `admin_sessions.session_id`
- 受保护管理端接口必须同时校验：
  - 签名
  - issuer
  - token type
  - 过期时间
  - 会话状态
  - 管理员状态
  - 当前数据库 RBAC

管理端前端可以消费权限快照改善体验，但后端 RBAC 仍是最终裁决。

## 当前管理端阶段边界

当前后端基础能力保留以下管理域：

- auth
- dashboard
- admin users
- roles and permissions
- admin sessions
- system configs
- audit logs
- risk logs

当前开放的管理端 API 和前端页面范围为：

- `Login`
- `Dashboard`
- `System Settings`
- `403`

`System Settings` 当前包含系统配置、管理员账号、管理员组权限和管理员会话。
当前不再保留用户端 API、产品、订单、支付、实例、工单、异步任务或用户端账号等业务域数据库契约。审计日志查询和高危日志查询当前不属于开放 API 契约。

## 审计与高危日志

- 普通后台操作写入 `admin_audit_logs`
- 高危后台操作同时写入 `admin_audit_logs` 和 `admin_risk_logs`
- 风险日志属于审计域，不单独拆成新的业务域

## 当前不在范围内的能力

以下能力不属于当前阶段契约：

- 用户端 API
- Worker
- 异步任务
- PVE 集成
- 支付集成
- 用户端业务流程
