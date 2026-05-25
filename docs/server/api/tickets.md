# 工单 API

本文档维护用户端工单和管理端工单运营接口。跨接口通用约定见 `docs/server/api/conventions.md`。

## 工单

工单提供用户与后台之间的文字和附件沟通能力，并支持管理端内部指派、转派、协作者、内部备注、内部 SLA、标签和优先级升级。工单可选关联当前用户自己的订单和实例，用于售后沟通和实例排障定位；不承诺支付、实例交付、PVE 操作、用户侧 SLA、实时推送、邮件或站内信通知。

### 用户端工单接口

#### `GET /api/tickets`

- 鉴权：用户端 Bearer Token
- 作用：分页查询当前用户自己的工单列表
- 查询参数支持：`page`、`per_page`、`status`、`category`、`priority`、`order_no`、`instance_no`
- 列表项包含工单编号、标题、分类、优先级、状态、公开标签、关联订单号、关联实例编号、最近消息时间和创建时间
- 约束：只能返回当前登录用户自己的工单

#### `POST /api/tickets`

- 鉴权：用户端 Bearer Token
- 请求格式：`multipart/form-data`
- 作用：创建工单
- 请求字段：`title`、`category`、`priority`、`content`、`order_no`、`instance_no`、`attachments`
- `category` 允许：`account`、`order`、`product`、`technical`、`billing`、`other`
- `priority` 允许：`low`、`normal`、`high`、`urgent`；为空时使用 `normal`
- `order_no` 可选；填写时必须属于当前登录用户
- `instance_no` 可选；填写时必须属于当前登录用户；同时填写 `order_no` 和 `instance_no` 时，该实例必须来源于同一订单
- `attachments` 可选，单条消息最多 5 个附件
- 成功数据包含工单详情
- 约束：
  - 创建后工单状态为 `waiting_admin`
  - 创建工单必须同时创建第一条用户消息
  - 只填写 `instance_no` 时，服务端使用实例来源订单回填工单关联订单字段
  - `order_no` 或 `instance_no` 不存在、不属于当前用户时按资源不存在处理；订单与实例不匹配时按参数错误处理
  - 附件必须通过文件大小、扩展名、声明 MIME、Magic Bytes 和危险扩展名校验
  - 工单、首条消息、附件记录和文件引用必须同事务写入
  - 不得信任前端传入的用户 ID、订单归属、实例归属、附件归属或状态

#### `GET /api/tickets/{ticket_no}`

- 鉴权：用户端 Bearer Token
- 作用：查看当前用户自己的工单详情
- 成功数据包含工单基础信息、关联订单号、关联实例编号、用户可见状态、公开标签、消息时间线和附件摘要
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
- 查询参数支持：`page`、`per_page`、`status`、`category`、`priority`、`ticket_no`、`order_no`、`instance_no`、`user_keyword`、`date_from`、`date_to`
- 扩展查询参数支持：`assignee_admin_id`、`tag_id`、`sla_status`
- `sla_status` 允许：`normal`、`first_response_overdue`、`resolution_overdue`
- 列表项包含工单编号、用户摘要、标题、分类、优先级、状态、处理人摘要、标签、内部 SLA 状态、关联订单号、关联实例编号、最近消息时间和创建时间

#### `GET /admin-api/tickets/{ticket_no}`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.tickets`
- 作用：查看工单详情
- 成功数据包含工单基础信息、用户摘要、可选订单号、可选实例编号、处理人摘要、协作者摘要、标签、内部 SLA 状态、消息时间线、附件摘要、内部备注和操作历史
- 约束：工单详情只返回业务实例编号，不返回 PVE 节点、存储、VMID、operation ID 或管理端实例失败详情；实例操作仍必须回到实例管理接口执行

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
