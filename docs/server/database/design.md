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
```

`users` 用于用户端登录账号。当前阶段只支持已有用户登录，不开放用户注册、资料编辑、实名、钱包或业务资料。

`user_sessions` 用于记录用户端登录会话。用户端 access token 的 `jti` 必须对应 `user_sessions.session_id`，服务端受保护用户接口需要校验 token 和会话状态。

### 文件管理

```text
file_attachments
```

`file_attachments` 用于存储管理端上传的图片和附件元信息。
文件物理存储在本地磁盘，数据库只记录元信息和关联关系。
状态使用字符串，不使用数据库 enum。
数据库中的 `storage_path` 只保存相对路径，格式为 `{YYYY}/{MM}/{DD}/{uuid}.{ext}`，不得保存本地存储根目录。
存储文件名强制使用随机 UUID，禁止用户控制存储路径；原始文件名只用于展示，入库前必须去除路径片段和空字节。
上传文件记录与上传审计日志必须在同一事务中写入；事务失败时应清理已落盘的物理文件。
删除文件采用软删除，状态变更与删除审计日志必须在同一事务中写入。

### 文件引用关系

```text
file_attachment_references
```

`file_attachment_references` 用于记录文件被哪些业务对象引用。
当前阶段先为公告、工单和页面配置预留引用能力，具体业务表接入时再补充映射字段或关联写入逻辑。
删除文件前必须先检查引用关系；存在引用时不允许删除。

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
- `system_configs.is_secret=1` 的配置不得通过接口返回明文
- `system_configs` 中 `site.name` 和 `site.logo_url` 是公开站点基础展示配置，分别控制 Web 左上角品牌文字和 Logo 图片 URL
- 站点品牌配置在系统设置中使用中文分组“站点设置”展示
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

当前仓库已经收口到“基础后台阶段”。
数据库契约只保留以下管理域：

- 认证
- RBAC
- 会话
- 系统配置
- 审计日志

以下业务域表不再属于当前数据库契约，后续如需恢复，必须先补新的迁移和文档确认：

- 用户注册、密码找回和账号资料编辑
- 产品目录
- 订单
- 支付与钱包
- 实例
- 异步任务
- 工单

## 关键唯一约束示例

- `admin_roles.code`
- `admin_permissions.code`
- `admin_sessions.session_id`
- `users.username`
- `users.email`
- `user_sessions.session_id`
- `system_configs.config_key`

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

## 一致性原则

- MariaDB 是基础后台事实来源
- Redis 只做缓存、限流、短 TTL 状态和辅助幂等
- 当前阶段不以 PVE、支付、订单、实例、工单或异步任务为现行数据库契约
