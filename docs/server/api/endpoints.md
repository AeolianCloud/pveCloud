# API 接口总览

本文档是后端 API 领域文档导航和实现边界提示。跨接口通用约定见 `docs/server/api/conventions.md`，目录索引见 `docs/server/api/README.md`。

## 实现边界提示

接口契约按访问边界区分：

- `/admin-api/*`：由 `server/internal/delivery/http/admin/*` 聚合路由，业务编排落在 `server/internal/usecase/admin/*`，GORM model 和可复用查询对象落在 `server/internal/repository/mysql/*`
- `/api/*`：由 `server/internal/delivery/http/web/*` 聚合路由，业务编排落在 `server/internal/usecase/web/*`，GORM model 和可复用查询对象落在 `server/internal/repository/mysql/*`

这里描述的是 API 契约，不直接替代具体代码结构；但当接口重新开放、迁移或新增时，路由注册、权限校验和实现目录应与上述边界保持一致。

## 领域接口文档

- `docs/server/api/platform.md`：系统检查、管理端认证会话、Dashboard、管理员、角色权限、管理员会话、系统配置。
- `docs/server/api/web-auth-realname.md`：用户端公开配置、认证、账号资料、密码、实名和实名供应商回调。
- `docs/server/api/admin-users-logs-files.md`：Web 用户管理、实名管理、审计日志、日志中心、文件管理。
- `docs/server/api/product-catalog.md`：产品、套餐、价格、销售地域、系统模板、网络类型、用户端公开产品目录。
- `docs/server/api/orders-payments-wallet.md`：用户端订单、支付、钱包，管理端钱包、支付和退款运营。
- `docs/server/api/invoices.md`：用户端发票和管理端发票运营。
- `docs/server/api/instances-tasks.md`：实例交付、管理端实例、异步任务、用户端实例和生命周期。
- `docs/server/api/tickets.md`：用户端工单和管理端工单运营。
- `docs/server/api/out-of-scope.md`：暂未开放或当前不在契约内的接口域。

## 维护规则

- 具体接口字段、请求参数、响应结构、权限和接口级约束必须写入对应领域文件。
- 本文件只维护导航、实现边界和拆分规则，不承载具体接口字段。
- 新领域文件创建后，同步更新 `docs/server/api/README.md` 和本文件。
