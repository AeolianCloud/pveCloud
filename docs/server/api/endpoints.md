# API 接口总览

本文档维护当前已确认的接口清单与主要契约口径。跨接口通用约定见 `docs/server/api/conventions.md`。

## 实现边界提示

接口契约按访问边界区分：

- `/admin-api/*`：由 `server/internal/delivery/http/admin/*` 聚合路由，业务编排落在 `server/internal/usecase/admin/*`，GORM model 和可复用查询对象落在 `server/internal/repository/mysql/*`
- `/api/*`：由 `server/internal/delivery/http/web/*` 聚合路由，业务编排落在 `server/internal/usecase/web/*`，GORM model 和可复用查询对象落在 `server/internal/repository/mysql/*`

这里描述的是 API 契约，不直接替代具体代码结构；但当接口重新开放、迁移或新增时，路由注册、权限校验和实现目录应与上述边界保持一致。

## 系统检查

### `GET /healthz`

- 鉴权：无
- 作用：检查 API 进程、MariaDB 和 Redis 是否可用

### `GET /admin-api/ping`

- 鉴权：无
- 作用：管理端 API 入口连通性检查

## 管理端认证与会话

### `GET /admin-api/auth/captcha`

- 鉴权：无
- 作用：获取管理端登录验证码
- 成功数据包含：`captcha_id`、验证码图片、有效期

### `POST /admin-api/auth/login`

- 鉴权：无
- 作用：管理员账号密码登录
- 请求字段：`username`、`password`、`captcha_id`、`captcha_code`
- 成功数据包含：
  - `access_token`
  - `token_type`
  - `expires_in`
  - `admin`
  - `role_ids`
  - `permission_codes`
  - `menus`
  - `session`

### `GET /admin-api/auth/me`

- 鉴权：管理端 Bearer Token
- 作用：恢复当前管理员、权限快照、后端菜单树与会话状态
- 成功数据包含：
  - `admin`
  - `role_ids`
  - `permission_codes`
  - `menus`
  - `session`
- `menus` 由 `admin_permissions` 中 `type=menu` 且当前管理员拥有的权限节点生成，前端侧栏按该树渲染。

### `POST /admin-api/auth/logout`

- 鉴权：管理端 Bearer Token
- 作用：注销当前会话

### `POST /admin-api/auth/refresh`

- 鉴权：管理端 Bearer Token
- 作用：轮换新 token 和新会话
- 成功响应结构与登录成功响应保持一致

## 管理端 Dashboard

### `GET /admin-api/dashboard`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.dashboard`
- 作用：获取当前基础后台首页数据
- 成功数据包含：
  - `admin`
  - `role_ids`
  - `permission_codes`
  - `menus`
  - `session`
  - `metrics`

当前阶段 Dashboard 只展示基础后台相关指标，不展示未开放业务模块数据。

## 管理员账号域

### `GET /admin-api/admin-users`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.admin-users`
- 作用：分页查询管理员账号

### `POST /admin-api/admin-users`

- 鉴权：管理端 Bearer Token
- 操作权限：`admin-user:create` 或 `admin-user:*`
- 作用：创建管理员账号
- 约束：若创建时分配角色，目标角色展开后的全部权限必须是当前操作者实时数据库权限集合的子集

### `GET /admin-api/admin-users/{id}`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.admin-users`
- 作用：查看管理员详情

### `PATCH /admin-api/admin-users/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`admin-user:update` 或 `admin-user:*`
- 作用：更新管理员信息、状态和角色
- 约束：管理员不能通过该接口修改自己的 `role_ids`；给其它管理员分配角色时，目标角色展开后的全部权限必须是当前操作者实时数据库权限集合的子集

### `POST /admin-api/admin-users/{id}/password`

- 鉴权：管理端 Bearer Token
- 操作权限：`admin-user:password-reset` 或 `admin-user:*`
- 作用：重置管理员密码

## 角色与权限域

### `GET /admin-api/admin-roles`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.admin-roles`
- 作用：查询角色列表

### `POST /admin-api/admin-roles`

