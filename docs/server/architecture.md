# 后端架构

`pveCloud` 后端当前以 Go 基础后台为主，现行契约范围聚焦管理端 API、用户端业务 API 和后台 Worker，并开放最小 Web 公开站点配置、用户账号自助接口、支付宝/微信侧用户实名接口、服务器产品目录展示接口、订单、支付、实例、异步任务和工单能力。

## 进程职责

- API：提供管理端 HTTP 接口、Web 公开站点配置接口、用户账号自助接口、支付宝/微信侧用户实名接口、服务器产品目录展示接口、订单接口、支付接口、实例接口、工单接口、管理端异步任务接口和健康检查
- Worker：领取并执行通用异步任务，负责实例 operation 同步、实例到期提醒、到期释放和通知发送

当前契约重新开放用户账号自助、用户实名、服务器产品目录、订单、支付、钱包、续费订单、实例、通用异步任务、通知和工单。支付一期只开放中国大陆商户的微信 Native 扫码、微信 H5、支付宝电脑网页支付、支付宝手机网页支付和钱包余额支付；发票、JSAPI/openid、小程序支付和部分退款仍不纳入当前阶段交付范围。实例交付和到期释放通过后端内部 MCP PVE client API 调用上游 PVE 适配服务，不向前端暴露 PVE 节点、存储、磁盘来源或 VMID。

## 路由边界

- 健康检查：`/healthz`
- 管理端 API：`/admin-api/*`
- Web 公开配置 API：`GET /api/site-config`
- Web 用户认证 API：`/api/auth/*`
- Web 用户资料和实名 API：`/api/user/*`
- Web 实名供应商回调 API：`POST /api/real-name/provider-callbacks/{provider}`
- Web 服务器产品目录 API：`GET /api/server-catalog`
- Web 订单 API：`/api/orders/*`
- Web 支付 API：`/api/orders/*/payments`、`/api/payments/*`、`/api/payment-callbacks/*`
- Web 实例 API：`/api/instances/*`
- Web 工单 API：`/api/tickets/*`

当前仓库除站点配置、用户账号自助、支付宝/微信侧用户实名、服务器产品目录、订单、支付、实例和工单外，不再把其它 `/api/*` 作为现行后端契约。管理端异步任务只通过 `/admin-api/async-tasks/*` 暴露，不对用户端开放。

## 事实来源

- MariaDB：基础后台事实来源
- Redis：运行时辅助依赖

Redis 可以保存缓存、限流、验证码、一次性 token、短锁和短期状态。它不能替代管理端会话最终状态、RBAC 关系或审计事实。

## 通用安全边界

跨端通用安全基线见 `docs/security.md`。

- 后端是鉴权、权限、资源归属和业务状态的最终裁决点。
- 前端权限判断、JWT 权限快照和请求参数不得替代当前数据库 RBAC、会话状态和资源归属校验。
- 安全增强如果改变接受输入、拒绝输入、返回内容、审计内容、授权方式、配置要求、存储方式或事务边界，必须先更新 `docs/security.md` 和对应 owner docs。

## 目标后端架构

后端目标形态采用“分层架构 + 领域子包”。当前使用 API 和 Worker 两类进程，不引入微服务，也不把每个功能都套成重型 DDD 模块。

这个结构优先解决当前代码的问题：全局 DTO、全局 model、admin/web 两边重复业务规则、handler/service/repository 边界不清。目标是让代码最好写、最好找、最好渐进迁移。

核心原则：

- 入口按运行进程划分：`cmd/api`、`cmd/worker`、`cmd/setup-admin`。
- `delivery/http` 负责 Gin、路由、中间件、请求绑定、响应写入和错误映射。
- `usecase` 负责业务用例、事务边界、权限裁决、幂等、跨仓储协作和用例输入输出类型；该层也必须按 `admin` 和 `web` 两条访问边界拆分，不在 `usecase` 根目录平铺业务包。
- `domain` 保存领域模型、状态机、领域规则和值对象，不依赖 Gin、GORM、Redis、配置结构或第三方 SDK。
- `repository/mysql` 负责 GORM model、可复用查询对象和 MariaDB 持久化实现；既有事务密集型用例允许在迁移期继续直接使用 GORM，但不得把 GORM 暴露到 handler。
- `integration` 负责外部系统协议适配，例如实名供应商、邮件和对象存储。
- 基础设施进入 `platform`，跨模块稳定基础能力进入 `shared`。
- 不再新增全局 `admin/dto`、`web/dto` 或按入口命名的全局 `models` 包。

