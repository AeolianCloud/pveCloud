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
    setup-admin/
  internal/
    admin/
      routes/
      middleware/
      modules/
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

后端不再保留“入口分端、业务层扁平堆放”的过渡结构。统一采用“先按边界分，再按模块分”的目录方案；当前只保留 `admin` 管理端边界；跨端核心业务进入 `domain`；基础设施进入 `platform`；仅稳定通用能力进入 `shared`。

## 结构原则

- `cmd/api`：API 入口
- `cmd/setup-admin`：初始化管理员工具
- `internal/admin`：管理端边界
- `internal/domain`：跨端核心业务领域
- `internal/platform`：基础设施与外部适配
- `internal/shared`：稳定通用能力

## 领域拆分基线

大型业务新增前，先完成对应文档中的领域边界和事务边界确认。代码拆分应遵守：

- 每个端内模块拥有自己的 handler、service、repository、dto 和测试
- 跨端共用且不带页面语义的规则，才允许进入 `domain`
- 跨领域写操作由明确的编排服务承接
- 外部系统适配放入 `platform/integrations/`，业务裁决留在模块或领域服务
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
air -c .air.toml
go run ./cmd/setup-admin -config config.yaml -username admin -email admin@example.com -password "123123"
```

## 验收基线

- `go mod tidy`
- `gofmt -w .`
- `go test ./...`
- API 能成功启动
- `/healthz`、`/admin-api/ping` 可访问
- Redis 不可用时，API 必须启动失败并输出明确错误

## 测试基线

新增或重构以下能力时，必须补对应单元测试或数据库集成测试：

- 管理端认证、验证码、登录限流、会话刷新和吊销
- RBAC 权限匹配、通配权限和路由鉴权
- 审计写入、敏感字段脱敏
- 系统配置读取、敏感配置隐藏和更新审计
- 用户端受保护 API、管理端写 API、公开回调 API 和任何会改变状态的 handler
- 外部系统集成适配，包括请求签名、响应验签、回调验签、错误码映射、超时和不可用分支
- 幂等、并发状态流转、唯一约束冲突、事务内审计或失败后补偿
- 敏感数据处理，包括证件号、真实姓名、token、secret、供应商完整响应不得进入响应、日志、审计或测试输出
- 订单、支付、钱包、实例等涉及金额、状态机或幂等的业务路径

### API 测试要求

只要新增或修改 API 行为，不能只跑 `go test ./...` 后人工点页面结束；必须至少补一层长期回归测试：

- handler 测试覆盖请求解析、鉴权入口、HTTP 状态码、统一响应包裹、错误码、公开/受保护路由边界和请求体大小限制。
- service 测试覆盖状态机、权限裁决、资源归属、幂等、事务、副作用编排、审计写入成功和失败分支。
- integration adapter 测试覆盖第三方请求构造、响应验签或可信性校验、回调验签、防重放、供应商错误码到本地状态的映射。
- repository 或数据库集成测试覆盖唯一索引、并发冲突、软删除、分页排序、空值兼容和迁移新增字段。
- DTO 或响应测试覆盖敏感字段不返回、字段命名不漂移、分页结构和公开配置白名单。

若某个 API 暂时无法补完整 handler 测试，必须在同一次变更中至少补 service 或 adapter 测试，并在回复或 PR 中说明缺口、原因和后续风险；不能把“后面再补测试”作为默认交付状态。

### 实名供应商回归测试矩阵

实名相关能力属于安全敏感路径，后续改动必须优先补正式 `*_test.go`：

- 用户提交：实名开关、供应商可用性、摘要密钥缺失、证件格式、重复提交、已通过拒绝覆盖、失败后重提次数。
- 供应商会话：配置完整性、外部创建失败后的本地状态、不得返回证件号码明文或供应商密钥。
- 同步查询：支付宝响应验签、微信/腾讯云 SDK 查询错误映射、供应商不可用保持可恢复状态、最终通过前本地重复证件检查。
- 回调处理：当前仅支付宝回调开放；必须覆盖验签失败、时间窗口、重放、未知会话、重复回调和请求体大小限制。
- 管理端同步：成功和失败都必须有 `admin_audit_logs` 锚点，审计详情不得包含证件号码明文、供应商完整响应或密钥。
- 并发冲突：两个申请同时收到供应商通过时，唯一约束冲突方必须落入明确拒绝或冲突失败状态，不得长期悬挂 `pending`。

数据库集成测试应使用可重复初始化的测试库或事务回滚策略，不能依赖开发机已有脏数据。

## 测试文件管理

`*_test.go` 属于后端代码资产，用来沉淀回归保护和契约验证。为功能修复、重构、权限、审计、配置等能力新增的测试文件，测试通过后应保留在仓库中，不应在验证完成后删除。
只有临时探测脚本、一次性样例、调试数据生成器这类不承载长期断言的文件，才应在验证后清理。这类临时文件不要命名成正式 `*_test.go`，避免和项目测试资产混淆。
