# 基础后台阶段进度总览

## 阶段目标

完成一个可登录、可授权、可追踪、可配置的基础后台底座。

## 相关文档

- 缺口分析：`docs/analysis/basic-admin-gap.md`
- 实施计划：`docs/plan/basic-admin-foundation.md`
- 管理端架构：`docs/admin/architecture.md`
- API 契约：`docs/server/api/`
- 数据库设计：`docs/server/database/design.md`

## 分阶段进度

- [x] Phase 0：安全日志基座
- [x] Phase 1：审计写入与脱敏收口
- [x] Phase 2：管理员与角色权限
- [x] Phase 3：登录会话与系统设置
- [x] Phase 4：Dashboard 与管理端体验收口
- [x] Phase 5：验收与发布准备

## 当前状态

基础后台阶段的后端能力已覆盖认证、会话、RBAC、Dashboard、管理员、角色、系统配置、审计写入与高危日志写入等管理域。

当前管理端前端实现已经按 Element Plus 后台结构收口，并明确当前开放范围为：

- `Login`
- `Dashboard`
- `System Settings`
- `403`

`System Settings` 当前承载系统配置和管理员设置能力；管理员设置内包含管理员账号、管理组权限和管理员会话管理。

审计日志查询和高危日志查询的前端页面、菜单、受保护路由、后端开放接口和权限码当前仍未开放。
相关数据结构和内部写入能力仍保留。

`docs/web/architecture.md` 当前用于用户端规划与契约准备；在仓库真正出现 `web/` 实现前，不应把它视为现有实现说明。

## 下一步维护原则

- 再次修改当前后台范围时，优先更新 `docs/admin/architecture.md`
- 再次修改阶段边界时，优先更新本文件和计划/分析文档
- 再次开放审计日志、高危日志或其它已移除前端页面前，必须先完成文档确认
