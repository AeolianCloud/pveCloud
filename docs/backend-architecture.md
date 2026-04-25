# 后端 Go 架构设计

## 项目定位

本项目是云服务器销售平台后端，负责官网用户端 API 与管理后台 API。

后端采用 Go 单体应用，不做微服务，不做复杂 DDD。当前阶段优先保证结构清晰、开发速度快、业务逻辑不重复，并为后续 PVE、支付、通知、异步任务保留扩展位置。

## 架构原则

- 官网/用户端 API 与管理后台 API 分开设计。
- 核心业务 Service 共享，避免订单、支付、实例规则重复实现。
- 后台特有业务使用 `admin_xxx_service.go` 单独承载。
- 数据模型统一放在 `models/`，不按前端拆两套模型。
- 外部系统对接统一放在 `integrations/`。
- 云服务器开通、续费、订单超时、支付补偿等异步流程放在 `jobs/`。

## 路由边界

```text
/api/*        官网/用户端 API
/admin-api/*  管理后台 API
```

用户端 API 示例：

```text
/api/auth/register
/api/auth/login
/api/products
/api/orders
/api/orders/:id/pay
/api/instances
/api/instances/:id/reboot
/api/tickets
/api/profile
```

管理端 API 示例：

```text
/admin-api/auth/login
/admin-api/dashboard
/admin-api/users
/admin-api/products
/admin-api/orders
/admin-api/payments
/admin-api/instances
/admin-api/tickets
/admin-api/admins
/admin-api/system
```

## 最终目录结构