- 鉴权：管理端 Bearer Token
- 操作权限：`admin-role:create` 或 `admin-role:*`
- 作用：创建角色
- 约束：提交的 `permission_codes` 必须是当前操作者实时数据库权限集合的子集，禁止通过角色创建授予操作者未拥有的权限

### `GET /admin-api/admin-roles/{id}`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.admin-roles`
- 作用：查看角色详情

### `PATCH /admin-api/admin-roles/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`admin-role:update` 或 `admin-role:*`
- 作用：更新角色信息、状态和权限
- 约束：提交的 `permission_codes` 必须是当前操作者实时数据库权限集合的子集，禁止通过角色编辑授予操作者未拥有的权限

### `GET /admin-api/admin-permissions`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.admin-roles`
- 作用：只读查询菜单和操作权限目录树
- 成功数据为树形节点数组，每个节点包含：`code`、`name`、`type`、`parent_code`、`path`、`icon`、`sort_order`、`description`、`children`

## 管理员会话域

### `GET /admin-api/admin-sessions`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.admin-sessions`
- 作用：分页查询管理员会话列表
- 查询参数支持：`page`、`per_page`、`keyword`、`status`

### `PATCH /admin-api/admin-sessions/{session_id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`admin-session:revoke` 或 `admin-session:*`
- 作用：吊销指定管理员会话
- 请求字段：`status`，当前固定为 `revoked`
- 约束：不得通过该接口吊销当前会话自身

## 系统配置域

### `GET /admin-api/system-configs`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.config`
- 作用：按配置分组查询系统配置
- 成功数据包含配置分组和配置项；配置项包含 `id`、`config_key`、`config_value`、`value_type`、`group_name`、`is_secret`、`has_value`、`description`
- 约束：
  - `is_secret=1` 的配置不得返回明文，`config_value` 必须为空或固定掩码，只通过 `has_value` 表示是否已配置
  - 支付宝、微信侧实名供应商密钥和证件摘要密钥都作为后台敏感配置管理

### `PATCH /admin-api/system-configs/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`system-config:update` 或 `system-config:*`
- 作用：更新系统配置
- 约束：
  - 更新 `is_secret=1` 配置时，仅非空新值会覆盖旧值；空值表示保留旧敏感值
  - `real_name.identity_digest_secret` 是外部供应商实名申请和证件摘要重复校验的敏感配置；缺少时外部供应商不可用，但不影响人工审核实名入口；已有当前 HMAC 版本实名申请后不得通过普通系统设置直接修改
  - 更新实名供应商启用开关时，服务端必须校验对应供应商必要后台配置是否完整

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

## Web 用户管理域

### `GET /admin-api/users`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.web-users`
- 作用：分页查询用户端账号列表
- 查询参数支持：`page`、`per_page`、`keyword`、`status`
- 成功数据包含：
  - `list`
  - `total`
  - `page`
  - `per_page`
  - `last_page`
- 列表项包含：id、username、email、display_name、status、created_at、updated_at
- 约束：不得返回 `password_hash`

### `POST /admin-api/users`

- 鉴权：管理端 Bearer Token
- 操作权限：`web-user:create` 或 `web-user:*`
- 作用：创建用户端账号
- 请求字段：`username`、`email`、`password`、`display_name`、`status`
- 约束：`username` 和 `email` 必须唯一；密码只保存 bcrypt 哈希

### `GET /admin-api/users/{id}`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.web-users`
- 作用：查看用户端账号详情
- 成功数据包含用户摘要，不包含 `password_hash`

### `PATCH /admin-api/users/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`web-user:update` 或 `web-user:*`
- 作用：编辑用户端账号邮箱、显示名称和状态
- 请求字段：`email`、`display_name`、`status`
- 约束：用户被设置为 `disabled` 后，后续 Web 受保护接口必须拒绝该用户 token

### `POST /admin-api/users/{id}/password`

- 鉴权：管理端 Bearer Token
- 操作权限：`web-user:password-reset` 或 `web-user:*`
- 作用：重置用户端账号密码
- 请求字段：`password`
- 约束：密码只保存 bcrypt 哈希，不返回明文或哈希

### `GET /admin-api/user-sessions`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.web-user-sessions`
- 作用：分页查询用户端登录会话
- 查询参数支持：`page`、`per_page`、`user_id`、`status`、`date_from`、`date_to`
- 成功数据包含：
  - `list`
  - `total`
  - `page`
  - `per_page`
  - `last_page`
