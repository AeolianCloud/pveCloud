# 实例、异步任务与生命周期 API

本文档维护实例交付、管理端实例、异步任务、用户端实例和实例生命周期相关接口。跨接口通用约定见 `docs/server/api/conventions.md`。

## 实例交付和实例管理

实例交付通过 MCP PVE client API 完成。pveCloud 不原样暴露 MCP `/api/pve/*` 路由，不向用户端返回 `node`、`storage`、`disk_source`、`snippets_storage`、`vmid` 或上游 operation ID。

当前只接入 MCP 已提供能力：

- `GET /api/pve/nodes`
- `GET /api/pve/nodes/{node}`
- `GET /api/pve/nodes/{node}/vms`
- `POST /api/pve/nodes/{node}/vms`
- `GET /api/pve/nodes/{node}/vms/{vmid}`
- `DELETE /api/pve/nodes/{node}/vms/{vmid}`
- `POST /api/pve/nodes/{node}/vms/{vmid}/start`
- `POST /api/pve/nodes/{node}/vms/{vmid}/stop`
- `GET /api/pve/storage`
- `GET /api/pve/operations/{id}`

当前不开放重启、重装、重置密码、控制台、快照、备份、迁移、监控、网络防火墙和资源池管理。

### 管理端交付映射

交付映射把产品目录选择映射到 MCP 创建 VM 参数。映射匹配键为 `plan_no`、`region_no`、`template_no` 和 `network_type_no`；`network_type_no` 为空字符串表示不限定网络类型。映射保存 `node`、`storage`、`disk_source`、`disk_format`、`disk_interface`、`snippets_storage`、CloudInit 非敏感参数和 VMID 分配范围。

CloudInit `ci_password` 当前不作为映射配置保存，也不通过接口返回；后续如需初始密码或重置密码，必须先补充一次性凭据展示、加密/脱敏存储和审计契约。

#### `GET /admin-api/instance-provision-mappings`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.instances`
- 作用：分页查询实例交付映射
- 查询参数支持：`page`、`per_page`、`status`、`plan_no`、`region_no`、`template_no`、`network_type_no`

#### `POST /admin-api/instance-provision-mappings`

- 鉴权：管理端 Bearer Token
- 操作权限：`instance:provision` 或 `instance:*`
- 作用：创建交付映射
- 请求字段包含映射匹配键、MCP 创建 VM 参数、VMID 范围和状态
- 约束：`next_vmid` 必须位于 `vmid_start` 和 `vmid_end` 范围内；同一 active 匹配范围只能存在一条有效映射
- 审计：`instance_mapping.create`

#### `PATCH /admin-api/instance-provision-mappings/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`instance:provision` 或 `instance:*`
- 作用：更新交付映射
- 约束：已用于实例交付的映射不得通过普通编辑回退 `next_vmid` 到已分配范围；禁用映射不影响历史实例
- 审计：`instance_mapping.update`

### 管理端 MCP 只读资源

以下接口仅用于后台配置交付映射和排障，返回内容必须经过服务端包装和必要字段筛选，不得向用户端开放。

#### `GET /admin-api/mcp-pve/nodes`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.instances`
- 作用：读取 MCP 节点列表
- 成功数据：数组；每项仅包含 `node`、`name`、`status`

#### `GET /admin-api/mcp-pve/nodes/{node}`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.instances`
- 作用：读取 MCP 节点详情
- 成功数据：仅包含 `node`、`name`、`status`

#### `GET /admin-api/mcp-pve/nodes/{node}/vms`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.instances`
- 作用：读取指定节点 VM 列表，用于排障和 VMID 占用核对
- 成功数据：数组；每项仅包含 `vmid`、`name`、`status`、`cpus`、`mem`、`maxmem`

#### `GET /admin-api/mcp-pve/storage`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.instances`
- 作用：读取 MCP 存储列表
- 成功数据：数组；每项仅包含 `storage`、`name`、`type`、`status`

### 管理端实例接口

#### `POST /admin-api/orders/{order_no}/provision`

