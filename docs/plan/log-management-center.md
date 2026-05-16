# 日志系统重构规划

## 目标

把当前单一的“系统设置 -> 日志管理”重新设计为完整日志系统，覆盖后端运行日志、管理端审计日志、用户端安全/业务日志，以及 admin/web 前端错误诊断日志。

本规划定义目标边界、菜单类目、数据来源、权限和分阶段路线。Phase 1 已同步 owner docs、迁移、API 契约和前端页面契约。

## 设计原则

- 日志按来源和用途分层，不把所有日志塞进 `admin_audit_logs`。
- 管理端操作审计、用户端安全/业务日志、运行访问日志、前端错误日志分开建模。
- 日志必须带请求链路标识，能与后端访问日志、错误日志和审计记录串联。
- 日志和审计不得记录 password、token、secret、验证码、证件号码明文、实名照片、供应商完整响应、请求 body 原文或真实配置。
- 前端日志只做错误诊断，不做全量用户行为埋点。
- 用户端日志属于用户安全和业务追踪，不复用管理端审计表。
- 第一阶段优先迁移和完善已有能力，后续再引入新表、新接口和采集器。

## Phase 1 当前现状

当前管理端已将日志入口迁出系统设置：

- 父级路由：`/logs`
- 操作审计路由：`/logs/admin-operations`
- 登录安全路由：`/logs/admin-security`
- 父级菜单权限：`page.logs`
- 操作审计菜单权限：`page.logs.admin-operations`
- 登录安全菜单权限：`page.logs.admin-security`
- 接口：`GET /admin-api/audit-logs`
- 数据来源：`admin_audit_logs`
- 旧 `/system/audit-logs` 保留为前端兼容重定向到 `/logs/admin-operations`

当前后端运行日志：

- 使用 `log/slog` JSON 输出到 stdout。
- 支持 `log.level`。
- 有 `X-Request-ID`、访问日志和 panic recovery 日志。
- 不入库，不提供管理端查询页面。

当前缺口：
- 用户安全/业务日志、前端错误日志和运行日志查询仍需继续推进。

## 日志分类

### 后端日志

后端日志服务于运维排障和系统运行观察。

分类：

- 运行日志：服务启动、停止、配置加载、依赖初始化、后台错误。
- 访问日志：HTTP 请求方法、路径、状态码、耗时、IP、请求 ID。
- 错误日志：panic、内部错误、依赖不可用。
- 慢请求日志：超过阈值的 HTTP 请求或关键内部调用。

当前阶段后端日志只输出 stdout。后续如需管理端查询，应接入外部日志平台、文件采集或独立运行日志存储，不进入 `admin_audit_logs`。

### 管理端日志

管理端日志服务于后台追责、权限审计和运营排查。

分类：

- 操作审计：管理员对配置、权限、文件、实名、产品、订单、工单等后台资源的写操作。
- 登录安全：管理员登录成功、登录失败、登录限流、验证码限流、退出、会话刷新、会话吊销。
- 敏感详情访问：查看审计详情、导出日志等日志自身的敏感操作，后续需要单独审计。

第一阶段继续使用 `admin_audit_logs` 承载操作审计和登录安全日志。

### 用户端日志

用户端日志服务于用户账号安全、客服排查和业务争议追踪。

建议单独建模，不复用 `admin_audit_logs`。

分类：

- 用户安全日志：登录成功、登录失败、注册、退出、token 刷新、密码重置申请、密码重置完成、会话失效。
- 用户业务日志：资料修改、实名提交、实名同步、订单创建、订单取消、工单创建、工单回复、工单关闭。

用户端日志不等同于前端埋点。它应由后端在 `/api/*` 关键业务流程中写入，前端只展示必要结果。

### 前端日志

前端日志服务于 admin/web 页面错误诊断。

分类：

- Admin 前端错误日志：管理端 JS error、unhandled promise rejection、资源加载失败、关键 API 异常摘要。
- Web 前端错误日志：用户端 JS error、unhandled promise rejection、资源加载失败、关键 API 异常摘要。

前端日志必须区分来源应用：

- `admin`
- `web`

允许采集：

