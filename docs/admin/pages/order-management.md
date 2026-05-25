# Order Management 页面契约

订单管理用于查看和处理用户端订单，页面只调用 `/admin-api/*`。

## 页面范围

- 订单列表
- 订单详情
- 用户摘要
- 产品、套餐、价格、销售地域和系统模板快照
- 后台备注
- 取消订单
- 关闭订单
- 触发实例交付入口
- 查看续费订单和支付状态
- 查看支付和退款摘要，并跳转到支付管理详情
- 查看发票申请摘要，并跳转到发票运营详情
- 人工确认续费订单

本页面不支持后台创建订单，不发起退款，不直接同步支付渠道，不处理发票状态流转，不包含库存扣减或通用 PVE 管理能力。实例交付通过实例管理域能力触发，只能基于已有用户端订单；续费订单由用户端创建，后台可人工确认并延长实例服务期。真实支付和退款处理入口在支付管理页，发票运营入口在发票运营页。

## 路由与权限

- 路由：`/orders`
- 菜单权限：`page.orders`
- 查看：`order:view` 或 `order:*`
- 更新后台备注和关闭订单：`order:update` 或 `order:*`
- 取消订单：`order:cancel` 或 `order:*`
- 触发实例交付：`instance:provision` 或 `instance:*`
- 人工确认续费：`order:update` 或 `order:*`

## 页面结构

订单管理是中等复杂度管理页，第一版可以使用单文件入口：

```text
admin/src/views/orders/index.vue
```

如果后续加入多 tab、多弹窗或复杂处理流，再拆分为页面容器结构。

## 行为约束

- 列表支持按订单状态、订单编号、用户关键字和创建时间范围筛选。
- 详情展示订单基础信息、用户摘要、状态、金额、用户备注、后台备注和下单快照。
- 后台备注只在管理端展示，不通过用户端订单详情返回。
- 管理端只能取消 `pending` 订单；关闭订单按后端状态机裁决。
- 非 `pending` 订单执行取消或关闭时，后端返回 `409xx` 状态冲突。
- 管理端可对 `order_type=purchase` 且 `pending` 的订单触发实例交付；交付后订单进入 `provisioning`，实例同步成功后进入 `fulfilled`。
- 管理端可对 `order_type=renewal` 且 `pending` 的续费订单人工确认；确认后订单 `payment_status=manual_confirmed`、状态进入 `fulfilled`，关联实例到期时间顺延。
- 取消、关闭和后台备注更新必须写入普通后台操作审计。
- 页面可展示最近支付/退款摘要、发票申请摘要和对应管理页跳转入口，但不得展示商户密钥、完整回调 payload、完整上游响应、发票 PDF 物理路径、PVE 节点、库存扣减或用户侧自动交付承诺。

## 关联接口

接口字段和响应结构以 `docs/server/api/endpoints.md` 为准。

- `GET /admin-api/orders`
- `GET /admin-api/orders/{order_no}`
- `PATCH /admin-api/orders/{order_no}/admin-note`
- `POST /admin-api/orders/{order_no}/cancel`
- `POST /admin-api/orders/{order_no}/close`
- `POST /admin-api/orders/{order_no}/provision`
- `POST /admin-api/orders/{order_no}/confirm-renewal`

## 验收重点

- 侧栏菜单来自后端 `menus`，本地路由只作为页面组件白名单。
- 无权限访问 `/orders` 时展示管理端 403 反馈。
- 低权限管理员看不到或无法触发无权限操作按钮。
- 订单列表分页、筛选和详情展示正常。
- 后台备注不会出现在用户端订单接口。
- 管理端不能创建订单。
