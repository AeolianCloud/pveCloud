# 异步任务与 Worker

API 进程负责创建任务。
Worker 进程负责执行任务。

任务最终状态以 `async_tasks` 表和对应迁移为准。

## 当前任务类型

```text
instance_create
instance_renew
order_expire
payment_check
instance_status_sync
```

## 执行原则

- Worker 只领取 `status=pending` 且 `run_at <= now()` 的任务
- 用 `locked_by`、`locked_until` 做抢占控制
- 执行前重新检查业务状态，防止重复执行或状态倒退
- 失败时记录 `last_error`、增加 `retry_count` 并计算下次执行时间
- 超过最大重试次数后标记为 `failed`

## 一致性原则

- Redis 可用于短锁、防抖和临时标记
- 任务领取、重试次数、最终状态和错误信息仍以 `async_tasks` 为准
- 不在长事务中调用外部系统

## 事务建议

实例开通至少拆成两段本地事务：

1. 领取任务、锁定订单、创建或复用实例占位、写入恢复锚点、标记订单开通中
2. 外部任务成功后，更新实例、订单和任务最终状态

## 目标

Worker 的目标不是“尽快把外部请求打出去”，而是让可重试、可恢复、可追踪的异步执行成立。
