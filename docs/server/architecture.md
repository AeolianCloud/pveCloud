# 后端架构

`pveCloud` 后端当前以 Go 基础后台为主，现行契约范围聚焦管理端 API，并开放最小 Web 公开站点配置、用户账号自助接口、支付宝/微信侧用户实名接口和服务器产品目录展示接口。

## 进程职责

- API：提供管理端 HTTP 接口、Web 公开站点配置接口、用户账号自助接口、支付宝/微信侧用户实名接口、服务器产品目录展示接口和健康检查

当前契约重新开放用户账号自助、用户实名和服务器产品目录；订单、支付、实例、Worker、工单和其他用户端业务流仍不纳入现阶段交付范围。这里的支付宝/微信仅用于实名核验，不代表支付能力已开放。

## 路由边界

- 健康检查：`/healthz`
- 管理端 API：`/admin-api/*`
- Web 公开配置 API：`GET /api/site-config`
- Web 用户认证 API：`/api/auth/*`
- Web 用户资料和实名 API：`/api/user/*`
- Web 实名供应商回调 API：`POST /api/real-name/provider-callbacks/{provider}`
- Web 服务器产品目录 API：`GET /api/server-catalog`

当前仓库除站点配置、用户账号自助、支付宝/微信侧用户实名和服务器产品目录外，不再把其它 `/api/*` 作为现行后端契约。

## 事实来源

- MariaDB：基础后台事实来源
- Redis：运行时辅助依赖

Redis 可以保存缓存、限流、验证码、一次性 token、短锁和短期状态。它不能替代管理端会话最终状态、RBAC 关系或审计事实。

## 通用安全边界

跨端通用安全基线见 `docs/security.md`。

- 后端是鉴权、权限、资源归属和业务状态的最终裁决点。
- 前端权限判断、JWT 权限快照和请求参数不得替代当前数据库 RBAC、会话状态和资源归属校验。
- 安全增强如果改变接受输入、拒绝输入、返回内容、审计内容、授权方式、配置要求、存储方式或事务边界，必须先更新 `docs/security.md` 和对应 owner docs。

## 目标目录原则

后端目录当前按“基础后台可用”口径维护：

- 先按边界分：`admin`、`web`
- 再按模块分：一个模块一个目录
- 基础设施进入 `platform`
- 只有稳定且无业务语义的能力才能进入 `shared`

目标结构如下：

```text
server/
  cmd/
    api/
    setup-admin/
  internal/
    admin/
      routes/
      middleware/
      modules/
        auth/
        dashboard/
        admin_user/
        admin_role/
        system_config/
        audit/
        file_attachment/
        real_name/
    web/
      routes/
      modules/
        site_config/
        real_name/
        product_catalog/
    domain/
    platform/
      bootstrap/
      database/
      cache/
      logger/
      integrations/
    shared/
      errors/
      jwt/
      password/
      response/
      validator/
  migrations/
```

## 目录职责

### `internal/admin`

管理端专属边界，承载 `/admin-api/*` 的路由、中间件和模块实现。

- `routes/`：管理端路由注册
- `middleware/`：管理端鉴权、权限校验等中间件
- `modules/*`：按模块组织的管理端 handler、service、repository、dto、test

### `internal/web`

用户端公开 API 边界，当前承载 `GET /api/site-config`、用户账号自助接口、支付宝/微信侧用户实名接口和 `GET /api/server-catalog`。

- `routes/`：Web 公开路由注册
- `modules/site_config`：读取公开站点基础展示配置
- `modules/auth`：用户注册、登录、登录态恢复、退出、自动刷新和密码找回
- `modules/user_profile`：当前登录用户资料和密码编辑
- `modules/real_name`：当前登录用户提交个人实名资料、创建支付宝/微信侧实名会话、同步实名状态和查看实名状态
- `modules/product_catalog`：读取公开服务器产品目录，只展示产品、套餐、价格、销售地域和服务器系统模板，不创建订单或实例

### `internal/platform`

基础设施层。

- `bootstrap/`：应用启动与依赖装配
- `database/`：数据库初始化和通用持久化基础设施
- `cache/`：缓存客户端与基础封装
- `logger/`：日志初始化与封装
- `integrations/`：第三方系统适配

### `internal/domain`

跨端核心业务领域。

- `realname/`：实名申请同步、状态裁决和供应商结果收敛

### `internal/shared`

仅存放稳定、无明确业务语义、被多边界长期复用的基础能力。

允许进入 `shared` 的典型内容：

- `errors`
- `jwt`
- `password`
- `response`
- `validator`

## 模块组织原则

当前示例：

- 管理端：`auth`、`dashboard`、`admin_user`、`admin_role`、`system_config`

## 领域边界

当前后端领域边界以管理端基础能力为主，同时已重新开放用户账号自助、支付宝/微信侧用户实名和服务器产品目录：

- 管理端认证
- 会话
- RBAC
- 系统配置
- 审计写入
- 文件管理

用户账号自助、支付宝/微信侧用户实名和服务器产品目录已经重新纳入当前阶段契约。订单、支付、实例、工单、异步任务和其它用户端业务流仍从当前阶段契约中收口。后续如需恢复，必须先更新文档与迁移。

## 鉴权与权限

### 管理端

- 使用管理端 JWT secret 和 issuer
- JWT 必须带 `jti`
- `jti` 对应 `admin_sessions.session_id`
- 受保护管理端接口必须同时校验：
  - 签名
  - issuer
  - token type
  - 过期时间
  - 会话状态
  - 管理员状态
  - 当前数据库 RBAC

管理端前端可以消费权限快照改善体验，但后端 RBAC 仍是最终裁决。

## 当前管理端阶段边界

当前后端基础能力保留以下管理域：

- auth
- dashboard
- admin users
- roles and permissions
- admin sessions
- system configs
- audit logs

当前开放的管理端 API 和前端页面范围为：

- `Login`
- `Dashboard`
- `System Settings`
- `File Management`
- `Web User Management`
- `Real Name Management`
- `Product Management`
- `403`

`System Settings` 当前包含系统配置、管理员账号、管理员组权限和管理员会话。
`System Settings` 下同时开放操作日志页面，用于查看普通后台操作日志。
当前重新开放服务器产品目录和用户实名相关数据库契约；实名供应商集成只覆盖支付宝/微信侧实名核验，不开放支付交易。订单、支付、实例、工单、异步任务等业务域仍不属于当前阶段。

## 审计日志

- 普通后台操作写入 `admin_audit_logs`
- 普通操作日志用于还原后台操作历史，应包含操作者快照、管理端会话、请求 ID、请求方法和请求路径
- 普通操作日志的请求上下文由管理端中间件统一采集并写入 Gin 上下文，业务模块只传业务动作、对象、前后快照和备注

## 文件管理

- 文件上传属于管理端边界，只开放在 `/admin-api/*`
- 上传文件必须经过扩展名、声明 MIME、Magic Bytes 和危险扩展名校验
- 上传读取必须限制最大字节数，避免超大文件被完整读入内存
- 原始文件名只用于展示，存储文件名必须由服务端随机生成
- 数据库只保存相对存储路径，存储根目录来自 `storage.local_path` 配置
- 上传记录和上传审计必须同事务写入；事务失败时清理已写入的物理文件
- 删除文件使用软删除，状态变更和删除审计必须同事务写入

## 当前不在范围内的能力

以下能力不属于当前阶段契约：

- 用户端订单、支付、实例和工单 API
- Worker
- 异步任务
- PVE 集成
- 支付集成
- 订单、实例和用户端交易流程
