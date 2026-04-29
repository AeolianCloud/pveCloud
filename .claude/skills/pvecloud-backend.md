---
name: pvecloud-backend
description: Backend implementation guardrails for pveCloud. Use when working on Go server code, handlers, services, integrations, or API endpoints.
---

# Backend Guardrails

## 先读什么

- `docs/server/README.md`
- `docs/server/architecture.md`
- `docs/server/api/`
- `docs/server/database/design.md`
- `server/migrations/`
- `server/config.example.yaml`

## 实现边界

- API 契约来自 `docs/server/api/`，不要只改 handler 或 DTO。
- 表结构契约最终来自 `server/migrations/`，不是口头约定。
- 配置项契约最终来自 `server/config.example.yaml`。
- 业务规则、状态机、事务边界和幂等规则写进 `docs/server/`。

## 代码守则

- 服务负责业务规则，handler 负责请求解析、权限声明和响应。
- 目录先按边界分到 `internal/admin`、`internal/web`、`internal/job`，再在边界内按模块分目录。
- 管理端与用户端模块不要继续落在扁平共享目录里。
- 真正跨端复用的核心规则才进入 `internal/domain`。
- 外部系统协议适配放在 `platform/integrations/`，业务裁决放在模块服务或领域服务。
- 不把 RBAC 最终授权逻辑下放到前端。
- 不把长耗时外部调用放进长事务。
- 幂等必须依赖业务唯一键、状态检查或任务键，不能只依赖前端防重复点击。

## 命名

- 使用明确业务域命名。
- 避免把具体业务逻辑埋进 `common`、`helper`、`manager`、`base` 之类泛名。
- 审计域统一使用 `audit` 命名。

## 验证

```powershell
cd server
gofmt -w .
go test ./...
```

## 测试文件保留

- `*_test.go`、临时验证脚本和一次性探测代码都默认视为验证产物，而不是必须长期保留的资产。
- 为修复、重构或新增能力临时创建的测试文件，测试通过后默认删除；只有维护者明确要求保留时才留下或提交。
- 如果某个测试文件承载长期回归保护价值，保留前也要以维护者明确要求为准，不自行默认留下。
- 临时探测脚本、一次性样例或本地数据生成文件不要伪装成正式测试；确需使用时放在临时目录或明确命名，并在验证后清理。