```text
server/
├─ cmd/
│  ├─ api/
│  │  └─ main.go                         # API 服务启动入口
│  └─ worker/
│     └─ main.go                         # 异步任务 Worker 启动入口
│
├─ internal/
│  ├─ bootstrap/
│  │  ├─ app.go                          # 应用初始化
│  │  ├─ config.go                       # 配置加载
│  │  ├─ database.go                     # 数据库连接
│  │  ├─ redis.go                        # Redis 连接，后期可选
│  │  └─ worker.go                       # Worker 装配
│  │
│  ├─ routes/
│  │  ├─ routes.go                       # 总路由注册
│  │  ├─ web_routes.go                   # 官网/用户端 API 路由
│  │  └─ admin_routes.go                 # 管理后台 API 路由
│  │
│  ├─ api/
│  │  ├─ web/                            # 官网 + 用户端接口
│  │  │  ├─ auth_handler.go              # 注册、登录
│  │  │  ├─ product_handler.go           # 套餐、价格、地域、镜像
│  │  │  ├─ order_handler.go             # 下单、支付、取消、续费
│  │  │  ├─ payment_handler.go           # 支付、充值、回调
│  │  │  ├─ instance_handler.go          # 我的云服务器、开关机、重启
│  │  │  ├─ ticket_handler.go            # 工单
│  │  │  └─ profile_handler.go           # 用户资料、账户余额
│  │  │
│  │  └─ admin/                          # 管理后台接口
│  │     ├─ auth_handler.go              # 管理员登录
│  │     ├─ dashboard_handler.go         # 仪表盘统计
│  │     ├─ user_handler.go              # 用户管理
│  │     ├─ product_handler.go           # 产品/套餐管理
│  │     ├─ order_handler.go             # 订单管理
│  │     ├─ payment_handler.go           # 支付/流水管理
│  │     ├─ instance_handler.go          # 实例管理
│  │     ├─ ticket_handler.go            # 工单管理
│  │     ├─ admin_handler.go             # 管理员/角色/权限
│  │     └─ system_handler.go            # 系统配置
│  │
│  ├─ services/
│  │  ├─ auth_service.go                 # 用户登录注册
│  │  ├─ admin_auth_service.go           # 管理员登录
│  │  ├─ user_service.go                 # 用户核心业务
│  │  ├─ admin_user_service.go           # 后台用户管理
│  │  ├─ product_service.go              # 产品查询、价格计算
│  │  ├─ admin_product_service.go        # 后台产品维护
│  │  ├─ order_service.go                # 下单、取消、续费、状态流转
│  │  ├─ admin_order_service.go          # 后台订单管理、人工处理
│  │  ├─ payment_service.go              # 支付、充值、回调处理
│  │  ├─ admin_payment_service.go        # 后台支付流水、人工入账
│  │  ├─ instance_service.go             # 实例开通、操作、续费
│  │  ├─ admin_instance_service.go       # 后台实例管理、强制操作
│  │  ├─ ticket_service.go               # 用户工单
│  │  ├─ admin_ticket_service.go         # 后台处理工单
│  │  ├─ dashboard_service.go            # 后台统计
│  │  └─ system_service.go               # 系统配置
│  │
│  ├─ models/
│  │  ├─ user.go                         # 用户表
│  │  ├─ admin.go                        # 管理员表
│  │  ├─ role.go                         # 角色权限
│  │  ├─ product.go                      # 产品/套餐
│  │  ├─ region.go                       # 地域/节点
│  │  ├─ image.go                        # 系统镜像
│  │  ├─ order.go                        # 订单
│  │  ├─ payment.go                      # 支付记录
│  │  ├─ wallet.go                       # 余额/流水
│  │  ├─ instance.go                     # 云服务器实例
│  │  ├─ ticket.go                       # 工单
│  │  ├─ audit_log.go                    # 后台操作日志
│  │  └─ system_config.go                # 系统配置
│  │
│  ├─ dto/
│  │  ├─ web/                            # 用户端请求/响应结构
│  │  │  ├─ auth_dto.go
│  │  │  ├─ product_dto.go
│  │  │  ├─ order_dto.go
│  │  │  ├─ payment_dto.go
│  │  │  ├─ instance_dto.go
│  │  │  ├─ ticket_dto.go
│  │  │  └─ profile_dto.go
│  │  │
│  │  └─ admin/                          # 管理端请求/响应结构
│  │     ├─ auth_dto.go
│  │     ├─ dashboard_dto.go
│  │     ├─ user_dto.go
│  │     ├─ product_dto.go
│  │     ├─ order_dto.go
│  │     ├─ payment_dto.go
│  │     ├─ instance_dto.go
│  │     ├─ ticket_dto.go
│  │     ├─ admin_dto.go
│  │     └─ system_dto.go
│  │
│  ├─ middleware/
│  │  ├─ user_auth.go                    # 用户端 JWT 鉴权
│  │  ├─ admin_auth.go                   # 管理端 JWT 鉴权
│  │  ├─ admin_permission.go             # 后台权限校验
│  │  ├─ admin_audit.go                  # 后台操作日志
│  │  ├─ cors.go
│  │  └─ recover.go
│  │
│  ├─ integrations/
│  │  ├─ pve/
│  │  │  ├─ client.go                    # Proxmox/PVE API 客户端
│  │  │  ├─ auth.go
│  │  │  ├─ vm.go                        # 创建、删除、查询 VM
│  │  │  └─ power.go                     # 开机、关机、重启
│  │  │
│  │  ├─ payment/
│  │  │  ├─ alipay.go                    # 支付宝，预留
│  │  │  ├─ wechat.go                    # 微信，预留
│  │  │  └─ notify.go                    # 支付回调验签
│  │  │
│  │  └─ notify/
│  │     ├─ email.go                     # 邮件通知
│  │     └─ sms.go                       # 短信通知，预留
│  │
│  ├─ jobs/
│  │  ├─ instance_create_job.go          # 实例开通任务
│  │  ├─ instance_renew_job.go           # 实例续费任务
│  │  ├─ order_expire_job.go             # 订单超时取消
│  │  └─ payment_check_job.go            # 支付状态补偿查询
│  │
│  └─ pkg/
│     ├─ response/
│     │  └─ response.go                  # 统一响应
│     ├─ errors/
│     │  └─ errors.go                    # 业务错误码
│     ├─ jwt/
│     │  └─ jwt.go                       # JWT 生成/解析
│     ├─ password/
│     │  └─ password.go                  # 密码哈希
│     ├─ validator/
│     │  └─ validator.go                 # 参数校验
│     ├─ pagination/
│     │  └─ pagination.go                # 分页
│     └─ logger/
│        └─ logger.go                    # 日志封装
│
├─ migrations/
│  └─ 001_init.sql                       # 初始化数据库结构
│
├─ storage/
│  └─ logs/                              # 日志目录，加入 .gitignore
│
├─ .env.example
├─ go.mod
└─ go.sum
```