- 应用版本
- 页面路径
- 浏览器和系统摘要
- 错误类型、错误消息摘要、脱敏后的堆栈摘要
- 请求 ID 或 trace ID
- API 路径、HTTP 状态、业务错误码

禁止采集：

- token、密码、验证码
- 请求体原文
- 响应体原文
- 证件号明文、实名照片、人脸照片
- 用户输入原文、敏感配置、供应商完整响应

## 日志管理中心菜单

新增一级菜单：

- 标题：日志管理中心
- 路径：`/logs`
- 菜单权限：`page.logs`
- 建议图标：`DocumentText` 或 `List`

二级类目规划：

| 类目 | 路径 | 菜单权限 | 数据来源 | 阶段 |
| --- | --- | --- | --- | --- |
| 操作审计 | `/logs/admin-operations` | `page.logs.admin-operations` | `admin_audit_logs` 非认证类后台操作 | Phase 1 |
| 登录安全 | `/logs/admin-security` | `page.logs.admin-security` | `admin_audit_logs` 中 `object_type=admin_auth` | Phase 1 |
| 用户安全日志 | `/logs/user-security` | `page.logs.user-security` | 新用户安全日志表 | Phase 2 |
| 用户业务日志 | `/logs/user-business` | `page.logs.user-business` | 新用户业务日志表 | Phase 2 |
| 前端错误日志 | `/logs/frontend-errors` | `page.logs.frontend-errors` | 新前端错误日志表或外部错误平台 | Phase 3 |
| 后端运行日志 | `/logs/backend-runtime` | `page.logs.backend-runtime` | stdout 采集、文件采集或外部日志平台 | Phase 4 |

第一阶段已将当前 `/system/audit-logs` 迁出系统设置，并保留旧路径到 `/logs/admin-operations` 的前端兼容重定向。

## 权限规划

菜单权限：

- `page.logs`
- `page.logs.admin-operations`
- `page.logs.admin-security`
- `page.logs.user-security`
- `page.logs.user-business`
- `page.logs.frontend-errors`
- `page.logs.backend-runtime`

操作权限：

- `audit-log:*`：管理端操作审计全权限
- `audit-log:sensitive-view`：查看管理端操作审计敏感详情
- `admin-security-log:*`：管理端登录安全日志全权限
- `admin-security-log:view`：查看管理端登录安全日志
- `user-security-log:*`：用户安全日志全权限
- `user-security-log:view`：查看用户安全日志
- `user-business-log:*`：用户业务日志全权限
- `user-business-log:view`：查看用户业务日志
- `frontend-error-log:*`：前端错误日志全权限
- `frontend-error-log:view`：查看前端错误日志
- `backend-runtime-log:*`：后端运行日志全权限
- `backend-runtime-log:view`：查看后端运行日志
- `log:sensitive-view`：统一敏感详情查看权限，是否替代现有 `audit-log:sensitive-view` 需单独确认
- `log:export`：日志导出权限，后续导出能力启用前再开放

日志导出、敏感详情查看、日志清理等高敏操作应写入管理端操作审计。

## 数据模型方向

### 现有表

`admin_audit_logs` 继续承载管理端操作审计和第一阶段登录安全日志。

第一阶段可增强查询条件，但不改表结构。

### 建议新增表

后续阶段建议新增：

- `user_security_logs`：用户认证、安全事件和会话相关日志。
- `user_business_logs`：用户关键业务轨迹，例如实名、订单、工单。
- `frontend_error_logs`：admin/web 前端错误诊断日志。

后端运行日志不优先建业务库表。应优先评估外部日志平台或文件采集，避免把高频访问日志压到 MariaDB。

## API 规划

### Phase 1 复用接口

- `GET /admin-api/audit-logs`

建议增强查询参数：

- `log_type=admin_operation|admin_security`
- `request_id`
- `ip`
- `module`

### 后续新增接口

管理端查询：

- `GET /admin-api/logs/admin-operations`
- `GET /admin-api/logs/admin-security`
- `GET /admin-api/logs/user-security`
- `GET /admin-api/logs/user-business`
- `GET /admin-api/logs/frontend-errors`
- `GET /admin-api/logs/backend-runtime`

