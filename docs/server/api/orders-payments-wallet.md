# 订单、支付与钱包 API

本文档维护用户端订单、支付、钱包以及管理端钱包、支付和退款运营接口。跨接口通用约定见 `docs/server/api/conventions.md`。

## 用户端订单

订单保存用户购买意向和后台处理所需快照。新购订单可由管理员人工触发交付，也可在真实支付成功后自动投递交付任务；续费订单用于延长已有实例服务期。支付流水、回调和退款以支付交易、退款交易和支付生效记录为准，订单支付字段只作为摘要。

### `POST /api/orders`

- 鉴权：用户端 Bearer Token
- 作用：基于固定套餐和用户选择的可选配置创建订单
- 请求字段：`plan_no`、`billing_cycle`、`region_no`、`template_no`、`network_type_no`、`quantity`、`client_token`、`user_note`
- `billing_cycle` 允许 `monthly`、`quarterly`、`semi_yearly`、`yearly`
- `region_no`、`template_no`、`network_type_no` 必须属于当前套餐可用配置
- `quantity` 当前固定为 `1`
- `user_note` 可选，最多 500 字
- 成功数据包含订单详情快照
- 约束：订单价格、地域、系统模板和网络类型必须在创建时从当前产品目录校验并保存快照
- 约束：网络类型当前只保存编号、编码和名称快照，不返回或保存 PVE 网络 ID
- 约束：当前阶段不接受自定义 CPU、内存、硬盘、带宽、公网 IP 数量、购买数量或登录密码模式
- 约束：创建订单不直接调用 MCP PVE client API，不直接创建实例

### `GET /api/orders`

- 鉴权：用户端 Bearer Token
- 作用：分页查询当前用户自己的订单列表
- 查询参数支持：`page`、`per_page`、`status`
- 列表项包含订单编号、订单类型、支付状态、状态、产品名称、套餐名称、计费周期、网络类型、关联实例编号、订单金额、最近支付摘要、是否可发起支付和创建/取消/关闭时间；订单状态允许 `pending`、`provisioning`、`fulfilled`、`error`、`cancelled`、`closed`

### `GET /api/orders/{order_no}`

- 鉴权：用户端 Bearer Token
- 作用：查看当前用户自己的订单详情
- 成功数据包含产品、套餐、价格、销售地域、系统模板、网络类型快照、最近支付摘要、支付入口信息和是否可发起支付

### `POST /api/orders/{order_no}/cancel`

- 鉴权：用户端 Bearer Token
- 作用：取消当前用户自己的 `pending` 订单
- 请求字段：`reason` 可选，最多 500 字
- 约束：仅 `pending` 订单可由用户取消；`provisioning` 和 `fulfilled` 订单不可由用户端取消

### `POST /api/instances/{instance_no}/renewal-orders`

- 鉴权：用户端 Bearer Token
- 作用：为当前用户自己的未释放实例创建续费订单
- 请求字段：`billing_cycle`、`client_token`
- `billing_cycle` 允许 `monthly`、`quarterly`、`semi_yearly`、`yearly`
- 成功数据包含续费订单详情
- 约束：
  - 只能为当前登录用户自己的实例创建续费订单
  - `released` 或 `releasing` 实例不可创建续费订单
  - 续费订单不创建新实例，不调用 MCP PVE client API
  - 续费价格必须由服务端按实例当前套餐价格重新计算
  - 同一用户同一 `client_token` 必须幂等，不得重复创建续费订单
  - 续费订单初始为 `payment_status=unpaid`；未配置支付渠道时仍可由管理端按人工流程确认

## 支付

支付一期只支持中国大陆商户的支付宝电脑网页支付、支付宝手机网页支付、微信 Native 扫码、微信 H5 和钱包余额支付。支付配置由后台系统设置维护；钱包配置由后台系统设置维护；用户端和公开接口不得返回商户密钥、签名串、完整回调 payload、内部任务 ID、PVE/MCP 细节或完整上游响应。

服务端支付适配要求：

- 支付宝使用成熟 Go SDK `github.com/smartwalle/alipay/v3` 对接网页支付、手机网页支付、交易查询、退款和通知验签。
- 微信支付使用官方 Go SDK `github.com/wechatpay-apiv3/wechatpay-go` 对接 Native 下单、H5 下单、支付通知解密验签、退款和查询。
- 生产路径不得使用 mock adapter；mock adapter 只允许用于单元测试和集成测试。
- 支付渠道配置不完整时，创建支付、同步查询和退款必须返回冲突或外部依赖不可用错误，不得创建不可处理的上游交易。