## 调用关系

用户端：

```text
web 前端
  ↓
/api/*
  ↓
internal/api/web
  ↓
internal/services
  ↓
internal/models
```

管理端：

```text
admin 前端
  ↓
/admin-api/*
  ↓
internal/api/admin
  ↓
internal/services
  ↓
internal/models
```

## Service 拆分规则

核心业务规则放普通 Service：

```text
order_service.go
payment_service.go
instance_service.go
product_service.go
```

后台专属动作放 Admin Service：

```text
admin_order_service.go
admin_payment_service.go
admin_instance_service.go
admin_product_service.go
```

示例：

- 价格计算、订单状态流转、支付成功处理、实例开通规则，只写在核心 Service。
- 人工改价、人工入账、强制取消订单、后台统计、批量上下架，只写在 Admin Service。

## 核心业务状态

数据库字段、表结构、索引设计单独设计。本节只定义后端业务语义，避免后续实现时状态含义不一致。

订单状态建议：

```text
pending        待支付
paid           已支付，等待按订单类型处理
provisioning   新购实例开通中
active         已完成，实例已交付
cancelled      已取消
expired        超时关闭
failed         处理失败，需要人工介入或补偿
refunded       已退款
```

支付状态建议：

```text
created        支付单已创建
pending        等待第三方支付结果
success        支付成功
failed         支付失败
closed         支付关闭
refunding      退款中
refunded       已退款
```

实例状态建议：

```text
creating       创建中
running        运行中
stopped        已关机
suspended      已暂停
expired        已到期
deleting       删除中
deleted        已删除
error          异常
```

异步任务状态建议：

```text
pending        待执行
running        执行中
success        执行成功
failed         执行失败，可重试或人工处理
cancelled      已取消
```

工单状态建议：

```text
open           已提交
pending_admin  等待客服/管理员处理
pending_user   等待用户回复
closed         已关闭
```

## 关键业务流程

### 下单流程

```text
校验用户登录
  ↓
校验套餐、地域、镜像、购买周期
  ↓
计算订单金额
  ↓
创建 pending 订单
  ↓
返回订单信息和支付入口
```

价格计算只能走 `product_service.go` 或 `order_service.go` 中的统一方法，不能由 handler 或前端传入最终金额后直接信任。

### 支付成功流程

```text
接收支付回调或主动查询支付结果
  ↓
验签和校验支付金额
  ↓
幂等检查支付单是否已处理
  ↓
更新支付状态为 success
  ↓
按支付场景和订单类型处理
  ├─ 新购订单：更新订单状态为 paid，并创建唯一实例开通任务
  ├─ 续费订单：更新订单状态为 paid，并延长到期时间或创建续费同步任务
  └─ 余额充值：写入钱包流水并增加余额，不创建实例任务
  ↓
返回第三方支付平台要求的成功响应
```

支付成功处理必须先区分 `payment_scene` 或 `order_type`，再在事务内完成订单、支付、余额流水或任务创建等关键变更。重复回调不能重复入账、重复创建实例或重复修改订单。

### 实例开通流程

```text
实例开通任务开始
  ↓
锁定任务并标记 running
  ↓
检查订单类型、订单状态和实例是否已经处理
  ↓
事务内将订单从 paid 更新为 provisioning
  ↓
创建或复用本地 instance 占位记录，状态为 creating
  ↓
分配并持久化 VMID、任务幂等键等恢复锚点
  ↓
调用 PVE 创建 VM，并记录 PVE task/UPID
  ↓
轮询或查询 PVE 任务结果
  ↓
写入实例最终信息，实例状态更新为 running
  ↓
订单状态更新为 active
  ↓
任务状态更新为 success
```

如果 PVE 创建失败，任务进入 `failed`，订单进入 `failed` 或保持 `provisioning` 并等待补偿策略。由于 VMID、幂等键和 PVE task/UPID 已经本地持久化，即使 PVE 已创建但本地写入后续信息失败，也可以通过补偿任务或后台人工同步找回远端 VM。后台应提供人工重试、取消、备注能力。

