# Phase 3：登录会话和系统设置

- [x] 更新登录会话和系统设置接口契约的细节字段、错误码和高危动作说明。
- [x] 实现登录会话列表、筛选和吊销他人会话后端能力。
- [x] 实现管理端登录会话页面。
- [x] 实现系统配置分组查询、普通配置更新和敏感配置更新后端能力。
- [x] 实现管理端系统设置页面。
- [x] 补齐会话失效、当前会话保护、敏感配置不回显和审计日志测试。

Notes:

当前会话退出仍使用 `POST /admin-api/auth/logout`，不通过会话列表吊销。

2026-04-26：已接入 `/admin-api/admin-sessions`、`/admin-api/admin-sessions/{id}/revoke`、`/admin-api/system-configs` 和 `/admin-api/system-configs/{id}`。吊销他人会话和修改敏感配置写入高危日志；敏感配置不返回明文。
