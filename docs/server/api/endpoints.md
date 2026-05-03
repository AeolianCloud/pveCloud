# API 接口总览

本文档维护当前已确认的接口清单与主要契约口径。跨接口通用约定见 `docs/server/api/conventions.md`。

## 实现边界提示

接口契约按访问边界区分：

- `/admin-api/*`：对应管理端后端实现边界 `server/internal/admin/*`

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

### `GET /admin-api/admin-users/{id}`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.admin-users`
- 作用：查看管理员详情

### `PATCH /admin-api/admin-users/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`admin-user:update` 或 `admin-user:*`
- 作用：更新管理员信息、状态和角色

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

### `GET /admin-api/admin-roles/{id}`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.system-settings.admin-roles`
- 作用：查看角色详情

### `PATCH /admin-api/admin-roles/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`admin-role:update` 或 `admin-role:*`
- 作用：更新角色信息、状态和权限

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

### `PATCH /admin-api/system-configs/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`system-config:update` 或 `system-config:*`
- 作用：更新系统配置

## 用户端公开配置域

### `GET /api/site-config`

- 鉴权：公开接口，无需登录
- 作用：读取 Web 端公开站点基础展示配置
- 数据来源：`system_configs` 中的非敏感配置
- 返回字段：
  - `site_name`：站点显示名称，来自 `site.name`，为空时服务端返回默认值 `pveCloud`
  - `logo_url`：站点 Logo 图片 URL，来自 `site.logo_url`，为空时返回空字符串
- 约束：不得返回 `is_secret=1` 的配置项，不得返回任意配置键列表

## 用户端认证域

### `POST /api/auth/login`

- 鉴权：公开接口，无需登录
- 作用：用户登录，创建用户端会话并签发用户端 access token
- 请求字段：
  - `account`：用户名或邮箱
  - `password`：密码
- 成功数据包含：
  - `access_token`
  - `token_type`：固定 `Bearer`
  - `expires_in`：有效期秒数
  - `user`：用户摘要，包含 `id`、`username`、`email`、`display_name`、`status`
  - `session`：当前会话摘要，包含 `session_id`、`issued_at`、`expires_at`
- 约束：仅 `status=active` 的用户允许登录；账号不存在或密码错误时返回未登录错误，用户被禁用时返回明确禁用错误

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
- 约束：当前会话已过期或已吊销时返回未登录错误；用户被禁用时返回明确禁用错误

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

产品目录只维护服务器产品展示和可售约束，不创建订单、不发起支付、不创建实例、不绑定 PVE 节点。

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

### `GET /api/server-catalog`

- 鉴权：公开接口，不要求用户登录
- 作用：返回 Web 可展示服务器产品目录聚合数据
- 返回范围：已上架且可见的服务器产品、套餐、周期价格、销售地域和服务器系统模板
- 展示约束：套餐需要至少有一个 active 周期价格、一个 active 且 visible 的销售地域、一个 active 且 visible 的服务器系统模板才进入公开目录
- 禁止返回：订单、支付、实例、库存扣减、PVE 节点、PVE 模板 ID 或资源池信息

## 暂未开放的管理域

密码、token、secret、验证码和敏感配置明文不得出现在任何接口响应中。

## 当前不在契约内的业务域

以下业务域仍不在当前 API 契约内：

- 用户端业务 API（公开站点配置、用户登录会话和服务器产品目录接口除外）
- 用户注册、密码找回和账号资料编辑
- 订单
- 支付
- 实例
- 工单
- 异步任务
