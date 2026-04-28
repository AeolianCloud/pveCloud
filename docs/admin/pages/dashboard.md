# Dashboard 页面契约

## 页面定位

`Dashboard` 是管理端登录后的默认受保护业务页。

## 页面范围

- 展示当前已开放的基础后台指标。
- 提供管理端基础运行状态入口感知。
- 不展示订单、支付、实例、工单或其它未开放业务模块指标。

## 路由

- 路径：`/dashboard`
- 页面需要登录态。
- 页面需要权限：`page.dashboard`

## 关联接口

以 `docs/server/api/` 中 Dashboard 相关契约为准。

`GET /admin-api/dashboard` 只负责 Dashboard 数据，不承担登录态恢复职责。

## 验收重点

- 未开放业务模块不出现在 Dashboard 指标中。
- 无权限用户无法进入页面。
- 刷新后登录态恢复不依赖 Dashboard 接口。
