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

基础后台阶段的后端能力已覆盖认证、会话、RBAC、Dashboard、管理员、角色、系统配置、审计与高危日志等管理域。

当前管理端前端实现已经按 Element Plus 后台结构收口，并明确只保留：

- `Login`
- `Dashboard`
- `403`

管理员、角色权限、登录会话、系统设置、审计日志和高危日志的前端页面、菜单和受保护路由已从当前实现中移除。
相关后端接口、权限码和数据结构仍保留。

`docs/web/architecture.md` 当前用于用户端规划与契约准备；在仓库真正出现 `web/` 实现前，不应把它视为现有实现说明。

## 下一步维护原则

- 再次修改当前后台范围时，优先更新 `docs/admin/architecture.md`
- 再次修改阶段边界时，优先更新本文件和计划/分析文档
- 再次开放已移除前端页面前，必须先完成文档确认
