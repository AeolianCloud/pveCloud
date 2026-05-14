# 数据库设计

可执行表结构最终以 `server/migrations/` 为准。
本文件记录当前基础后台阶段的数据库契约。

## 基础环境

```text
database: pvecloud
engine: MariaDB 11.4.x / InnoDB
charset: utf8mb4
collation: utf8mb4_unicode_ci
```

## 设计原则

- 主键使用 `BIGINT UNSIGNED AUTO_INCREMENT`
- 状态字段使用字符串，不使用数据库 enum
- 表和字段写明 `COMMENT`
- 常规时间字段使用 `DATETIME(3)`
- 对外展示优先使用业务编号，不直接暴露自增 ID
- JSON 字段只用于低频配置片段或审计快照

## 当前表分组

### 基础后台账号、菜单与权限

```text
admin_users
admin_roles
admin_permissions
admin_user_roles
admin_role_permissions
admin_sessions
```

### 基础后台配置与审计

```text
system_configs
admin_audit_logs
```

### 用户端认证

```text
users
user_sessions
user_password_reset_tokens
```

`users` 用于用户端账号。当前阶段开放用户注册、登录、资料编辑、密码修改、个人实名、订单 MVP 和工单 MVP，不开放钱包、支付、实例或其它业务资料。

`user_sessions` 用于记录用户端登录会话。用户端 access token 的 `jti` 必须对应 `user_sessions.session_id`，服务端受保护用户接口需要校验 token 和会话状态。

`user_password_reset_tokens` 用于记录用户端密码找回的一次性 token。数据库只保存 token 哈希，不保存 token 原文；token 原文只通过邮件重置链接发送给用户。

密码重置 token 状态使用字符串字段，不使用数据库 enum。可用状态包括：

- `active`：可使用
- `used`：已使用
- `revoked`：已吊销
- `expired`：已过期

同一用户同时只能保留一个可用的 active 密码重置 token。用户申请新的密码重置 token 时，应吊销旧 active token 或复用未过期请求。密码重置成功后必须将 token 标记为 used，并吊销该用户所有 active 用户端会话。

### 文件管理

```text
file_attachments
```

`file_attachments` 用于存储管理端和用户端上传的图片和附件元信息。
文件物理存储在本地磁盘，数据库只记录元信息和关联关系。
状态使用字符串，不使用数据库 enum。
数据库中的 `storage_path` 只保存相对路径，格式为 `{YYYY}/{MM}/{DD}/{uuid}.{ext}`，不得保存本地存储根目录。
存储文件名强制使用随机 UUID，禁止用户控制存储路径；原始文件名只用于展示，入库前必须去除路径片段和空字节。
管理端上传文件记录与上传审计日志必须在同一事务中写入；事务失败时应清理已落盘的物理文件。用户端工单附件上传不写入后台审计，但必须写入上传者用户 ID，并与工单消息附件引用同事务建立业务引用。
删除文件采用软删除，状态变更与删除审计日志必须在同一事务中写入。

### 文件引用关系

```text
file_attachment_references
```

`file_attachment_references` 用于记录文件被哪些业务对象引用。
工单附件必须写入 `ref_type=ticket_message` 的引用记录，用于阻止仍被工单引用的附件被删除。
删除文件前必须先检查引用关系；存在引用时不允许删除。

### 用户实名

```text
user_real_name_applications
```

`user_real_name_applications` 用于保存用户端个人实名申请、外部供应商核验会话和人工审核申请。当前只开放个人实名，并支持支付宝、微信侧实名供应商和后台人工审核；不开放企业实名或后台代填实名资料。

实名状态使用字符串字段，不使用数据库 enum。可用申请状态包括：

- `pending`：待外部核验结果
- `approved`：实名通过
- `rejected`：实名拒绝或核验失败

`pending` 可表示等待用户完成外部供应商核验、等待供应商回调/查询结果，或等待后台人工审核。

证件号码不得明文存储。数据库只保存脱敏展示值；外部供应商模式必须保存带后台敏感配置密钥的证件号码查询摘要，接口只返回脱敏展示值。人工审核模式在未配置摘要密钥时允许 `id_number_digest` 为空。证件摘要不得使用无盐 SHA-256。外部供应商模式新增记录必须保存当前 HMAC 摘要版本；历史 `sha256-legacy` 摘要只用于兼容旧已通过记录的重复校验。

实名供应商字段保存当前申请来源，允许值包括：

- `alipay`
- `wechat`
- `manual`

供应商会话字段保存外部会话标识、供应商状态、供应商结果码、供应商结果说明、开始时间、完成时间、响应摘要和链路号。不得保存外部供应商完整响应、证件号码明文、人脸图片、token 或签名密钥。

