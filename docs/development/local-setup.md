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
- `web`：用户端前端；当前承载公开站点、产品目录、用户账号自助、支付宝/微信侧个人实名入口、订单、真实支付、续费订单、实例和工单

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

## 本地数据库分工

本地开发库和自动化集成测试库必须分开使用，避免测试创建、删除临时库时影响开发数据。

- 开发库：沿用 `server/config.yaml` 中的 `database` 配置，供 API、Worker、管理端和用户端联调使用。
- 测试库：使用本仓库 `compose.test-db.yml` 启动的本地 MariaDB 容器，只用于 `PVECLOUD_TEST_MYSQL_DSN` 驱动的事务/锁/幂等集成测试。
- 不要把 `server/config.yaml` 指向测试库端口 `3307`；应用开发仍连接开发库。
- 不要把 `PVECLOUD_TEST_MYSQL_DSN` 指向开发库；集成测试会创建并删除 `pvecloud_it_*` 临时数据库。

测试库容器长期运行在本机 Docker 中：

```powershell
docker compose -f compose.test-db.yml up -d
```

测试库固定连接信息：

```text
host: 127.0.0.1
port: 3307
database: pvecloud_test_seed
user: root
password: pvecloud_test_root
```

运行后端全量测试，包括需要真实 MariaDB 行锁、事务、生成列和接口路由语义的集成测试：

```powershell
cd server
$env:PVECLOUD_TEST_MYSQL_DSN = "root:pvecloud_test_root@tcp(127.0.0.1:3307)/pvecloud_test_seed?parseTime=true&loc=Local"
$env:GOCACHE = "/tmp/pvecloud-go-build"
go test ./...
```

Linux shell 等价命令：

```bash
cd server
PVECLOUD_TEST_MYSQL_DSN='root:pvecloud_test_root@tcp(127.0.0.1:3307)/pvecloud_test_seed?parseTime=true&loc=Local' \
GOCACHE=/tmp/pvecloud-go-build \
go test ./...
```

未设置 `PVECLOUD_TEST_MYSQL_DSN` 时，数据库集成测试会自动跳过，普通 `go test ./...` 不依赖 Docker 或 MariaDB 测试容器；但后端接口和关键链路变更的最终验收必须使用上面的 Docker MariaDB 测试库命令。

## 后端接口测试验证

后端接口和关键后端链路的回归测试以本地 Docker MariaDB 测试库为默认验证环境，不使用开发库或真实配置库承载自动化测试。

- 新增或修改任一后端接口时，必须补对应长期回归测试；管理端接口、用户端接口、公开回调和健康检查都纳入后端接口测试范围。
- 接口测试应优先覆盖 handler 层，验证路由、鉴权入口、请求解析、统一响应包裹、错误码、资源归属和关键成功/失败分支；涉及事务、锁、唯一约束、异步任务或迁移字段时，同步补 service 或 repository 集成测试。
- 本地验证先启动 `compose.test-db.yml`，再设置 `PVECLOUD_TEST_MYSQL_DSN` 运行后端测试；测试失败时先修复代码或测试数据初始化，再重新执行同一组命令直到通过。
- 普通不带 `PVECLOUD_TEST_MYSQL_DSN` 的 `go test ./...` 只用于快速单元测试和跳过集成测试的 smoke check，不能替代本地 Docker MariaDB 下的后端接口验证。
- CI 后续应按同一口径提供 MariaDB 11.4 service container，并设置 `PVECLOUD_TEST_MYSQL_DSN` 执行后端测试；CI 日志不得输出真实生产数据库 DSN，job 只使用临时 service container 凭据。

如需停止测试库但保留数据：

```powershell
docker compose -f compose.test-db.yml stop
```

如需彻底删除测试库数据卷：

```powershell
docker compose -f compose.test-db.yml down -v
```

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

说明：`web/` 已存在。当前用户端开放公开站点配置、用户账号自助、支付宝/微信侧用户实名、服务器产品目录、订单、真实支付、续费订单、实例和工单接口。

## 本地一键启动脚本约定

仓库根目录可维护 `scripts/dev.mjs` 用于开发环境同时启动 API 和前端服务。

脚本边界：

- 使用 Node.js，保持跨平台
- 不覆盖已有 `server/config.yaml`
- 默认启动 API、`admin` 和 `web`
- 不负责启动 MariaDB 和 Redis