目标结构如下：

```text
server/
  cmd/
    api/
    worker/
    setup-admin/
  internal/
    app/
      api/
      worker/
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
        order/
        instance/
        asynctask/
        ticket/
        dto/
        support/
      web/
        auth/
        userprofile/
        siteconfig/
        realname/
        catalog/
        order/
        instance/
        ticket/
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
      order/
      instance/
      ticket/
    repository/
      mysql/
        iam/
        user/
        audit/
        systemconfig/
        file/
        realname/
        catalog/
        order/
        instance/
        ticket/
    integration/
      realname/
      mcppve/
      mail/
      storage/
    platform/
      config/
      database/
      cache/
      logger/
    shared/
      errors/
      jwt/
      password/
      pagination/
      requestcontext/
      response/
      validator/
  migrations/
```

## 目录职责

### `internal/app`

进程装配层。只负责读取配置、初始化基础设施、组装 handler/usecase/repository/integration 依赖、创建 HTTP 服务，不承载业务规则。

### `internal/delivery/http`

HTTP 传输层。承载 Gin router、全局中间件、管理端和用户端路由组合。

- `router/`：创建 Gin 引擎、挂载健康检查、`/admin-api/*` 和 `/api/*`
- `middleware/`：请求 ID、恢复、访问日志、CORS、鉴权入口和限流入口；鉴权中间件允许读取数据库会话与 RBAC 事实，但不得承载业务状态流转
- `admin/`：管理端 handler 和路由聚合，只挂载 `/admin-api/*`
- `web/`：用户端 handler 和路由聚合，只挂载 `/api/*`

该层不得直接访问数据库，不写业务状态机。

### `internal/usecase`

应用用例层。先按访问边界拆为 `admin` 与 `web`，再在边界内按领域子包组织业务编排。

- `usecase/admin/*`：只服务 `/admin-api/*` 管理端用例。
- `usecase/web/*`：只服务 `/api/*` 用户端用例。
- `usecase` 根目录不新增业务用例包；跨端共享规则优先进入 `domain/*`，外部协议进入 `integration/*`，稳定通用能力进入 `shared/*`。
- `usecase/admin/*` 和 `usecase/web/*` 不互相导入；需要复用时通过 `domain`、`repository/mysql` 或 `integration` 的稳定边界完成。

当前用例边界：

- `usecase/admin/auth`、`adminuser`、`adminrole`、`adminsession`：管理端认证、管理员账号、管理员角色、权限码、管理员会话
- `usecase/admin/webuser`：管理端用户账号管理
- `usecase/admin/audit`：后台普通操作审计
- `usecase/admin/systemconfig`：系统配置维护
- `usecase/admin/fileattachment`：文件附件、文件引用和本地存储编排
- `usecase/admin/realname`：实名申请管理、供应商同步和人工审核
- `usecase/admin/productcatalog`：服务器产品、套餐、价格、销售地域、系统模板维护
- `usecase/admin/order`：订单列表、详情、后台备注、取消和关闭
- `usecase/admin/instance`：实例交付映射、订单交付、实例列表、实例详情、MCP 只读资源、实例操作和同步
- `usecase/admin/ticket`：工单列表、详情、回复、关闭和附件访问
- `usecase/web/auth`、`userprofile`：用户账号、用户会话、用户资料、密码找回
- `usecase/web/siteconfig`：公开站点配置读取
- `usecase/web/realname`：当前用户个人实名申请、供应商会话和同步
- `usecase/web/catalog`：公开服务器产品目录
- `usecase/web/order`：当前用户创建订单、订单列表、订单详情和取消订单
- `usecase/web/instance`：当前用户实例列表、实例详情、开机和关机
- `usecase/web/ticket`：当前用户创建工单、工单列表、工单详情、回复、关闭和附件访问