当前人工审核流程不写入实名图片附件；历史人工实名附件如已存在，只作为历史数据保留，不提供新的上传或预览入口。通用文件管理权限不得绕过历史实名附件的敏感数据边界。

同一用户存在 `pending` 申请时不得重复提交；已 `approved` 后不得由用户端自行覆盖实名资料。证件号码与其它已通过实名用户重复时必须拒绝通过或提交。

外部供应商回调和主动同步必须按 `verification_provider + provider_application_id` 幂等处理。已通过证件摘要唯一性由已通过摘要生成列和唯一索引兜底；HMAC 摘要切换期间，同一证件号码的 legacy 摘要和当前 HMAC 摘要不同，服务端仍必须同时查询两个摘要版本防止历史重复通过。

### 服务器产品目录

```text
products
product_plans
plan_prices
sales_regions
server_os_templates
plan_regions
plan_os_templates
```

服务器产品目录用于维护 Web 可展示的固定服务器套餐，不包含订单、支付、实例、库存扣减或 PVE 节点绑定。

`products` 表示产品主数据，当前只开放 `type=server` 的云服务器产品。

`product_plans` 表示固定服务器套餐，保存 CPU、内存、磁盘、带宽、流量、公网 IP、虚拟化和架构等销售规格。

`plan_prices` 保存套餐周期价格，金额字段使用分为单位，不使用浮点数。

`sales_regions` 表示销售地域，只用于展示和可售约束，不等同于 PVE 节点、集群或资源池。

`server_os_templates` 表示服务器系统模板，避免与图片、Logo、附件等 image 概念混淆；当前不绑定 PVE 模板 ID。

`plan_regions` 和 `plan_os_templates` 分别维护套餐可用销售地域和可用服务器系统模板。

产品目录状态使用字符串字段，不使用数据库 enum。产品和套餐对外展示使用 `product_no`、`plan_no`、`template_no`、`region_no` 等业务编号，不直接暴露自增 ID。

产品目录删除采用受限硬删除，不新增软删除字段。产品存在套餐时不可删除；销售地域或服务器系统模板仍被套餐关联时不可删除；套餐删除必须同事务删除 `plan_prices`、`plan_regions` 和 `plan_os_templates` 中的关联数据。历史订单保存产品目录快照，不通过外键引用当前产品目录，因此历史订单不阻止产品目录删除。

### 订单 MVP

```text
orders
```

`orders` 用于保存用户端基于服务器产品目录创建的订单最终事实。订单只表示购买意向和后台人工处理入口，不代表支付成功、实例开通或资源交付。

订单状态使用字符串字段，不使用数据库 enum。第一阶段只允许以下状态：

- `pending`：订单已创建，等待运营处理或后续人工流程。
- `cancelled`：订单已取消，可由当前用户或管理员取消。
- `closed`：订单已关闭，只能由管理员关闭。

本阶段不保存或使用 `paid`、`provisioning`、`fulfilled` 等支付和交付状态。

订单对外展示使用 `order_no`，不直接暴露自增 ID。金额字段使用分为单位，不使用浮点数。

创建订单时必须保存以下快照，后续产品目录变化不得改变历史订单事实：

- 产品编号、类型、名称、简介
- 套餐编号、编码、名称、规格
- 计费周期、价格金额、原价金额和币种
- 销售地域编号、编码、名称
- 系统模板编号、编码、名称、系统族、发行版、版本和架构

订单创建必须基于当前产品目录重新校验可售组合并计算金额，不能信任前端提交的金额、产品名称、套餐名称、地域名称或系统模板名称。

当 `real_name.required_for_order=true` 时，订单创建必须要求当前用户实名状态为 `approved`。用户端取消订单只能取消自己的 `pending` 订单。管理端只处理用户端订单，不支持后台创建订单。

订单创建幂等依赖客户端提交的 `client_token` 和当前用户组合唯一约束。重复提交同一个有效幂等键时，应返回同一订单或明确的 `409xx` 冲突，不得重复创建订单。

订单状态变更必须检查当前状态并在事务中写入。管理端取消、关闭和后台备注变更必须与普通后台审计写入保持同事务。

### 工单 MVP

```text
tickets
ticket_messages
ticket_message_attachments
ticket_tags
ticket_tag_bindings
ticket_internal_notes
ticket_collaborators
ticket_events
```

`tickets` 用于保存用户端提交的工单主记录。工单只表示用户与后台之间的沟通事项和后台内部处理协作，不代表支付、实例交付、PVE 操作或自动处理承诺。内部 SLA 仅用于后台处理时限，不作为用户端服务承诺。

