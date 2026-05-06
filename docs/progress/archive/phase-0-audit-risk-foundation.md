# Phase 0：审计日志基座

> 历史记录：本文件只记录 Phase 0 当时的审计和高危日志阶段背景。当前日志、数据库和 API 契约以 `docs/server/`、`docs/server/api/` 和 `server/migrations/` 为准。

- [x] 历史阶段曾建立审计与高危日志相关 API 契约
- [x] 当前契约已收口为普通审计日志，`admin_risk_logs` 不再属于数据库契约
- [x] 审计日志查询接口当前未开放，内部写入能力保留
- [x] 明确阶段边界与日志域职责
- [x] 当前阶段文档以 `docs/server/`、`docs/server/api/` 与 `server/migrations/` 为准

## Notes

本文件记录历史阶段背景。
当前持久契约已经删除高危操作日志表和高危日志权限，只保留普通审计日志能力。

后续如果要重新引入独立高危日志，必须先重新更新数据库迁移、API/架构文档和验收口径。
