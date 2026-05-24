# System Settings 页面契约

## 页面定位

`System Settings` 是当前管理端开放的系统设置父级菜单。

当前承载：

- 系统配置
- 管理员设置

不承载：

- 日志管理中心。日志管理中心由 `docs/admin/pages/log-management-center.md` 定义。

## 路由结构

父级菜单：

- 路径：`/system`
- 标题：系统设置
- 作为侧栏父级菜单展示

子页面：

- 系统配置：`/system/settings`
- 管理员设置：`/system/admin-users`

当前不为系统设置继续拆更多侧栏层级。

## 系统配置

页面职责：

- 按分组展示 `system_configs` 表中的配置项。
- 支持编辑允许更新的配置项。
- `is_secret=1` 的配置值不得展示明文。
- `site.name` 和 `site.logo_url` 用于 Web 左上角品牌区域，分别控制显示文字和 Logo 图片 URL。
- `site.logo_url` 可为空；为空时 Web 使用前端默认标识。
- `site.logo_url` 在管理端系统配置页不使用自由文本输入；应通过图片上传控件上传 Logo，并将上传接口返回的 URL 写入该配置项。
- 上传 Logo 复用 `POST /admin-api/files/upload`，仅允许图片类文件；上传权限按 `file:upload` 或 `file:*` 判断，配置保存权限仍按 `system-config:update` 或 `system-config:*` 判断。
- 站点品牌配置在后台系统配置中按中文分组“站点设置”展示。
- Web 用户认证验证码开关在后台系统配置中按中文分组“用户认证”展示。
- 用户实名业务开关、可选供应商、供应商启用状态、支付宝/微信侧接入参数、密钥、回调地址和证件摘要密钥在后台系统配置中按中文分组“实名设置”展示。
- 支付宝、微信/腾讯云密钥和证件摘要密钥使用 `is_secret=1`；页面不得回显明文，只展示是否已配置，并允许通过重新填写覆盖。
- 支付业务开关、支付订单过期时间、支付宝支付和微信支付商户参数在后台系统配置中按中文分组“支付设置”展示。
- 支付宝支付私钥、公钥、微信支付 API v3 key、商户私钥或平台公钥等密钥类配置使用 `is_secret=1`；页面不得回显明文，只展示是否已配置，并允许通过重新填写覆盖。
- `value_type=bool` 的配置项必须使用明确布尔编辑控件，不使用自由文本输入。
- 保存布尔配置时始终提交字符串 `true` / `false`。
- 保存 `is_secret=1` 配置时，空值表示保留旧值，非空值表示覆盖旧值。

当前阶段系统配置至少包含以下用户认证开关：

- `web.auth.login_captcha_enabled`
- `web.auth.register_captcha_enabled`
- `web.auth.password_reset_request_captcha_enabled`
- `web.auth.password_reset_confirm_captcha_enabled`

当前阶段系统配置至少包含以下实名配置：

- `real_name.enabled`
- `real_name.manual_review_enabled`
- `real_name.required_for_order`
- `real_name.allowed_providers`
- `real_name.default_provider`
- `real_name.identity_digest_secret`
- `real_name.callback_base_url`
- `real_name.resubmit_enabled`
- `real_name.max_submit_attempts`
- `real_name.review_notice`
- `real_name.alipay.enabled`
- `real_name.alipay.app_id`
- `real_name.alipay.gateway_url`
- `real_name.alipay.app_private_key`
- `real_name.alipay.alipay_public_key`
- `real_name.alipay.return_url`
- `real_name.alipay.notify_url`
- `real_name.wechat.enabled`
- `real_name.wechat.secret_id`
- `real_name.wechat.secret_key`
- `real_name.wechat.region`
- `real_name.wechat.endpoint`
- `real_name.wechat.rule_id`
- `real_name.wechat.redirect_url`

当前阶段系统配置至少包含以下支付配置：

- `payment.enabled`
- `payment.default_expire_minutes`
- `payment.callback_base_url`
- `payment.alipay.enabled`
- `payment.alipay.app_id`
- `payment.alipay.gateway_url`
- `payment.alipay.app_private_key`
- `payment.alipay.alipay_public_key`
- `payment.alipay.notify_url`
- `payment.alipay.return_url`
- `payment.wechat.enabled`
- `payment.wechat.app_id`
- `payment.wechat.mch_id`
- `payment.wechat.api_v3_key`
- `payment.wechat.mch_private_key`
- `payment.wechat.mch_certificate_serial_no`
- `payment.wechat.platform_public_key_id`
- `payment.wechat.platform_public_key`
- `payment.wechat.notify_url`
- `payment.wechat.h5_scene_info`

当前阶段系统配置至少包含以下钱包配置：

- `wallet.enabled`
- `wallet.recharge_min_cents`
- `wallet.recharge_max_cents`

供应商启用约束：