工单状态使用字符串字段，不使用数据库 enum。第一阶段只允许以下状态：

- `waiting_admin`：等待后台处理。创建工单或用户回复后进入该状态。
- `waiting_user`：等待用户反馈。管理员回复后进入该状态。
- `closed`：工单已关闭。用户或管理员均可关闭未关闭工单。

工单分类固定为 `account`、`order`、`product`、`technical`、`billing`、`other`。工单优先级固定为 `low`、`normal`、`high`、`urgent`，默认 `normal`。

内部 SLA 按工单优先级在创建时写入首次响应和解决截止时间。默认规则为：`low` 首响 48 小时、解决 7 天；`normal` 首响 24 小时、解决 5 天；`high` 首响 8 小时、解决 3 天；`urgent` 首响 2 小时、解决 24 小时。管理员首次回复时写入首次响应时间。关闭工单时视为解决并写入解决时间。逾期状态由当前时间和截止/完成时间计算，不依赖后台定时任务。

工单对外展示使用 `ticket_no`，不直接暴露自增 ID。工单可选关联 `orders.id`，并保存 `order_no` 快照；用户端关联订单时必须校验订单属于当前登录用户。

工单当前处理人使用 `assignee_admin_id` 关联 `admin_users.id`。可指派对象只允许 `active` 且具备 `page.tickets`、`ticket:reply` 或 `ticket:*` 的管理员。关闭后的工单不得指派、转派、维护协作者或升级优先级。

`ticket_messages` 保存工单消息。消息发送方使用 `sender_type=user|admin` 和对应发送者 ID 表示。关闭后的工单不得继续写入新消息。

`ticket_message_attachments` 保存工单消息与 `file_attachments` 的关联。单条消息最多 5 个附件。工单附件同时必须写入 `file_attachment_references`，`ref_type=ticket_message`，引用 ID 使用消息 ID。

`ticket_tags` 保存管理端标签字典。标签分 `public` 和 `internal`；公开标签可返回用户端，内部标签只返回管理端。停用标签不可新绑定，但历史绑定仍可展示。

`ticket_tag_bindings` 保存工单与标签的多对多关系。标签绑定变更必须整体替换当前绑定集合，并记录事件和后台审计。

`ticket_internal_notes` 保存管理端内部备注。内部备注只追加，不编辑、不删除，不返回用户端。关闭后的工单仍允许追加内部备注。

`ticket_collaborators` 保存工单协作者。协作者只返回管理端。关闭后的工单不得增删协作者。

`ticket_events` 保存工单内部操作历史，记录指派、转派、协作者变更、内部备注、优先级升级、标签绑定、回复和关闭等后台处理事件。

优先级升级只能从低紧急度升到更高紧急度，必须填写原因，不支持降级。升级后未完成的 SLA 截止时间只允许提前，不允许延后。

工单创建、回复、附件关联和工单主表最近消息时间更新必须在本地数据库事务中完成。管理端回复、关闭、指派、转派、协作者维护、内部备注、优先级升级、标签绑定和标签字典变更必须与普通后台审计写入保持同事务。

## 管理端关键规则

- 管理端专用表使用 `admin_` 前缀
- `admin_permissions` 是管理端菜单和操作权限的唯一目录来源
- 权限节点分为 `type=menu` 和 `type=action`
- 菜单权限使用 `page.<menu>.<feature>`，控制菜单可见、路由访问和页面主数据读取
- 操作权限使用 `resource:action`，控制按钮、写接口、危险操作和敏感详情读取
- 操作权限必须通过 `parent_code` 挂到明确菜单节点
- 菜单树通过 `parent_code`、`path`、`icon`、`sort_order` 和 `visible_in_menu` 生成
- 管理端会话最终状态以 `admin_sessions` 为准
- `super_admin` 角色应始终拥有当前 `admin_permissions` 中定义的全部权限
- JWT 中的角色和权限快照只用于登录响应与前端体验，不替代服务端当前 RBAC 校验
- 角色权限分配和管理员角色分配属于二次授权入口，服务端必须基于当前数据库 RBAC 限制分配范围，禁止普通管理员分配自己未拥有的权限集合
- `system_configs.is_secret=1` 的配置不得通过接口返回明文
- `system_configs` 中 `site.name` 和 `site.logo_url` 是公开站点基础展示配置，分别控制 Web 左上角品牌文字和 Logo 图片 URL
- 站点品牌配置在系统设置中使用中文分组“站点设置”展示
- `system_configs` 中 `web.auth.login_captcha_enabled`、`web.auth.register_captcha_enabled`、`web.auth.password_reset_request_captcha_enabled`、`web.auth.password_reset_confirm_captcha_enabled` 是用户端认证验证码开关，使用中文分组“用户认证”展示
- 上述 4 个验证码开关使用 `value_type=bool`，`config_value` 统一保存字符串 `true` 或 `false`
- `system_configs` 中 `real_name.*` 是用户实名业务开关、供应商选择、供应商接入参数、第三方密钥、回调地址、返回地址和证件摘要密钥，使用中文分组“实名设置”展示
- 支付宝和微信侧实名供应商配置全部来自 `system_configs`，不使用 `server/config.yaml` 或 `server/config.example.yaml` 管理
- 实名布尔配置使用 `value_type=bool`，数值配置使用 `value_type=int`，允许供应商列表、URL、密钥和说明文案使用 `value_type=string`
- `real_name.identity_digest_secret`、`real_name.alipay.app_private_key`、`real_name.alipay.alipay_public_key`、`real_name.wechat.secret_id` 和 `real_name.wechat.secret_key` 等敏感实名配置必须 `is_secret=1`；后台 API 不得回显明文，公开站点配置接口不得返回这些配置
- 普通操作日志用于查看后台操作历史，应保存操作者快照和请求上下文，避免只依赖当前管理员资料反查
- 普通操作日志的请求上下文由管理端中间件统一采集，业务模块不得重复从每个模块内拼装 IP、会话、请求路径等通用信息

