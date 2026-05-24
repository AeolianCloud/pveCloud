# Payment Management 页面契约

支付管理用于查看真实支付流水、退款流水和支付详情，并处理退款、渠道状态同步和真实支付后自动交付失败重试。页面只调用 `/admin-api/*`。

## 页面范围

- 支付流水列表
- 退款流水列表
- 支付详情
- 订单和用户摘要
- 支付生效记录摘要
- 发起全额退款
- 主动同步渠道支付状态
- 重试新购订单自动交付失败

本页面不后台创建订单，不创建用户支付，不释放实例，不展示商户密钥、完整回调 payload、完整上游响应、PVE/MCP Bearer Token 或 Worker 锁详情。新购已交付订单退款前的实例释放仍由实例管理页完成。

## 路由与权限

- 路由：`/payments`
- 菜单权限：`page.payments`
- 查看：`payment:view` 或 `payment:*`
- 发起退款：`payment:refund` 或 `payment:*`
- 同步渠道状态：`payment:sync` 或 `payment:*`
- 重试自动交付：`payment:retry-provision` 或 `payment:*`

## 页面结构

支付管理包含支付流水、退款流水、详情抽屉和危险操作确认，按复杂管理页结构实现：

```text
admin/src/views/payments/
  index.vue
  types.ts
  components/
```

## 行为约束

- 支付列表支持按供应商、方式、状态、支付编号、订单编号、用户关键字和创建时间范围筛选。
- 退款列表支持按供应商、状态、退款编号、支付编号、订单编号和创建时间范围筛选。
- 详情展示订单摘要、用户摘要、支付交易、退款摘要和支付生效记录摘要。
- 发起退款必须二次确认，并展示“一期仅全额退款”的明确结果；前端只做体验提示，后端仍最终校验订单、实例和支付状态。
- 新购已交付订单若实例未释放，不展示可退款主按钮；后端仍必须拒绝未释放实例退款。
- 续费退款采用渠道成功后本地回滚；页面在 `pending` 退款期间展示处理中状态，不提前显示服务期已扣回。
- `status=error` 且 `payment_status=paid` 的新购订单可展示自动交付重试入口；其它订单不得展示该入口。
- 低权限管理员不能看到或触发退款、同步和重试按钮；后端权限仍是最终裁决。

## 关联接口

接口字段和响应结构以 `docs/server/api/endpoints.md` 为准。

- `GET /admin-api/payments`
- `GET /admin-api/payments/{payment_no}`
- `POST /admin-api/payments/{payment_no}/sync`
- `POST /admin-api/payments/{payment_no}/refunds`
- `GET /admin-api/refunds`
- `POST /admin-api/payments/{payment_no}/retry-provision`

## 验收重点

- 侧栏菜单来自后端 `menus`，本地路由只作为页面组件白名单。
- 无权限访问 `/payments` 时展示管理端 403 反馈。
- 敏感配置、完整回调 payload 和完整上游响应不在页面展示、控制台日志或错误提示中出现。
- 退款、同步和重试必须有确认反馈和失败提示。
- 支付管理不调用 `/api/*`，不复用用户端运行时代码。
