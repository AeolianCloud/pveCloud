# 用户运营、日志与文件 API

本文档维护管理端用户运营、实名审核、审计日志、日志中心和文件管理相关接口。跨接口通用约定见 `docs/server/api/conventions.md`。

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
- 菜单权限：操作审计页面使用 `page.logs.admin-operations`；登录安全页面使用 `page.logs.admin-security`
- 敏感详情权限：`audit-log:sensitive-view` 或 `audit-log:*`
- 作用：分页查询普通后台审计日志，可用于日志管理中心的操作审计页面和登录安全页面
- 查询参数支持：`page`、`per_page`、`log_type`、`admin_id`、`action`、`object_type`、`object_id`、`date_from`、`date_to`
- `log_type` 可选值：
  - `admin_operation`：查询非认证类后台操作审计记录，排除 `object_type=admin_auth`
  - `admin_security`：查询管理端登录安全记录，固定 `object_type=admin_auth`
- 成功数据包含：
  - `list`
  - `total`
  - `page`
  - `per_page`
  - `last_page`

列表项包含操作者摘要、会话 ID、请求 ID、请求方法、请求路径、操作动作、对象类型、对象 ID、IP、备注和创建时间。
未具备敏感详情权限时，`before_data`、`after_data` 和 `user_agent` 不返回。

登录安全页面不新增独立接口或表，使用本接口并传 `log_type=admin_security` 查询认证相关日志；如需按动作类型筛选，继续使用单个 `action` 查询参数。Phase 1 中 `admin-security-log:view` 作为登录安全页面查询语义权限，后端路由仍复用本接口。

## 日志管理中心

### `GET /admin-api/logs/user-security`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.logs.user-security`
- 操作权限：`user-security-log:view` 或 `user-security-log:*`
- 作用：分页查询用户安全日志

### `GET /admin-api/logs/user-business`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.logs.user-business`
- 操作权限：`user-business-log:view` 或 `user-business-log:*`
- 作用：分页查询用户业务日志

### `GET /admin-api/logs/frontend-errors`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.logs.frontend-errors`
- 操作权限：`frontend-error-log:view` 或 `frontend-error-log:*`
- 作用：分页查询前端错误日志

### `GET /admin-api/logs/backend-runtime`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.logs.backend-runtime`
- 操作权限：`backend-runtime-log:view` 或 `backend-runtime-log:*`
- 作用：分页查询后端运行日志
- 约束：当前实现优先查询结构化运行日志表；stdout 纯文本采集边界由部署和外部日志平台负责

### `POST /admin-api/client-logs/errors`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.logs.frontend-errors`
- 作用：admin 前端错误上报
- 约束：字段必须截断、脱敏和限流，不得上传 token、密码、验证码、请求体原文或供应商完整响应

### `POST /api/client-logs/errors`

- 鉴权：公开或用户端登录态均可
- 作用：web 前端错误上报
- 约束：字段必须截断、脱敏和限流，不得上传 token、密码、验证码、请求体原文或供应商完整响应

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
