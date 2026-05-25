# 数据库设计

可执行表结构最终以 `server/migrations/` 为准。
本文件记录当前管理端、用户端业务、实例交付、异步任务和运维生命周期相关数据库契约。

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

### 日志管理中心

```text
user_security_logs
user_business_logs
frontend_error_logs
backend_runtime_logs
log_export_records
```

### 用户端认证

```text
users
user_sessions
user_password_reset_tokens
```

`users` 用于用户端账号。当前阶段开放用户注册、登录、资料编辑、密码修改、个人实名、订单、支付、钱包、续费订单、实例和工单，不开放发票或其它业务资料。

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

服务器产品目录用于维护 Web 可展示的固定服务器套餐，不包含支付流水、库存扣减或 PVE 节点直接绑定。实例交付通过独立交付映射把产品目录选择映射到 MCP PVE client API 参数。

`products` 表示产品主数据，当前只开放 `type=server` 的云服务器产品。

`product_plans` 表示固定服务器套餐，保存 CPU、内存、磁盘、带宽、流量、公网 IP、虚拟化和架构等销售规格。

`plan_prices` 保存套餐周期价格，金额字段使用分为单位，不使用浮点数。

`sales_regions` 表示销售地域，只用于展示和可售约束，不等同于 PVE 节点、集群或资源池；实例交付时由 `instance_provision_mappings` 选择上游节点。

`server_os_templates` 表示服务器系统模板，避免与图片、Logo、附件等 image 概念混淆；当前不直接绑定 PVE 模板 ID，实例交付时由 `instance_provision_mappings` 选择上游磁盘来源。

`plan_regions` 和 `plan_os_templates` 分别维护套餐可用销售地域和可用服务器系统模板。

产品目录状态使用字符串字段，不使用数据库 enum。产品和套餐对外展示使用 `product_no`、`plan_no`、`template_no`、`region_no` 等业务编号，不直接暴露自增 ID。

产品目录删除采用受限硬删除，不新增软删除字段。产品存在套餐时不可删除；销售地域或服务器系统模板仍被套餐关联时不可删除；套餐删除必须同事务删除 `plan_prices`、`plan_regions` 和 `plan_os_templates` 中的关联数据。历史订单保存产品目录快照，不通过外键引用当前产品目录，因此历史订单不阻止产品目录删除。

### 订单

```text
orders
```

`orders` 用于保存用户端基于服务器产品目录创建的订单最终事实。订单表示购买意向和后台处理入口，不代表支付成功；管理员触发交付后可关联一条实例记录。

订单状态使用字符串字段，不使用数据库 enum。当前允许以下状态：

- `pending`：订单已创建，等待运营处理或后续人工流程。
- `provisioning`：管理员已触发实例交付，实例创建或同步处理中。
- `fulfilled`：实例交付完成，至少存在一条已成功创建的实例记录。
- `error`：真实支付后的自动交付失败，需要管理端处理或重试。
- `cancelled`：订单已取消，可由当前用户或管理员取消。
- `closed`：订单已关闭，只能由管理员关闭。

订单状态不使用 `paid` 表示支付完成；支付相关事实只进入独立的 `payment_status` 预留字段。

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

订单状态变更必须检查当前状态并在事务中写入。管理端取消、关闭、后台备注变更和交付触发必须与普通后台审计写入保持同事务。

订单类型允许 `purchase` 和 `renewal`。`purchase` 表示新购订单，可由管理端人工触发交付，也可由真实支付成功后自动投递交付任务；`renewal` 表示实例续费订单，必须关联当前用户自己的未释放实例，不创建新实例。

订单支付字段包括 `payment_status`、`paid_at`、`payment_provider`、`payment_trade_no` 和 `payment_callback_payload`，只作为摘要字段。真实支付流水、回调摘要、退款和支付生效事实以支付相关表为准。`payment_status` 允许 `unpaid`、`paid`、`manual_confirmed`、`refunded`；管理端人工续费确认继续使用 `manual_confirmed`，真实支付成功使用 `paid`。

### 支付

```text
payment_transactions
refund_transactions
payment_effects
```