- `real_name.allowed_providers` 只控制用户端可选列表，具体供应商还必须满足对应 `real_name.<provider>.enabled=true`。
- 启用支付宝前，必须填写支付宝应用 ID、网关、私钥、公钥、返回地址，以及全局回调基础地址或支付宝异步通知地址。
- 启用微信侧实名前，必须填写腾讯云 SecretId、SecretKey、地域、端点、规则 ID 和返回地址；当前微信/腾讯云结果通过服务端同步查询确认，不开放异步回调。
- `real_name.manual_review_enabled=true` 时，支付宝/微信侧实名不可用后用户端默认进入人工审核。
- `real_name.identity_digest_secret` 只作为外部供应商实名和证件摘要重复校验配置；缺失时外部供应商不可用，但不影响人工审核实名入口。已有当前 HMAC 版本实名申请后，页面不允许通过普通系统设置直接修改该密钥。
- `wallet.enabled=true` 只开放钱包入口和余额支付入口；充值仍依赖 `payment.enabled=true` 且支付宝或微信支付配置完整。
- `wallet.recharge_min_cents` 和 `wallet.recharge_max_cents` 必须为正整数，且最小值不得大于最大值。
- 启用支付总开关前，至少需要启用并配置完整一个支付供应商。
- 启用支付宝支付前，必须填写应用 ID、网关、应用私钥、支付宝公钥、异步通知地址和同步返回地址。
- 启用微信支付前，必须填写应用 ID、商户号、API v3 key、商户私钥、商户证书序列号、平台公钥或平台证书、异步通知地址；使用平台公钥模式时还必须填写平台公钥 ID；微信 H5 还必须填写 H5 场景信息。
- 微信支付平台公钥、平台公钥 ID 或平台证书轮换时，页面只支持覆盖写入，不展示明文；服务端应允许新旧签名材料在有效期内短期并存，过期后由维护者清理旧配置。
- 支付密钥配置只用于服务端渠道调用和回调验签，不进入用户端公开配置、前端环境变量、审计详情或运行日志。

权限建议：

- 页面入口：`page.system-settings.config`
- 页面可见资源：`system-config:view` 或 `system-config:*`
- 更新：`system-config:update` 或 `system-config:*`
- Logo 上传：`file:upload` 或 `file:*`

关联接口：

- `GET /admin-api/system-configs`
- `PATCH /admin-api/system-configs/{id}`
- `POST /admin-api/files/upload`（仅用于 `site.logo_url` 上传图片并回填 URL）

具体字段、响应和错误码以 `docs/server/api/` 为准。

## 管理员设置

页面职责：

- 在同一页面内承载管理员账号、管理组权限和管理员会话三块能力。
- 可以用标签页、分区或其它明确的信息架构切分三块能力。
- 不恢复 `/admin-users` 和 `/admin-roles` 独立侧栏菜单。
- 不新增管理员会话独立侧栏菜单或受保护路由。

管理员账号能力：

- 管理员列表
- 创建管理员
- 编辑管理员
- 状态切换
- 密码重置

管理组权限能力：

- 管理组列表
- 创建管理组
- 编辑管理组
- 状态切换
- 权限码分配

管理员会话能力：

- 按管理员设置页第三个 tab 展示 `admin_sessions` 对应的会话列表
- 展示会话状态、签发时间、过期时间、最近访问时间、最近访问 IP、User-Agent 等核心信息
- 区分当前会话与其它会话
- 支持吊销其它管理员会话
- 不支持从该 tab 吊销当前会话自身

权限建议：

- 管理员账号 tab 入口：`page.system-settings.admin-users`
- 管理员列表资源：`admin-user:view` 或 `admin-user:*`
- 新建管理员：`admin-user:create` 或 `admin-user:*`
- 编辑管理员与状态切换：`admin-user:update` 或 `admin-user:*`
- 重置管理员密码：`admin-user:password-reset` 或 `admin-user:*`
- 管理组权限 tab 入口：`page.system-settings.admin-roles`
- 管理组列表资源：`admin-role:view` 或 `admin-role:*`
- 新建管理组：`admin-role:create` 或 `admin-role:*`
- 编辑管理组、状态切换、权限分配：`admin-role:update` 或 `admin-role:*`
- 管理员会话 tab 入口：`page.system-settings.admin-sessions`
- 管理员会话列表资源：`admin-session:view` 或 `admin-session:*`
- 吊销管理员会话：`admin-session:revoke` 或 `admin-session:*`

关联接口：

- `GET /admin-api/admin-users`
- `POST /admin-api/admin-users`
- `GET /admin-api/admin-users/{id}`
- `PATCH /admin-api/admin-users/{id}`
- `POST /admin-api/admin-users/{id}/password`
- `GET /admin-api/admin-roles`
- `POST /admin-api/admin-roles`
- `GET /admin-api/admin-roles/{id}`
- `PATCH /admin-api/admin-roles/{id}`
- `GET /admin-api/admin-permissions`
- `GET /admin-api/admin-sessions`
- `PATCH /admin-api/admin-sessions/{session_id}`

具体字段、响应和错误码以 `docs/server/api/` 为准。

## 验收重点

- 系统设置只调用 `/admin-api/*`。
- 页面入口、标签页和按钮显隐都通过统一权限能力判断。
- 页面模板中不散写 `permissionCodes.includes(...)`。
- 敏感配置不展示明文。
- `site.logo_url` 不展示自由文本输入，应通过图片上传后回填 URL。
- 管理员、管理组和管理员会话能力都不恢复独立侧栏菜单。
- 管理员会话 tab 需要对当前会话提供明确标识，并阻止自吊销误操作。
- 系统设置不再提供日志管理入口。
