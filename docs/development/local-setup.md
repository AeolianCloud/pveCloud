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
- `admin`：管理端前端
- `web`：用户端前端；当前承载公开站点、产品目录、用户账号自助和个人实名入口，不代表订单、支付、实例或工单 API 已开放

## 配置

- 示例配置：`server/config.example.yaml`
- 真实配置：`server/config.yaml`
- `server/config.yaml` 默认不提交；维护者明确要求时可纳入提交
- 新增配置项时先更新示例配置
- 用户端密码找回依赖 `mail` 配置；本地未配置 SMTP 时，密码找回申请应返回服务不可用提示，不应生成可用重置 token

## 推荐启动顺序

1. MariaDB
2. Redis
3. API
4. `admin`
5. `web`

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

说明：`web/` 已存在。当前用户端只开放公开站点配置、用户账号自助、用户实名和服务器产品目录接口，不代表订单、支付、实例或工单 API 已开放。

## 本地一键启动脚本约定

仓库根目录可维护 `scripts/dev.mjs` 用于开发环境同时启动 API 和前端服务。

脚本边界：

- 使用 Node.js，保持跨平台
- 不覆盖已有 `server/config.yaml`
- 默认启动 API、`admin` 和 `web`
- 不负责启动 MariaDB 和 Redis