- 鉴权：管理端 Bearer Token
- 操作权限：`instance:provision` 或 `instance:*`
- 作用：从 `pending` 订单触发实例交付
- 成功数据包含实例详情和当前交付操作摘要
- 约束：
  - 订单必须存在且状态为 `pending`
  - 必须存在匹配的 active 交付映射
  - 服务端必须在本地事务中分配 `next_vmid`、创建 `instances` 和 `instance_operations` 初始记录，并把订单置为 `provisioning`
  - 外部 MCP 创建 VM 调用不得放在长事务中；本地记录必须能在上游失败后进入可排查状态
  - 重复对同一订单触发交付时，如果已有实例，应返回已有实例或 `409xx` 状态冲突，不得重复创建 VM
- 审计：`instance.provision`

#### `POST /admin-api/orders/{order_no}/confirm-renewal`

- 鉴权：管理端 Bearer Token
- 操作权限：`order:update` 或 `order:*`
- 作用：人工确认续费订单并延长关联实例服务期
- 请求字段：`remark` 可选，最多 500 字
- 约束：
  - 订单必须是 `order_type=renewal`
  - 订单必须处于 `pending`
  - 关联实例必须存在且不为 `released`
  - 服务端必须同事务更新订单 `payment_status=manual_confirmed`、`status=fulfilled`、`paid_at` 和实例 `expires_at`
  - 续期时若实例尚未到期，从原 `expires_at` 起顺延；若已到期，从当前时间起顺延
  - 确认续费必须写入后台操作审计

#### `GET /admin-api/instances`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.instances`
- 作用：分页查询实例列表
- 查询参数支持：`page`、`per_page`、`status`、`instance_no`、`order_no`、`user_keyword`、`date_from`、`date_to`
- 列表项包含实例编号、用户摘要、订单号、实例状态、产品/套餐/地域/系统模板快照、管理端可见的 `node` 和 `vmid`、创建时间和释放时间
- 列表项同时包含服务开始时间、到期时间、到期提醒时间、自动释放计划时间和因到期释放完成时间

#### `GET /admin-api/instances/{instance_no}`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.instances`
- 作用：查看实例详情
- 成功数据包含实例快照、管理端可见的 MCP 资源标识、最近错误、操作记录、订单摘要、服务期和续费记录摘要

#### `POST /admin-api/instances/{instance_no}/start`

- 鉴权：管理端 Bearer Token
- 操作权限：`instance:operate` 或 `instance:*`
- 作用：启动实例
- 约束：只允许对 `stopped` 或可由 MCP 幂等接受的实例发起；服务端必须创建操作记录并调用 MCP start
- 审计：`instance.start`

#### `POST /admin-api/instances/{instance_no}/stop`

- 鉴权：管理端 Bearer Token
- 操作权限：`instance:operate` 或 `instance:*`
- 作用：停止实例
- 约束：只允许对 `running` 或可由 MCP 幂等接受的实例发起；服务端必须创建操作记录并调用 MCP stop
- 审计：`instance.stop`

#### `POST /admin-api/instances/{instance_no}/release`

- 鉴权：管理端 Bearer Token
- 操作权限：`instance:release` 或 `instance:*`
- 作用：释放实例并调用 MCP 删除 VM
- 约束：释放中的实例不可重复释放；释放完成后状态为 `released`，本地实例记录保留
- 审计：`instance.release`

#### `POST /admin-api/instances/{instance_no}/sync`

- 鉴权：管理端 Bearer Token
- 操作权限：`instance:sync` 或 `instance:*`
- 作用：同步实例最近 MCP operation 和 VM 状态
- 约束：若存在未完成 operation，优先查询 MCP operation；operation 成功后再查询 VM 当前状态并映射到本地实例状态
- 约束：operation 未完成、缺少可查询 operation ID 或无法确认成功时，服务端不得仅凭 VM 查询提前推进实例或订单状态
- 审计：`instance.sync`

#### `PATCH /admin-api/instances/{instance_no}/expires-at`

