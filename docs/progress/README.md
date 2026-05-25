# 进度文档说明

`docs/progress/` 是项目阶段账本，用来记录阶段状态、完成情况、验收口径和范围变化历史。

这里的文档不是最终接口契约、页面契约或数据库契约。

## 职责

- 记录某个阶段做了什么
- 记录阶段验收是否完成
- 记录范围变化的背景和时间点
- 帮助维护者和 AI 恢复上下文
- 防止后续误恢复已经收口或尚未开放的能力

## 权威顺序

如果进度文档和当前契约冲突，以以下位置为准：

- API 契约：`docs/server/api/`
- 后端架构与业务规则：`docs/server/`
- 管理端前端契约：`docs/admin/`
- 用户端前端契约：`docs/web/`
- 数据库可执行契约：`server/migrations/`
- 配置示例契约：`server/config.example.yaml`

确认权威契约后，再同步修正进度文档。

## 生命周期

阶段进行中：

- `MASTER.md` 记录当前总览
- 进行中的阶段文件放在 `docs/progress/` 根目录，按需要记录阶段进度、验收和重要说明

当前进行中的阶段入口：

- 当前暂无已确认的进行中代码阶段入口；工单关联实例排障入口已完成代码层验证，当前契约回到对应 owner docs 和迁移。
- 最近阶段复盘入口：`docs/progress/worker-instance-lifecycle-code-start.md`

阶段完成后：

- 当前事实必须沉淀到对应权威文档
- `MASTER.md` 只保留简短当前状态
- 旧阶段文件归档到 `docs/progress/archive/`
- 归档文件只用于恢复历史背景；当前契约必须回到 API、架构、页面、数据库迁移和配置示例等权威位置

## 归档记录

基础后台历史阶段记录已归档到：

- `docs/progress/archive/phase-0-audit-risk-foundation.md`
- `docs/progress/archive/phase-1-audit-write-masking.md`
- `docs/progress/archive/phase-2-admin-rbac.md`
- `docs/progress/archive/phase-3-session-config.md`
- `docs/progress/archive/phase-4-dashboard-ux.md`
- `docs/progress/archive/phase-5-acceptance.md`

## 维护规则

- 不因为功能完成就直接删除进度记录
- 不把进度记录当作新增功能的唯一依据
- 不在进度文档里写详细 API 字段、数据库结构或前端实现细节
- 修改页面范围、接口契约、权限、配置或迁移后，应同步检查相关进度文档是否需要更新或归档
