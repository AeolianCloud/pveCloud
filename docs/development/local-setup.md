# 本地开发说明

## 依赖

- Go 1.26.2
- Bun
- MariaDB 11.4.9
- Redis 可预留，第一阶段不是必需依赖

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

## 启动顺序

1. MariaDB。
2. Redis，如果当前功能需要。
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
go run ./cmd/api -config config.yaml
go run ./cmd/worker -config config.yaml
```

管理端：

```powershell
cd admin
bun install
bun dev
```
