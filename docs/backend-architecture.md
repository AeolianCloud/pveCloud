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
│  └─ api/
│     └─ main.go                         # API 服务启动入口
│
├─ internal/
│  ├─ bootstrap/
│  │  ├─ app.go                          # 应用初始化
│  │  ├─ config.go                       # 配置加载
│  │  ├─ database.go                     # 数据库连接
│  │  └─ redis.go                        # Redis 连接，后期可选
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

## 技术建议

- Web 框架：Gin。
- ORM：GORM。
- 数据库：MySQL。
- 缓存：Redis，前期可只预留。
- 鉴权：JWT。
- 配置：`.env`。
- 异步任务：前期可用 Go goroutine + 数据库任务表，后期再接 Redis 队列。

## 暂不引入

- 微服务。
- 复杂 DDD。
- 全局 repository 层。
- 事件总线。
- Kubernetes 目录结构。