### `POST /api/orders/{order_no}/payments`

- 鉴权：用户端 Bearer Token
- 作用：为当前用户自己的订单创建支付交易
- 请求字段：`provider`、`method`、`client_token`
- `provider` 允许：`alipay`、`wechat`、`wallet`
- `method` 允许：`alipay_page`、`alipay_wap`、`wechat_native`、`wechat_h5`、`wallet_balance`
- 成功数据包含：`payment_no`、`order_no`、`provider`、`method`、`amount_cents`、`currency`、`status`、`expires_at`、`redirect_url`、`qr_code_url`
- 约束：
  - 只能为当前登录用户自己的订单创建支付
  - 订单必须处于 `pending` 且 `payment_status=unpaid`
  - 金额和币种必须由服务端读取订单事实计算，不信任前端提交
  - 同一订单、供应商、方式和 `client_token` 必须幂等；重复提交返回已有支付
  - 支付宝/微信未启用或配置不完整时返回 `409xx` 或 `700xx`，不得创建不可处理交易
  - 钱包余额支付必须启用 `wallet.enabled`，但不要求支付宝/微信渠道可用
  - 渠道下单在本地支付交易创建后执行；上游交易创建失败时本地支付保持 `failed` 或可同步恢复状态，并保存脱敏错误摘要
  - 创建支付宝/微信支付不得直接交付实例或延长服务期，支付成功以后续回调或主动查询确认为准
  - 钱包余额支付使用同一接口，`provider=wallet`、`method=wallet_balance`；服务端必须同事务锁定订单和钱包账户，余额不足返回冲突，余额支付成功后直接推进订单生效

### `GET /api/payments/{payment_no}`

- 鉴权：用户端 Bearer Token
- 作用：查询当前用户自己的支付状态和订单处理状态
- 成功数据包含支付编号、订单编号、供应商、方式、金额、币种、支付状态、过期时间、支付完成时间、订单状态、订单支付状态、关联实例编号和用户可见的失败摘要
- 约束：
  - 只能查询当前登录用户自己的支付
  - 不返回商户密钥、完整回调 payload、内部任务 ID、Worker 锁、PVE/MCP 细节或完整上游响应

### `POST /api/payment-callbacks/alipay`

- 鉴权：公开回调，无 Bearer Token
- 作用：接收支付宝支付、钱包充值和退款异步通知
- 约束：
  - 必须验签通过后处理
  - 必须从验签后的通知解析本地支付编号、供应商交易号、交易状态和金额
  - 必须校验本地支付交易、订单、金额、币种和供应商交易号
  - 重复回调必须幂等，不能重复交付、续费或退款回滚
  - 验签失败、金额不一致、未知交易或状态冲突不得推进本地状态
  - 仅保存回调摘要，不保存完整表单、签名串、密钥或完整上游响应

### `POST /api/payment-callbacks/wechat`

- 鉴权：公开回调，无 Bearer Token
- 作用：接收微信支付、钱包充值和退款异步通知
- 约束：
  - 必须校验微信平台签名、证书/平台公钥、通知解密结果和时间窗口
  - 必须从 SDK 解密后的通知解析本地支付编号、供应商交易号、交易状态和金额
  - 必须校验本地支付交易、订单、金额、币种和供应商交易号
  - 重复回调必须幂等，不能重复交付、续费或退款回滚
  - 验签失败、金额不一致、未知交易或状态冲突不得推进本地状态
  - 仅保存回调摘要，不保存完整通知密文、签名串、密钥或完整上游响应

### 支付成功后的业务处理

- 新购订单支付成功后，服务端必须同事务更新支付交易和订单支付摘要，并投递 `payment_order_provision` 任务；任务读取现有实例交付规则创建实例。
- 自动交付成功后订单进入 `fulfilled`；自动交付失败后订单进入 `error`，保留 `payment_status=paid`，由管理端支付管理页重试入队。
- 续费订单支付成功后，服务端必须复用管理端人工确认续费的同一计算规则：若实例尚未到期，从原 `expires_at` 顺延；若已到期，从当前时间顺延。
- 续费支付成功必须写入支付生效记录，记录支付、订单、实例、续费前 `expires_at`、续费后 `expires_at` 和生效时间。

## 钱包