前端错误上报：

- `POST /admin-api/client-logs/errors`：admin 前端错误上报
- `POST /api/client-logs/errors`：web 前端错误上报

前端错误上报接口必须限流、截断字段、脱敏、拒绝大 payload，并且不得要求前端上传敏感上下文。

## 页面能力

### 操作审计

- 展示管理员后台操作审计。
- 支持管理员、模块、动作、对象、请求 ID、IP 和时间筛选。
- 展示动作中文名，同时保留 action 原始值可复制。
- 敏感详情按权限展示。

### 登录安全

- 展示管理端登录、失败、限流、验证码限流、退出和刷新。
- 支持管理员、动作、IP、请求 ID 和时间筛选。
- 后续可增加异常 IP、连续失败、账号风险聚合。

### 用户安全日志

- 展示用户登录、注册、退出、密码重置、会话刷新等安全事件。
- 支持用户、动作、IP、请求 ID 和时间筛选。
- 可作为客服排查和账号安全依据。

### 用户业务日志

- 展示用户关键业务轨迹。
- 支持用户、业务模块、对象编号、动作和时间筛选。
- 第一批覆盖实名、订单、工单。

### 前端错误日志

- 展示 admin/web 前端错误。
- 支持应用来源、页面路径、错误类型、API 路径、状态码、请求 ID 和时间筛选。
- 仅展示脱敏后的错误摘要。

### 后端运行日志

- 展示服务运行、访问、错误、panic 和慢请求。
- 具体能力取决于采集方案。
- 不在 Phase 1 实现。

## 分阶段实施

### Phase 1：日志管理中心和现有能力迁移

- 更新 admin owner docs，把日志从系统设置拆出为独立一级菜单。
- 更新路由与权限文档，新增 `/logs` 父路由和两个第一阶段子路由。
- 新增权限迁移，创建 `page.logs`、`page.logs.admin-operations`、`page.logs.admin-security`。
- 将 `audit-log:*`、`audit-log:sensitive-view` 挂到新菜单节点。
- 前端迁移当前日志页面，拆为操作审计和登录安全两个类目。
- 旧 `/system/audit-logs` 前端重定向到 `/logs/admin-operations`。

### Phase 2：用户端安全/业务日志

- 设计 `user_security_logs` 和 `user_business_logs`。
- 在 `/api/*` 认证、账号、实名、订单、工单关键流程写入日志。
- 新增管理端查询接口和页面。
- 明确用户是否可在用户中心查看自己的登录记录；如果开放，需要另行设计 `/api/*` 用户自查接口。

### Phase 3：admin/web 前端错误日志

- 设计 `frontend_error_logs` 或选定外部错误平台。
- 分别为 admin 和 web 增加错误采集器。
- 增加 `/admin-api/client-logs/errors` 和 `/api/client-logs/errors`。
- 完成限流、字段截断、脱敏和 request ID 串联。

### Phase 4：后端运行访问日志查询

- 确认采集方案：外部日志平台优先，其次文件采集；不优先入业务库。
- 设计查询 API、权限、留存、脱敏和审计策略。
- 后续再实现管理端运行日志查询。

### Phase 5：导出、留存和清理

- 设计导出权限、导出审计、导出字段脱敏。
- 定义不同日志类型的留存策略。
- 定义归档、清理和备份权限。

## 当前执行范围

这次持续推进时，Phase 2-5 统一按以下收口：

- Phase 2：用户安全/业务日志落库、写入和管理端查询页面。
- Phase 3：admin/web 前端错误上报、管理端查询页面和脱敏展示。
- Phase 4：后端运行日志的结构化查询页面和可配置采集边界。
- Phase 5：导出审计、留存配置和清理接口。

## 后续需要确认

- Web 用户安全/业务日志是否进入 Phase 2，而不是和 Phase 1 混做。
- 前端错误日志是否使用自建表，还是接入外部错误平台。
- 后端运行日志是否坚持 stdout + 外部采集，不进入 MariaDB。

后续阶段确认后再更新对应 `docs/admin/`、`docs/web/`、`docs/server/api/`、`docs/server/database/`、迁移和实现。
