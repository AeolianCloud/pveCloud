# 项目阶段进度总览

## 已完成阶段

基础后台阶段已完成一个可登录、可授权、可追踪、可配置的后台底座。

## 相关文档

- 缺口分析：`docs/analysis/basic-admin-gap.md`
- 实施计划：`docs/plan/basic-admin-foundation.md`
- Web 基础前台计划：`docs/plan/web-foundation.md`
- 管理端架构：`docs/admin/architecture.md`
- 用户端架构：`docs/web/architecture.md`
- API 契约：`docs/server/api/`
- 数据库设计：`docs/server/database/design.md`

## 基础后台分阶段进度

- [x] Phase 0：安全日志基础
- [x] Phase 1：审计写入与脱敏收口
- [x] Phase 2：管理员与角色权限
- [x] Phase 3：登录会话与系统设置
- [x] Phase 4：Dashboard 与管理端体验收口
- [x] Phase 5：验收与发布准备

## 当前状态

基础后台阶段的后端能力已覆盖认证、会话、RBAC、Dashboard、管理员、角色、系统配置、审计写入和文件管理等管理域。
当前管理端前端实现范围为：

- `Login`
- `Dashboard`
- `System Settings`
- `File Management`
- `403`

`System Settings` 当前承载系统配置和管理员设置能力；管理员设置内包含管理员账号、管理员组权限和管理员会话管理。
当前数据库契约已经收口，不再保留用户端 API、用户端账号、产品、订单、支付、实例、异步任务和工单等业务域表结构。
操作日志查询已按系统设置子页面口径开放；审计日志内部写入能力仍保留。

## 当前推进阶段

当前准备进入 Web 基础前台阶段。

本阶段目标：创建最小可运行用户端前台，为后续产品、账号、订单、支付、实例和工单业务接入提供前端壳、路由和请求边界准备。

本阶段范围：

- Home
- Products 占位页
- Pricing 占位页
- Login / Register 占位页
- User Center 占位页
- 404

本阶段不开放：

- 用户端业务接口（公开站点配置和用户登录会话接口除外）
- 用户端账号、产品、订单、支付、实例、工单和异步任务
- 数据库业务域迁移
- 管理端新增业务页面

## 下一步维护原则

- 再次修改当前后台范围时，优先更新 `docs/admin/architecture.md`
- 修改 Web 前端页面范围、路由、请求封装或状态语义时，优先更新 `docs/web/architecture.md`
- 再次修改阶段边界时，优先更新本文档和计划/分析文档
- 再次开放公开站点配置以外的用户端 API、用户端账号、产品、订单、支付、实例、工单、异步任务或其它管理端页面前，必须先完成文档确认
