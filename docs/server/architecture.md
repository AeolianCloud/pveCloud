# 后端架构

`pveCloud` 后端是一个以 Go 为核心的单体应用，包含 API 进程和 Worker 进程。

## 进程职责

- API：提供用户端和管理端 HTTP 接口
- Worker：执行异步任务

API 负责接收请求、做同步业务编排和创建任务。
Worker 负责长耗时、可重试或依赖外部系统的任务执行。

## 路由边界

- 用户端 API：`/api/*`
- 管理端 API：`/admin-api/*`

这两个边界必须同时体现在接口文档、路由注册、中间件和前端请求包装中。

## 事实来源

- MariaDB：业务事实来源
- Redis：运行时辅助依赖
- PVE、支付、通知：外部系统

Redis 可以保存缓存、限流、验证码、一次性 token、短锁和短期状态。
它不能替代管理端会话最终状态、RBAC 关系、订单状态、支付结果、实例状态、异步任务最终状态或审计事实。

## 代码结构原则

- `services/` 负责业务规则
- `api/` 负责 handler
- `middleware/` 负责鉴权、权限与通用请求链逻辑
- `integrations/` 负责外部协议适配
- `models/` 负责持久化模型
- `dto/` 负责输入输出对象
- `pkg/` 负责稳定基础能力

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

但当前管理端前端页面范围已经收缩为：

- `Login`
- `Dashboard`
- `403`

这意味着“后端能力存在”不等于“前端页面应该存在”。

## 审计与高危日志

- 普通后台操作写入 `admin_audit_logs`
- 高危后台操作同时写入 `admin_audit_logs` 和 `admin_risk_logs`
- 风险日志属于审计域，不单独拆成新的业务域

## 异步任务原则

- API 只创建任务，不做长耗时外部执行
- Worker 从 `async_tasks` 读取任务并推进状态
- 外部调用要依赖本地恢复锚点和补偿逻辑
- 不在长事务中调用 PVE、支付或通知系统