`usecase` 可以依赖 `domain`、仓储接口、外部集成接口和 `shared`。它不依赖 Gin，也不直接返回 HTTP DTO。

### `internal/domain`

领域层。按领域子包保存实体、值对象、状态机、纯业务规则和领域错误。

`domain` 不依赖 Gin、GORM、Redis、配置结构、HTTP DTO 或第三方 SDK。跨管理端和用户端共享的实名状态裁决、产品状态规则、会话状态规则等应优先进入这里。

### `internal/repository`

持久化层。当前只维护 `repository/mysql`，按领域子包放置 GORM model、查询对象和 repository 实现。

- GORM model 只能在 `repository/mysql/*` 中定义和使用。
- 对外响应不得直接返回 GORM model。
- repository 不负责权限裁决、审计语义或跨领域业务编排。

### `internal/integration`

外部系统适配层。保存第三方协议、SDK 包装和外部错误映射。

- `realname/`：支付宝/微信侧实名供应商适配
- `mcppve/`：MCP PVE client API 适配，仅封装当前上游已提供的节点、存储、VM 和异步操作接口
- `mail/`：邮件发送适配
- `storage/`：本地或对象存储适配

该层只做协议适配，不做业务状态裁决。

### `internal/platform`

基础设施层。

- `config/`：配置读取、默认值和配置校验
- `database/`：数据库初始化和连接池
- `cache/`：缓存客户端与基础封装
- `logger/`：日志初始化与封装

### `internal/shared`

仅存放稳定、无明确业务语义、被多边界长期复用的基础能力。

允许进入 `shared` 的典型内容：

- `errors`
- `jwt`
- `pagination`
- `password`
- `requestcontext`
- `response`
- `validator`

`shared` 不得依赖 `delivery`、`usecase`、`domain`、`repository`、`integration` 或 `platform` 中的业务适配。

## 依赖方向

允许的依赖方向：

```text
cmd -> internal/app -> delivery/http
delivery/http/admin -> usecase/admin
delivery/http/web -> usecase/web
delivery/http/auth middleware -> repository/mysql models
usecase/admin -> domain
usecase/admin -> repository ports
usecase/admin -> integration ports
usecase/web -> domain
usecase/web -> repository ports
usecase/web -> integration ports
repository/mysql -> domain
integration -> external SDK/client
internal/app -> platform
all layers -> shared
```

禁止事项：

- `domain` 依赖 Gin、GORM、Redis、配置结构或第三方 SDK。
- `delivery/http` handler 直接读写数据库；鉴权中间件只能读取和维护会话活跃状态。
- `repository/mysql` 调用 `usecase` 或写业务状态机。
- `usecase/admin/*` 调用 `usecase/web/*`，或 `usecase/web/*` 调用 `usecase/admin/*`。
- 在 `usecase` 根目录新增承载业务流程的包。
- 在 `delivery/http` 中写业务状态变更。
- 在 `platform` 中引入具体业务规则。
- 新增泛名 `common`、`helper`、`manager`、`base` 包承载业务逻辑。

跨领域协作必须在所属访问边界内由明确的 `usecase/admin/*` 或 `usecase/web/*` 用例完成。跨领域写操作需要明确事务拥有方；跨领域读模型可以在 owning usecase 提供查询用例，不能散落在调用方手写 SQL。跨端共享事实和规则不得通过 admin/web usecase 互调实现，应下沉到 `domain`、`repository/mysql` 或 `integration`。

## 模型和 DTO 规则

