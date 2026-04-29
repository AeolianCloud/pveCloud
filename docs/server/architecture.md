# 后端架构

`pveCloud` 后端是一个以 Go 为核心的单体应用，包含 API 进程和 Worker 进程。

## 进程职责

- API：提供用户端和管理端 HTTP 接口
- Worker：执行异步任务

API 负责接收请求、做同步业务编排和创建任务。Worker 负责长耗时、可重试或依赖外部系统的任务执行。

## 路由边界

- 用户端 API：`/api/*`
- 管理端 API：`/admin-api/*`

这两个边界必须同时体现在接口文档、路由注册、中间件和前端请求包装中。

## 事实来源

- MariaDB：业务事实来源
- Redis：运行时辅助依赖
- PVE、支付、通知：外部系统

Redis 可以保存缓存、限流、验证码、一次性 token、短锁和短期状态。它不能替代管理端会话最终状态、RBAC 关系、订单状态、支付结果、实例状态、异步任务最终状态或审计事实。

## 目标目录原则

后端目录不再采用“入口分端、业务层扁平共用”的过渡结构。统一采用：

- 先按边界分：`admin`、`web`、`job`
- 再按模块分：一个模块一个目录
- 真正跨端复用的核心业务再下沉到 `domain`
- 基础设施进入 `platform`
- 只有稳定且无业务语义的能力才能进入 `shared`

目标结构如下：

```text
server/
  cmd/
    api/
    worker/
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
    web/
      routes/
      middleware/
      modules/
        auth/
        user/
        order/
        instance/
        ticket/
        system/
    job/
      handlers/
    domain/
      catalog/
      billing/
      instance/
      ticket/
    platform/
      bootstrap/
      database/
      cache/
      logger/
      integrations/
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

管理端模块内允许把以下代码放在一起：

- `handler.go`
- `service.go`
- `repository.go`
- `dto.go`
- `types.go`
- `*_test.go`

如果模块继续变大，再在模块内部拆 `handlers/`、`services/`、`repositories/`、`dto/` 等子目录。

### `internal/web`

用户端专属边界，承载 `/api/*` 的路由、中间件和模块实现。

- `routes/`：用户端路由注册
- `middleware/`：用户端鉴权等中间件
- `modules/*`：按模块组织的用户端 handler、service、repository、dto、test

`web` 目录不能依赖 `admin` 模块内部实现来完成请求逻辑。两端如有共用业务规则，应下沉到 `domain`。

### `internal/job`

异步任务边界，承载任务领取、调度、状态推进和具体任务 handler。

- `dispatcher.go`、`store.go`、`types.go`：任务调度与存储核心
- `handlers/`：按任务类型拆分的执行处理器

Worker 主循环只做调度，不承载具体业务逻辑。

### `internal/domain`

真正跨 `admin` 和 `web` 的核心业务领域，面向长期复用。

适合进入 `domain` 的内容：

- 订单、支付、实例、工单等跨端共享的业务规则
- 多端共同使用的领域模型与状态机
- 不依赖某一端页面语义的编排逻辑

不适合进入 `domain` 的内容：

- 管理端页面专属 handler 和查询拼装
- 用户端页面专属展示逻辑
- 仅服务某一个 API 边界的 DTO

### `internal/platform`

基础设施和外部系统适配层。

- `bootstrap/`：应用启动与依赖装配
- `database/`：数据库初始化和通用持久化基础设施
- `cache/`：缓存客户端与基础封装
- `logger/`：日志初始化与封装
- `integrations/`：PVE、支付、通知等外部系统适配

这里负责“怎么接系统”，不负责“业务应该怎么做”。

### `internal/shared`

仅存放稳定、无明确业务语义、被多边界长期复用的基础能力。

允许进入 `shared` 的典型内容：

- `errors`
- `jwt`
- `password`
- `response`
- `validator`

禁止把具体业务规则、页面语义、权限判断、订单流程等逻辑放入 `shared`。

## 模块组织原则

每个模块优先按业务语义命名，而不是机械按页面路由命名。

推荐示例：

- 管理端：`auth`、`dashboard`、`admin_user`、`admin_role`、`system_config`
- 用户端：`user`、`order`、`instance`、`ticket`

如果一个页面只是多个业务模块的组合，目录仍按业务模块拆，不按页面容器强行合并。

## 领域边界

新增业务进入前，必须先在文档中确认领域边界、事务边界和异步任务边界。

推荐后续按以下业务域建设：

- `admin`：管理端认证、会话、RBAC、系统配置、审计写入
- `catalog`：产品、套餐、地域、节点、镜像和价格
- `billing`：订单、支付、钱包、退款和人工入账
- `instance`：实例生命周期、本地状态机和 PVE 执行锚点
- `ticket`：工单与消息
- `job`：异步任务领取、重试、状态推进和任务 handler 注册

跨领域协作应通过服务方法、任务或明确的事务编排完成，不直接在一个服务中散落修改多个领域的数据表。

## 鉴权与权限

### 用户端

- 使用用户端 JWT secret 和 issuer
- 作用范围是 `/api/*`

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

当前后端基础能力仍保留以下管理域：

- auth
- dashboard
- admin users
- roles and permissions
- admin sessions
- system configs
- audit logs
- risk logs

但当前开放的管理端 API 和前端页面范围已经收缩为：

- `Login`
- `Dashboard`
- `System Settings`
- `403`

`System Settings` 当前包含系统配置、管理员账号、管理组权限和管理员会话。

这意味着“数据结构或内部服务能力存在”不等于“接口、权限码、菜单或前端页面已经开放”。
审计日志查询和高危日志查询当前不属于开放 API 契约。

## 审计与高危日志

- 普通后台操作写入 `admin_audit_logs`
- 高危后台操作同时写入 `admin_audit_logs` 和 `admin_risk_logs`
- 风险日志属于审计域，不单独拆成新的业务域

## 异步任务原则

- API 只创建任务，不做长耗时外部执行
- Worker 从 `async_tasks` 读取任务并推进状态
- 外部调用要依赖本地恢复锚点和补偿逻辑
- 不在长事务中调用 PVE、支付或通知系统
- Worker 的领取、锁定、重试和失败状态以 `docs/server/jobs.md` 为准
