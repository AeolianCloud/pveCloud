# 用户端认证、账号与实名 API

本文档维护用户端公开配置、登录注册、账号资料、密码和实名相关接口，以及实名供应商回调入口。跨接口通用约定见 `docs/server/api/conventions.md`。

## 用户端公开配置域

### `GET /api/site-config`

- 鉴权：公开接口，无需登录
- 作用：读取 Web 端公开站点基础展示配置
- 数据来源：`system_configs` 中的公开配置白名单
- 返回字段：
  - `site_name`：站点显示名称，来自 `site.name`，为空时服务端返回默认值 `pveCloud`
  - `logo_url`：站点 Logo 图片公开展示 URL，来自 `site.logo_url`；当后台保存值为 `/admin-api/files/{id}` 或 `/admin-api/files/{id}/download` 时，服务端必须转换为公开读取地址 `/api/site-logo/{id}`；为空时返回空字符串
  - `login_captcha_enabled`：登录页验证码开关，来自 `web.auth.login_captcha_enabled`
  - `register_captcha_enabled`：注册页验证码开关，来自 `web.auth.register_captcha_enabled`
  - `password_reset_request_captcha_enabled`：忘记密码申请页验证码开关，来自 `web.auth.password_reset_request_captcha_enabled`
  - `password_reset_confirm_captcha_enabled`：重置密码确认页验证码开关，来自 `web.auth.password_reset_confirm_captcha_enabled`
  - `real_name`：实名公开配置对象，来自实名公开配置白名单，包含 `enabled`、`required_for_order`、`allowed_providers`、`default_provider`、`resubmit_enabled`、`max_submit_attempts`、`review_notice`
- 约束：
  - 不得返回 `is_secret=1` 的配置项，不得返回供应商密钥、网关、回调地址、返回地址、规则 ID 或任意配置键列表
  - 不得向用户端返回 `/admin-api/*` 管理端受保护地址
  - `real_name.allowed_providers` 对外返回时必须过滤未启用、必要后台配置不完整或缺少证件摘要密钥的外部供应商；无可用外部供应商且人工审核兜底已启用时返回 `manual`
  - `real_name.default_provider` 对外返回时必须是过滤后的可用实名方式；无可用外部供应商时返回 `manual`
  - `real_name.enabled` 对外返回时必须满足后台 `real_name.enabled=true` 且至少存在一个可用实名方式；缺少外部供应商、证件摘要密钥或回调地址时不得关闭已启用的人工审核实名入口

### `GET /api/site-logo/{id}`

- 鉴权：公开接口，无需登录
- 作用：读取公开站点 Logo 图片，仅用于 `GET /api/site-config` 返回的 Logo 展示地址
- 路径参数：
  - `id`：文件附件 ID
- 约束：
  - 只允许访问当前 `site.logo_url` 指向的附件 ID
  - 只允许图片类 MIME 类型
  - 不得开放任意文件下载能力，不得复用为通用公开附件下载接口
  - 响应不得暴露本地物理路径或管理端鉴权地址

## 用户端认证域

### `GET /api/auth/login-captcha`

- 鉴权：公开接口，无需登录
- 作用：获取登录页图形验证码
- 成功数据包含：`captcha_id`、`image`、`expires_in`
- 约束：
  - 仅当 `web.auth.login_captcha_enabled=true` 时开放
  - 验证码按登录场景单独生成，不能用于其它认证流程
  - 服务端对 `IP + scene` 独立限流

### `GET /api/auth/register-captcha`

- 鉴权：公开接口，无需登录
- 作用：获取注册页图形验证码
- 成功数据包含：`captcha_id`、`image`、`expires_in`
- 约束：
  - 仅当 `web.auth.register_captcha_enabled=true` 时开放
  - 验证码按注册场景单独生成，不能用于其它认证流程
  - 服务端对 `IP + scene` 独立限流

### `GET /api/auth/password-reset-request-captcha`