- 用例输入输出类型放在 `usecase/admin/dto`、`usecase/web/dto` 或对应边界内的具体 `usecase/admin/<domain>`、`usecase/web/<domain>`。
- Handler 只做请求绑定、校验和响应写入，不在 delivery 层定义业务输出结构。
- 领域实体和值对象放在 `domain/<domain>`。
- GORM model、表字段映射和可复用 SQL 查询对象放在 `repository/mysql/<domain>`。
- 不再新增全局 DTO 包。
- 不再把用户端、管理端和产品域的持久化结构继续放入 `admin/models` 这类按入口命名的包。
- 对外响应不得直接返回 GORM model。

## 迁移优先级

当前代码迁移到目标架构时按风险从低到高推进：

1. 清理空临时目录、本地构建产物和历史兼容但未使用的入口。
2. 将配置、数据库、Redis、日志等基础设施迁入 `platform`，保留兼容薄包装直到调用方迁完。
3. 将管理端和用户端路由聚合迁入 `delivery/http/admin` 与 `delivery/http/web`。
4. 将全局 DTO 移入 `usecase/admin/dto` 或 `usecase/web/dto`，并把用例服务按访问边界拆到 `usecase/admin/*` 与 `usecase/web/*`。
5. 将 `admin/models` 移入 `repository/mysql/models`，后续按风险拆为更细的 `repository/mysql/<domain>` model。
6. 将实名同步、供应商结果收敛和配置解析分别收口到 `usecase/admin/realname`、`usecase/web/realname`、`domain/realname` 与 `integration/realname`，纯领域裁决继续下沉到 `domain/realname`。
7. 最后收敛启动装配到 `internal/app/api`，保留 `cmd/api` 只作为薄入口。

当前迁移状态：入口装配、HTTP 路由/中间件、用例输入输出类型、配置、数据库连接、Redis 连接、实名/邮件外部适配和 GORM model 已迁入目标层级；过渡目录 `internal/admin` 与 `internal/web` 已移除。下一步必须把当前 `usecase` 根目录下的业务包继续拆入 `usecase/admin/*` 与 `usecase/web/*`，并清空根目录业务包。现阶段仍允许按风险渐进拆分更细的 `domain/*` 与 `repository/mysql/*` 子包，但不得再新增旧式入口分端目录。

## 领域边界

当前后端领域边界以管理端基础能力为主，同时已重新开放用户账号自助、支付宝/微信侧用户实名、服务器产品目录、订单、支付、续费订单、实例、异步任务、通知和工单：

- 管理端认证
- 会话
- RBAC
- 系统配置
- 审计写入
- 文件管理

用户账号自助、支付宝/微信侧用户实名、服务器产品目录、订单、支付、钱包、续费订单、实例、通用异步任务、通知和工单已经重新纳入当前阶段契约。发票、部分退款、JSAPI/openid、小程序支付和其它用户端业务流仍从当前阶段契约中收口。实例只覆盖 MCP PVE client API 当前具备的创建、查询、启动、停止、删除和异步操作查询能力。

## 鉴权与权限

### 管理端

- 使用管理端 JWT secret 和 issuer
- JWT 必须带 `jti`
- `jti` 对应 `admin_sessions.session_id`
- 受保护管理端接口必须同时校验：
  - 签名
  - issuer
  - token type
  - 过期时间
  - 会话状态
  - 管理员状态
  - 当前数据库 RBAC

管理端前端可以消费权限快照改善体验，但后端 RBAC 仍是最终裁决。

## 当前管理端阶段边界

当前后端基础能力保留以下管理域：

- auth
- dashboard
- admin users
- roles and permissions
- admin sessions
- system configs
- audit logs

当前开放的管理端 API 和前端页面范围为：

- `Login`
- `Dashboard`
- `System Settings`
- `File Management`
- `Web User Management`
- `Real Name Management`
- `Product Management`
- `Order Management`
- `Payment Management`
- `Async Tasks`
- `Ticket Management`
- `Instance Management`
- `403`

