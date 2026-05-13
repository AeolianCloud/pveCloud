# Ticket Management 页面契约

## 页面定位

`Ticket Management` 是管理端工单管理页面，用于后台查看和处理用户端提交的工单。

本页面只调用 `/admin-api/*`。

## 页面范围

- 工单列表
- 工单详情
- 工单消息时间线
- 管理员回复
- 关闭工单
- 工单附件下载或预览

本页面不支持后台创建用户工单，不支持工单指派、转派、SLA、内部备注、支付、实例开通、PVE 节点或自动交付操作。

## 路由与权限

- 路由：`/tickets`
- 菜单权限：`page.tickets`
- 回复工单：`ticket:reply` 或 `ticket:*`
- 关闭工单：`ticket:close` 或 `ticket:*`

## 页面结构

工单管理是中等复杂度管理页，第一版可以使用单文件入口：

```text
admin/src/views/tickets/index.vue
```

如果后续加入指派、多 tab、内部备注或多弹窗处理流，再拆分为页面容器结构。

## 行为规则

- 列表支持按工单状态、分类、优先级、工单编号、订单编号、用户关键字和创建时间范围筛选。
- 详情展示工单基础信息、用户摘要、可选订单号、状态、分类、优先级、消息时间线和附件。
- 管理员只能回复未关闭工单。
- 管理员只能关闭未关闭工单。
- 关闭后的工单不可继续回复。
- 管理端回复和关闭操作必须以后端审计为准，前端只展示操作入口。
- 附件下载或预览必须通过工单附件接口，不得复用公网匿名地址。

## 关联接口

- `GET /admin-api/tickets`
- `GET /admin-api/tickets/{ticket_no}`
- `POST /admin-api/tickets/{ticket_no}/messages`
- `POST /admin-api/tickets/{ticket_no}/close`
- `GET /admin-api/tickets/{ticket_no}/attachments/{file_id}/download`

具体字段、响应和错误码以 `docs/server/api/` 为准。

## 验收重点

- 无权限访问 `/tickets` 时展示管理端 403 反馈。
- 工单列表分页、筛选和详情展示正常。
- 管理员无回复权限时不展示回复入口。
- 管理员无关闭权限时不展示关闭入口。
- 关闭后的工单不展示回复入口。
- 附件下载和预览不暴露本地磁盘路径。
- 页面不出现支付确认、实例开通、PVE 节点、库存扣减、SLA 或自动交付操作。