- 鉴权：公开接口，无需登录
- 作用：获取忘记密码申请页图形验证码
- 成功数据包含：`captcha_id`、`image`、`expires_in`
- 约束：
  - 仅当 `web.auth.password_reset_request_captcha_enabled=true` 时开放
  - 验证码按忘记密码申请场景单独生成，不能用于其它认证流程
  - 服务端对 `IP + scene` 独立限流

### `GET /api/auth/password-reset-confirm-captcha`

- 鉴权：公开接口，无需登录
- 作用：获取重置密码确认页图形验证码
- 成功数据包含：`captcha_id`、`image`、`expires_in`
- 约束：
  - 仅当 `web.auth.password_reset_confirm_captcha_enabled=true` 时开放
  - 验证码按重置密码确认场景单独生成，不能用于其它认证流程
  - 服务端对 `IP + scene` 独立限流

### `POST /api/auth/login`

- 鉴权：公开接口，无需登录
- 作用：用户登录，创建用户端会话并签发用户端 access token
- 请求字段：
  - `account`：用户名或邮箱
  - `password`：密码
  - `captcha_id`：可选；当 `web.auth.login_captcha_enabled=true` 时必填
  - `captcha_code`：可选；当 `web.auth.login_captcha_enabled=true` 时必填
- 成功数据包含：
  - `access_token`
  - `token_type`：固定 `Bearer`
  - `expires_in`：有效期秒数
  - `user`：用户摘要，包含 `id`、`username`、`email`、`display_name`、`status`
  - `session`：当前会话摘要，包含 `session_id`、`issued_at`、`expires_at`
- 约束：
  - 当 `web.auth.login_captcha_enabled=true` 时，验证码字段必须存在且校验通过
  - 当 `web.auth.login_captcha_enabled=false` 时，验证码字段忽略
  - 验证码错误、过期、缺失时返回明确错误
  - 仅 `status=active` 的用户允许登录；账号不存在或密码错误时返回未登录错误，用户被禁用时返回明确禁用错误

### `POST /api/auth/register`

- 鉴权：公开接口，无需登录
- 作用：注册用户端账号，创建用户端会话并签发用户端 access token
- 请求字段：
  - `username`：用户名，必须唯一
  - `email`：邮箱，必须唯一
  - `password`：密码
  - `captcha_id`：可选；当 `web.auth.register_captcha_enabled=true` 时必填
  - `captcha_code`：可选；当 `web.auth.register_captcha_enabled=true` 时必填
- 成功数据同登录接口
- 约束：
  - 当 `web.auth.register_captcha_enabled=true` 时，验证码字段必须存在且校验通过
  - 当 `web.auth.register_captcha_enabled=false` 时，验证码字段忽略
  - 验证码错误、过期、缺失时返回明确错误
  - 注册成功后 `users.status` 默认为 `active`
  - `username` 和 `email` 必须唯一；重复时返回状态冲突错误
  - 密码只保存 bcrypt 哈希，不返回明文或哈希
  - 注册不创建订单、实例、钱包、余额或其它业务资源

### `GET /api/auth/me`

- 鉴权：用户端 Bearer Token
- 作用：恢复当前用户登录态
- 成功数据包含当前用户真实摘要和当前会话摘要；用户被禁用后登录态恢复返回明确禁用错误

### `POST /api/auth/logout`

- 鉴权：用户端 Bearer Token
- 作用：吊销当前用户会话
- 成功数据：空对象

### `POST /api/auth/refresh`

- 鉴权：用户端 Bearer Token
- 作用：轮换当前用户 access token，创建新的用户端会话，并吊销旧用户端会话
- 成功数据同登录接口
- 约束：
  - 当前会话已过期或已吊销时返回未登录错误
  - 用户被禁用时返回明确禁用错误
  - refresh 成功后旧会话状态改为 `revoked`，`revoke_reason=refresh`
  - refresh 必须具备幂等保护；同一旧会话不能并发创建多个新 active 会话

### `POST /api/auth/password-reset/request`

- 鉴权：公开接口，无需登录
- 作用：申请密码重置邮件
- 请求字段：
  - `email`：用户邮箱
  - `captcha_id`：可选；当 `web.auth.password_reset_request_captcha_enabled=true` 时必填
  - `captcha_code`：可选；当 `web.auth.password_reset_request_captcha_enabled=true` 时必填
