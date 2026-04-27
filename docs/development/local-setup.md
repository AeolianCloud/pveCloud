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
- Worker：异步任务执行
- `admin`：管理端前端
- `web`：未来用户端前端；当前仓库尚未落地实现

## 配置

- 示例配置：`server/config.example.yaml`
- 真实配置：`server/config.yaml`
- `server/config.yaml` 不提交
- 新增配置项时先更新示例配置

## 推荐启动顺序

1. MariaDB
2. Redis
3. API
4. Worker
5. `admin`
6. 如果未来存在 `web/`，再启动 `web`

## 检查接口

```text
GET http://localhost:8080/healthz
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

## 本地一键启动脚本约定

仓库根目录可维护 `scripts/dev.mjs` 用于开发环境同时启动 API、Worker 和前端服务。

脚本边界：

- 使用 Node.js，保持跨平台
- 不覆盖已有 `server/config.yaml`
- 默认启动 API、Worker、`admin`
- 只有在真实存在 `web/package.json` 时才启动 `web`
- 不负责启动 MariaDB 和 Redis
