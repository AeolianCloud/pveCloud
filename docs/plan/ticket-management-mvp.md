# 工单系统 MVP 计划

> 历史说明：本文记录工单 MVP 的初始开放范围。当前工单已进入增强阶段，管理端内部指派、转派、协作者、内部备注、内部 SLA、标签和优先级升级的现行契约以 `docs/admin/pages/ticket-management.md`、`docs/server/api/tickets.md`、`docs/server/database/design.md` 和 `server/migrations/033_ticket_enhancement.sql` 为准。

## 目标

开放工单系统 MVP，让用户端可以围绕账号、产品、订单和通用售后问题提交工单，与后台客服或运营人员进行文字和附件沟通。

本阶段只开放工单沟通能力，不开放支付、实例交付、PVE 操作、客服指派、SLA 承诺、邮件通知或站内信通知。

## 范围

- 用户端创建通用工单，或可选关联当前登录用户自己的订单。
- 用户端查看自己的工单列表、详情、消息记录和附件。
- 用户端可以回复自己的未关闭工单，并关闭自己的未关闭工单。
- 管理端查看所有工单列表、详情、消息记录和附件。
- 管理端可以回复未关闭工单，并关闭未关闭工单。
- 用户端和管理端回复都支持附件，附件复用现有文件上传安全规则和本地存储配置。

## 非范围

- 支付、退款、钱包、发票或余额。
- 实例开通、资源交付、PVE 节点、资源池或库存扣减。
- 工单指派、转派、内部协作、客服排班或 SLA。
- 邮件、短信、站内信、WebSocket 或实时推送通知。
- 富文本、图片粘贴、压缩包上传或公网匿名附件下载。

## 状态模型

工单状态只包含：

- `waiting_admin`：等待后台处理。创建工单或用户回复后进入该状态。
- `waiting_user`：等待用户反馈。管理员回复后进入该状态。
- `closed`：工单已关闭。用户或管理员均可关闭未关闭工单。

关闭后的工单不可继续回复。需要继续沟通时由用户创建新工单。

## 分类与优先级

工单分类固定为：

- `account`：账号问题
- `order`：订单问题
- `product`：产品咨询
- `technical`：技术支持
- `billing`：账务问题
- `other`：其它问题

工单优先级固定为：

- `low`
- `normal`
- `high`
- `urgent`

默认优先级为 `normal`。

## 附件规则

- 单条工单消息最多包含 5 个附件。
- 单文件大小、允许 MIME 类型、扩展名、Magic Bytes 和危险扩展名校验沿用 `server/config.example.yaml` 中 `storage` 配置与现有文件安全规则。
- 附件记录复用 `file_attachments`，并通过 `file_attachment_references` 记录 `ref_type=ticket_message` 引用，避免仍被工单引用的文件被删除。
- 用户端附件只能由工单所属用户访问。
- 管理端附件只能由具备工单页面权限的管理员访问。
- 不新增公开通用文件下载接口。

## 数据库契约

新增：

- `tickets`
- `ticket_messages`
- `ticket_message_attachments`

关键规则：

- 主键使用 `BIGINT UNSIGNED AUTO_INCREMENT`。
- 对外展示使用 `ticket_no`，不得向用户端暴露自增 ID。
- 状态、分类和优先级使用字符串，不使用数据库 enum。
- 工单可选关联 `orders.id`，同时保存 `order_no` 快照用于筛选和展示。
- 用户端关联订单时，后端必须校验订单属于当前登录用户。
- 工单回复和附件引用写入必须在同一事务中完成。
- 管理端回复和关闭工单必须写入后台审计。

## API 契约

用户端接口只挂载 `/api/*`：

- `GET /api/tickets`：查询当前用户自己的工单列表。
- `POST /api/tickets`：创建工单。
- `GET /api/tickets/{ticket_no}`：查看当前用户自己的工单详情。
- `POST /api/tickets/{ticket_no}/messages`：回复当前用户自己的未关闭工单。
- `POST /api/tickets/{ticket_no}/close`：关闭当前用户自己的未关闭工单。
- `GET /api/tickets/{ticket_no}/attachments/{file_id}/download`：下载或预览当前用户自己工单中的附件。

管理端接口只挂载 `/admin-api/*`：

- `GET /admin-api/tickets`：查询工单列表。
- `GET /admin-api/tickets/{ticket_no}`：查看工单详情。
- `POST /admin-api/tickets/{ticket_no}/messages`：管理员回复未关闭工单。
- `POST /admin-api/tickets/{ticket_no}/close`：管理员关闭未关闭工单。
- `GET /admin-api/tickets/{ticket_no}/attachments/{file_id}/download`：下载或预览工单附件。

## 管理端权限

新增管理端页面和权限：

- 菜单权限：`page.tickets`
- 回复权限：`ticket:reply`
- 关闭权限：`ticket:close`
- 全权限：`ticket:*`

历史 MVP 阶段未新增 `ticket:assign`；当前增强阶段已在 owner docs 和 `033_ticket_enhancement.sql` 中开放指派、协作、内部备注、标签和优先级升级权限。

## 页面契约

用户端：

- 用户中心展示工单入口。
- 新增工单列表页、工单创建页和工单详情页。
- 用户端只展示当前登录用户自己的工单。
- 工单详情按消息时间线展示用户和管理员回复及附件。

管理端：

- 新增工单管理页面。
- 页面支持列表筛选、详情查看、回复、关闭和附件下载/预览。
- 页面不支持指派、转派、SLA、内部备注或创建用户工单。

## 实施阶段

1. 更新 owner docs 和迁移契约，并暂停确认。
2. 实现后端工单领域、仓储、用户端用例、管理端用例和 HTTP 路由。
3. 实现用户端工单列表、创建、详情和用户中心入口。
4. 实现管理端工单管理页面和权限接入。
5. 执行验证并回扫文档与实现漂移。

## 验证

- 后端：工单相关 handler/service/repository 测试，环境允许时执行 `go test ./...`。
- 管理端：`bun run build`
- 用户端：`bun run build`

重点回归：

- 未登录访问用户端工单接口。
- 用户关联他人订单创建工单。
- 跨用户查看、回复、关闭或下载附件。
- 低权限管理员访问工单列表、回复、关闭和附件下载。
- 关闭后再次回复。
- 单条消息超过附件数量限制。
- 非法附件类型、伪造 MIME、超大文件和路径穿越。
- 管理端回复和关闭审计写入。