- 成功数据：空对象
- 约束：
  - 当 `web.auth.password_reset_request_captcha_enabled=true` 时，验证码字段必须存在且校验通过
  - 当 `web.auth.password_reset_request_captcha_enabled=false` 时，验证码字段忽略
  - 验证码错误、过期、缺失时返回明确错误
  - 无论邮箱是否存在，都返回统一成功响应，避免暴露账号存在性
  - 仅当邮箱对应 `status=active` 用户时，服务端创建一次性密码重置 token 并发送重置链接
  - token 原文只出现在邮件链接中，数据库只保存 token 哈希
  - 同一用户短时间重复申请时，应吊销旧的 active token 或复用未过期请求，不得产生多个可用 token
  - 未配置邮件发送能力时返回服务端配置错误，不创建 token

### `POST /api/auth/password-reset/confirm`

- 鉴权：公开接口，无需登录
- 作用：通过一次性 token 重置用户端账号密码
- 请求字段：
  - `token`：密码重置 token 原文
  - `password`：新密码
  - `captcha_id`：可选；当 `web.auth.password_reset_confirm_captcha_enabled=true` 时必填
  - `captcha_code`：可选；当 `web.auth.password_reset_confirm_captcha_enabled=true` 时必填
- 成功数据：空对象
- 约束：
  - 当 `web.auth.password_reset_confirm_captcha_enabled=true` 时，验证码字段必须存在且校验通过
  - 当 `web.auth.password_reset_confirm_captcha_enabled=false` 时，验证码字段忽略
  - 验证码错误、过期、缺失时返回明确错误
  - token 必须存在、未过期、未使用且状态为 `active`
  - token 对应用户必须仍为 `status=active`；用户已被禁用时拒绝重置并吊销该 token
  - 密码只保存 bcrypt 哈希，不返回明文或哈希
  - 重置成功后 token 状态改为 `used`，记录 `used_at`
  - 重置成功后吊销该用户所有 active 用户端会话，`revoke_reason=password_reset`
  - token 不存在、过期、已使用或已吊销时返回状态冲突或未授权错误，不泄露 token 对应用户信息

## 用户端账号资料域

### `PATCH /api/user/profile`

- 鉴权：用户端 Bearer Token
- 作用：当前登录用户编辑自己的基础资料
- 请求字段：
  - `email`：邮箱
  - `display_name`：显示名称，可为空
- 成功数据包含当前用户真实摘要和当前会话摘要
- 约束：
  - `username` 不允许通过用户端修改
  - `email` 必须唯一；与其它用户重复时返回状态冲突错误
  - 用户被禁用时返回明确禁用错误
  - 不允许修改状态、密码哈希、余额、角色或任何业务资源

### `POST /api/user/password`

- 鉴权：用户端 Bearer Token
- 作用：当前登录用户修改自己的密码
- 请求字段：
  - `current_password`：当前密码
  - `password`：新密码
- 成功数据：空对象
- 约束：
  - 当前密码错误时返回明确校验错误
  - 新密码只保存 bcrypt 哈希，不返回明文或哈希
  - 修改成功后吊销该用户除当前会话外的其它 active 用户端会话，`revoke_reason=password_change`
  - 当前会话保持有效，避免用户修改密码后被立即踢出

## 用户端实名域

### `POST /api/user/real-name`

- 鉴权：用户端 Bearer Token
- 作用：当前登录用户创建个人实名核验申请；有可用支付宝/微信侧供应商时返回外部核验入口，没有可用外部供应商时创建人工审核申请
- 请求字段：
  - `real_name`：真实姓名
  - `id_type`：证件类型，当前仅允许 `id_card`
  - `id_number`：证件号码
  - `provider`：实名方式，允许值来自服务端过滤后的 `real_name.allowed_providers`，当前支持 `alipay`、`wechat`、`manual`；为空时使用 `real_name.default_provider`