### 续费流程

```text
校验实例归属和当前状态
  ↓
计算续费金额
  ↓
创建续费订单
  ↓
支付成功后延长到期时间
  ↓
必要时创建实例恢复或续费同步任务
```

续费和新购可以共用订单能力，但订单类型必须区分，避免支付成功后错误触发新实例开通。

### 订单超时流程

```text
定时扫描 pending 订单
  ↓
判断是否超过支付有效期
  ↓
关闭本地订单
  ↓
必要时关闭第三方支付单
```

订单超时任务必须跳过已支付、已取消、已超时的订单。关闭第三方支付单失败时应记录日志，并允许补偿任务重试。

## 幂等和事务规则

- 所有支付回调必须以第三方交易号、本地支付单号或幂等键做重复处理保护。
- 支付成功、人工入账、退款、实例开通、实例删除都必须设计为可重复调用但不会重复产生副作用。
- 金额相关字段使用整数分或定点小数，禁止使用浮点数参与金额计算。
- 订单金额以服务端计算结果为准，前端传入的金额只能作为展示或二次校验参考。
- 涉及订单、支付、钱包流水、任务创建的链路需要明确事务边界。
- 外部接口调用不要长时间占用数据库事务。通常先完成本地状态和任务创建，再由异步任务调用外部系统。
- 后台人工操作必须记录操作者、操作对象、前后状态和备注。

## 鉴权和权限

用户端使用用户 JWT，管理端使用管理员 JWT。两类 token 的签发主体、过期时间、claims 和中间件分开，避免用户 token 误用于管理端。

用户端建议 claims：

```text
user_id
token_type=user
issued_at
expires_at
```

管理端建议 claims：

```text
admin_id
token_type=admin
role_ids
permission_codes
issued_at
expires_at
```

管理端权限采用权限码控制，示例：

```text
dashboard:view
user:view
user:update
product:create
product:update
order:view
order:cancel
payment:view
payment:manual_credit
instance:view
instance:operate
ticket:reply
admin:manage
system:update
audit:view
```

权限校验放在 `middleware/admin_permission.go`。handler 只声明接口需要的权限码，不在 handler 内手写复杂权限判断。

## 统一响应和错误码

所有 API 使用统一 JSON 响应，用户端和管理端保持同一基础格式。

成功响应：

```json
{
  "code": 0,
  "message": "ok",
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
  "message": "ok",
  "data": {
    "items": [],
    "page": 1,
    "page_size": 20,
    "total": 0
  }
}
```

错误码建议按业务域分段：

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

## 异步任务规则

前期使用独立 `cmd/worker` 进程 + 数据库任务表实现轻量任务队列。数据库表结构后续单独设计，这里只定义任务行为。

任务需要具备：

- 任务类型：实例开通、实例续费、订单超时、支付补偿、实例状态同步。
- 幂等键：同一订单或实例的同一类任务不能无限重复创建。
- 重试次数：失败后按最大次数重试。
- 下次执行时间：支持延迟重试。
- 锁定信息：避免多个 worker 同时执行同一任务。
- 错误信息：保留最近一次失败原因，便于后台排查。

任务执行原则：

- worker 进程启动后循环拉取可执行任务。
- 拉取任务时必须做抢占锁定。
- API 进程只负责创建任务，不在请求链路中直接执行长耗时任务。
- 执行任务前再次检查业务状态，避免过期任务产生错误副作用。
- 调用外部系统失败时记录错误并进入重试。
- 超过最大重试次数后标记 failed，并在后台展示给管理员处理。

## PVE 集成边界

PVE 相关代码统一放在 `integrations/pve/`，Service 不直接拼 PVE HTTP 请求。

需要封装的能力：

- 登录和 token/cookie 管理。
- 查询节点、存储、模板、VM 状态。
- 创建 VM、删除 VM、重装 VM。
- 开机、关机、重启、强制关机。
- 查询 PVE task/UPID 执行结果。

PVE 调用注意事项：

