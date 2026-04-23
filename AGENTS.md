# pveCloud — Codex Agent 工作指南
## 项目概述

PVE Cloud 是云资源销售平台，覆盖商品目录、下单、支付、开通、实例操作和用户/管理员前端 MVP 闭环。

- **后端**：Go 模块化单体，三入口 `public-api` / `admin-api` / `worker`
- **数据库**：MariaDB（业务事实唯一来源）+ Redis（缓存/会话/幂等辅助）
- **前端**：`web/` 和 `admin/`，均为 Bun + Vue 3 SPA
- **模块 path**：`github.com/AeolianCloud/pveCloud/server`

## 目录结构

```
pveCloud/
  server/
    cmd/public-api/     # 用户侧 HTTP :8080
    cmd/admin-api/      # 管理侧 HTTP :8081
    cmd/worker/         # 异步任务 worker :8082
    config/config.yaml  # 唯一配置，YAML only
    internal/
      bootstrap/        # 装配层
      common/           # database, response, errors, testutil
      auth/
      user/
      adminuser/
      catalog/
      billing/
      order/
      payment/
      task/
      instance/
      resource/
      audit/
      notification/
  web/
  admin/
  docs/
    adr/
    superpowers/specs/
    superpowers/plans/
  docker-compose.yml
```

## 环境准备

```bash
# 启动基础设施
docker compose up -d

# 验证后端可编译
go -C server build ./cmd/public-api ./cmd/admin-api ./cmd/worker

# 运行后端测试
go -C server test ./...

# 前端
bun --cwd web run test && bun --cwd web run build
bun --cwd admin run test && bun --cwd admin run build
```

## 架构约定

1. **MariaDB 是唯一业务事实源**。Redis 只做缓存/会话/幂等，不存业务状态。
2. **装配顺序**：配置 → DB/Redis → repository → service → handler，不得跳层。
3. **三条强事务边界**（必须在单个 `WithTx` 内完成）：
   - 下单：`orders` + `billing_records` + `payment_orders` + `resource_reservations`
   - 支付成功：回调日志 + `payment_orders` 状态 + `orders` 状态 + 唯一任务创建
   - 开通成功：`instances` + `instance_services` + `orders` 状态 + 任务完成标记
4. **配置只用 YAML**：`server/config/config.yaml`，无 `.env`，无环境变量覆盖。
5. **事务 helper**：`internal/common/database.WithTx(ctx, db, fn)`，所有跨表写操作必须通过此函数。
6. **路由**：`net/http ServeMux`，不引入第三方路由框架。
7. **无 ORM**：`database/sql` 直接使用。

## 编码规范

- 错误统一用 `internal/common/errors` 包装，handler 层统一用 `internal/common/response` 返回
- repository 层写 integration test，需要真实 MariaDB 连接
- service 层测试可 mock repository 接口
- 不写解释性注释，只在 WHY 不明显时写一行
- 不加 feature flag，不做向后兼容 shim

## 提交规范

- 每个 Task 单独提交，不混入其他 Task 改动
- 提交信息格式：`feat: <简短描述>`
- 测试必须在提交前通过：`go -C server test ./...`

## ADR 摘要

- **ADR-001**：MariaDB `tasks` 表是任务唯一权威，Redis 只做加速辅助
- **ADR-002**：下单事务内写 `resource_reservations` 预留容量，开通成功后转正式记录