`System Settings` 当前包含系统配置、管理员账号、管理员组权限和管理员会话。
`System Settings` 下同时开放操作日志页面，用于查看普通后台操作日志。
当前重新开放服务器产品目录、用户实名、订单、支付、钱包、实例、工单、通用异步任务和实例生命周期相关数据库契约；实名供应商集成只覆盖支付宝/微信侧实名核验，支付供应商集成只覆盖微信支付、支付宝支付的一期网页/扫码交易能力和钱包余额支付。实例集成只覆盖 MCP PVE client API 已提供能力，不开放通用 PVE 运维管理。

## 订单与实例交付

- 用户端订单只挂载 `/api/orders/*`，管理端订单只挂载 `/admin-api/orders/*`。
- 用户端可基于当前公开服务器产品目录创建订单、查看自己的订单列表和详情，并取消自己的 `pending` 订单。
- 管理端只查看和处理用户端订单，不支持后台创建订单。
- 管理端可查看订单列表和详情，编辑后台备注，取消或关闭订单。
- 订单状态包含 `pending`、`provisioning`、`fulfilled`、`error`、`cancelled`、`closed`。
- 订单类型包含 `purchase` 和 `renewal`；续费订单只延长已有实例服务期，不创建新实例。
- 支付状态包含 `unpaid`、`paid`、`manual_confirmed`、`refunded`；真实支付流水以支付交易表为准，订单支付字段只作为列表和详情摘要。
- 订单金额使用分为单位，创建订单时由后端基于当前产品、套餐、计费周期、销售地域和系统模板重新计算。
- 订单必须保存产品、套餐、价格、销售地域和系统模板快照，后续产品目录变化不得改变历史订单事实。
- 当 `real_name.required_for_order=true` 时，订单创建必须要求当前用户实名状态为 `approved`。
- 新购订单不扣减库存。管理员仍可按人工流程触发交付；真实支付成功的新购订单由支付成功处理投递自动交付任务，任务失败后订单进入 `error` 且保留 `payment_status=paid`，等待管理端重试。
- 续费订单由用户端创建，管理端人工确认后延长实例 `expires_at`；真实支付成功后必须复用同一套续期确认规则，并写入支付生效记录。
- 用户端订单接口不返回 PVE 节点、存储、磁盘来源、VMID 或上游 operation ID。
- 管理端备注、取消、关闭、人工续费确认和触发交付必须写入普通后台操作审计。

## 支付与退款

- 用户端支付只挂载 `/api/orders/{order_no}/payments` 和 `/api/payments/{payment_no}`；用户端钱包只挂载 `/api/wallet/*`；公开回调只挂载 `/api/payment-callbacks/alipay` 和 `/api/payment-callbacks/wechat`。
- 管理端支付只挂载 `/admin-api/payments/*` 和 `/admin-api/refunds/*`，并通过独立支付管理菜单开放。
- 管理端钱包只挂载 `/admin-api/wallets/*`、`/admin-api/wallet-ledger` 和 `/admin-api/wallet-recharges`，并通过只读钱包管理菜单开放。
- 支付供应商允许值为 `alipay`、`wechat` 和 `wallet`；支付方式允许值为 `alipay_page`、`alipay_wap`、`wechat_native`、`wechat_h5` 和 `wallet_balance`。
- 支付宝适配使用成熟 Go SDK `github.com/smartwalle/alipay/v3`；微信支付适配使用微信支付 API v3 官方 Go SDK `github.com/wechatpay-apiv3/wechatpay-go`。生产路径不得使用 mock adapter，自研签名和验签只允许出现在测试辅助中。
- 支付交易状态包含 `pending`、`paid`、`closed`、`failed`、`refunded`；退款状态包含 `pending`、`succeeded`、`failed`。
- 用户只能为自己的 `pending` 且 `payment_status=unpaid` 订单创建支付；支付金额必须等于订单应付金额，币种一期固定为 `CNY`。
- 支付创建幂等依赖 `order_no + provider + method + client_token`；重复提交同一幂等键返回已有支付，不重复创建上游交易。
- 支付回调不要求 Bearer Token，但必须通过供应商验签、金额校验、交易归属校验和本地状态校验；重复回调只返回成功确认，不重复交付、续费或退款。
- 支付渠道启用前必须完成配置完整性校验：支付宝需要应用 ID、网关、应用私钥、支付宝公钥、异步通知地址和同步返回地址；微信需要应用 ID、商户号、API v3 key、商户私钥、商户证书序列号、平台公钥或平台证书、异步通知地址，微信 H5 还需要 H5 场景信息。缺少必要项时不得创建上游交易、主动查询或发起退款。
- 渠道下单、主动查询和退款属于外部副作用，不放入持有订单或支付行锁的长事务；本地支付交易创建、状态推进、退款本地回滚、支付生效记录和任务投递仍必须在本地事务中完成。
- 支付成功后在本地事务中锁定订单和支付交易：订单支付摘要更新为 `payment_status=paid`、`paid_at`、`payment_provider` 和 `payment_trade_no`；真实流水和回调摘要只写入支付表。
- 新购支付成功后投递自动交付任务；续费支付成功后延长实例服务期并写入支付生效记录，记录续费前后 `expires_at`。
- 一期只支持全额退款。新购已交付订单必须先释放实例后才能退款；续费退款必须先校验支付生效记录仍可扣回。
- 支付宝/微信退款采用“渠道成功后本地回滚”：先创建 `pending` 退款并调用渠道退款；渠道退款成功或查询确认后，再同事务扣回续费时间、更新退款/支付/订单状态和写审计。渠道失败不得扣用户服务期。
- 钱包余额支付的退款不调用外部渠道，必须同事务锁定钱包账户、退款、支付、订单和必要的实例/生效记录，把支付金额退回钱包余额并写入钱包流水；真实支付订单不允许退款到钱包余额。
- 支付、退款、配置变更、渠道同步和自动交付失败重试必须写后台审计；任务 payload/result、审计和日志只保存摘要，不保存商户密钥、签名串、完整回调 payload 或完整上游响应。

