# 异步任务和 Worker

API 进程只创建任务，Worker 进程执行任务。任务状态和可执行结构以 `async_tasks` 表及迁移 SQL 为准。

## 任务类型

```text
instance_create
instance_renew
order_expire
payment_check
instance_status_sync
```

## 拉取和锁定

- Worker 拉取 `status=pending` 且 `run_at <= now()` 的任务。
- 使用 `locked_by` 和 `locked_until` 抢占任务。
- 执行前重新检查业务状态，避免重复执行或状态倒退。
- 失败时写入 `last_error`，递增 `retry_count`，并计算下一次 `run_at`。
- 超过最大重试次数后标记为 `failed`，并在管理端暴露给运维人员。
- Redis 可用于任务触发防抖、幂等短锁或临时执行标记，但任务领取、重试次数、最终状态和错误信息仍以 `async_tasks` 表为准。

## 事务边界

实例开通至少拆为两个本地事务：

1. 领取任务、锁定订单、创建或复用实例占位、持久化 VMID/幂等锚点，并标记订单开通中。
2. PVE 任务成功后，更新实例、订单和任务最终状态。

不要在长数据库事务中调用 PVE、支付或通知服务。
