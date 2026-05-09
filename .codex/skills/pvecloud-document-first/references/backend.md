# Backend Guardrails

本文档定义后端实现守则，不承载接口字段或表结构契约。

## 先读什么？

- `docs/server/README.md`
- `docs/security.md`（涉及鉴权、权限、脱敏、限流、审计、上传、敏感数据、会话或配置安全时）
- `docs/server/architecture.md`
- `docs/server/api/`
- `docs/server/database/design.md`
- `server/migrations/`
- `server/config.example.yaml`

## 实现边界

- API 契约来自 `docs/server/api/`，不要只改 handler 或 DTO。
- 响应包装、分页形状、错误码、错误语义和鉴权失败语义必须回到 `docs/server/api/conventions.md`；不要在单个 handler 里临时发明口径。
- 表结构契约最终来自 `server/migrations/`，不是口头约定。
- 配置项契约最终来自 `server/config.example.yaml`。
- 业务规则、状态机、事务边界和幂等规则写进 `docs/server/`。
- 管理端接口由 `server/internal/delivery/http/admin/*` 聚合路由，用户端接口由 `server/internal/delivery/http/web/*` 聚合路由。
- 管理端业务编排落在 `server/internal/usecase/admin/*`，用户端业务编排落在 `server/internal/usecase/web/*`，领域规则落在 `server/internal/domain/*`，GORM model 和可复用查询对象落在 `server/internal/repository/mysql/*`。
- `admin-api` 和 `api` 不要混在同一组 handler 路径里。
- 真正跨端复用的领域规则进入对应 `domain`，不要复制到 admin/web 两份实现，也不要为了复用把业务流程放回 `usecase` 根目录。

## SDK 与依赖优先

- 涉及支付、实名、短信、邮件、对象存储、云服务、OAuth、验证码、开放平台回调、签名验签或其它第三方协议时，先找与当前契约匹配的官方 SDK 或成熟社区库。
- 能用稳定依赖完成的，不重复手写协议拼装、鉴权头、签名算法、回调验签、分页器、错误码映射或序列化基础层。
- 只有在没有合适依赖、依赖不维护、许可证、安全或兼容性不满足要求，或维护者明确要求自研时，才自行实现，并在实现说明里写明取舍。

## 代码守则

- 服务负责业务规则，handler 负责请求解析、权限声明和响应。
- 目录按技术层分为 `delivery`、`usecase`、`domain`、`repository`、`integration`、`platform`、`shared`，每层下再按领域子包分组。
- 管理端和用户端不仅是不同 HTTP delivery，应用用例也必须拆到 `usecase/admin/*` 和 `usecase/web/*`，不要继续落在扁平共享目录里。
- `usecase/admin/*` 和 `usecase/web/*` 不互相导入；共享规则进入对应 `domain`，共享外部协议进入 `integration`。
- 真正跨端复用的核心规则进入对应 `domain`。
- 外部系统协议适配放在 `integration/`，业务裁决放在 usecase 或 domain。
- auth、session 和 RBAC 的最终裁决必须在后端完成；前端权限判断只能用于可见性和体验优化。
- 不把长耗时外部调用放进长事务。
- 幂等必须依赖业务唯一键、状态检查或任务键，不能只依赖前端防重复点击。
- 不要为了赶进度先落一个端的接口再口头约定补另一个端。

## 安全与日志

- 安全基线来自 `docs/security.md`；具体接口安全要求来自 `docs/server/api/`；不要只在代码里临时增加安全口径。
- handler、service、repository 和集成适配层不得记录 secret、token、password、session、验证码、SMTP 凭据、数据库密码、Redis 密码、对象存储密钥或敏感原文。
- 对外响应、审计日志和业务日志中的手机号、邮箱、证件号、姓名、地址、密钥片段等敏感字段必须按 owner docs 的脱敏或摘要口径处理。
- 不把前端传入的权限、角色、用户 ID、租户 ID 或状态当成可信最终事实；必须通过后端上下文、会话、RBAC 和数据库状态重新裁决。
- 新增安全校验、脱敏、限流、会话失效、审计写入或权限规则时，先更新对应 API、架构、数据库或配置契约并停确认。
- 做安全相关代码时，至少反向检查未登录、低权限、跨用户资源、重复提交、非法状态、敏感响应和日志泄露路径。

## AI 代码层安全检查

审查或修改后端代码时，AI 必须定向检索并检查：