- 列表项包含用户摘要、session_id、status、issued_at、expires_at、revoked_at、revoke_reason、last_seen_at、last_seen_ip、user_agent、created_at

### `PATCH /admin-api/user-sessions/{session_id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`web-user-session:revoke` 或 `web-user-session:*`
- 作用：吊销指定用户端登录会话
- 请求字段：`status`，当前固定为 `revoked`
- 约束：仅 active 状态会话可吊销；吊销后对应 Web token 后续访问必须失效

## 实名管理域

### `GET /admin-api/real-name-applications`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.real-name-management`
- 作用：分页查询用户实名申请
- 查询参数支持：`page`、`per_page`、`keyword`、`status`、`id_type`、`provider`、`provider_status`、`date_from`、`date_to`
- 成功数据包含：`list`、`total`、`page`、`per_page`、`last_page`
- 列表项包含：申请编号、用户摘要、真实姓名、证件类型、脱敏证件号码、实名供应商、供应商状态、状态、提交次数、提交时间、核验完成时间、失败原因
- 约束：不得返回证件号码明文或供应商完整响应

### `GET /admin-api/real-name-applications/{id}`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.real-name-management`
- 作用：查看用户实名申请详情
- 成功数据包含：申请编号、用户摘要、真实姓名、证件类型、脱敏证件号码、实名供应商、供应商会话摘要、供应商状态、供应商结果码、供应商结果说明、状态、提交次数、核验完成时间、失败原因、创建时间、更新时间
- 约束：
  - 不得返回证件号码明文、供应商完整响应或供应商密钥

### `POST /admin-api/real-name-applications/{id}/sync`

- 鉴权：管理端 Bearer Token
- 操作权限：`real-name:sync` 或 `real-name:*`
- 作用：后台触发服务端重新查询支付宝或微信实名供应商结果
- 成功数据包含最新实名申请详情
- 约束：
  - 只有外部供应商申请允许同步
  - 同步调用不得放入长事务；供应商查询完成后再以本地事务更新状态和审计
  - 同步操作必须写入 `admin_audit_logs`，动作使用 `real_name.sync`

### `POST /admin-api/real-name-applications/{id}/review`

- 鉴权：管理端 Bearer Token
- 操作权限：`real-name:review` 或 `real-name:*`
- 作用：后台通过或拒绝人工审核实名申请
- 请求字段：
  - `status`：固定允许 `approved` 或 `rejected`
  - `reason`：拒绝原因；`status=rejected` 时必填
- 成功数据包含最新实名申请详情
- 约束：
  - 只有 `verification_provider=manual` 且 `status=pending` 的申请允许审核
  - 审核通过时不得补写证件号码明文
  - 审核操作必须写入 `admin_audit_logs`，动作使用 `real_name.review`

## 日志管理域

### `GET /admin-api/audit-logs`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.audit-logs`
- 敏感详情权限：`audit-log:sensitive-view` 或 `audit-log:*`
- 作用：分页查询普通后台审计日志，可用于日志管理页面的操作日志 tab 和登录日志 tab
- 查询参数支持：`page`、`per_page`、`admin_id`、`action`、`object_type`、`object_id`、`date_from`、`date_to`
- 成功数据包含：
  - `list`
  - `total`
  - `page`
  - `per_page`
  - `last_page`

列表项包含操作者摘要、会话 ID、请求 ID、请求方法、请求路径、操作动作、对象类型、对象 ID、IP、备注和创建时间。
未具备敏感详情权限时，`before_data`、`after_data` 和 `user_agent` 不返回。

登录日志 tab 不新增独立接口或表，使用本接口并固定 `object_type=admin_auth` 查询认证相关日志；如需按动作类型筛选，继续使用单个 `action` 查询参数。

## 文件管理域

### `POST /admin-api/files/upload`

