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

- 用户端账号
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
- `system_configs.config_key`

## 一致性原则

- MariaDB 是基础后台事实来源
- Redis 只做缓存、限流、短 TTL 状态和辅助幂等
- 当前基础后台阶段不以 PVE、支付、订单、实例、工单或异步任务为现行数据库契约
