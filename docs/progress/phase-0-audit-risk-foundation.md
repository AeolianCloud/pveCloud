# Phase 0：安全日志基座

- [x] 更新 `docs/admin/architecture.md`，明确审计日志和高危操作日志页面范围。
- [x] 更新 `docs/server/api/endpoints.md`，补充 `/admin-api/audit-logs` 和 `/admin-api/risk-logs` 契约。
- [x] 更新 `docs/server/database/design.md` 和迁移，补入 `admin_risk_logs`、双写规则和 `audit:sensitive_view`。
- [x] 补入后端查询模型、DTO、服务、Handler 和路由。
- [x] 补入管理端审计/高危日志页面、请求类型、路由、菜单和图标。

Notes:

当前阶段成果仍在工作区未提交改动中。进入后续实现或提交前需要重新运行后端测试和管理端构建。