- 成功数据包含：
  - `application`：最新实名申请摘要
  - `provider_action`：核验动作，包含 `provider`、`action_type`、`redirect_url`、`expires_at`；外部供应商模式不得包含供应商密钥、完整签名参数或证件号码明文；人工审核模式 `provider=manual`、`action_type=manual_review`、`redirect_url` 为空
- 约束：
  - 仅当 `real_name.enabled=true` 时允许创建申请
  - `provider=manual` 时创建人工审核申请
  - `provider=alipay` 或 `provider=wechat` 时，供应商必须已在后台系统配置中启用、对应供应商必要配置完整，且 `real_name.identity_digest_secret` 已配置
  - 当前用户存在 `pending` 申请时拒绝重复创建；用户返回页面后应调用同步接口查询结果
  - 当前用户已 `approved` 时拒绝用户端覆盖实名资料
  - 拒绝后是否允许重新提交由 `real_name.resubmit_enabled` 和 `real_name.max_submit_attempts` 决定
  - 证件号码必须通过格式校验；外部供应商模式下不得与其它已通过实名用户重复，兼容历史数据期间重复校验必须同时覆盖当前 HMAC 摘要和历史 `sha256-legacy` 摘要
  - 证件号码不得明文落库；外部供应商模式保存带后台敏感配置密钥的查询摘要和脱敏展示值，人工审核模式在未配置摘要密钥时只保存脱敏展示值，接口不得返回明文
  - 创建外部供应商会话失败时，不得把申请误标记为可继续核验或已通过
  - 服务端调用支付宝或微信/腾讯云接口时不得记录供应商密钥、证件号码明文、真实姓名明文或完整响应
  - 当前人工审核流程不接收实名图片附件，提交请求不得包含图片附件字段

### `GET /api/user/real-name`

- 鉴权：用户端 Bearer Token
- 作用：读取当前登录用户的实名状态和最新实名申请摘要
- 成功数据包含：
  - `status`：`unverified`、`pending`、`approved`、`rejected`
  - `application`：最新实名申请摘要；无申请时为空
  - `config`：实名提交相关公开配置快照
- 申请摘要包含：申请编号、真实姓名、证件类型、脱敏证件号码、实名供应商、供应商状态、状态、失败原因、提交次数、提交时间、核验完成时间
- 约束：不得返回证件号码明文、供应商完整响应、供应商密钥或后台敏感信息

### `POST /api/user/real-name/sync`

- 鉴权：用户端 Bearer Token
- 作用：当前登录用户从支付宝或微信侧返回后，触发服务端查询最新供应商核验结果
- 请求字段：
  - `application_no`：申请编号；为空时同步当前用户最新 `pending` 申请
- 成功数据包含最新实名状态响应
- 约束：
  - 只能同步当前登录用户自己的实名申请
  - 只有 `pending` 外部供应商申请允许同步
  - 同步由服务端调用供应商查询接口并验签或校验响应可信性，前端 URL 参数不能作为通过依据
  - 供应商未出最终结果时保持 `pending`
  - 供应商已通过但证件号码与其它已通过用户重复时，本地不得通过当前申请；兼容历史数据期间，重复校验必须同时覆盖当前 HMAC 摘要和历史 `sha256-legacy` 摘要

### `POST /api/real-name/provider-callbacks/{provider}`

- 鉴权：公开回调接口；必须执行供应商签名校验、时间窗口校验和幂等检查
- 作用：接收已开放供应商的实名核验异步通知；当前仅开放支付宝回调，微信/腾讯云回调暂未开放，微信结果通过同步查询确认
- 路径参数：
  - `provider`：当前仅 `alipay` 可处理；`wechat` 返回参数错误并要求使用同步查询
- 约束：
  - 该接口不使用用户 Bearer Token
  - 回调必须通过供应商验签后才可更新本地申请
  - 回调必须通过本地申请编号和供应商会话编号定位申请
  - 重复回调必须幂等；已进入 `approved` 或 `rejected` 的申请不得被低可信或旧状态覆盖
  - 回调处理不得记录证件号码明文、真实姓名明文、供应商密钥或完整响应
