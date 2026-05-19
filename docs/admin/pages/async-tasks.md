# Async Tasks 页面契约

异步任务页面用于管理端查看 Worker 后台任务执行情况，并对失败任务进行人工重试。页面只调用 `/admin-api/*`。

## 页面范围

- 异步任务列表
- 任务详情
- 失败任务重试
- 按任务类型、状态、业务对象和时间范围筛选

本页面不直接执行实例操作，不替代实例管理页面；所有业务状态最终以对应业务接口返回为准。

## 路由与权限

- 路由：`/async-tasks`
- 菜单权限：`page.async-tasks`
- 查看：`page.async-tasks`
- 重试：`async-task:retry` 或 `async-task:*`

## 行为约束

- 列表展示任务编号、类型、状态、业务对象、计划执行时间、尝试次数、最近错误、锁定 Worker、创建时间和完成时间。
- 任务详情不得展示敏感 `payload`、SMTP 凭据、MCP Bearer Token、用户敏感明文或完整上游响应。
- 仅 `failed` 任务展示重试入口。
- 重试任务必须二次确认，并以服务端返回状态为准。

## 关联接口

- `GET /admin-api/async-tasks`
- `POST /admin-api/async-tasks/{task_no}/retry`

## 验收重点

- 无权限访问 `/async-tasks` 时展示管理端 403 反馈。
- 低权限管理员看不到或无法触发重试按钮。
- 任务列表、筛选、详情和失败重试正常。
- 页面不泄露任务内部敏感 payload 或完整上游响应。