- 鉴权：管理端 Bearer Token
- 操作权限：`file:upload` 或 `file:*`
- 作用：上传单个文件（图片/附件）
- 请求格式：`multipart/form-data`
- 请求字段：`file`（文件流）
- 安全校验：
  - 扩展名白名单校验（jpg/png/gif/webp/pdf）
  - 声明 MIME 类型必须在白名单内
  - Magic Bytes 文件头必须匹配扩展名和声明 MIME 类型，防止伪装文件
  - 危险文件类型黑名单拦截（php/exe/sh/bat/js/html 等）
  - 单文件最大 10MB（可配置）
  - 上传读取必须限制最大字节数，避免超大文件被完整读入内存
  - 路径穿越防护：原始文件名只保留 basename，存储文件名强制使用随机 UUID
- 成功数据包含：
  - `id`：附件 ID
  - `original_name`：原始文件名
  - `mime_type`：MIME 类型
  - `size`：文件大小（字节）
  - `url`：文件访问 URL
  - `created_at`：上传时间
- 存储：数据库只保存相对存储路径，不保存本地根目录
- 审计：文件记录和审计日志必须在同一事务中写入；事务失败时清理已写入的物理文件

### `GET /admin-api/files`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.file-management`
- 作用：分页查询文件列表
- 查询参数支持：`page`、`per_page`、`keyword`、`mime_type`、`uploader_id`、`date_from`、`date_to`
- 成功数据包含：
  - `list`：文件列表
  - `total`：总数
  - `page`：当前页
  - `per_page`：每页数量
  - `last_page`：最后一页
- 列表项包含：id、original_name、mime_type、size、uploader 信息、created_at

### `GET /admin-api/files/{id}`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.file-management`
- 作用：查看文件详情
- 成功数据包含：完整文件元信息、引用信息、可用操作信息

### `GET /admin-api/files/{id}/download`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.file-management`
- 作用：安全下载或预览文件
- 约束：
  - 仅允许已授权管理员访问
  - 仅返回非删除状态文件
  - 如果文件被历史实名申请引用，通用文件下载接口必须拒绝访问，不提供预览或下载
  - 下载响应不得暴露物理存储路径
  - 图片和 PDF 可直接预览，其它类型走下载

### `GET /admin-api/files/{id}/references`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.file-management`
- 作用：查看文件引用关系，用于详情抽屉和删除前校验
- 成功数据包含：
  - `file_id`
  - `reference_count`
  - `references`
- `references` 用于展示被哪些业务记录引用，后续公告、工单、页面配置等业务域可复用

### `DELETE /admin-api/files/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`file:delete` 或 `file:*`
- 作用：删除文件（软删除，物理文件保留）
- 约束：
  - 若文件仍被业务记录引用，必须阻止删除并返回明确错误
  - 删除前应先通过引用接口或服务端校验确认无引用
- 审计：软删除状态和审计日志必须在同一事务中写入（action: `file.delete`）

## 产品目录

产品目录维护服务器产品展示和可售约束。订单 MVP 创建订单时读取产品目录并保存快照，但产品目录本身不发起支付、不创建实例、不绑定 PVE 节点。网络类型当前只作为后台可维护的套餐可选项、Web 下单选择项和订单快照；后续对接 PVE 时可基于网络类型编码映射真实网络。

### `GET /admin-api/products`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.products`
- 作用：分页查看产品主数据
- 支持按 `type`、`status`、`keyword` 查询

### `POST /admin-api/products`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:create` 或 `product:*`
- 作用：创建产品，当前 `type` 仅允许 `server`
- 审计：`product.create`

### `PUT /admin-api/products/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:update` 或 `product:*`
- 作用：编辑产品名称、slug、介绍、可见性和排序
- 审计：`product.update`

### `PATCH /admin-api/products/{id}/status`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:publish` 或 `product:*`
- 作用：切换产品 `draft`、`active`、`inactive` 状态
- 审计：`product.status.update`

### `DELETE /admin-api/products/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:delete` 或 `product:*`
- 作用：删除产品主数据
- 约束：产品存在套餐时不得删除，应先处理套餐；历史订单只依赖订单快照，不阻止删除
- 审计：`product.delete`

### `GET /admin-api/product-plans`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.products`
- 作用：分页查看服务器套餐和规格
- 支持按 `product_id`、`status`、`keyword` 查询

