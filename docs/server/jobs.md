# 异步任务与 Worker

API 进程负责创建任务。Worker 进程负责执行任务。
任务最终状态以 `async_tasks` 表和对应迁移为准。

## 当前任务类型

```text
instance_create
instance_renew
order_expire
payment_check
instance_status_sync
```

## 代码落点

Worker 相关代码统一按以下边界组织：

- `server/internal/job/`：任务调度、存储、状态推进
- `server/internal/job/handlers/`：按任务类型拆分的处理器
- `server/internal/domain/*`：被任务复用的跨端核心业务规则

不要把具体任务逻辑继续堆在启动装配代码里，也不要把任务实现散落到无边界的公共目录。

## 执行原则

- Worker 只领取 `status=pending` 且 `run_at <= now()` 的任务
- 用 `locked_by`、`locked_until` 做抢占控制
- 执行前重新检查业务状态，防止重复执行或状态倒退
- 失败时记录 `last_error`、增加 `retry_count` 并计算下次执行时间
- 超过最大重试次数后标记为 `failed`

## 任务状态

`async_tasks.status` 当前使用以下状态：

```text
pending
running
succeeded
failed
```

- `pending`：待领取或等待下次重试
- `running`：已被 Worker 领取且锁仍有效
- `succeeded`：任务执行完成
- `failed`：超过最大重试次数或遇到不可重试错误

Worker 只能领取 `pending` 且 `run_at <= now()` 的任务。
如果发现 `running` 任务的 `locked_until < now()`，后续恢复策略必须先写入文档并确认，不能静默抢占。

## 领取规则

单轮领取按以下顺序选择任务：

1. `priority DESC`
2. `run_at ASC`
3. `id ASC`

领取必须在数据库事务内完成，并以行锁或等价数据库机制保证多 Worker 不会重复领取同一任务。
领取成功后必须写入：

- `status = running`
- `locked_by = worker.id`
- `locked_until = now() + worker.lock_ttl`

未领取到任务时，Worker 按 `worker.poll_interval_seconds` 等待下一轮。

## Handler 注册

Worker 主循环只负责调度，不承载具体业务逻辑。

每个任务类型必须注册一个 handler：

```text
task_type -> handler
```

handler 输入为任务记录和已解析 payload。
handler 输出分三类：

- 成功：标记 `succeeded`，写入 `result`、`finished_at`，清空锁
- 可重试失败：增加 `retry_count`，写入 `last_error`，计算下一次 `run_at`，状态回到 `pending`
- 不可重试失败：标记 `failed`，写入 `last_error`、`finished_at`，清空锁

遇到未注册的 `task_type`，视为不可重试失败。

## 重试规则

默认重试间隔使用指数退避：

```text
min(2^retry_count * worker.poll_interval_seconds, 30 minutes)
```

具体任务可以在文档中定义更严格的重试上限或不可重试错误。
当 `retry_count + 1 >= max_retries` 时，本次失败后标记为 `failed`。

## 幂等规则

任务创建必须提供稳定的 `idempotency_key`。
同一业务动作重复创建任务时，应复用或命中 `async_tasks.idempotency_key` 唯一约束，而不是创建多条并发任务。

handler 必须在执行前重新读取业务对象并检查状态：

- 目标状态已经完成时，任务可直接标记成功
- 目标状态不允许继续时，任务应标记不可重试失败或按业务规则跳过
- 外部系统调用前后都要保存本地恢复锚点

## 一致性原则

- Redis 可用于短锁、防抖和临时标记
- 任务领取、重试次数、最终状态和错误信息仍以 `async_tasks` 为准
- 不在长事务中调用外部系统
- 外部系统调用不得发生在持有数据库长事务或任务领取事务期间

## 事务建议

实例开通至少拆成两段本地事务：

1. 领取任务、锁定订单、创建或复用实例占位、写入恢复锚点、标记订单开通中
2. 外部任务成功后，更新实例、订单和任务最终状态

## 测试要求

Worker 落地时至少覆盖以下测试：

- 单 Worker 能按优先级领取到期任务
- 多 Worker 并发领取不会重复处理同一任务
- 成功任务写入 `succeeded`、`result`、`finished_at`
- 可重试失败增加 `retry_count` 并回到 `pending`
- 超过最大重试次数后进入 `failed`
- 未注册任务类型进入 `failed`
- handler 重入时能通过业务状态检查保持幂等

## 目标

Worker 的目标不是“尽快把外部请求打出去”，而是让可重试、可恢复、可追踪的异步执行成立。