`payment_transactions` 保存用户端为订单创建的支付交易。支付编号使用 `payment_no` 对外展示，不直接暴露自增 ID。供应商允许 `alipay`、`wechat` 和 `wallet`；方式允许 `alipay_page`、`alipay_wap`、`wechat_native`、`wechat_h5` 和 `wallet_balance`。状态允许 `pending`、`paid`、`closed`、`failed`、`refunded`。金额字段使用分为单位，币种一期固定为 `CNY`。

同一订单、供应商、方式和用户端 `client_token` 必须唯一，用于支付创建幂等。供应商交易号按 `provider + upstream_trade_no` 唯一；为空时允许多条未完成交易。支付记录可保存二维码 URL、跳转 URL、过期时间、支付完成时间、关闭时间、失败时间、渠道查询摘要和回调摘要；不得保存商户私钥、API v3 key、签名串、完整回调 payload 或完整上游响应。

`refund_transactions` 保存全额退款流水。退款编号使用 `refund_no` 对外展示，一期同一支付最多一条退款记录，通过 `payment_id` 唯一约束兜底。状态允许 `pending`、`succeeded`、`failed`。退款先创建本地 `pending` 记录，再调用渠道；渠道退款成功或查询确认后，才在本地事务中回滚续费生效、更新支付和订单状态。渠道失败不得扣回用户服务期。供应商退款号按 `provider + upstream_refund_no` 唯一。

`payment_effects` 保存支付成功后的业务生效记录。新购支付记录关联后续交付出的实例编号；续费支付记录保存续费前 `before_expires_at` 和续费后 `after_expires_at`，用于审计和退款回滚。状态允许 `active` 和 `reverted`。同一支付最多一条生效记录；退款成功回滚后记录退款编号和 `reverted_at`。

支付相关写入必须明确事务边界：本地支付状态、订单摘要、支付生效记录和任务投递使用本地事务；渠道下单、退款、主动查询和异步通知处理不得在长事务中保存完整上游响应。回调处理必须锁定支付和订单记录，按本地状态幂等推进。钱包余额支付不调用外部渠道，必须同事务锁定订单和钱包账户完成扣款、支付交易、生效记录和钱包流水写入。

### 钱包

```text
wallet_accounts
wallet_ledger_entries
wallet_recharges
```

`wallet_accounts` 保存用户钱包账户当前余额。钱包编号使用 `wallet_no` 对外展示，不直接暴露自增 ID。钱包按 `user_id + currency` 唯一，v1 币种固定为 `CNY`，状态允许 `active` 和 `disabled`。当前余额使用 `available_balance_cents`，累计充值、消费和退回钱包金额分别保存在统计字段中，全部使用分为单位。

`wallet_ledger_entries` 保存钱包余额变动流水。流水是追加式账本，不更新、不删除。方向允许 `credit` 和 `debit`，类型允许 `recharge`、`payment` 和 `refund`。每条流水必须保存变动金额、变动前余额、变动后余额、关联对象类型/编号和幂等键；同一钱包下同一幂等键只能写入一次，防止重复回调、重复余额支付或重复退款导致余额重复变化。

`wallet_recharges` 保存钱包充值记录。充值编号使用 `recharge_no` 对外展示。充值只允许通过 `alipay` 或 `wechat` 创建上游交易，方式允许 `alipay_page`、`alipay_wap`、`wechat_native` 和 `wechat_h5`。状态允许 `pending`、`paid`、`closed`、`failed`。同一钱包、供应商、方式和用户端 `client_token` 必须唯一；供应商交易号按 `provider + upstream_trade_no` 唯一。

钱包充值回调必须锁定充值记录和钱包账户，只有 `pending` 充值可以推进为 `paid` 并入账；重复回调不得重复写流水或重复增加余额。余额支付必须锁定订单和钱包账户，只有订单 `status=pending` 且 `payment_status=unpaid`、钱包 `status=active` 且余额充足时才能扣款。余额支付退款必须锁定退款、支付、订单、钱包账户和必要的支付生效记录，退回钱包余额后写入 `refund` 类型流水。

### 实例交付

```text
instance_provision_mappings
instances
instance_operations
```