- 鉴权：管理端 Bearer Token
- 操作权限：`instance:renew` 或 `instance:*`
- 作用：后台手动调整实例到期时间
- 请求字段：`expires_at`、`remark`
- 约束：
  - `released` 实例不可调整到期时间
  - `expires_at` 必须是有效时间且不得早于当前时间
  - 调整必须写入后台操作审计

### 管理端异步任务接口

#### `GET /admin-api/async-tasks`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.async-tasks`
- 作用：分页查询异步任务
- 查询参数支持：`page`、`per_page`、`task_type`、`status`、`object_type`、`object_no`、`date_from`、`date_to`
- 成功数据包含任务编号、类型、状态、业务对象、计划执行时间、尝试次数、最大次数、最近错误、锁定 Worker 和完成时间
- 约束：不得返回包含敏感字段的完整 `payload` 或完整上游响应

#### `POST /admin-api/async-tasks/{task_no}/retry`

- 鉴权：管理端 Bearer Token
- 操作权限：`async-task:retry` 或 `async-task:*`
- 作用：人工重试失败任务
- 约束：
  - 仅 `failed` 任务可重试
  - 重试时必须清空锁定字段，状态回到 `pending`，并通过后台操作审计记录本次人工重试
  - 当前不单独维护人工重试计数字段；任务执行次数仍以 `attempts` 表示 Worker 实际领取执行次数
  - 必须写入后台操作审计

### 用户端实例接口

#### `GET /api/instances`

- 鉴权：用户端 Bearer Token
- 作用：分页查询当前用户自己的实例列表
- 查询参数支持：`page`、`per_page`、`status`
- 列表项包含实例编号、订单号、实例状态、产品/套餐/地域/系统模板快照、创建时间和释放时间
- 列表项同时包含服务开始时间、到期时间和到期状态
- 约束：不得返回 `node`、`storage`、`disk_source`、`vmid`、operation ID 或管理端失败详情

#### `GET /api/instances/{instance_no}`

- 鉴权：用户端 Bearer Token
- 作用：查看当前用户自己的实例详情
- 成功数据包含实例快照、订单号、状态和用户可见的最近操作摘要
- 成功数据包含服务期、到期提醒、续费可用状态和最近续费订单摘要
- 约束：只能查看当前登录用户自己的实例；他人实例不得通过错误文案泄露存在性

#### `POST /api/instances/{instance_no}/start`

- 鉴权：用户端 Bearer Token
- 作用：启动当前用户自己的实例
- 约束：只能操作当前登录用户自己的实例；释放中或已释放实例不可操作；重复提交必须依赖本地状态和操作记录幂等保护

#### `POST /api/instances/{instance_no}/stop`

- 鉴权：用户端 Bearer Token
- 作用：停止当前用户自己的实例
- 约束：只能操作当前登录用户自己的实例；释放中或已释放实例不可操作；重复提交必须依赖本地状态和操作记录幂等保护

## 异步任务、通知和实例生命周期

异步任务由独立 Worker 执行，不对用户端开放。API 进程只负责在本地事务提交后投递任务。

当前开放任务：

- `instance_operation_sync`
- `instance_expiry_notice`
- `instance_expiry_release`
- `notification_email_send`
- `notification_sms_placeholder`

实例生命周期规则：

- 实例首次交付完成时写入 `service_started_at` 和 `expires_at`。
- 到期前按 `instance_lifecycle.expire_notice_before_seconds` 投递提醒任务。
- 邮件提醒使用 SMTP 发送；短信提醒本阶段只生成占位任务和通知记录，不接真实短信供应商。
- 到期后按 `instance_lifecycle.expire_release_after_seconds` 计算自动释放计划。
- `instance_lifecycle.auto_release_enabled=false` 时不得自动释放上游 VM。
- 自动释放只能调用当前 MCP 已有 DELETE VM 能力，不得实现 MCP 未提供的重启、重装、重置密码、控制台、快照、备份、迁移、监控或防火墙能力。
