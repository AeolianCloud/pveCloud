# 后端 Go 技术说明

本文件描述 `server/` 的技术栈、目录结构、运行方式和基础验收方式。

## 技术栈

| 领域 | 选择 |
| --- | --- |
| Go | 1.26.2 |
| HTTP | Gin |
| ORM | GORM |
| DB driver | `gorm.io/driver/mysql` |
| Config | YAML |
| Redis client | `github.com/redis/go-redis/v9` |
| JWT | `github.com/golang-jwt/jwt/v5` |
| Validation | `github.com/go-playground/validator/v10` |
| Logging | `log/slog` JSON logs |
| Tests | `github.com/stretchr/testify` |
| Dev reload | Air |

## 目录结构

```text
server/
  cmd/
    api/
    worker/
    setup-admin/
  internal/
    api/
      admin/
      web/
    bootstrap/
    dto/
      admin/
    middleware/
    models/
    pkg/
    routes/
    services/
  migrations/
  storage/logs/
  config.example.yaml
```

## 结构原则

- `cmd/api`：API 入口
- `cmd/worker`：Worker 入口
- `cmd/setup-admin`：初始化管理员工具
- `bootstrap/`：应用初始化
- `api/`：handler 层
- `services/`：业务服务层
- `middleware/`：请求链中间件
- `routes/`：路由注册
- `pkg/`：稳定基础能力

## 配置原则

- 真实配置默认路径：`server/config.yaml`
- 配置示例契约：`server/config.example.yaml`
- 新增配置项时先更新示例配置
- Redis 是运行时基础依赖，不提供生产降级模式

## 常用命令

```powershell
cd server
Copy-Item config.example.yaml config.yaml
go mod tidy
gofmt -w .
go test ./...
go run ./cmd/api -config config.yaml
go run ./cmd/worker -config config.yaml
air -c .air.toml
air -c .air.worker.toml
go run ./cmd/setup-admin -config config.yaml -username admin -email admin@example.com -password "123123"
```

## 验收基线

- `go mod tidy`
- `gofmt -w .`
- `go test ./...`
- API 能成功启动
- Worker 能成功启动
- `/healthz`、`/api/ping`、`/admin-api/ping` 可访问
- Redis 不可用时，API 与 Worker 必须启动失败并输出明确错误
