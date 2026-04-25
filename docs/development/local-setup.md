# 本地开发说明

本文件记录本地开发环境约定。连接信息通过后端 YAML 配置文件提供。

## 基础依赖

- Go 1.26.2：后端服务。
- Bun：前端 web/admin 应用。
- MariaDB 11.4.9：主数据库。
- Redis：前期可选，后续用于缓存、会话或队列增强。

## 数据库

目标数据库名：

```text
pvecloud
```

初始化 SQL：

```text
server/migrations/001_init.sql
```

本地导入时使用自己的 MariaDB 账号执行初始化脚本。

## 配置

后端通过 `server/config.yaml` 加载配置，示例文件为 `server/config.example.yaml`。建议配置分组：

```text
app        应用名称、环境、监听端口
database   MariaDB 连接配置
redis      Redis 连接配置，前期可选
jwt        用户端和管理端 JWT 配置、过期时间
pve        PVE 地址、账号、认证信息
payment    支付渠道配置
mail       邮件配置
sms        短信配置，前期可选
log        日志级别、日志路径
```

`config.example.yaml` 可以保留字段示例，真实 `config.yaml` 不提交仓库。

## 启动顺序

第一阶段后端代码生成后，建议按这个顺序启动：

1. MariaDB。
2. Redis，可选。
3. `cmd/api`。
4. `cmd/worker`。
5. web/admin 前端 dev server。

API 进程只创建任务，worker 进程负责执行实例开通、续费同步、订单超时和支付补偿等异步任务。

## 后端本地命令

后端目录为 `server/`。首次启动前复制配置示例：

```powershell
Copy-Item config.example.yaml config.yaml
```

常用命令：

```powershell
go mod tidy
go test ./...
go run ./cmd/api -config config.yaml
go run ./cmd/worker -config config.yaml
```

默认 API 监听 `app.addr`，未配置时使用 `:8080`。健康检查地址：

```text
GET http://localhost:8080/healthz
GET http://localhost:8080/api/ping
GET http://localhost:8080/admin-api/ping
```

## 前端开发

`web` 和 `admin` 完全独立开发、独立构建、独立维护依赖和类型。不要新增公共 `shared/` 前端包，也不要跨工程 import 另一个前端的代码。
