# 本地开发说明

## 基础依赖

- Go 1.26.2
- Bun
- Air
- MariaDB 11.4.x
- Redis

## 角色分工

- MariaDB：业务事实来源
- Redis：缓存、限流、验证码、短 TTL 状态、短锁和辅助幂等
- API：HTTP 接口
- Worker：后台异步任务、实例生命周期和通知任务
- `admin`：管理端前端
- `web`：用户端前端；当前承载公开站点、产品目录、用户账号自助、支付宝/微信侧个人实名入口、订单、续费订单、实例和工单，不代表真实支付 API 已开放

## 配置

- 示例配置：`server/config.example.yaml`
- 真实配置：`server/config.yaml`
- `server/config.yaml` 默认不提交；维护者明确要求时可纳入提交
- 新增配置项时先更新示例配置
- `app.timezone` 使用 IANA 时区名；API 启动后以该配置设置 Go 进程默认时区，不依赖宿主机系统时区
- 用户端密码找回依赖 `mail` 配置；本地未配置 SMTP 时，密码找回申请应返回服务不可用提示，不应生成可用重置 token
- 支付宝/微信侧实名依赖后台系统设置中的实名供应商配置；本地未配置供应商密钥时，应保持 `real_name.enabled=false`、关闭对应 `real_name.<provider>.enabled`，或不把对应供应商加入 `real_name.allowed_providers`

## 推荐启动顺序

1. MariaDB
2. Redis
3. API
4. Worker（需要验证异步任务、到期提醒或自动释放时启动）
5. `admin`
6. `web`

## 检查接口

```text
GET http://localhost:8080/healthz
GET http://localhost:8080/admin-api/ping
```

## 常用命令

后端：

```powershell
cd server
Copy-Item config.example.yaml config.yaml
go test ./...
air -c .air.toml
```

Worker：

```powershell
cd server
go run ./cmd/worker
```

说明：本地只验证普通 API 时可以保持 `worker.enabled=false` 且不启动 Worker；验证实例 operation 自动同步、到期提醒、到期释放或通知任务时，需要在真实 `server/config.yaml` 中启用 Worker 和相关实例生命周期配置。

后端 API 变更验证：

```powershell
cd server
gofmt -w .
go test ./...
```

说明：修改 handler、service、路由、DTO、错误码、鉴权、权限、审计、事务、幂等、外部集成或安全校验时，必须评估是否需要正式 `*_test.go` 回归测试；测试范围和测试文件保留口径按 `docs/server/go-technical.md` 执行。只构建或手工点接口不算完整验收。

管理端：

```powershell
cd admin
bun install
bun dev
```

用户端：

```powershell
cd web
bun install
bun dev
```

用户端构建验证：

```powershell
cd web
bun run build
```

说明：`web/` 已存在。当前用户端开放公开站点配置、用户账号自助、支付宝/微信侧用户实名、服务器产品目录、订单、续费订单、实例和工单接口，不代表真实支付 API 已开放。

## 本地一键启动脚本约定

仓库根目录可维护 `scripts/dev.mjs` 用于开发环境同时启动 API 和前端服务。

脚本边界：

- 使用 Node.js，保持跨平台
- 不覆盖已有 `server/config.yaml`
- 默认启动 API、`admin` 和 `web`
- 不负责启动 MariaDB 和 Redis