实例交付通过 MCP PVE client API 调用上游 PVE 适配服务。pveCloud 不保存通用 PVE 节点、存储或资源池目录，只保存业务实例、交付映射和操作记录。

`instance_provision_mappings` 保存交付映射，使用 `plan_no`、`region_no`、`template_no` 和 `network_type_no` 匹配订单快照；`network_type_no` 为空字符串表示不限定网络类型。映射保存 MCP 创建 VM 所需的 `node`、`storage`、`disk_source`、`disk_format`、`disk_interface`、`snippets_storage`、CloudInit 非敏感字段和 VMID 分配范围。`next_vmid` 必须在本地事务中分配并递增；服务端不得依赖前端传入 VMID。

CloudInit `ci_password` 当前不落库、不作为映射配置保存，也不通过接口返回。后续如需初始密码或重置密码，必须先补充一次性凭据展示、加密或脱敏存储、审计和日志保护契约。

`instances` 保存云主机实例最终事实。实例对外展示使用 `instance_no`，不直接暴露自增 ID。用户端只返回实例编号、订单号、状态和产品/套餐/地域/系统模板等业务快照；`external_node`、`external_vmid`、`external_resource_location` 只允许管理端和服务端内部使用。

实例状态使用字符串字段，不使用数据库 enum。当前允许以下状态：

- `creating`：已触发 MCP 创建 VM，等待上游 operation 完成。
- `running`：实例处于运行状态。
- `stopped`：实例处于停止状态。
- `error`：最近一次创建、同步或操作失败，需要管理端处理。
- `releasing`：已触发释放，等待上游删除 VM 完成。
- `released`：实例已释放，本地记录保留。

`instances.order_id` 当前使用唯一约束，表示一个订单最多交付一台实例；订单数量仍固定为 `1`。`instances(external_node, external_vmid)` 必须唯一，避免同一上游 VM 被重复绑定。

实例服务期字段用于到期、提醒和释放：

- `service_started_at`：服务期开始时间，实例首次从 `creating` 同步到 `running` 或 `stopped` 时写入。
- `expires_at`：服务到期时间，由订单 `billing_cycle` 计算，续费确认后顺延。
- `expire_notice_sent_at`：最近一次到期提醒发送完成时间。
- `expire_release_scheduled_at`：到期后自动释放计划时间。
- `expire_released_at`：因到期自动释放完成时间。

自动释放必须通过任务执行并调用现有 MCP 删除 VM 能力；当配置关闭自动释放时，只允许发送到期提醒和展示到期状态，不得释放上游 VM。

`instance_operations` 保存实例异步操作记录，包括 `provision`、`start`、`stop`、`release` 和 `sync`。MCP 返回的 operation ID、Operation-Location、resourceLocation、失败码和失败说明保存为排障事实。操作状态只允许 `running`、`succeeded`、`failed`。

产生外部副作用的操作必须明确事务边界：本地实例、操作记录、订单状态和后台审计写入使用本地事务；MCP 网络调用不得放进长事务。上游调用失败后必须把本地实例或操作记录置为可恢复、可排查状态，不得静默丢失。

### 异步任务与通知

```text
async_tasks
notifications
```

`async_tasks` 保存通用后台任务。任务类型首批允许 `instance_operation_sync`、`instance_expiry_notice`、`instance_expiry_release`、`notification_email_send`、`notification_sms_placeholder`。任务状态只允许 `pending`、`running`、`succeeded`、`failed`、`cancelled`。任务通过 `task_type` 和内部幂等投影约束同一 `idempotency_key` 只能存在一条未取消任务；取消任务时释放幂等投影，重试失败任务时复用原任务行。Worker 领取时必须写入 `locked_by`、`locked_until`，避免并发重复执行。

`notifications` 保存通知发送记录和用户可见/后台可查的通知事实。通知通道首批允许 `email` 和 `sms`；`email` 可复用 SMTP 发送，`sms` 当前只做占位记录，不接真实短信供应商。通知内容不得保存密码、token、MCP Bearer Token、SMTP 凭据或完整上游响应。

实例生命周期任务必须以业务状态作为最终幂等判断：已经释放的实例不得重复释放；已经延长到期时间的实例不得执行旧的到期释放任务；已成功发送的同一到期提醒不得重复发送。

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

