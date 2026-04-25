# 本地开发说明

## 依赖

- Go 1.26.2
- Air
- Bun
- MariaDB 11.4.9
- Redis，用于缓存、限流、短 TTL 状态、验证码、一次性 token、幂等短锁和防重复提交标记

## 数据库

目标数据库：

```text
pvecloud
```

初始化 SQL：

```text
server/migrations/001_init.sql
```

使用本地 MariaDB 凭据，不提交真实账号密码。

## 后端配置

- 示例配置：`server/config.example.yaml`
- 真实配置：`server/config.yaml`
- `server/config.yaml` 保持忽略。
- 配置以 YAML 为主，不使用 `.env` 作为主配置来源。
- 示例配置中的每个配置组和配置项都应保留中文注释，说明用途、单位、默认值语义和安全注意事项。
- 新增配置项时先更新 `server/config.example.yaml`，再同步调整读取逻辑和本地开发说明。

## 启动顺序

1. MariaDB。
2. Redis。
3. API 进程。
4. Worker 进程。
5. `admin` 和后续 `web` Vite dev server。

## 检查接口

```text
GET http://localhost:8080/healthz
GET http://localhost:8080/openapi.yaml
GET http://localhost:8080/api/ping
GET http://localhost:8080/admin-api/ping
```

## 常用命令

后端：

```powershell
cd server
Copy-Item config.example.yaml config.yaml
go test ./...
air -c .air.toml
air -c .air.worker.toml
```

管理端：

```powershell
cd admin
bun install
bun dev
```

一键本地启动脚本拟维护在仓库根目录 `scripts/dev.mjs`，用于开发环境同时启动后端 API 和前端开发服务：

```powershell
node ./scripts/dev.mjs
```

脚本约定：

- 使用 Node.js 编写，保持跨平台路径处理，不依赖 PowerShell 专有语法。
- 不写入真实密钥，不覆盖已存在的 `server/config.yaml`。
- 如 `server/config.yaml` 不存在，提示开发者先从 `server/config.example.yaml` 复制并按本机环境修改。
- 在 `server/` 下使用 `air -c .air.toml` 启动 API 热重载进程。
- 在 `server/` 下使用 `air -c .air.worker.toml` 启动 Worker 热重载进程。
- 在 `admin/` 下执行 `bun install`（依赖已存在时由 Bun 自行复用）并启动 `bun dev`。
- 如果后续存在 `web/package.json`，在 `web/` 下按同样方式启动用户端 Vite dev server。
- 默认启动 API、Worker、`admin` 和已存在的 `web`。
- 脚本不提供运行参数，避免不同开发者启动组合不一致。
- MariaDB 和 Redis 仍由开发者本机或容器环境提前启动。
