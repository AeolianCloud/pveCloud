# 后端架构深化计划

本文档用于交接和记录后端架构深化工作。上一轮已经完成目录重构和主要反向依赖清理，本轮继续完成仓储下沉、领域规则下沉、事务封装和长期回归测试补强。

## 当前基线

已完成：

- 旧入口分端目录和旧 platform 子目录已迁出。
- API 入口装配已迁入 `server/internal/app/api`。
- HTTP 路由、handler 和中间件已迁入 `server/internal/delivery/http`。
- 管理端和用户端业务编排已迁入 `server/internal/usecase`，但仍需继续拆为 `server/internal/usecase/admin/*` 与 `server/internal/usecase/web/*` 两条边界。
- 用例输入输出类型已迁入 `server/internal/usecase/admin/dto` 和 `server/internal/usecase/web/dto`。
- 外部邮件和实名适配已迁入 `server/internal/integration`。
- 配置、数据库、Redis 和日志基础设施已迁入 `server/internal/platform`。
- GORM model 已按领域拆入 `server/internal/repository/mysql/<domain>`。
- `usecase` 不再依赖 `delivery/http`，也不再依赖 Gin。

接手和最终交付前必须跑：

```powershell
cd server
gofmt -w .
go test ./...
```

如果该基线不通过，先修复基线，不进入或交付下面的深化项。

## 本轮完成目标

### 1. 仓储层继续下沉

状态：已完成。

目标：让 usecase 不再直接拼 GORM 查询，把可复用查询、写入、锁定读取和关联表操作沉到 `repository/mysql/<domain>`。

优先顺序：

1. `systemconfig`：范围小，先作为样板。
2. `audit`：查询和写入边界清晰。
3. `usecase/admin/adminsession`、`usecase/admin/auth`：会话读写和过期状态维护。
4. `adminrole`、`adminuser`：RBAC 关系表操作和权限集合查询。
5. `usecase/web/auth`、`usecase/web/userprofile`、`usecase/admin/webuser`：用户账号、会话、密码重置和管理端用户管理。
6. `fileattachment`：文件记录、引用计数、上传者摘要。
7. `productcatalog`、`catalog`：产品、套餐、价格、地域、系统模板。
8. `usecase/web/realname`、`usecase/admin/realname`、`domain/realname`：实名申请、供应商会话、人工审核和同步。

要求：

- `usecase/admin/*` 和 `usecase/web/*` 分别保留各自访问边界内的业务编排、权限裁决、事务边界和错误语义。
- repository 只做持久化查询和写入，不做权限裁决、审计语义或跨领域业务流程。
- 不改变现有 API 路由、字段、错误码、权限码、数据库表结构和配置项。
- 每完成一个领域，运行 `go test ./...`。

完成记录：

- `iam`：管理员、角色、权限、菜单、会话和 RBAC 关系表查询/写入已进入 `repository/mysql/iam`。
- `user`：用户账号、用户会话、密码重置 token 查询/写入已进入 `repository/mysql/user`。
- `file`：文件附件、引用计数和软删除写入已进入 `repository/mysql/file`。
- `catalog`：产品、套餐、价格、地域、模板和关联表操作已进入 `repository/mysql/catalog`。
- `realname`：实名申请列表、详情、锁定读取、状态写入、重复摘要检查已进入 `repository/mysql/realname`。
- `dashboard`：后台首页静态指标计数已进入 `repository/mysql/dashboard`。
- `usecase` 直接 GORM 查询/写入扫描已清零；后续目录拆分时仍需保持 `usecase/admin/*` 与 `usecase/web/*` 不直接拼可复用 GORM 查询。

### 2. GORM model 按领域拆分

状态：已完成。

目标：把 `repository/mysql/models` 从全局包拆到领域子包，避免所有业务共享一个巨大 model 包。

建议拆分：

- `repository/mysql/iam`：`AdminUser`、`AdminRole`、`AdminPermission`、`AdminSession` 以及 RBAC 关系查询。
- `repository/mysql/user`：`User`、`UserSession`、`UserPasswordResetToken`。
- `repository/mysql/audit`：`AdminAuditLog`。
- `repository/mysql/systemconfig`：`SystemConfig`。
- `repository/mysql/file`：`FileAttachment`、`FileAttachmentReference`。
- `repository/mysql/realname`：`UserRealNameApplication`。
- `repository/mysql/catalog`：`Product`、`ProductPlan`、`PlanPrice`、`SalesRegion`、`ServerOSTemplate`、`PlanRegion`、`PlanOSTemplate`。

要求：

- 先迁 model，再迁 repository，再改 usecase import。
- 保持 `TableName()` 不变。
- 不新增迁移，不改表字段。
- 每个领域迁完后检查没有继续引用旧 `repository/mysql/models`。

完成记录：

- 旧 `repository/mysql/models` 已删除。
- 当前按 `iam`、`user`、`audit`、`systemconfig`、`file`、`realname`、`catalog` 分包保留 model。
- `repository/mysql/models` 引用扫描已清零。

### 3. 领域规则下沉到 domain

状态：已完成。

目标：把纯业务规则从 usecase 中提取到 `domain/<domain>`，让 usecase 只负责编排。

优先抽取：

- `domain/realname`：实名状态转换、供应商结果映射、重复实名策略、人工审核可用性。
- `domain/iam`：管理员状态、会话状态、RBAC 权限集合判断。
- `domain/user`：用户状态、会话状态、密码重置 token 状态。
- `domain/catalog`：产品、套餐、价格、地域、模板状态规则和可见性规则。
- `domain/file`：文件状态、危险扩展名和可删除性策略。