### `POST /admin-api/product-plans`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:create` 或 `product:*`
- 作用：创建固定服务器套餐
- 约束：套餐只保存固定规格，不提供自定义配置器
- 审计：`product_plan.create`

### `PUT /admin-api/product-plans/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:update` 或 `product:*`
- 作用：编辑套餐规格、介绍、推荐、可见性和排序
- 审计：`product_plan.update`

### `PATCH /admin-api/product-plans/{id}/status`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:publish` 或 `product:*`
- 作用：切换套餐 `draft`、`active`、`inactive`、`sold_out` 状态
- 审计：`product_plan.status.update`

### `DELETE /admin-api/product-plans/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:delete` 或 `product:*`
- 作用：删除固定服务器套餐
- 约束：同事务删除套餐周期价格、套餐销售地域关联和套餐系统模板关联；历史订单只依赖订单快照，不阻止删除
- 审计：`product_plan.delete`

### `PUT /admin-api/product-plans/{id}/prices`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:update` 或 `product:*`
- 作用：覆盖保存套餐周期价格
- 金额单位：分，不使用浮点数
- 支持周期：`monthly`、`quarterly`、`semi_yearly`、`yearly`
- 审计：`product_plan.prices.update`

### `GET /admin-api/product-plans/{id}/prices`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.products`
- 作用：读取套餐当前周期价格，用于产品管理页面回显

### `PUT /admin-api/product-plans/{id}/regions`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:update` 或 `product:*`
- 作用：覆盖保存套餐可用销售地域
- 审计：`product_plan.regions.update`

### `GET /admin-api/product-plans/{id}/regions`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.products`
- 作用：读取套餐当前可用销售地域，用于产品管理页面回显

### `PUT /admin-api/product-plans/{id}/os-templates`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:update` 或 `product:*`
- 作用：覆盖保存套餐可用服务器系统模板
- 审计：`product_plan.os_templates.update`

### `GET /admin-api/product-plans/{id}/os-templates`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.products`
- 作用：读取套餐当前可用服务器系统模板，用于产品管理页面回显

### `PUT /admin-api/product-plans/{id}/network-types`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:update` 或 `product:*`
- 作用：覆盖保存套餐可用网络类型
- 审计：`product_plan.network_types.update`

### `GET /admin-api/product-plans/{id}/network-types`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.products`
- 作用：读取套餐当前可用网络类型，用于产品管理页面回显

### `GET /admin-api/sales-regions`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.products`
- 作用：查看销售地域列表

### `POST /admin-api/sales-regions`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:create` 或 `product:*`
- 作用：创建销售地域。销售地域不绑定 PVE 节点。
- 审计：`sales_region.create`

### `PUT /admin-api/sales-regions/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:update` 或 `product:*`
- 作用：编辑销售地域
- 审计：`sales_region.update`

### `DELETE /admin-api/sales-regions/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:delete` 或 `product:*`
- 作用：删除销售地域
- 约束：销售地域仍被套餐关联时不得删除；历史订单只依赖订单快照，不阻止删除
- 审计：`sales_region.delete`

### `GET /admin-api/server-os-templates`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.products`
- 作用：查看服务器系统模板列表

### `POST /admin-api/server-os-templates`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:create` 或 `product:*`
- 作用：创建服务器系统模板。当前不绑定 PVE 模板。
- 审计：`server_os_template.create`

### `PUT /admin-api/server-os-templates/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:update` 或 `product:*`
- 作用：编辑服务器系统模板
- 审计：`server_os_template.update`

### `DELETE /admin-api/server-os-templates/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:delete` 或 `product:*`
- 作用：删除服务器系统模板
- 约束：系统模板仍被套餐关联时不得删除；历史订单只依赖订单快照，不阻止删除
- 审计：`server_os_template.delete`

### `GET /admin-api/network-types`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.products`
- 作用：查看网络类型列表

### `POST /admin-api/network-types`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:create` 或 `product:*`
- 作用：创建网络类型。当前不绑定 PVE 网络。
- 审计：`network_type.create`

### `PUT /admin-api/network-types/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:update` 或 `product:*`
- 作用：编辑网络类型
- 审计：`network_type.update`

