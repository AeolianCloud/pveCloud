# Dashboard 页面契约

## 页面定位

`Dashboard` 是管理端登录后的默认受保护业务页。

## 页面范围

- 展示当前已开放的基础后台指标，包括启用管理员、启用角色、活跃会话和今日操作日志。
- 展示当前已开放业务域的运营待办和异常摘要，包括订单、支付、退款、实例、异步任务、工单和发票。
- 提供管理端基础运行状态、当前会话、可访问菜单和业务处理入口感知。
- Dashboard 只做只读汇总和入口跳转，不直接执行订单处理、退款、实例操作、任务重试、工单回复或发票流转。

## 业务指标

Dashboard 首批展示以下业务指标：

- 待处理订单：`orders.status = pending`
- 交付异常订单：`orders.status = error`
- 异常实例：`instances.status = error`
- 失败异步任务：`async_tasks.status = failed`
- 待后台处理工单：`tickets.status = waiting_admin`
- 待处理发票：`invoice_applications.status IN (pending, processing)`
- 支付异常：`payment_transactions.status = failed` 与 `refund_transactions.status IN (pending, failed)` 的合计

每个业务指标可配置目标页面路径、目标页面权限和严重度。当前管理员具备目标页面权限时，页面展示可点击跳转；不具备目标页面权限时只展示数量，不展示跳转操作。

业务指标不得展示商户密钥、完整回调 payload、完整上游响应、PVE/MCP Bearer Token、PVE 节点、VMID、Worker payload/result 或其它敏感详情。

## 路由

- 路径：`/dashboard`
- 页面需要登录态。
- 页面需要权限：`page.dashboard`

## 关联接口

以 `docs/server/api/` 中 Dashboard 相关契约为准。

`GET /admin-api/dashboard` 只负责 Dashboard 数据，不承担登录态恢复职责。

## 验收重点

- 基础后台指标和业务指标均来自 `GET /admin-api/dashboard`，刷新后能同步最新汇总。
- 业务指标只展示本契约列出的当前已开放业务域，不展示未开放的专票、红冲、部分退款、提现、自动对账、实例控制台、重装、快照或备份等能力。
- 无目标页面权限时，对应业务指标不能提供跳转入口；后端目标业务接口仍按各自权限最终裁决。
- Dashboard 页面不能直接执行业务写操作，也不能绕过对应业务页面的权限、状态和二次确认。
- 无权限用户无法进入页面。
- 刷新后登录态恢复不依赖 Dashboard 接口。
