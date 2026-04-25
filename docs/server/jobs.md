# 异步任务和 Worker

本文件对应未来代码目录：

```text
server/cmd/worker/
server/internal/jobs/
```

## 基本原则

- API 进程只创建任务，不直接执行长耗时任务。
- worker 进程循环拉取可执行任务。
- 任务状态保存在 `async_tasks` 表。
- 任务必须有 `idempotency_key`，防止重复创建和重复执行副作用。

## 第一阶段任务类型

```text
instance_create       实例开通
instance_renew        实例续费同步
order_expire          订单超时关闭
payment_check         支付状态补偿查询
instance_status_sync  实例状态同步
```

## 执行规则

1. 拉取 `pending` 且 `run_at <= now()` 的任务。
2. 抢占锁定任务，写入 `locked_by` 和 `locked_until`。
3. 执行前重新检查业务状态。
4. 调用外部系统前先确保本地恢复锚点已持久化。
5. 失败时写入 `last_error`，增加 `retry_count`，计算下次 `run_at`。
6. 超过最大重试次数后标记 `failed`，交给后台人工处理。

## 实例开通恢复锚点

实例开通任务必须先持久化：

```text
instances.vmid
instances.provisioning_key
instances.pve_task_upid
```

这样 PVE 已经创建 VM 但本地流程中断时，可以通过补偿任务找回远端 VM。