### `DELETE /admin-api/network-types/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:delete` 或 `product:*`
- 作用：删除网络类型
- 约束：网络类型仍被套餐关联时不得删除；历史订单只依赖订单快照，不阻止删除
- 审计：`network_type.delete`

### `GET /api/server-catalog`

- 鉴权：公开接口，不要求用户登录
- 作用：返回 Web 可展示服务器产品目录聚合数据
- 返回范围：已上架且可见的服务器产品、套餐、周期价格、销售地域、服务器系统模板和网络类型
- 展示约束：套餐需要至少有一个 active 周期价格、一个 active 且 visible 的销售地域、一个 active 且 visible 的服务器系统模板、一个 active 且 visible 的网络类型才进入公开目录
- 禁止返回：支付、实例、库存扣减、PVE 节点、PVE 模板 ID、PVE 网络 ID 或资源池信息

## 用户端订单 MVP

订单 MVP 只保存用户购买意向和后台人工处理所需快照，不发起支付、不创建实例、不下发 PVE 资源。

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

### `GET /api/orders`

- 鉴权：用户端 Bearer Token
- 作用：分页查询当前用户自己的订单列表
- 查询参数支持：`page`、`per_page`、`status`
- 列表项包含订单编号、状态、产品名称、套餐名称、计费周期、网络类型、订单金额和创建/取消/关闭时间

### `GET /api/orders/{order_no}`

- 鉴权：用户端 Bearer Token
- 作用：查看当前用户自己的订单详情
- 成功数据包含产品、套餐、价格、销售地域、系统模板和网络类型快照

### `POST /api/orders/{order_no}/cancel`

- 鉴权：用户端 Bearer Token
- 作用：取消当前用户自己的 `pending` 订单
- 请求字段：`reason` 可选，最多 500 字

## 工单

工单提供用户与后台之间的文字和附件沟通能力，并支持管理端内部指派、转派、协作者、内部备注、内部 SLA、标签和优先级升级。不承诺支付、实例交付、PVE 操作、用户侧 SLA、实时推送、邮件或站内信通知。

### 用户端工单接口

#### `GET /api/tickets`

- 鉴权：用户端 Bearer Token
- 作用：分页查询当前用户自己的工单列表
- 查询参数支持：`page`、`per_page`、`status`、`category`、`priority`、`order_no`
- 列表项包含工单编号、标题、分类、优先级、状态、公开标签、关联订单号、最近消息时间和创建时间
- 约束：只能返回当前登录用户自己的工单

#### `POST /api/tickets`

- 鉴权：用户端 Bearer Token
- 请求格式：`multipart/form-data`
- 作用：创建工单
- 请求字段：`title`、`category`、`priority`、`content`、`order_no`、`attachments`
- `category` 允许：`account`、`order`、`product`、`technical`、`billing`、`other`
- `priority` 允许：`low`、`normal`、`high`、`urgent`；为空时使用 `normal`
- `order_no` 可选；填写时必须属于当前登录用户
- `attachments` 可选，单条消息最多 5 个附件
- 成功数据包含工单详情
- 约束：
  - 创建后工单状态为 `waiting_admin`
  - 创建工单必须同时创建第一条用户消息
  - 附件必须通过文件大小、扩展名、声明 MIME、Magic Bytes 和危险扩展名校验
  - 工单、首条消息、附件记录和文件引用必须同事务写入
  - 不得信任前端传入的用户 ID、订单归属、附件归属或状态

#### `GET /api/tickets/{ticket_no}`

- 鉴权：用户端 Bearer Token
- 作用：查看当前用户自己的工单详情
- 成功数据包含工单基础信息、用户可见状态、公开标签、消息时间线和附件摘要
- 约束：只能查看当前登录用户自己的工单；他人工单不得通过错误文案泄露存在性

#### `POST /api/tickets/{ticket_no}/messages`

- 鉴权：用户端 Bearer Token
- 请求格式：`multipart/form-data`
- 作用：回复当前用户自己的未关闭工单
- 请求字段：`content`、`attachments`
- `attachments` 可选，单条消息最多 5 个附件
- 成功数据包含最新工单详情
- 约束：
  - 只能回复当前登录用户自己的工单
  - `closed` 工单不可回复
  - 用户回复后工单状态变为 `waiting_admin`
  - 消息、附件记录、文件引用和工单最近消息时间必须同事务写入

