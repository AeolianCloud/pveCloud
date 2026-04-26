# 后端 Go 技术文档

## 技术栈

| 领域 | 选择 |
| --- | --- |
| Go | 1.26.2 |
| HTTP | Gin |
| ORM | GORM |
| DB driver | `gorm.io/driver/mysql` |
| Config | YAML |
| Dev reload | Air |
| Logging | standard `log/slog` JSON logs |
| JWT | `github.com/golang-jwt/jwt/v5` |
| Redis | `github.com/redis/go-redis/v9` |
| Password | bcrypt from `golang.org/x/crypto/bcrypt` |
| Validation | `github.com/go-playground/validator/v10` |
| Tests | `github.com/stretchr/testify` |
| Jobs | `cmd/worker` + `async_tasks` |

避免引入 runtime hot reload、全局 repository 层、event bus、微服务框架、代码生成框架和 Kubernetes 目录结构。

## 目录结构

```text
server/
├─ cmd/
│  ├─ api/
│  ├─ worker/
│  └─ setup-admin/
├─ internal/
│  ├─ bootstrap/
│  ├─ routes/
│  ├─ api/
│  │  ├─ web/
│  │  └─ admin/
│  ├─ middleware/
│  ├─ services/
│  ├─ models/
│  ├─ dto/
│  │  ├─ web/
│  │  └─ admin/
│  ├─ jobs/
│  ├─ integrations/
│  └─ pkg/
├─ migrations/
├─ storage/logs/
├─ config.example.yaml
├─ go.mod
└─ go.sum
```

## 配置

- 真实配置默认路径是 `server/config.yaml`，保持忽略，不提交。
- API 和 Worker 支持 `-config config.yaml`。
- 配置示例维护在 `server/config.example.yaml`。
- 当前配置组：`app`、`database`、`redis`、`jwt`、`worker`、`log`。
- 后续可增加：`pve`、`payment`、`mail`、`sms`。
- 不支持运行时热更新配置，配置变更后重启进程。
- Redis 是后端运行时基础依赖，用于缓存、限流、短 TTL 状态、验证码、一次性 token、幂等短锁和防重复提交标记；业务事实、管理端会话有效性、权限和异步任务最终状态仍以 MariaDB 为准。
- Redis 客户端统一由 `bootstrap.App` 注入，业务代码不自行创建连接；所有 key 通过统一 helper 拼接 `redis.key_prefix`。
- API 和 Worker 都通过 `bootstrap.NewApp` 初始化 Redis 并执行 `PING`；Redis 连接失败时进程直接启动失败，不提供生产降级分支。

## 本地命令

```powershell
cd server
Copy-Item config.example.yaml config.yaml
go mod tidy
gofmt -w .
go test ./...
go run ./cmd/api -config config.yaml
go run ./cmd/worker -config config.yaml
air -c .air.toml
go run ./cmd/setup-admin -config config.yaml -username admin -email admin@example.com -password "123123"
```

## 验收基线

- `go mod tidy` 成功。
- `gofmt -w .` 后没有非预期差异。
- `go test ./...` 成功。
- API 可以启动。
- Redis 未启动或配置错误时，API 和 Worker 必须启动失败并输出明确错误。
- `/healthz`、`/api/ping`、`/admin-api/ping` 可访问。
- Worker 可以启动和停止。
