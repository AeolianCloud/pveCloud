# 后端 Go 技术说明

本文档描述 `server/` 的技术栈、目录结构、运行方式和基础验收方式。

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
    admin/
      routes/
      middleware/
      modules/
    web/
      routes/
      middleware/
      modules/
    job/
      handlers/
    domain/
    platform/
      bootstrap/
      database/
      cache/
      logger/
      integrations/
    shared/
  migrations/
  storage/logs/
  config.example.yaml
```

后端不再保留“入口分端、业务层扁平堆放”的过渡结构。统一采用“先按边界分，再按模块分”的目录方案：`admin` 和 `web` 各自维护自己的路由、中间件和模块目录；`job` 承载异步任务；跨端核心业务进入 `domain`；基础设施进入 `platform`；仅稳定通用能力进入 `shared`。

## 结构原则

- `cmd/api`：API 入口
- `cmd/worker`：Worker 入口
- `cmd/setup-admin`：初始化管理员工具
- `internal/admin`：管理端边界
- `internal/web`：用户端边界
- `internal/job`：异步任务边界
- `internal/domain`：跨端核心业务领域
- `internal/platform`：基础设施与外部适配
- `internal/shared`：稳定通用能力

## 领域拆分基线

大型业务新增前，先完成对应文档中的领域边界和事务边界确认。代码拆分应遵守：

- 每个端内模块拥有自己的 handler、service、repository、dto 和测试
- 跨端共用且不带页面语义的规则，才允许进入 `domain`
- 跨领域写操作由明确的编排服务或异步任务承接
- 外部系统适配放入 `platform/integrations/`，业务裁决留在模块或领域服务
- Worker 任务逻辑放入 `job/handlers/`，不写在 Worker 主循环里
- 通用工具进入 `shared/` 前必须具备稳定复用语义，不能把业务逻辑放入泛名目录

## 模块目录建议

模块默认先保持紧凑结构：

```text
module_xxx/
  handler.go
  service.go
  repository.go
  dto.go
  types.go
  service_test.go
```

当模块继续膨胀，再在模块内部拆分：

```text
module_xxx/
  handlers/
  services/
  repositories/
  dto/
  tests/
```

优先保证“进入模块目录就能找到该模块所有实现”，而不是把所有 service、dto、repository 平铺到全局根目录。

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

## 测试基线

新增或重构以下能力时，必须补对应单元测试或数据库集成测试：

- 管理端认证、验证码、登录限流、会话刷新和吊销
- RBAC 权限匹配、通配权限和路由鉴权
- 审计写入、高危日志写入、敏感字段脱敏
- 系统配置读取、敏感配置隐藏和更新审计
- Worker 任务领取、锁定、成功、失败、重试和超过最大重试次数
- 订单、支付、钱包、实例等涉及金额、状态机或幂等的业务路径

数据库集成测试应使用可重复初始化的测试库或事务回滚策略，不能依赖开发机已有脏数据。

## 测试文件管理

`*_test.go` 属于后端代码资产，用来沉淀回归保护和契约验证。为功能修复、重构、Worker 状态机、权限、审计、配置等能力新增的测试文件，测试通过后应保留在仓库中，不应在验证完成后删除。

只有临时探测脚本、一次性样例、调试数据生成器这类不承载长期断言的文件，才应在验证后清理。这类临时文件不要命名成正式 `*_test.go`，避免和项目测试资产混淆。