#### `POST /api/tickets/{ticket_no}/close`

- 鉴权：用户端 Bearer Token
- 作用：关闭当前用户自己的未关闭工单
- 请求字段：`reason` 可选，最多 500 字
- 成功数据包含最新工单详情
- 约束：只能关闭当前登录用户自己的未关闭工单，关闭后不可继续回复

#### `GET /api/tickets/{ticket_no}/attachments/{file_id}/download`

- 鉴权：用户端 Bearer Token
- 作用：下载或预览当前用户自己工单消息中的附件
- 约束：
  - 必须校验工单属于当前登录用户
  - 必须校验附件属于该工单消息
  - 下载响应不得暴露物理存储路径
  - 图片和 PDF 可直接预览，其它允许类型走下载
  - 文件名进入响应头前必须安全编码
  - 受保护下载响应不得被共享缓存长期保存

### 管理端工单接口

#### `GET /admin-api/tickets`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.tickets`
- 作用：分页查询工单列表
- 查询参数支持：`page`、`per_page`、`status`、`category`、`priority`、`ticket_no`、`order_no`、`user_keyword`、`date_from`、`date_to`
- 扩展查询参数支持：`assignee_admin_id`、`tag_id`、`sla_status`
- `sla_status` 允许：`normal`、`first_response_overdue`、`resolution_overdue`
- 列表项包含工单编号、用户摘要、标题、分类、优先级、状态、处理人摘要、标签、内部 SLA 状态、关联订单号、最近消息时间和创建时间

#### `GET /admin-api/tickets/{ticket_no}`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.tickets`
- 作用：查看工单详情
- 成功数据包含工单基础信息、用户摘要、可选订单号、处理人摘要、协作者摘要、标签、内部 SLA 状态、消息时间线、附件摘要、内部备注和操作历史

#### `POST /admin-api/tickets/{ticket_no}/messages`

- 鉴权：管理端 Bearer Token
- 操作权限：`ticket:reply` 或 `ticket:*`
- 请求格式：`multipart/form-data`
- 作用：管理员回复未关闭工单
- 请求字段：`content`、`attachments`
- `attachments` 可选，单条消息最多 5 个附件
- 成功数据包含最新工单详情
- 约束：
  - `closed` 工单不可回复
  - 管理员回复后工单状态变为 `waiting_user`
  - 消息、附件记录、文件引用、工单最近消息时间和后台审计必须同事务写入
  - 审计动作使用 `ticket.reply`

#### `POST /admin-api/tickets/{ticket_no}/close`

- 鉴权：管理端 Bearer Token
- 操作权限：`ticket:close` 或 `ticket:*`
- 作用：管理员关闭未关闭工单
- 请求字段：`reason` 可选，最多 500 字
- 成功数据包含最新工单详情
- 约束：
  - `closed` 工单不可重复关闭
  - 关闭工单和后台审计必须同事务写入
  - 审计动作使用 `ticket.close`

#### `GET /admin-api/tickets/assignee-candidates`

- 鉴权：管理端 Bearer Token
- 操作权限：`ticket:assign` 或 `ticket:*`
- 作用：查询可指派管理员候选
- 查询参数支持：`keyword`、`page`、`per_page`
- 列表项包含管理员 ID、用户名、显示名和邮箱
- 约束：只返回 `active` 且具备 `page.tickets`、`ticket:reply` 或 `ticket:*` 的管理员

#### `POST /admin-api/tickets/{ticket_no}/assign`

- 鉴权：管理端 Bearer Token
- 操作权限：`ticket:assign` 或 `ticket:*`
- 作用：指派或转派未关闭工单
- 请求字段：`assignee_admin_id` 必填，`reason` 可选，最多 500 字
- 成功数据包含最新工单详情
- 约束：
  - `closed` 工单不可指派或转派
  - 目标管理员必须是可指派候选
  - 当前无处理人时记录为指派；已有处理人变更时记录为转派
  - 指派或转派、操作历史和后台审计必须同事务写入
  - 审计动作使用 `ticket.assign`

#### `POST /admin-api/tickets/{ticket_no}/collaborators`

