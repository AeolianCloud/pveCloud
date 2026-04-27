# Backend Guardrails

本文件定义后端实现守则，不承载接口字段或表结构契约。

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
- 外部系统协议适配放在 `integrations/`，业务裁决放在 `services/`。
- 不把 RBAC 最终授权逻辑下放到前端。
- 不把长耗时外部调用放进长事务。
- 幂等必须依赖业务唯一键、状态检查或任务键，不能只依赖前端防重复点击。

## 命名

- 使用明确业务域命名。
- 避免把具体业务逻辑埋进 `common`、`helper`、`manager`、`base` 之类泛名。
- 审计域统一使用 `audit` 命名。

## 验证

常用最小验证：

```powershell
cd server
gofmt -w .
go test ./...
```