`tickets` 用于保存用户端提交的工单主记录。工单只表示用户与后台之间的沟通事项和后台内部处理协作，不代表支付、PVE 操作或自动处理承诺。内部 SLA 仅用于后台处理时限，不作为用户端服务承诺。

工单状态使用字符串字段，不使用数据库 enum。第一阶段只允许以下状态：

- `waiting_admin`：等待后台处理。创建工单或用户回复后进入该状态。
- `waiting_user`：等待用户反馈。管理员回复后进入该状态。
- `closed`：工单已关闭。用户或管理员均可关闭未关闭工单。

工单分类固定为 `account`、`order`、`product`、`technical`、`billing`、`other`。工单优先级固定为 `low`、`normal`、`high`、`urgent`，默认 `normal`。

内部 SLA 按工单优先级在创建时写入首次响应和解决截止时间。默认规则为：`low` 首响 48 小时、解决 7 天；`normal` 首响 24 小时、解决 5 天；`high` 首响 8 小时、解决 3 天；`urgent` 首响 2 小时、解决 24 小时。管理员首次回复时写入首次响应时间。关闭工单时视为解决并写入解决时间。逾期状态由当前时间和截止/完成时间计算，不依赖后台定时任务。

工单对外展示使用 `ticket_no`，不直接暴露自增 ID。工单可选关联 `orders.id` 并保存 `order_no` 快照，也可选关联 `instances.id` 并保存 `instance_no` 快照；用户端关联订单或实例时必须校验订单、实例属于当前登录用户。

只填写 `instance_no` 创建工单时，服务端必须使用实例来源订单回填 `order_id` 和 `order_no`。同时填写 `order_no` 和 `instance_no` 时，实例必须来源于同一订单，否则拒绝创建。工单关联实例只用于排障沟通和筛选，不改变订单、实例或实例操作状态，不保存 PVE 节点、存储、VMID、operation ID 或上游错误详情。

工单当前处理人使用 `assignee_admin_id` 关联 `admin_users.id`。可指派对象只允许 `active` 且具备 `page.tickets`、`ticket:reply` 或 `ticket:*` 的管理员。关闭后的工单不得指派、转派、维护协作者或升级优先级。

`ticket_messages` 保存工单消息。消息发送方使用 `sender_type=user|admin` 和对应发送者 ID 表示。关闭后的工单不得继续写入新消息。

`ticket_message_attachments` 保存工单消息与 `file_attachments` 的关联。单条消息最多 5 个附件。工单附件同时必须写入 `file_attachment_references`，`ref_type=ticket_message`，引用 ID 使用消息 ID。

`ticket_tags` 保存管理端标签字典。标签分 `public` 和 `internal`；公开标签可返回用户端，内部标签只返回管理端。停用标签不可新绑定，但历史绑定仍可展示。

`ticket_tag_bindings` 保存工单与标签的多对多关系。标签绑定变更必须整体替换当前绑定集合，并记录事件和后台审计。

`ticket_internal_notes` 保存管理端内部备注。内部备注只追加，不编辑、不删除，不返回用户端。关闭后的工单仍允许追加内部备注。

`ticket_collaborators` 保存工单协作者。协作者只返回管理端。关闭后的工单不得增删协作者。

`ticket_events` 保存工单内部操作历史，记录指派、转派、协作者变更、内部备注、优先级升级、标签绑定、回复和关闭等后台处理事件。

优先级升级只能从低紧急度升到更高紧急度，必须填写原因，不支持降级。升级后未完成的 SLA 截止时间只允许提前，不允许延后。

工单创建、回复、附件关联和工单主表最近消息时间更新必须在本地数据库事务中完成。创建工单时只读取订单和实例做归属与一致性校验，不修改订单或实例，因此不需要锁定订单或实例行；同一请求内新建的工单主记录、首条消息、附件关联和文件引用必须同事务写入。管理端回复、关闭、指派、转派、协作者维护、内部备注、优先级升级、标签绑定和标签字典变更必须与普通后台审计写入保持同事务。

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
- 日志管理中心使用独立一级菜单 `page.logs`；Phase 1 子页面为 `page.logs.admin-operations` 和 `page.logs.admin-security`
- 后续日志管理中心子页面包括 `page.logs.user-security`、`page.logs.user-business`、`page.logs.frontend-errors` 和 `page.logs.backend-runtime`

