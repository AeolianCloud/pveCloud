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
- 工单指派和转派
- 客服协作者
- 内部备注
- 内部 SLA 时限和逾期状态
- 工单标签和标签字典
- 优先级升级

本页面不支持后台创建用户工单，不支持支付、实例开通、PVE 节点或自动交付操作；页面可以展示工单关联的业务实例编号并跳转到实例管理排障，实际实例操作仍必须在实例管理页面完成。

## 路由与权限

- 路由：`/tickets`
- 菜单权限：`page.tickets`
- 回复工单：`ticket:reply` 或 `ticket:*`
- 关闭工单：`ticket:close` 或 `ticket:*`
- 指派和转派：`ticket:assign` 或 `ticket:*`
- 协作者维护：`ticket:collaborate` 或 `ticket:*`
- 内部备注：`ticket:note` 或 `ticket:*`
- 优先级升级：`ticket:priority` 或 `ticket:*`
- 工单标签绑定：`ticket:tag` 或 `ticket:*`
- 标签字典管理：`ticket:tag-manage` 或 `ticket:*`

## 页面结构

工单管理是复杂管理页，应使用页面容器结构：

```text
admin/src/views/tickets/
  index.vue
  types.ts
  components/
```

## 行为规则

- 列表支持按工单状态、分类、优先级、工单编号、订单编号、实例编号、用户关键字和创建时间范围筛选。
- 列表增加处理人、标签、SLA 逾期状态筛选和展示。
- 详情展示工单基础信息、用户摘要、可选订单号、可选实例编号、状态、分类、优先级、处理人、协作者、标签、内部 SLA 时限、消息时间线、附件、内部备注和操作历史。
- 关联实例只展示业务实例编号；不得在工单页面展示 MCP/PVE 节点、存储、VMID、operation ID 或上游原始错误。
- 跳转实例管理入口只有具备 `page.instances` 时展示；是否允许开机、关机、释放或同步继续由实例管理页面和后端实例权限裁决。
- 管理员只能回复未关闭工单。
- 管理员只能关闭未关闭工单。
- 关闭后的工单不可继续回复。
- 关闭后的工单不可指派、转派、维护协作者或升级优先级，但可以追加内部备注。
- 指派空处理人表示首次指派；已有处理人变更表示转派。
- 可指派对象只包含 `active` 且具备 `page.tickets`、`ticket:reply` 或 `ticket:*` 的管理员。
- 内部备注只追加，不编辑、不删除，只在管理端展示。
- 内部 SLA 只用于后台处理时限和逾期提示，不作为用户端承诺展示。
- 优先级升级只能从低紧急度升到更高紧急度，必须填写原因；不支持降级。
- 标签分公开和内部。公开标签可以返回给用户端，内部标签只在管理端展示。
- 管理端回复和关闭操作必须以后端审计为准，前端只展示操作入口。
- 附件下载或预览必须通过工单附件接口，不得复用公网匿名地址。

## 关联接口

- `GET /admin-api/tickets`
- `GET /admin-api/tickets/{ticket_no}`
- `POST /admin-api/tickets/{ticket_no}/messages`
- `POST /admin-api/tickets/{ticket_no}/close`
- `GET /admin-api/tickets/{ticket_no}/attachments/{file_id}/download`
- `GET /admin-api/tickets/assignee-candidates`
- `POST /admin-api/tickets/{ticket_no}/assign`
- `POST /admin-api/tickets/{ticket_no}/collaborators`
- `DELETE /admin-api/tickets/{ticket_no}/collaborators/{admin_id}`
- `POST /admin-api/tickets/{ticket_no}/internal-notes`
- `POST /admin-api/tickets/{ticket_no}/priority`
- `PUT /admin-api/tickets/{ticket_no}/tags`
- `GET /admin-api/ticket-tags`
- `POST /admin-api/ticket-tags`
- `PATCH /admin-api/ticket-tags/{id}`

具体字段、响应和错误码以 `docs/server/api/` 为准。

## 验收重点

- 无权限访问 `/tickets` 时展示管理端 403 反馈。
- 工单列表分页、筛选和详情展示正常。
- 管理员无回复权限时不展示回复入口。
- 管理员无关闭权限时不展示关闭入口。
- 管理员无对应操作权限时不展示指派、协作、内部备注、优先级升级、标签绑定或标签字典管理入口。
- 关闭后的工单不展示回复入口。
- 用户端不能看到处理人、协作者、内部备注、内部 SLA 或内部标签。
- 附件下载和预览不暴露本地磁盘路径。
- 页面不出现支付确认、实例开通、实例电源操作、PVE 节点、库存扣减或自动交付操作。