## 普通操作日志上下文

`admin_audit_logs` 除业务动作、对象和前后快照外，还记录以下后台请求上下文：

- `admin_username`：操作发生时的管理员用户名快照
- `admin_display_name`：操作发生时的管理员显示名快照
- `session_id`：触发操作的管理端会话标识
- `request_id`：请求链路 ID，用于串联访问日志、错误日志和审计日志
- `request_method`：后台请求方法
- `request_path`：后台请求路径

这些字段只增强普通操作日志的可读性和排查能力。

## 当前阶段说明

当前仓库已经从“基础后台阶段”继续演进到用户账号自助、用户实名、服务器产品目录、订单和工单阶段。
数据库契约保留以下管理域：

- 认证
- RBAC
- 会话
- 系统配置
- 审计日志
- 用户端账号与会话
- 用户实名
- 服务器产品目录
- 订单 MVP
- 工单 MVP

以下业务域表不属于当前数据库契约，后续如需恢复，必须先补新的迁移和文档确认：

- 支付与钱包
- 实例
- 异步任务

## 关键唯一约束示例

- `admin_roles.code`
- `admin_permissions.code`
- `admin_sessions.session_id`
- `users.username`
- `users.email`
- `user_sessions.session_id`
- `user_password_reset_tokens.token_hash`
- `user_real_name_applications.application_no`
- `user_real_name_applications.approved_id_number_digest`，只约束 `status=approved` 的证件摘要唯一，防止并发核验通过同一证件号码
- `system_configs.config_key`
- `products.product_no`
- `products.slug`
- `product_plans.plan_no`
- `product_plans.code`
- `plan_prices(plan_id, billing_cycle)`
- `sales_regions.region_no`
- `sales_regions.code`
- `server_os_templates.template_no`
- `server_os_templates.code`
- `orders.order_no`
- `orders(user_id, client_token)`，只约束有效幂等键
- `tickets.ticket_no`
- `ticket_message_attachments(message_id, file_id)`

## 管理端权限新增口径

Web 用户管理需要新增以下管理端权限目录：

- `page.web-users`
- `web-user:*`
- `web-user:create`
- `web-user:update`
- `web-user:password-reset`
- `page.web-user-sessions`
- `web-user-session:*`
- `web-user-session:revoke`

其中 `page.web-user-sessions` 是 `page.web-users` 下的非侧栏 tab 权限，不作为独立菜单展示。

产品目录需要新增以下管理端权限目录：

- `page.products`
- `product:*`
- `product:view`
- `product:create`
- `product:update`
- `product:publish`
- `product:delete`

实名管理需要新增以下管理端权限目录：

- `page.real-name-management`
- `real-name:*`
- `real-name:view`
- `real-name:sync`

订单管理需要新增以下管理端权限目录：

- `page.orders`
- `order:*`
- `order:view`
- `order:update`
- `order:cancel`

工单管理需要新增以下管理端权限目录：

- `page.tickets`
- `ticket:*`
- `ticket:reply`
- `ticket:close`

## 一致性原则

- MariaDB 是基础后台事实来源
- Redis 只做缓存、限流、短 TTL 状态和辅助幂等
- 当前阶段不以 PVE、支付、实例或异步任务为现行数据库契约
- 未来创建订单时必须复制产品、套餐、价格、销售地域和服务器系统模板快照，不能只依赖当前产品表引用