钱包 v1 只开放用户端充值、余额支付和管理端只读查看。钱包币种固定为 `CNY`，金额字段使用分为单位。钱包充值复用支付宝/微信支付渠道；订单余额支付复用 `payment_transactions` 保存支付事实，`provider=wallet`、`method=wallet_balance`。真实支付订单退款必须原路退回支付宝/微信渠道；余额支付订单退款退回钱包余额，不支持真实支付退款入钱包。

钱包账户按用户和币种唯一，首次读取、充值或余额支付时可懒创建。钱包余额最终事实为 `wallet_accounts.available_balance_cents`；钱包流水为追加式审计账本，不提供更新或删除接口。

### `GET /api/wallet`

- 鉴权：用户端 Bearer Token
- 作用：查看当前登录用户自己的钱包摘要
- 成功数据包含钱包编号、币种、状态、可用余额、累计充值、累计消费、累计退回钱包和创建时间
- 约束：只能查看当前用户自己的钱包；未存在钱包时可返回零余额钱包摘要或先懒创建钱包账户

### `GET /api/wallet/ledger`

- 鉴权：用户端 Bearer Token
- 作用：分页查询当前登录用户自己的钱包流水
- 查询参数支持：`page`、`per_page`、`direction`、`entry_type`、`related_no`、`date_from`、`date_to`
- 列表项包含流水编号、方向、类型、金额、变动前余额、变动后余额、币种、关联对象、摘要和创建时间
- 约束：只返回当前用户自己的流水；不返回内部自增 ID、后台操作者信息或渠道完整响应

### `POST /api/wallet/recharges`

- 鉴权：用户端 Bearer Token
- 作用：创建钱包充值交易
- 请求字段：`provider`、`method`、`amount_cents`、`client_token`
- `provider` 允许：`alipay`、`wechat`
- `method` 允许：`alipay_page`、`alipay_wap`、`wechat_native`、`wechat_h5`
- 成功数据包含充值编号、钱包编号、供应商、方式、金额、币种、状态、过期时间、跳转地址和二维码地址
- 约束：
  - 必须启用 `wallet.enabled`
  - 充值金额必须落在 `wallet.recharge_min_cents` 和 `wallet.recharge_max_cents` 之间
  - 同一钱包、供应商、方式和 `client_token` 必须幂等；重复提交返回已有充值
  - 渠道下单在本地充值记录创建后执行；上游创建失败时本地充值进入 `failed` 或可同步恢复状态，并保存脱敏错误摘要
  - 创建充值不得直接增加钱包余额，充值成功以后续回调或主动查询确认为准

### `GET /api/wallet/recharges/{recharge_no}`

- 鉴权：用户端 Bearer Token
- 作用：查询当前登录用户自己的充值状态
- 成功数据包含充值编号、钱包编号、供应商、方式、金额、币种、状态、过期时间、完成时间、跳转地址、二维码地址和用户可见失败摘要
- 约束：只能查询当前用户自己的充值；不返回完整回调 payload、签名串、内部锁或完整上游响应

### 钱包充值回调和余额变更规则

- 支付宝/微信回调入口继续使用 `POST /api/payment-callbacks/alipay` 和 `POST /api/payment-callbacks/wechat`。
- 服务端必须能根据回调中的本地交易编号或上游交易号区分订单支付与钱包充值。
- 充值回调必须验签、校验金额、币种、供应商和本地充值状态；只有 `pending` 充值可推进为 `paid`。
- 充值入账必须在同一事务中锁定充值记录和钱包账户，更新充值状态、增加钱包余额、写入 `wallet_ledger_entries`；重复回调不得重复入账。
- 余额支付扣款、余额支付退款退回钱包和充值入账都必须写钱包流水；流水必须包含幂等键，重复执行不得重复改变余额。

## 管理端钱包管理

钱包管理用于只读查看用户钱包、余额、充值和流水。v1 不支持管理端人工加款、扣款、冻结、解冻、提现、余额转账或退款入钱包。

### `GET /admin-api/wallets`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.wallets`
- 作用：分页查询用户钱包
- 查询参数支持：`page`、`per_page`、`user_keyword`、`status`、`wallet_no`
- 列表项包含钱包编号、用户摘要、币种、状态、可用余额、累计充值、累计消费、累计退回钱包和创建时间
- 约束：不返回用户敏感资料明文，不返回内部自增 ID

