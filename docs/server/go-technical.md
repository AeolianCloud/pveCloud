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
| 开发热重载 | Air | 本地开发监听 Go、YAML 和 OpenAPI 文件变化，自动重启 API 或 worker 进程 |
| 日志 | 标准库 `log/slog` | 输出 JSON 日志，便于部署平台采集 |
| OpenAPI | `github.com/getkin/kin-openapi` | API 规范文件校验和 OpenAPI 3.x 文档加载 |
| JWT | `github.com/golang-jwt/jwt/v5` | 用户端和管理端 token 签发、解析和类型隔离 |
| 密码哈希 | `golang.org/x/crypto/bcrypt` | 用户和管理员密码哈希，不保存明文密码 |
| 参数校验 | `github.com/go-playground/validator/v10` | DTO 基础字段校验，业务校验仍放 service |
| 测试断言 | `github.com/stretchr/testify` | 单元测试和集成测试断言工具 |
| 任务 | `cmd/worker` + `async_tasks` 表 | 第一阶段使用数据库任务队列 |

暂不引入运行时热加载、全局 repository 层、事件总线、微服务框架、代码生成框架和 Kubernetes 目录结构。

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
│  ├─ openapi/                # OpenAPI 3.x API 规范文件加载和校验
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
openapi    OpenAPI 文档开关和规范文件路径
log        日志级别
```

本地 `config.yaml` 和 `server/config.yaml` 必须加入 `.gitignore`。

配置变更规则：

- 后端进程启动时读取 YAML 配置。
- 运行中不做配置的局部热更新，配置变更必须重启进程后生效。
- 本地开发使用 Air 做热重载，由 Air 监听 Go、YAML、OpenAPI 文件变化并重启进程。
- 生产环境由进程管理器或发布流程负责重启，不在业务进程内监听配置文件。

## 开发热重载

本地开发使用 Air：

```powershell
go install github.com/air-verse/air@latest
air -c .air.toml
```

默认 `.air.toml` 编译并启动 API 进程：

```text
go build -o ./tmp/api.exe ./cmd/api
./tmp/api.exe -config config.yaml
```

worker 调试可以临时修改 `.air.toml` 的启动命令为：

```text
go run ./cmd/worker -config config.yaml
```

Air 只作为本地开发工具，不进入生产运行链路。

## OpenAPI 约定

OpenAPI 采用规范文件先行，接口说明、参数、响应、鉴权和错误返回都写入 OpenAPI 3.x 规范文件。文件路径为：

```text
docs/server/api/openapi.yaml
```

后端启动时在启用 OpenAPI 的情况下加载并校验该文件。API 进程提供只读规范入口：

```text
GET /openapi.yaml
```

接口实现前必须先补充或更新 OpenAPI 文档；OpenAPI 只描述 HTTP 契约，不承载业务流程、事务边界和数据库设计，这些内容仍分别写入架构、数据库和集成文档。

API handler 可以写块注释标签，例如 `@route`、`@request`、`@response`、`@auth`，用于让维护者在代码中快速理解接口行为。这些标签必须与 OpenAPI 3.x 规范文件保持一致，不能成为第二套独立接口契约。

## 认证、密码和校验

JWT、密码哈希和 DTO 校验属于后端基础能力：

- 用户端 JWT 和管理端 JWT 使用不同 issuer、secret 和 `token_type`。
- 密码使用 bcrypt 哈希，禁止保存或日志输出明文密码。
- DTO 使用 validator 做必填、长度、格式和枚举等基础校验。
- 资源归属、状态流转、金额计算、权限码判断等业务校验必须放 service 或 middleware。

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

日志、错误、返回信息和注释规则：

- 日志消息、错误信息、API `message`、命令行提示统一使用中文。
- 函数、类型、接口和包级变量注释采用块注释标签风格，使用 `/** ... */`，第一段用中文说明用途；必要时补充 `@param`、`@return`、`@error`、`@sideeffect` 等标签。
- API handler 注释可以使用 `@route`、`@request`、`@response`、`@auth` 等标签；响应示例必须与 OpenAPI 3.x 文件一致。
- 代码内部注释使用详细中文，重点解释不直观逻辑、业务规则、幂等、事务、权限、补偿、并发和外部系统边界。
- 不写重复代码含义的空注释，例如“设置变量”“调用函数”这类注释不要写。
- API 契约以 `docs/server/api/openapi.yaml` 的 OpenAPI 3.x 规范文件为准，Go 代码注释只做就近辅助说明。
- 第三方协议强制要求英文返回时，按协议返回，并在代码附近用中文注释说明原因。

API handler 注释示例：

```go
/**
 * Ping 返回用户端 API 入口连通性结果。
 *
 * @route GET /api/ping
 * @response 200 {"code":0,"message":"成功","data":{"scope":"api","pong":true}}
 */
```

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
air -c .air.toml
```

## 第一阶段初始化验收点

后端基础初始化完成时，至少满足：

- `go mod tidy` 成功。
- `gofmt -w .` 后无格式差异。
- `go test ./...` 成功。
- `go run ./cmd/api` 能启动 API 进程。
- `GET /healthz` 能返回统一 JSON；数据库异常时返回非 2xx。
- `GET /api/ping` 和 `GET /admin-api/ping` 能区分两个 API 入口。
- `GET /openapi.yaml` 能返回当前 OpenAPI 规范文件。
- `go run ./cmd/worker` 能启动 worker 进程，并能响应退出信号。
- `air -c .air.toml` 能在本地开发时热重载 API 进程。

如果本地环境没有安装 Go，可以先完成文件初始化，但必须在最终交付说明中明确尚未执行 Go 命令。