- 鉴权：管理端 Bearer Token
- 操作权限：`ticket:collaborate` 或 `ticket:*`
- 作用：添加工单协作者
- 请求字段：`admin_id` 必填
- 成功数据包含最新工单详情
- 约束：`closed` 工单不可新增协作者；目标管理员必须是可指派候选；重复添加应保持幂等

#### `DELETE /admin-api/tickets/{ticket_no}/collaborators/{admin_id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`ticket:collaborate` 或 `ticket:*`
- 作用：移除工单协作者
- 成功数据包含最新工单详情
- 约束：`closed` 工单不可移除协作者；不存在的协作者移除请求应保持幂等

#### `POST /admin-api/tickets/{ticket_no}/internal-notes`

- 鉴权：管理端 Bearer Token
- 操作权限：`ticket:note` 或 `ticket:*`
- 作用：追加内部备注
- 请求字段：`content` 必填，最多 5000 字
- 成功数据包含最新工单详情
- 约束：内部备注只追加、不编辑、不删除，不返回用户端；备注、操作历史和后台审计必须同事务写入

#### `POST /admin-api/tickets/{ticket_no}/priority`

- 鉴权：管理端 Bearer Token
- 操作权限：`ticket:priority` 或 `ticket:*`
- 作用：升级工单优先级
- 请求字段：`priority` 必填，允许 `low`、`normal`、`high`、`urgent`；`reason` 必填，最多 500 字
- 成功数据包含最新工单详情
- 约束：
  - 只能从低紧急度升到更高紧急度，不支持降级或同级更新
  - `closed` 工单不可升级优先级
  - 未完成的 SLA 截止时间只允许提前，不允许延后
  - 优先级升级、操作历史和后台审计必须同事务写入
  - 审计动作使用 `ticket.priority_upgrade`

#### `PUT /admin-api/tickets/{ticket_no}/tags`

- 鉴权：管理端 Bearer Token
- 操作权限：`ticket:tag` 或 `ticket:*`
- 作用：整体替换工单标签绑定
- 请求字段：`tag_ids` 必填，最多 20 个
- 成功数据包含最新工单详情
- 约束：只允许绑定启用标签；绑定变更、操作历史和后台审计必须同事务写入

#### `GET /admin-api/ticket-tags`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.tickets`
- 作用：查询工单标签字典
- 查询参数支持：`page`、`per_page`、`keyword`、`visibility`、`status`
- 列表项包含标签 ID、名称、颜色、可见性、状态、排序和创建时间

#### `POST /admin-api/ticket-tags`

- 鉴权：管理端 Bearer Token
- 操作权限：`ticket:tag-manage` 或 `ticket:*`
- 作用：创建工单标签
- 请求字段：`name` 必填，最多 40 字；`color` 可选；`visibility` 必填，允许 `public`、`internal`；`status` 必填，允许 `active`、`disabled`；`sort_order` 可选
- 成功数据包含标签详情
- 约束：标签名称全局唯一；创建和后台审计必须同事务写入

#### `PATCH /admin-api/ticket-tags/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`ticket:tag-manage` 或 `ticket:*`
- 作用：更新工单标签
- 请求字段：`name`、`color`、`visibility`、`status`、`sort_order` 均可选
- 成功数据包含标签详情
- 约束：标签名称全局唯一；停用标签不可新绑定，但历史绑定仍可展示；更新和后台审计必须同事务写入

#### `GET /admin-api/tickets/{ticket_no}/attachments/{file_id}/download`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.tickets`
- 作用：下载或预览工单消息附件
- 约束：
  - 必须校验附件属于该工单消息
  - 下载响应不得暴露物理存储路径
  - 图片和 PDF 可直接预览，其它允许类型走下载
  - 文件名进入响应头前必须安全编码
  - 受保护下载响应不得被共享缓存长期保存

## 暂未开放的管理域

密码、token、secret、验证码和敏感配置明文不得出现在任何接口响应中。

## 当前不在契约内的业务域

以下业务域仍不在当前 API 契约内：

- 用户端业务 API（公开站点配置、用户账号自助、用户实名、服务器产品目录、订单 MVP 和工单 MVP 接口除外）
- 支付
- 实例
- 异步任务