## 钱包

- 钱包 v1 只支持用户充值、余额支付订单、余额支付退款退回钱包和管理端只读查看。
- 钱包币种固定为 `CNY`，金额使用分为单位，不使用浮点数；钱包账户按用户和币种唯一，首次读取、充值或余额支付时懒创建。
- 钱包当前余额以 `wallet_accounts.available_balance_cents` 为最终事实，`wallet_ledger_entries` 是追加式账本，不更新、不删除，用于审计和排查。
- 用户充值通过支付宝/微信现有支付 adapter 创建上游交易；充值成功回调必须锁定充值记录和钱包账户，只允许 `pending -> paid` 入账一次，重复回调幂等跳过。
- 余额支付必须锁定订单和钱包账户，校验订单 `pending`、`payment_status=unpaid`、钱包 `active` 且余额足够；扣款、写入 `payment_transactions`、写钱包流水和推进订单生效必须在同一事务内完成。
- 余额支付成功的新购订单仍投递 `payment_order_provision`，续费订单仍复用当前续费生效规则并写支付生效记录。
- 管理端钱包 v1 只读，不支持人工加款、扣款、冻结、提现、退款入钱包或余额转账。

## 实例

- 用户端实例只挂载 `/api/instances/*`，管理端实例只挂载 `/admin-api/instances/*`。
- MCP PVE client API 只作为后端内部上游，不注册为 pveCloud 对外 `/api/pve/*` 路由。
- 管理端可从 `pending` 订单触发交付，服务端读取 `instance_provision_mappings` 分配 VMID，并调用 MCP `POST /api/pve/nodes/{node}/vms`。
- 实例状态包含 `creating`、`running`、`stopped`、`error`、`releasing`、`released`。
- 用户端可查看自己的实例列表和详情，可对 `stopped` 实例开机，可对 `running` 实例关机。
- 管理端可查看全部实例、触发开机、关机、释放和同步，可查看内部 `node`、`vmid`、operation 状态和失败原因。
- 释放实例调用 MCP 删除 VM；释放后的实例保留本地记录，不复用 `instance_no`。
- 异步操作通过 `instance_operations` 保存，本地状态以 MariaDB 为最终事实；MCP operation 查询只用于同步上游结果。
- 实例服务期通过 `service_started_at`、`expires_at` 和到期释放相关字段管理。到期提醒、自动释放和 operation 同步由 Worker 执行。
- 自动释放必须受 `instance_lifecycle.auto_release_enabled` 控制；关闭时不得删除上游 VM。
- 当前不开放 MCP 未提供的重启、重装、重置密码、控制台、快照、备份、迁移、监控、网络防火墙和资源池管理。

