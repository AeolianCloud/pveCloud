# Order Detail 页面契约

## 页面定位

`Order Detail` 是用户端订单详情页，对应路由 `/user/orders/:orderNo`。

## 行为范围

- 页面是受保护路由，未登录访问时跳转 `/login` 并携带站内 `redirect`。
- 页面只展示当前登录用户自己的订单详情。
- 用户可以取消自己的 `pending` 订单。
- 页面展示新购订单和续费订单。

## 展示内容

- 订单编号
- 订单类型：`purchase`、`renewal`
- 支付占位状态：`unpaid`、`paid`、`manual_confirmed`
- 订单状态：`pending`、`provisioning`、`fulfilled`、`cancelled`、`closed`
- 订单金额和币种
- 用户备注
- 产品快照
- 套餐规格快照
- 计费周期和价格快照
- 销售地域快照
- 系统模板快照
- 网络类型快照
- 创建时间、取消时间、关闭时间
- 关联实例编号和续费确认时间

页面不得展示后台备注、真实支付入口、PVE 节点、资源池、库存扣减或上游自动开通进度。若订单已交付或关联实例，可展示跳转到实例详情的业务入口。

## 关联接口

- `GET /api/orders/{order_no}` - 当前用户订单详情
- `POST /api/orders/{order_no}/cancel` - 当前用户取消自己的订单

具体字段、响应和错误码以 `docs/server/api/` 为准。

用户端只调用 `/api/*`，不得调用 `/admin-api/*`。

## 验收重点

- 跨用户订单不可访问。
- 订单快照展示以后端订单详情为准，不从当前产品目录重新拼接。
- 只能取消 `pending` 订单。
- 页面不出现后台备注、真实支付入口或自动交付承诺。