## 普通操作日志上下文

`admin_audit_logs` 除业务动作、对象和前后快照外，还记录以下后台请求上下文：

- `admin_username`：操作发生时的管理员用户名快照
- `admin_display_name`：操作发生时的管理员显示名快照
- `session_id`：触发操作的管理端会话标识
- `request_id`：请求链路 ID，用于串联访问日志、错误日志和审计日志
- `request_method`：后台请求方法
- `request_path`：后台请求路径

这些字段只增强普通操作日志的可读性和排查能力。

## 日志管理中心表

### 用户安全日志

`user_security_logs` 用于保存用户端登录、退出、刷新、密码重置和限流等安全事件。表中允许未知账号失败场景的 `user_id` 为空，重点记录事件结果、请求链路和客户端摘要。

### 用户业务日志

`user_business_logs` 用于保存用户实名、订单和工单等关键业务事件。表中保存模块、动作、对象类型和对象 ID，供用户安全/业务日志管理页面查询。

### 前端错误日志

`frontend_error_logs` 用于保存 admin/web 前端上报的错误摘要。只允许脱敏后的页面路径、错误消息、堆栈摘要和关联 API 信息入库。

### 后端运行日志

`backend_runtime_logs` 用于保存结构化运行日志快照，覆盖访问日志、panic 和关键运行错误的可查询摘要。它不替代 stdout 结构化日志输出。

### 日志导出记录

`log_export_records` 用于记录日志导出行为和导出条件，为后续导出审计与留存策略提供锚点。

## 当前阶段说明

当前仓库已经从“基础后台阶段”继续演进到用户账号自助、用户实名、服务器产品目录、订单、支付、钱包、续费订单、实例、异步任务、通知和工单阶段。
数据库契约保留以下管理域：

- 认证
- RBAC
- 会话
- 系统配置
- 审计日志
- 用户端账号与会话
- 用户实名
- 服务器产品目录
- 订单
- 支付
- 钱包
- 实例
- 异步任务
- 通知
- 工单

以下业务域表不属于当前数据库契约，后续如需恢复，必须先补新的迁移和文档确认：

- 发票

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
- `instance_provision_mappings.mapping_no`
- `instance_provision_mappings(plan_no, region_no, template_no, network_type_no, status)`
- `instances.instance_no`
- `instances.order_id`
- `instances(external_node, external_vmid)`
- `instance_operations.operation_no`
- `async_tasks.task_no`
- `async_tasks(task_type, idempotency_active_key)`，只约束未取消任务的有效幂等键
- `notifications.notification_no`
- `tickets.ticket_no`
- `ticket_message_attachments(message_id, file_id)`
- `user_security_logs(user_id, created_at)`
- `user_business_logs(user_id, created_at)`
- `frontend_error_logs(source_app, created_at)`
- `backend_runtime_logs(category, created_at)`
- `log_export_records(admin_id, created_at)`

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

实例管理需要新增以下管理端权限目录：

- `page.instances`
- `instance:*`
- `instance:view`
- `instance:provision`
- `instance:operate`
- `instance:release`
- `instance:sync`
- `instance:renew`

异步任务需要新增以下管理端权限目录：

- `page.async-tasks`
- `async-task:*`
- `async-task:retry`

支付管理需要新增以下管理端权限目录：

- `page.payments`
- `payment:*`
- `payment:view`
- `payment:refund`
- `payment:sync`
- `payment:retry-provision`

## 一致性原则

- MariaDB 是基础后台事实来源
- Redis 只做缓存、限流、短 TTL 状态和辅助幂等
- 当前阶段只以 MCP PVE client API 支撑的实例交付、基础操作、到期释放、真实支付一期、钱包 v1 和异步任务为现行数据库契约，不以通用 PVE 运维管理、发票、提现、人工调账、余额转账或部分退款为现行数据库契约
- 未来创建订单时必须复制产品、套餐、价格、销售地域和服务器系统模板快照，不能只依赖当前产品表引用