要求：

- domain 不依赖 Gin、GORM、Redis、配置结构、HTTP DTO 或第三方 SDK。
- domain 函数只接收普通值对象或领域实体，返回明确结果或领域错误。
- 不改变任何对外行为；先补测试再迁关键规则。

完成记录：

- `domain/iam`：管理员状态、会话状态、RBAC 权限匹配封装。
- `domain/user`：用户状态、用户会话状态、密码重置 token 状态规则。
- `domain/file`：危险扩展名、扩展名/MIME/magic bytes 匹配、路径安全和可删除策略。
- `domain/realname`：实名供应商失败文案、重复摘要冲突、回调重放和人工审核基础规则。
- `domain/catalog`：公开产品目录可见性和可展示套餐组成规则。
- `domain/systemconfig`：基础配置值校验、敏感空值保留、实名供应商配置完整性规则。

### 4. 事务边界收口

状态：已完成。

目标：让事务拥有方清晰，避免 repository 自行开启跨领域事务。

建议：

- 新增 `platform/tx` 或 `repository/mysql/tx` 事务封装。
- `usecase/admin/*` 和 `usecase/web/*` 声明各自事务边界，repository 接收 transaction handle 或接口。
- 外部调用、邮件、实名供应商请求不得放入长事务。
- 审计写入需要和业务状态同事务时，必须明确由 usecase 编排。

要求：

- 不改变现有提交/回滚语义。
- 改事务代码时必须补对应 service 或 repository 测试。

完成记录：

- 新增 `repository/mysql/tx` 事务封装。
- `usecase/admin/*` 和 `usecase/web/*` 声明事务边界，repository 接收事务 handle。
- 补充 `WithinContext`，恢复事务创建时携带请求上下文。

### 5. 鉴权中间件去数据库细节

状态：已完成。

目标：管理端和用户端鉴权中间件不再直接持有 GORM 查询细节。

建议：

- 抽 `usecase/admin/auth` 或 `repository/mysql/iam` 提供会话校验、管理员加载、RBAC 权限加载。
- 抽 `usecase/web/auth` 或 `repository/mysql/user` 提供用户会话校验和用户加载。
- 中间件只负责读取 Bearer token、调用校验端口、写入 Gin 上下文和统一错误响应。

要求：

- 不改变未登录、过期、禁用、无权限的错误语义。
- 反向检查 `/admin-api/*` 和 `/api/*` 路由组隔离。

完成记录：

- 管理端中间件只读取 Bearer token、调用 `usecase/admin/auth` 的认证服务，再写入 Gin 上下文。
- 用户端中间件只读取 Bearer token、调用 `usecase/web/auth` 的认证服务，再写入 Gin 上下文。
- 会话、管理员/用户状态和 RBAC 当前数据库事实由 usecase + repository 校验。

### 6. 应用装配继续收口

状态：已完成。

目标：减少 routes 文件中的 service/repository 构造，让 `internal/app/api` 成为依赖装配中心。

建议：

- `app/api` 内部创建 repositories、`usecase/admin/*`、`usecase/web/*` 和 handlers。
- `delivery/http/admin/routes` 和 `delivery/http/web/routes` 只接收 handlers 或 route set。
- 保持 `cmd/api` 作为薄入口。

要求：

- 不引入反射式 DI 容器。
- 装配结构必须清晰可读，不为了抽象而抽象。

完成记录：

- 新增 `internal/app/api/routes.go` 聚合 route set。
- `delivery/http/admin/routes` 和 `delivery/http/web/routes` 只挂载 route set，不再构造 service/repository。

### 7. 测试补强

状态：已完成。

优先补长期回归测试：

- RBAC 权限集合和通配权限。
- 管理端 JWT、会话刷新、会话吊销、过期处理。
- 用户端注册、登录、刷新、密码找回、会话失效。
- 系统配置敏感项隐藏、空值保留、实名配置完整性校验。
- 文件上传扩展名、MIME、magic bytes、路径安全和删除引用检查。
- 实名提交、同步、回调验签、重复回调、并发通过冲突。
- 产品目录公开可见性和价格/地域/模板过滤。

要求：

- 新增测试文件必须是长期回归测试，不保留临时探测脚本。
- 每个高风险领域至少补 service 或 repository 层测试；公开回调和受保护 API 优先补 handler 或 integration adapter 测试。

完成记录：

- `domain/iam`：RBAC 权限集合和通配权限、管理员会话状态。
- `domain/user`：用户会话状态、密码重置 token 状态。
- `domain/systemconfig`、`usecase/admin/systemconfig` 与 `usecase/web/siteconfig`：敏感配置隐藏、空值保留、实名配置完整性校验和公开配置读取。
- `domain/file`：扩展名、MIME、magic bytes、路径安全和删除引用策略。
- `domain/realname`：供应商失败文案、回调重放、重复摘要冲突。
- `domain/catalog` 与 `usecase/web/catalog`：公开目录可见性和价格/地域/模板过滤。

## 验收标准

最终交付前必须全部通过：

```powershell
cd server
gofmt -w .
go test ./...
```

并执行以下扫描：

执行旧目录引用扫描、HTTP delivery 反向依赖扫描和 Gin import/符号依赖扫描。

预期：

- 第一条没有旧目录引用。
- 第二条没有 usecase/domain/repository/integration/platform/shared 反向依赖 HTTP delivery。
- 第三条没有 usecase/domain/repository/integration/platform 依赖 Gin。

如果改动涉及 API、权限、错误码、数据库、配置或业务行为，必须先回到对应 owner docs 或机器契约更新并等待确认。
