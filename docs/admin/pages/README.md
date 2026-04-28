# Admin 页面文档索引

本目录按页面记录管理端前端页面契约。

页面文档只描述页面范围、行为、权限入口和关联接口索引，不复制完整 API 字段。

## 当前开放页面

- `Login`：`docs/admin/pages/login.md`
- `Dashboard`：`docs/admin/pages/dashboard.md`
- `System Settings`：`docs/admin/pages/system-settings.md`
- `403`：`docs/admin/pages/403.md`

## 当前未开放页面

以下后端能力或接口可以存在，但当前管理端前端不提供独立页面、菜单入口或受保护路由：

- 登录会话管理
- 审计日志
- 高危操作日志
- 产品套餐
- 订单
- 支付
- 实例
- 工单
- 用户端业务流

重新开放上述页面前，必须先更新：

- `docs/admin/architecture.md`
- 本目录对应页面文档
- `docs/admin/routing-permissions.md`
- 相关 API 文档
- 必要时更新 `docs/plan/` 与 `docs/progress/`
