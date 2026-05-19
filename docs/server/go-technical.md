# 后端 Go 技术说明

本文档描述 `server/` 的技术栈、目录结构、运行方式和基础验收方式。

## 技术栈

| 领域 | 选择 |
| --- | --- |
| Go | 1.26.2 |
| HTTP | Gin |
| 架构形态 | 分层架构 + 领域子包 |
| ORM | GORM；GORM model 和可复用查询对象放在 `repository/mysql`，handler 不直接使用 GORM |
| DB | MariaDB |
| DB driver | `gorm.io/driver/mysql` |
| Migration | `server/migrations/*.sql` 作为可执行数据库契约 |
| Transaction | 应用用例层声明事务边界；迁移期可直接使用 GORM 事务，后续再收口事务封装 |
| Config | YAML，使用 `KnownFields` 拒绝未知字段 |
| Redis client | `github.com/redis/go-redis/v9` |
| JWT | `github.com/golang-jwt/jwt/v5` |
| Validation | `github.com/go-playground/validator/v10` |
| Logging | `log/slog` JSON logs |
| DI | 手写构造和 provider 装配，不引入反射式 DI 容器 |
| Tests | `github.com/stretchr/testify`；数据库集成测试可引入 `testcontainers-go` |
| Observability | `log/slog`、request ID；后续跨服务追踪再引入 OpenTelemetry |
| Dev reload | Air |

## 目录结构

```text
server/
  cmd/
    api/
    setup-admin/
  internal/
    app/
      api/
      setupadmin/
    delivery/
      http/
        router/
        middleware/
        admin/
        web/
    usecase/
      admin/
        auth/
        dashboard/
        adminuser/
        adminrole/
        adminsession/
        webuser/
        audit/
        systemconfig/
        fileattachment/
        realname/
        productcatalog/
        dto/
        support/
      web/
        auth/
        userprofile/
        siteconfig/
        realname/
        catalog/
        dto/
        support/
    domain/
      iam/
      user/
      audit/
      systemconfig/
      file/
      realname/
      catalog/
    repository/
      mysql/
        iam/
        user/
        audit/
        systemconfig/
        file/
        realname/
        catalog/
    integration/
      realname/
      mail/
      storage/
    platform/
      config/
      database/
      cache/
      logger/
    shared/
  migrations/
  storage/logs/
  config.example.yaml
```

后端不再保留“入口分端、业务层扁平堆放”的过渡结构。统一采用“按技术层分目录，每层下按领域子包分组”的目录方案；HTTP 管理端和用户端作为传输边界，应用用例层也按 `usecase/admin/*` 和 `usecase/web/*` 分离；基础设施进入 `platform`；仅稳定通用能力进入 `shared`。

## 结构原则

- `cmd/api`：API 入口
- `cmd/setup-admin`：初始化管理员工具
- `internal/app`：应用装配和进程启动依赖图
- `internal/delivery/http`：Gin、路由、中间件、管理端和用户端 HTTP 边界
- `internal/usecase`：业务用例、事务边界、权限裁决和编排
- `internal/domain`：领域实体、状态机、值对象和纯业务规则
- `internal/repository`：持久化实现，当前以 `repository/mysql` 为主
- `internal/integration`：外部系统协议适配
- `internal/platform`：配置、数据库、事务、缓存和日志基础设施
- `internal/shared`：稳定通用能力

## 领域拆分基线

大型业务新增前，先完成对应文档中的领域边界和事务边界确认。代码拆分应遵守：

- 每一层都按领域子包组织；其中 `usecase` 必须先按访问边界拆为 `usecase/admin/*` 与 `usecase/web/*`，例如 `usecase/admin/realname`、`usecase/web/realname`、`domain/realname`、`repository/mysql/realname`
- 管理端和用户端能力通过 `delivery/http/admin/<domain>` 与 `delivery/http/web/<domain>` 暴露
- 管理端和用户端业务编排分别放在 `usecase/admin/<domain>` 与 `usecase/web/<domain>`，两边不得互相导入
- 用例输入输出类型放在 `usecase/admin/dto`、`usecase/web/dto` 或对应边界内的具体用例包，delivery handler 只负责请求绑定、校验和响应写入
- 跨端共用且不带页面语义的规则，优先进入对应 `domain/<domain>`；不要为了复用把 admin/web 流程混放到 `usecase` 根目录
- 跨领域写操作由明确的 usecase 编排服务承接
- 外部系统适配放入 `integration/`，业务裁决留在 usecase 或 domain
- 通用工具进入 `shared/` 前必须具备稳定复用语义，不能把业务逻辑放入泛名目录
- GORM model 和可复用 SQL 查询对象放在 `repository/mysql`；迁移期事务密集型 usecase 可直接使用 GORM，但不得把 GORM 暴露到 handler
- Handler 不直接调用 GORM，不直接拼 SQL，不直接写业务状态机

## 领域子包建议

每个领域按需要在各层建立同名子包，不强制一次建满。以实名为例：

```text
delivery/http/admin/realname/
  handler.go
  routes.go
delivery/http/web/realname/
  handler.go
  routes.go
usecase/admin/realname/
  service.go
  ports.go
  dto.go
usecase/web/realname/
  service.go
  ports.go
  dto.go
domain/realname/
  entity.go
  policy.go
  value.go
repository/mysql/realname/
  model.go
  repository.go
integration/realname/
  client.go
```

较小领域可以只建立需要的层。例如公开站点配置可先落在 `delivery/http/web/siteconfig` 与 `usecase/web/siteconfig`，不单独创建 `domain/siteconfig`。

当某一层继续膨胀，再按用例拆分：

```text
usecase/admin/realname/
  commands/
  queries/
  policies/
usecase/web/realname/
  commands/
  queries/
repository/mysql/realname/
  repositories/
  readmodels/
delivery/http/admin/realname/
  handlers/
  requests/
  responses/
```

优先保证“看层能明白职责，看子包能定位领域”，而不是把所有 handler、service、repository 或 GORM model 平铺到全局根目录。

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

实名相关能力属于安全敏感路径，后续改动优先评估是否需要正式 `*_test.go`：

- 用户提交：实名开关、供应商可用性、摘要密钥缺失、证件格式、重复提交、已通过拒绝覆盖、失败后重提次数。
- 供应商会话：配置完整性、外部创建失败后的本地状态、不得返回证件号码明文或供应商密钥。
- 同步查询：支付宝响应验签、微信/腾讯云 SDK 查询错误映射、供应商不可用保持可恢复状态、最终通过前本地重复证件检查。
- 回调处理：当前仅支付宝回调开放；必须覆盖验签失败、时间窗口、重放、未知会话、重复回调和请求体大小限制。
- 管理端同步：成功和失败都必须有 `admin_audit_logs` 锚点，审计详情不得包含证件号码明文、供应商完整响应或密钥。
- 并发冲突：两个申请同时收到供应商通过时，唯一约束冲突方必须落入明确拒绝或冲突失败状态，不得长期悬挂 `pending`。

数据库集成测试应使用可重复初始化的测试库或事务回滚策略，不能依赖开发机已有脏数据。

## 测试文件管理

正式长期回归 `*_test.go` 属于测试资产，新增后默认保留，不按临时文件清理。
修改 handler、service、路由、DTO、错误码、鉴权、权限、审计、事务、幂等、外部集成或安全校验时，新增的正式 `*_test.go` 应随实现一起提交验证。
只有明确用于一次性探测、临时复现、调试数据生成或维护者要求不保留的测试文件，才在验证完成后清理。