- VMID 分配必须由后端统一管理，避免并发创建冲突。
- 创建 VM 后要查询 PVE 任务结果，不要只按 HTTP 请求成功判断开通成功。
- PVE 节点不可用、资源不足、模板不存在、网络配置失败都要能返回明确错误。
- PVE 返回成功但本地写入失败时，需要补偿任务或后台人工同步。
- 用户端实例操作需要校验实例归属，管理端强制操作需要记录审计日志。

## 支付集成边界

支付相关代码统一放在 `integrations/payment/`，支付业务编排放在 `payment_service.go`。

需要封装的能力：

- 创建支付单。
- 查询支付状态。
- 支付回调验签。
- 关闭支付单。
- 退款申请和退款查询，前期可预留。

支付处理注意事项：

- 回调接口必须跳过普通 JWT 鉴权，但必须做签名校验。
- 回调响应格式要符合第三方支付平台要求。
- 本地支付金额、订单金额、第三方回调金额必须一致。
- 第三方交易号和本地支付单号都要记录，便于对账。
- 支付补偿任务用于主动查询长时间 pending 的支付单。

## 日志和审计

系统日志用于排查服务运行问题，审计日志用于追踪后台人员操作，两者分开处理。

系统日志建议记录：

- 请求 ID。
- 请求路径、方法、状态码、耗时。
- 当前用户或管理员 ID。
- 关键错误堆栈或错误码。
- 外部系统请求摘要和耗时。

审计日志建议覆盖：

- 管理员登录。
- 人工改价、取消订单、退款处理。
- 人工入账或扣减余额。
- 实例强制操作。
- 产品、套餐、地域、镜像配置变更。
- 管理员、角色、权限变更。
- 系统配置和支付配置变更。

敏感信息如密码、token、支付密钥、PVE 密码不能写入日志。

## 配置管理

配置通过 `.env` 加载，`.env.example` 保留必要示例但不包含真实密钥。

建议配置分组：

```text
APP_*        应用名称、环境、监听端口
DB_*         MySQL 连接配置
REDIS_*      Redis 连接配置，前期可选
JWT_*        用户端和管理端 JWT 密钥、过期时间
PVE_*        PVE 地址、账号、认证信息
PAYMENT_*    支付渠道配置
MAIL_*       邮件配置
SMS_*        短信配置，前期可选
LOG_*        日志级别、日志路径
```

用户端 JWT 和管理端 JWT 建议使用不同密钥或至少使用不同 issuer/token_type，避免混用。

## 参数校验

DTO 负责定义请求结构和基础校验规则，Service 负责业务校验。

示例：

- DTO 校验：必填字段、字符串长度、数字范围、枚举合法性。
- Service 校验：用户是否拥有实例、订单是否可支付、套餐是否可售、余额是否足够、状态是否允许流转。

handler 的职责是绑定参数、调用 Service、返回响应，不承载复杂业务判断。

## 测试重点

优先补以下测试：

- 价格计算和订单金额。
- 订单状态流转。
- 支付回调幂等。
- 支付成功后任务创建。
- 余额流水入账和人工入账。
- 实例开通任务重试。
- 用户端实例归属校验。
- 管理端权限校验。
- PVE 客户端错误映射。

第一期可以先做 Service 单元测试和少量 handler 集成测试。外部支付和 PVE 使用 mock client，不直接依赖真实外部环境。

## 开发落地顺序

建议按以下顺序实现：

```text
1. bootstrap、config、database、logger、response、errors
2. 用户 JWT、管理员 JWT、鉴权中间件
3. 产品、套餐、地域、镜像的查询和后台维护
4. 订单创建、价格计算、订单取消和订单查询
5. 支付单创建、支付回调、支付补偿
6. 异步任务 worker 和实例开通任务
7. PVE client 和实例操作
8. 工单、审计日志、系统配置
9. 后台仪表盘统计
```

## 技术建议

- Web 框架：Gin。
- ORM：GORM。
- 数据库：MySQL。
- 缓存：Redis，前期可只预留。
- 鉴权：JWT。
- 配置：`.env`。
- 异步任务：前期使用独立 `cmd/worker` 进程 + 数据库任务表，后期再接 Redis 队列。

## 暂不引入

- 微服务。
- 复杂 DDD。
- 全局 repository 层。
- 事件总线。
- Kubernetes 目录结构。