### `GET /admin-api/wallets/{wallet_no}`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.wallets`
- 作用：查看钱包详情、最近流水和最近充值摘要
- 成功数据不得包含渠道完整响应、签名串、用户敏感资料明文或内部锁详情

### `GET /admin-api/wallet-ledger`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.wallets`
- 作用：分页查询钱包流水
- 查询参数支持：`page`、`per_page`、`wallet_no`、`user_keyword`、`direction`、`entry_type`、`related_no`、`date_from`、`date_to`
- 列表项包含流水编号、钱包编号、用户摘要、方向、类型、金额、变动前后余额、关联对象、摘要和创建时间

### `GET /admin-api/wallet-recharges`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.wallets`
- 作用：分页查询钱包充值记录
- 查询参数支持：`page`、`per_page`、`wallet_no`、`user_keyword`、`provider`、`method`、`status`、`recharge_no`、`date_from`、`date_to`
- 列表项包含充值编号、钱包编号、用户摘要、供应商、方式、金额、币种、状态、过期时间、完成时间和创建时间

## 管理端支付管理

支付管理用于查看支付流水、退款流水、支付详情、发起全额退款、同步渠道状态和重试自动交付失败订单。

### `GET /admin-api/payments`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.payments`
- 作用：分页查询支付流水
- 查询参数支持：`page`、`per_page`、`provider`、`method`、`status`、`order_no`、`payment_no`、`user_keyword`、`date_from`、`date_to`
- 列表项包含支付编号、订单编号、用户摘要、供应商、方式、金额、币种、状态、支付完成时间、过期时间和创建时间
- 约束：不返回完整回调 payload、完整上游响应或商户密钥

### `GET /admin-api/payments/{payment_no}`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.payments`
- 作用：查看支付详情、订单摘要、退款摘要和支付生效记录摘要
- 成功数据不得包含商户密钥、完整回调 payload、完整上游响应、PVE/MCP 密钥或 Worker 锁详情

### `POST /admin-api/payments/{payment_no}/sync`

- 鉴权：管理端 Bearer Token
- 操作权限：`payment:sync` 或 `payment:*`
- 作用：主动查询渠道支付状态并按验签后的渠道结果更新本地状态
- 约束：只允许同步 `pending`、`failed` 或渠道状态不确定的支付；同步必须调用对应渠道查询接口并校验返回的供应商交易号、金额和币种；同步成功后按支付成功业务处理推进订单
- 审计：`payment.sync`

### `POST /admin-api/payments/{payment_no}/refunds`

- 鉴权：管理端 Bearer Token
- 操作权限：`payment:refund` 或 `payment:*`
- 作用：为已支付交易发起全额退款
- 请求字段：`reason` 必填，最多 500 字
- 约束：
  - 一期只允许全额退款，同一支付最多存在一条退款记录
  - 支付必须处于 `paid`
  - 支付关联订单存在 `pending`、`processing` 或 `issued` 发票申请时不得退款；v1 不支持红冲或开票后在线退款
  - 新购已交付订单必须先释放实例后才能退款；未交付新购订单可直接退款
  - 续费订单必须存在可回滚的支付生效记录
  - 服务端先创建 `pending` 退款记录并调用渠道退款；渠道成功或查询确认后，再同事务回滚本地支付生效、更新退款/支付/订单状态和写审计
  - 退款请求必须复用支付交易的供应商交易号和退款编号作为幂等锚点；渠道返回处理中或不可确认时，本地退款保持 `pending`，不得提前扣回续费时间
  - 渠道退款失败时退款状态为 `failed`，不得扣回用户服务期
- 审计：`payment.refund.create`、`payment.refund.succeeded` 或 `payment.refund.failed`

### `GET /admin-api/refunds`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.payments`
- 作用：分页查询退款流水
- 查询参数支持：`page`、`per_page`、`provider`、`status`、`order_no`、`payment_no`、`refund_no`、`date_from`、`date_to`
- 列表项包含退款编号、支付编号、订单编号、用户摘要、供应商、金额、币种、状态、发起管理员、完成时间和创建时间

### `POST /admin-api/payments/{payment_no}/retry-provision`

- 鉴权：管理端 Bearer Token
- 操作权限：`payment:retry-provision` 或 `payment:*`
- 作用：针对 `status=error` 且 `payment_status=paid` 的新购订单重新投递自动交付任务
- 约束：仅新购订单可重试；若订单已存在实例或状态已变化，返回 `409xx` 或当前业务结果，不得重复创建实例
- 审计：`payment.provision.retry`
