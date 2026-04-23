# pveCloud — Claude Code 工作指南
## 项目概述

PVE Cloud 是一个云资源销售平台，覆盖商品目录、下单、支付、开通、实例操作和用户/管理员前端的 MVP 闭环。

- **后端**：Go 模块化单体，三入口 `public-api` / `admin-api` / `worker`
- **数据库**：MariaDB（业务事实唯一来源）+ Redis（缓存/会话/幂等辅助）
- **前端**：`web/`（用户端）和 `admin/`（管理端），均为 Bun + Vue 3 SPA
- **模块 path**：`github.com/AeolianCloud/pveCloud/server`

## 目录结构

```
pveCloud/
  server/
    cmd/public-api/     # 用户侧 HTTP 入口 :8080
    cmd/admin-api/      # 管理侧 HTTP 入口 :8081
    cmd/worker/         # 异步任务 worker :8082
    config/config.yaml  # 唯一配置文件，YAML only，无 .env
    internal/
      bootstrap/        # 装配：配置 -> DB/Redis -> repo -> service -> handler
      common/           # 共享：database, response, errors, testutil
      auth/             # JWT 鉴权
      user/             # 用户注册登录
      adminuser/        # 管理员登录
      catalog/          # 商品目录、容量预留
      billing/          # 账单快照
      order/            # 订单状态机
      payment/          # 支付单、回调、幂等
      task/             # 异步任务 claim/execute/retry
      instance/         # 实例和服务周期
      resource/         # mock provider（稳定 adapter 合约）
      audit/            # 业务事件记录
      notification/     # 通知入口
  web/                  # 用户端 SPA
  admin/                # 管理端 SPA
  docs/
    adr/                # 架构决策记录
    superpowers/specs/  # 设计文档
    superpowers/plans/  # 实施计划
  docker-compose.yml    # MariaDB + Redis
  start.bat             # 本地一键启动（Windows）
```

## 开发命令

```bash
# 基础设施
docker compose up -d

# 后端
go -C server test ./...
go -C server build ./cmd/public-api ./cmd/admin-api ./cmd/worker
go -C server run ./cmd/public-api

# 前端
bun --cwd web run test && bun --cwd web run build
bun --cwd admin run test && bun --cwd admin run build

# 健康检查
curl http://127.0.0.1:8080/healthz
```

## 架构约定（必须遵守）

1. **MariaDB 是唯一业务事实源**。Redis 只做缓存/会话/幂等辅助，不存业务状态。
2. **装配顺序固定**：配置 → DB/Redis → repository → service → handler。不得跳层。
3. **三条强事务边界**：
   - 下单事务：`orders` + `billing_records` + `payment_orders` + 预留关系
   - 支付成功事务：回调日志 + `payment_orders` 状态 + `orders` 状态 + 创建唯一任务
   - 开通成功事务：`instances` + `instance_services` + `orders` 状态 + 任务完成
4. **配置只用 YAML**：`server/config/config.yaml`，无 `.env`，无环境变量覆盖。
5. **事务 helper**：`internal/common/database.WithTx(ctx, db, fn)`，repository 统一通过此方式注入事务。
6. **net/http ServeMux**：不引入第三方路由框架。

## 编码规范

- Go：标准库优先，`database/sql` 直接使用，不引入 ORM
- 错误处理：`internal/common/errors` 统一包装，handler 层统一用 `internal/common/response` 返回
- 测试：repository 层写 integration test（需真实 DB），service 层可 mock repository
- 不写注释，除非 WHY 不明显（隐藏约束、绕过特定 bug、会让读者意外的行为）
- 不加 feature flag，不做向后兼容 shim，直接改代码

## ADR 摘要

- **ADR-001**：任务事实源 — MariaDB `tasks` 表是唯一权威，Redis 只做加速
- **ADR-002**：容量预留 — 下单时在事务内写 `resource_reservations`，开通成功后转正式记录
