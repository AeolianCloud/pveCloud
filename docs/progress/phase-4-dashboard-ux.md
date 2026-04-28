# Phase 4：Dashboard 与管理端体验收口

- [x] 更新 Dashboard 指标契约，限定只展示当前阶段真实开放指标
- [x] 完成 Dashboard 真实基础指标
- [x] 完成 `403` 无权限页和页面级权限反馈
- [x] 完成 Element Plus 入口、后台壳、路由中心化权限系统和当前开放页面收口
- [x] 在 `admin/` 内部整理最小复用模式，不创建跨前端共享包

## Notes

Dashboard 不展示产品、订单、支付、实例、工单等未开放业务指标。

2026-04-27：管理端前端完成以 Element Plus 为基础的经典后台结构重组，并收口为 `Login`、`Dashboard`、`403` 三页。

2026-04-28：当前开放范围更新为 `Login`、`Dashboard`、`System Settings`、`403`；`System Settings` 承载系统配置和管理员设置。