## 异步任务与 Worker

- API 进程负责任务投递，Worker 进程负责领取并执行 `async_tasks`。
- Worker 首批执行实例 operation 同步、实例到期提醒、到期释放、支付成功后新购自动交付、退款状态同步、邮件通知和短信占位任务。
- Worker 不注册 HTTP 路由，不被反向代理公开。
- 管理端通过 `/admin-api/async-tasks/*` 查看和重试失败任务。
- 任务 payload、result 和日志不得保存 secret、token、SMTP 凭据、MCP Bearer Token 或完整上游响应。

## 工单 MVP

- 用户端工单只挂载 `/api/tickets/*`，管理端工单只挂载 `/admin-api/tickets/*`。
- 用户端可创建通用工单，或可选关联当前用户自己的订单。
- 用户端可查看、回复和关闭自己的工单。
- 管理端可查看全部工单、回复未关闭工单、关闭未关闭工单，并访问工单附件。
- 管理端可对未关闭工单执行指派、转派、协作者维护、优先级升级和标签绑定。
- 管理端可对工单追加内部备注；内部备注不返回用户端。
- 工单状态只包含 `waiting_admin`、`waiting_user`、`closed`。
- 工单分类只包含 `account`、`order`、`product`、`technical`、`billing`、`other`。
- 工单优先级只包含 `low`、`normal`、`high`、`urgent`，默认 `normal`。
- 内部 SLA 按优先级固定计算首次响应和解决截止时间，仅管理端展示逾期状态，不作为用户端承诺。
- 工单标签由管理端标签字典维护，公开标签可返回用户端，内部标签只返回管理端。
- 工单附件复用文件附件存储和安全校验；附件访问必须经过工单归属或管理端权限校验。
- 用户端关联订单时必须校验订单属于当前登录用户。
- 工单回复和附件引用必须同事务写入。
- 管理端回复、关闭、指派、转派、协作者维护、内部备注、优先级升级、标签绑定和标签字典变更必须写入普通后台操作审计。
- 工单不承载支付处理、PVE 节点、资源池、库存扣减或自动交付能力；如需关联实例排障，必须后续先补工单与实例关联契约。

## 审计日志

- 普通后台操作写入 `admin_audit_logs`
- 普通操作日志用于还原后台操作历史，应包含操作者快照、管理端会话、请求 ID、请求方法和请求路径
- 普通操作日志的请求上下文由管理端中间件统一采集并写入 Gin 上下文，业务模块只传业务动作、对象、前后快照和备注
- 运行时日志、访问日志、后台操作审计和管理端日志管理的边界见 `logging.md`

## 文件管理

- 文件上传属于管理端边界，只开放在 `/admin-api/*`
- 上传文件必须经过扩展名、声明 MIME、Magic Bytes 和危险扩展名校验
- 上传读取必须限制最大字节数，避免超大文件被完整读入内存
- 原始文件名只用于展示，存储文件名必须由服务端随机生成
- 数据库只保存相对存储路径，存储根目录来自 `storage.local_path` 配置
- 上传记录和上传审计必须同事务写入；事务失败时清理已写入的物理文件
- 工单附件可以由用户端或管理端上传，但下载和预览必须通过工单附件接口按资源归属或管理端权限裁决
- 删除文件使用软删除，状态变更和删除审计必须同事务写入

## 当前不在范围内的能力

以下能力不属于当前阶段契约：

- 通用 PVE 运维管理
- 真实短信供应商集成
- 发票、JSAPI/openid、小程序支付、部分退款、提现、人工调账、余额转账和自动对账批处理
