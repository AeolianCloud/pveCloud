# 后端 Go 技术文档

本文档定义 pveCloud 后端 Go 工程的技术口径、目录约定、依赖选择、启动方式和第一阶段验收点。任何后端代码初始化、依赖引入、目录调整和公共包变更，都必须先与本文档保持一致。

## 技术栈

后端采用 Go 单体应用，不做微服务，不做复杂 DDD。

| 分类 | 选择 | 说明 |
| --- | --- | --- |
| Go 版本 | Go 1.26.2 | 后端统一使用该版本，避免不同开发环境的语言和工具链差异 |
| Web 框架 | Gin | 负责 HTTP 路由、中间件、请求绑定和响应输出 |
| ORM | GORM | 连接 MariaDB，后续模型映射以 `server/migrations/001_init.sql` 为准 |
| 数据库驱动 | `gorm.io/driver/mysql` | MariaDB 使用 MySQL 协议连接 |
| 配置 | YAML 文件 | 本地和部署统一使用 YAML 配置文件，不使用 `.env` 或环境变量作为主配置来源 |
| 日志 | 标准库 `log/slog` | 输出 JSON 日志，便于部署平台采集 |
| 任务 | `cmd/worker` + `async_tasks` 表 | 第一阶段使用数据库任务队列 |

暂不引入全局 repository 层、事件总线、微服务框架、代码生成框架和 Kubernetes 目录结构。

## Go 模块

后端 Go 模块位于：

```text
server/
```

模块名：

```text
github.com/AeolianCloud/pveCloud/server
```

依赖新增规则：

- 依赖必须解决明确问题，不能为了预留而提前引入。
- 公共依赖优先选择成熟、维护活跃、文档清晰的库。
- 新增依赖需要能说明用途、替代方案和影响范围。
- 业务代码不得直接依赖前端目录。

## 目录约定

第一阶段先建立后端基础骨架：

```text
server/
├─ cmd/
│  ├─ api/                    # API 进程入口
│  └─ worker/                 # 异步任务进程入口
├─ internal/
│  ├─ bootstrap/              # 配置、数据库、应用装配
│  ├─ routes/                 # 路由注册
│  ├─ api/
│  │  ├─ web/                 # 用户端 handler
│  │  └─ admin/               # 管理端 handler
│  ├─ middleware/             # HTTP 中间件
│  ├─ services/               # 业务服务，后续按模块补齐
│  ├─ models/                 # GORM 模型，字段以迁移 SQL 为准
│  ├─ jobs/                   # worker 任务执行器
│  ├─ integrations/           # PVE、支付、通知等外部系统适配
│  └─ pkg/                    # 后端内部公共工具
├─ migrations/                # MariaDB 迁移 SQL
├─ storage/logs/              # 本地日志目录，不提交日志文件
├─ config.example.yaml
├─ go.mod
└─ go.sum
```

`internal/pkg` 只放后端内部稳定公共能力，例如统一响应、错误码、JWT、分页、参数校验和日志封装。不要把业务规则下沉到 `pkg`。

## 进程边界

API 进程：

- 启动入口为 `server/cmd/api`。
- 负责 HTTP 请求、鉴权、参数校验、调用 service 和创建异步任务。
- 不在请求链路中直接执行 PVE 创建、支付补偿、订单超时扫描等长耗时任务。

Worker 进程：

- 启动入口为 `server/cmd/worker`。
- 负责从 `async_tasks` 拉取任务并执行。
- 执行任务前必须重新检查业务状态，避免过期任务产生副作用。

## 配置约定

配置通过 YAML 文件读取。`config.example.yaml` 只保存示例，不保存真实密钥。真实配置文件默认路径为：

```text
server/config.yaml
```

API 和 worker 进程都必须支持通过启动参数指定配置文件：

```text
-config config.yaml
```

配置分组：

```text
app        应用名称、环境、监听地址、优雅退出时间
database   MariaDB 连接和连接池配置
jwt        用户端和管理端 JWT 密钥、issuer、过期时间
worker     worker 标识、轮询间隔、锁超时、批量大小
log        日志级别
```

本地 `config.yaml` 和 `server/config.yaml` 必须加入 `.gitignore`。

## HTTP 约定

用户端和管理端路由边界：

```text
/api/*
/admin-api/*
```

初始化阶段允许提供以下无需鉴权的检查接口：

```text
GET /healthz
GET /api/ping
GET /admin-api/ping
```

`/healthz` 必须做轻量数据库 ping。数据库不可用时返回非 2xx 状态。

所有业务响应使用统一格式，具体规则见 `docs/server/api/conventions.md`。

## 错误和响应

handler 不直接拼接错误响应。错误码统一定义在后端错误包，响应统一通过 response 包输出。

错误码分段沿用 API 约定：

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

## 数据库约定

数据库结构以 `server/migrations/001_init.sql` 和 `docs/server/database/design.md` 为准。

第一阶段代码初始化只建立连接能力，不自动迁移表结构。表结构变更必须先改数据库设计文档，再新增或修改迁移 SQL。

GORM 模型后续补齐时必须遵守：

- 字段名、类型、索引和注释以迁移 SQL 为准。
- 金额字段使用整数分，字段名后缀为 `_cents`。
- 状态字段使用字符串常量，不使用数据库 enum。
- 不直接暴露自增 ID 给前端，业务展示使用 `order_no`、`payment_no`、`instance_no` 等编号。

## 日志和中间件

基础中间件：

- request id：读取或生成 `X-Request-ID`。
- access log：记录请求方法、路径、状态码、耗时、客户端 IP。
- recover：捕获 panic 并输出统一错误响应。
- cors：本地和前后端联调使用，后续部署可收紧来源。

系统日志输出 JSON；后台审计日志不写普通日志文件，必须后续落 `admin_audit_logs` 表。

## 本地命令

后端目录：

```powershell
cd server
```

首次配置：

```powershell
Copy-Item config.example.yaml config.yaml
```

常用命令：

```powershell
go mod tidy
gofmt -w .
go test ./...
go run ./cmd/api -config config.yaml
go run ./cmd/worker -config config.yaml
```

## 第一阶段初始化验收点

后端基础初始化完成时，至少满足：

- `go mod tidy` 成功。
- `gofmt -w .` 后无格式差异。
- `go test ./...` 成功。
- `go run ./cmd/api` 能启动 API 进程。
- `GET /healthz` 能返回统一 JSON；数据库异常时返回非 2xx。
- `GET /api/ping` 和 `GET /admin-api/ping` 能区分两个 API 入口。
- `go run ./cmd/worker` 能启动 worker 进程，并能响应退出信号。

如果本地环境没有安装 Go，可以先完成文件初始化，但必须在最终交付说明中明确尚未执行 Go 命令。