- SQL 注入：搜索 `Raw`、`Exec`、`Query`、`Where`、`Order`、`fmt.Sprintf`、字符串拼接 SQL、动态 `IN`、动态排序和动态字段名；动态值必须参数化，动态字段必须白名单。
- Mass assignment / 过度绑定：检查 `Bind`、`ShouldBind`、DTO 到 model 映射、`Updates`、`Save` 和 map 更新；不得让请求体直接写入受保护字段。
- 越权与 IDOR：检查 handler/service 是否从登录上下文取主体；不得信任 body/query/path 中的 `user_id`、`admin_id`、`role_ids`、`permission_codes`、状态或资源归属。
- 鉴权绕过：检查新路由是否挂到正确 middleware，`/admin-api/*` 和 `/api/*` 是否分组隔离，公开接口是否有 owner docs 支撑。
- 路径穿越与文件安全：检查 `filepath.Join`、下载/预览/删除路径、上传文件名、压缩解压、相对路径保存和存储根目录限制。
- 文件响应安全：检查下载响应头、`Content-Disposition` 文件名转义、SVG/HTML/压缩包/超大图片开放范围、对象存储 ACL 和缓存控制。
- 命令注入：搜索 `exec.Command`、shell、脚本调用和外部工具参数；不得把用户输入拼进命令字符串。
- SSRF：检查服务端主动请求用户提供 URL、Webhook、图片、回调地址或对象存储地址时是否有协议、域名/IP、端口和内网限制。
- 敏感信息泄露：检查 error wrap、日志字段、审计快照、响应 DTO、测试输出是否包含 token、secret、password、验证码、证件号明文或真实配置。
- 侧信道泄露：检查账号枚举、资源存在性、403/404 差异、错误文案、列表计数和响应耗时是否泄露不该暴露的事实。
- 弱随机与 token：检查会话、验证码、密码重置、文件存储名和一次性 token 是否使用安全随机数、哈希存储、TTL 和单次使用状态。
- 并发与幂等：检查状态变更、审批、发布、密码重置、会话吊销和文件删除是否有事务、唯一约束、状态条件或短锁兜底。
- API 滥用：检查批量接口、导出、搜索、大分页、文件上传和高危操作是否有限制、限流和审计。
- 审计可靠性：检查高危操作审计、敏感详情权限、审计失败策略、日志脱敏和审计日志不可普通物理删除。

如果发现风险但 owner docs 没有规定修复后的拒绝语义、错误码、审计、配置、存储或事务边界，先回到文档先行门禁。

## 并发、事务与副作用

- 并发安全依赖数据库唯一约束、事务隔离、状态检查、任务键或短锁组合；不要只依赖内存状态或前端禁用按钮。
- 会产生外部副作用的操作必须明确事务内外边界；外部调用失败、事务回滚或提交后补偿失败时，要有可恢复的状态或重试锚点。
- 事务内只放必须原子完成的本地数据库操作；邮件、对象存储、第三方 API、长耗时任务和网络调用不要放进长事务。
- 幂等规则、锁粒度、重试语义和失败后的可见状态属于业务行为；新增或改变时先回到 owner docs。

## 测试边界

- handler 测试优先覆盖请求解析、鉴权入口、响应包装、错误码映射和路由边界。
- service 测试优先覆盖业务规则、状态流转、权限裁决、事务/幂等和副作用编排。
- repository 测试优先覆盖查询条件、唯一约束、迁移兼容、分页排序和空值/软删除语义。
- integration adapter 测试优先覆盖外部请求签名、响应验签或可信性校验、回调验签、防重放、错误码映射和超时不可用分支。
- 修改 API、handler、DTO、路由、错误码、鉴权、权限、审计、事务、幂等、外部集成或安全校验时，必须先判断需要补哪一层正式 `*_test.go`；不能只靠手工审查或前端构建。
- 公开回调、用户受保护 API、管理端写 API 和状态流转接口至少要有 handler、service 或 integration adapter 中的一层长期回归测试；缺测试必须在最终说明中列为风险。
- 修复权限、安全、SQL 注入、过度绑定、路径穿越、SSRF、命令注入、审计、事务、幂等、配置或错误处理时，应优先沉淀正式 `*_test.go` 回归保护；未补测试要说明原因和风险。

## 命名

- 使用明确业务域命名。
- 避免把具体业务逻辑塞进 `common`、`helper`、`manager`、`base` 之类泛名。
- 审计域统一使用 `audit` 命名。

## 验证

常用最小验证：

```powershell
cd server
gofmt -w .
go test ./...
```

## 测试文件管理

- AI 不默认保留新建的后端 `*_test.go`。
- 是否新增并保留后端 `*_test.go`，由维护者要求、任务验收口径或 owner docs 明确决定。
- 如果 AI 认为某个变更需要长期回归测试，应先说明理由；未获确认时，不把新增 `*_test.go` 作为默认交付内容。
- 临时探测脚本、一次性样例、本地数据生成文件和未确认保留的测试文件，在验证完成后默认清理。
