# AI 项目上下文

这份文档给 AI 助手快速建立项目上下文。它不替代具体设计文档，只提炼当前已经确认的决策，避免后续实现时走偏。

## 当前阶段

项目当前处于初始化设计阶段：

- 后端架构文档已建立。
- 后端 Go 技术文档已建立。
- 前端架构文档已建立，并拆分为 `web`、`admin` 两个完全独立的工程域入口。
- 数据库设计文档已建立。
- MariaDB 初始化 SQL 已建立。
- 后端、前端主体代码尚未生成。

后续实现必须按“文档确认 -> 结构生成 -> 核心闭环 -> 管理后台 -> 用户端”的顺序推进。

## 目录意图

```text
pveCloud/
├─ AGENTS.md
├─ docs/
│  ├─ README.md
│  ├─ ai/
│  │  └─ context.md
│  ├─ process/
│  │  └─ document-first.md
│  ├─ server/
│  │  ├─ README.md
│  │  ├─ architecture.md
│  │  ├─ go-technical.md
│  │  ├─ api/
│  │  │  └─ conventions.md
│  │  ├─ database/
│  │  │  └─ design.md
│  │  ├─ integrations/
│  │  │  └─ README.md
│  │  └─ jobs.md
│  ├─ web/
│  │  └─ architecture.md
│  ├─ admin/
│  │  └─ architecture.md
│  ├─ development/
│  │  └─ local-setup.md
│  └─ operations/
│     └─ deployment.md
└─ server/
   └─ migrations/
      └─ 001_init.sql
```

未来代码目录和文档目录保持镜像：

```text
server/ -> docs/server/
web/    -> docs/web/
admin/  -> docs/admin/
```

`web` 和 `admin` 不使用公共前端包。请求封装、API 类型、状态枚举、格式化工具、路由和状态管理都分别放在各自工程内维护。

## 数据库口径

```text
database: pvecloud
engine: MariaDB 11.4.9
charset: utf8mb4
collation: utf8mb4_unicode_ci
```

连接地址、用户名和密码属于部署配置。

## 后端核心边界

- Go 单体应用，不做微服务，不做复杂 DDD。
- Go 版本固定为 1.26.2。
- 后端配置使用 YAML 文件，不使用 `.env` 或环境变量作为主配置来源。
- 用户端 API：`/api/*`。
- 管理端 API：`/admin-api/*`。
- API 进程只创建异步任务。
- `cmd/worker` 进程执行实例开通、续费同步、订单超时和支付补偿。
- PVE 是外部资源系统，本地 MariaDB 是业务事实源。
- 支付成功必须按 `payment_scene` 和 `order_type` 分支。

## 权限边界

用户端和管理端 token 分开：

- 用户端 JWT：`user_id`、`token_type=user`。
- 管理端 JWT：`admin_id`、`token_type=admin`、角色和权限码。

管理端权限使用 RBAC：

```text
admin_users
  -> admin_user_roles
admin_roles
  -> admin_role_permissions
admin_permissions
```

接口只声明需要的权限码，权限判断统一放在 admin permission middleware。

## 第一阶段实现建议

第一阶段优先打通后端核心闭环：

1. 补齐并评审相关文档。
2. 生成 Go 后端基础结构。
3. 加载配置、连接 MariaDB、统一 logger、统一 response/errors。
4. 实现用户和管理员 JWT。
5. 实现产品、套餐、地域、镜像查询。
6. 实现订单创建和价格计算。
7. 实现支付单创建和支付成功回调的幂等处理。
8. 实现数据库任务队列和 worker。
9. 用 mock PVE client 打通实例开通。
10. 补核心 service 单元测试。

真实支付和真实 PVE 可以在 mock 闭环稳定后再接。
