# 后端架构

pveCloud 后端是 Go 单体应用，包含 API 进程和独立 Worker 进程。API 进程负责用户端和管理端 HTTP 接口，Worker 负责异步任务执行。

## 架构决策

- 后端采用 Go monolith，不引入微服务或复杂 DDD 分层。
- 用户端 API 使用 `/api/*`。
- 管理端 API 使用 `/admin-api/*`。
- API 进程只创建异步任务，长耗时任务由 `cmd/worker` 执行。
- 核心业务服务共享，避免订单、支付、实例等规则在用户端和管理端重复实现。
- 管理端专用动作使用 `admin_xxx_service.go`。
- 业务模型共享在 `server/internal/models/`，不按前端拆模型。
- 外部协议客户端放在 `server/internal/integrations/`。
- 稳定基础设施放在 `server/internal/pkg/`，例如 response、errors、JWT、password、pagination、validator、logger。

## 鉴权和权限

- 用户 JWT claims 包含 `user_id`、`token_type=user`、签发和过期字段。
- 管理员 JWT claims 包含 `admin_id`、`token_type=admin`、`role_ids`、`permission_codes`、签发和过期字段。
- 用户和管理端 JWT 使用不同 secret 和 issuer。
- 管理端 RBAC 关系：

```text
admin_users -> admin_user_roles -> admin_roles -> admin_role_permissions -> admin_permissions
```

- 管理端权限检查由权限中间件执行。
- Handler 只声明所需权限码，不在业务逻辑里手写权限判断。
- 后端 RBAC 是最终权限边界，前端隐藏菜单只用于改善体验。

### 管理端会话

- 管理端登录成功后创建 `admin_sessions` 记录，并签发带 `jti` 的管理端 JWT；`jti` 与 `admin_sessions.session_id` 一一对应。
- 管理端 JWT 仍可返回角色 ID 和权限码快照给前端渲染，但受保护接口必须以当前数据库 RBAC 和会话状态作为最终授权依据。
- 管理端认证中间件负责校验 token 签名、issuer、token type、过期时间、会话状态和管理员账号状态，并把当前管理员 ID、角色 ID、权限码和会话 ID 写入请求上下文。
- `POST /admin-api/auth/logout` 只吊销当前会话；`POST /admin-api/auth/refresh` 使用当前会话换取新 token，并在同一事务内吊销旧会话。
- `GET /admin-api/auth/me` 是前端刷新页面和恢复登录态时的权威自检接口。
- 登录成功、登录失败、退出登录、刷新 token 和会话失效应写入 `admin_audit_logs`，便于后台安全追踪。
- 管理端登录前必须获取图形验证码；验证码答案只保存在 Redis 短 TTL key 中，登录校验无论成功或失败都应让当前验证码失效，前端失败后重新拉取验证码。
- 登录失败限流按 `IP + 账号标识哈希` 做短窗口限制，使用 Redis 保存 15 分钟失败计数；Redis 不作为审计事实来源，登录失败仍写入 `admin_audit_logs`。

## Redis 基础能力

- Redis 是后端运行时基础依赖，不是只服务登录功能。
- Redis 用于短 TTL 和高频临时状态：登录失败限流、通用 API 限流、验证码、一次性 token、临时缓存、幂等短锁和防重复提交标记。
- 所有 Redis key 必须统一使用 `redis.key_prefix` 前缀，并按业务域组织，例如 `pvecloud:admin:login_fail:<hash>`、`pvecloud:rate:<scope>:<hash>`、`pvecloud:verify:<scene>:<hash>`。
- Redis 中的数据必须可过期；不能把订单、支付、钱包、实例、审计、管理端会话有效性或 RBAC 权限关系只保存在 Redis。
- MariaDB 仍是业务事实来源；Worker 队列仍以 `async_tasks` 表为准，Redis 只可作为短锁或防重复辅助，不替代任务最终状态。
- Redis 不可用时 API 和 Worker 启动失败，避免限流、验证码、缓存或幂等能力静默失效；生产环境不得绕过 Redis 降级运行。

## 核心业务状态

订单：

```text
pending, paid, provisioning, active, cancelled, expired, failed, refunded
```

支付：

```text
created, pending, success, failed, closed, refunding, refunded
```

实例：

```text
creating, running, stopped, suspended, expired, deleting, deleted, error
```

异步任务：

```text
pending, running, success, failed, cancelled
```

工单：

```text
open, pending_admin, pending_user, closed
```

## 支付和订单规则

- 订单金额由后端服务计算，不信任前端传入金额。
- 第一阶段一个订单对应一个实例，不给 `orders` 添加 `quantity`。
- 镜像选择必须按地域和已配置的 PVE 模板映射过滤。
- 支付成功按 `payment_scene` 和 `order_type` 分支：
  - `payment_scene=order`、`order_type=new`：订单标记为 `paid`，创建唯一实例开通任务。
  - `payment_scene=order`、`order_type=renew`：订单标记为 `paid`，延长到期时间或创建续费同步任务。
  - `payment_scene=topup`：写钱包流水并增加余额，不创建实例任务。
- 支付回调、人工入账、退款、实例开通和实例删除必须幂等。

## 实例开通规则

调用 PVE 前先持久化本地恢复锚点：

```text
instances.vmid
instances.provisioning_key
instances.pve_task_upid
```

不要在长数据库事务中调用 PVE。流程应先落本地状态，再由任务执行外部调用，外部失败通过补偿或重试恢复。
