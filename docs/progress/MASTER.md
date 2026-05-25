# 项目阶段进度总览

## 已完成阶段

基础后台阶段已完成一个可登录、可授权、可追踪、可配置的后台底座。

## 相关文档

- 缺口分析：`docs/analysis/basic-admin-gap.md`
- 实施计划：`docs/plan/basic-admin-foundation.md`
- Web 基础前台计划：`docs/plan/web-foundation.md`
- 管理端架构：`docs/admin/architecture.md`
- 用户端架构：`docs/web/architecture.md`
- API 契约：`docs/server/api/`
- 数据库设计：`docs/server/database/design.md`

## 基础后台分阶段进度

基础后台历史阶段记录已归档到 `docs/progress/archive/`：

- [x] Phase 0：安全日志基础
- [x] Phase 1：审计写入与脱敏收口
- [x] Phase 2：管理员与角色权限
- [x] Phase 3：登录会话与系统设置
- [x] Phase 4：Dashboard 与管理端体验收口
- [x] Phase 5：验收与发布准备

归档阶段记录只用于恢复历史背景；当前契约以 `docs/admin/`、`docs/server/`、`docs/server/api/`、`server/migrations/` 和 `server/config.example.yaml` 为准。

## 当前状态

基础后台阶段的后端能力已覆盖认证、会话、RBAC、Dashboard、管理员、角色、系统配置、审计写入和文件管理等管理域。
当前管理端前端开放页面以 `docs/admin/pages/README.md` 为准，路由和权限以 `docs/admin/routing-permissions.md` 为准。

`System Settings` 当前承载系统配置和管理员设置能力；管理员设置内包含管理员账号、管理员组权限和管理员会话管理。
当前数据库契约已重新开放用户端账号会话、用户实名、服务器产品目录、订单、真实支付一期、钱包 v1、续费订单、实例交付、通用异步任务、通知和工单；发票等业务域仍不在当前阶段。
操作日志查询已按系统设置子页面口径开放；审计日志内部写入能力仍保留。

## 当前推进阶段

MCP PVE 实例交付阶段已进入实现复盘状态；实例生命周期、续费订单和 Worker 异步任务已进入实现收尾状态；真实支付闭环一期和钱包 v1 已完成代码提交，待真实渠道端到端验收与生产告警配置确认。

当前已开放管理端交付映射、MCP 只读资源、订单触发交付、实例列表/详情/开机/关机/释放/同步；已开放用户端实例列表、详情、启动和停止。真实支付闭环一期已开放用户端支付发起/查询/回调、管理端支付管理、退款、支付成功自动交付和续费支付生效记录。钱包 v1 已开放用户充值、余额支付、余额支付退款退回钱包和管理端只读钱包管理。工单关联实例排障入口已开放用户端工单关联当前用户自己的业务实例编号、管理端工单按实例编号筛选和从工单跳转实例管理定位排障；不开放新增 PVE 运维能力。
真实支付闭环一期上线前，需按运维文档完成支付宝沙箱、微信支付沙箱或小额真实商户号端到端验签、回调和退款闭环验证，并纳入支付创建失败、回调验签失败、退款 `pending`/`failed` 的告警口径。

2026-05-24 本地技术验证记录：后端在 Docker MariaDB 测试库下执行 `PVECLOUD_TEST_MYSQL_DSN=... GOCACHE=/tmp/pvecloud-go-build go test -count=1 ./...` 通过；后端入口 `go build ./cmd/api ./cmd/worker ./cmd/setup-admin` 通过；`admin` 与 `web` 均执行 `bun run build` 通过。该记录只代表代码层、构建层和自动化测试层通过，不替代支付宝沙箱、微信支付沙箱或小额真实商户号的端到端验签、回调和退款闭环验收；生产日志采集系统仍需按 `payment_alert` 事件配置外部告警规则。

2026-05-25 本地技术验证记录：工单关联实例排障入口完成代码层验证；后端执行 `GOCACHE=/tmp/pvecloud-go-build go test ./internal/usecase/web/ticket -v`、`PVECLOUD_TEST_MYSQL_DSN=... GOCACHE=/tmp/pvecloud-go-build go test ./...` 和 `GOCACHE=/tmp/pvecloud-go-build go build ./cmd/api ./cmd/worker ./cmd/setup-admin` 通过；`admin` 与 `web` 均执行 `bash -ic 'bun run build'` 通过。

2026-05-25 支付告警稳定化记录：`payment_alert` recorder 已收敛为只记录 owner docs 允许的四类事件，避免误写未知事件导致外部监控规则无法识别；后端执行 `PVECLOUD_TEST_MYSQL_DSN=... GOCACHE=/tmp/pvecloud-go-build go test ./internal/usecase/paymentalert ./internal/usecase/web/payment ./internal/usecase/admin/payment -v`、`PVECLOUD_TEST_MYSQL_DSN=... GOCACHE=/tmp/pvecloud-go-build go test ./...` 和 `GOCACHE=/tmp/pvecloud-go-build go build ./cmd/api ./cmd/worker ./cmd/setup-admin` 通过。

本阶段交接记录见 `docs/progress/mcp-pve-instance-handoff.md`。
Worker、实例生命周期和续费阶段复盘记录见 `docs/progress/worker-instance-lifecycle-code-start.md`。
当前暂无已确认的新功能代码开工入口；下一阶段需维护者先确认功能范围，并按文档先行流程更新对应 owner docs 或机器契约。

当前实例、支付和钱包阶段仍不包含：

- 发票、JSAPI/openid、小程序支付、部分退款、提现、人工调账、余额转账和自动对账批处理
- 重启、重装、重置密码、控制台、快照、备份、迁移、监控和防火墙
- PVE 资源池、库存扣减和通用 PVE 运维管理

工单关联实例排障入口已完成；后续如需继续增强工单或实例排障，只允许在确认 owner docs 后扩展，不得借工单页面开放新增 PVE 运维能力。

## 下一步维护原则

- 再次修改当前后台范围时，优先更新 `docs/admin/architecture.md`
- 修改 Web 前端页面范围、路由、请求封装或状态语义时，优先更新 `docs/web/architecture.md`
- 再次修改阶段边界时，优先更新本文档和计划/分析文档
- 再次开放真实支付网关、实例新增 MCP 能力或其它管理端页面前，必须先完成文档确认

## 服务器产品目录阶段已完成

- 新增服务器产品目录数据库契约和迁移：产品、套餐、周期价格、销售地域、服务器系统模板及关联表
- 新增管理端产品管理菜单、权限、API 和页面
- 管理端产品管理支持产品、套餐、价格、销售地域、服务器系统模板和套餐关联维护
- 管理端套餐列表提供公开检查，提示缺产品公开、套餐公开、启用价格、销售地域或系统模板时 Web 不展示
- 新增 Web 公开接口 `GET /api/server-catalog`
- Web 页面读取公开产品目录并展示产品、套餐、价格、销售地域、服务器系统模板、简介和状态
- 订单 MVP 只表示购买意向和后台处理入口；实例交付能力以后续实例阶段契约为准，Web 不展示支付、PVE 节点、资源池、库存扣减或自动开通承诺

## 服务器产品目录阶段验收记录

- `GET /api/server-catalog` 已可返回公开服务器产品目录
- `admin` 构建已通过
- `web` 构建已通过
- 后端产品目录相关包测试已通过
- 完整 `go test ./...` 曾因环境下载 `github.com/stretchr/testify` 超时失败，非产品目录包编译错误

## 后续阶段入口

后续若要开放发票、JSAPI/openid、小程序支付、部分退款、提现、人工调账、余额转账、自动对账批处理、实例更多 MCP 运维能力等真实用户端业务，必须先补齐并确认：

- 相关用户端 `/api/*` 和管理端 `/admin-api/*` 接口契约
- 新增能力所需数据库契约、迁移和配置示例
- 支付、退款、钱包、发票或实例新增能力的业务流程说明
- 对应管理端和用户端页面范围
- 用户端路由、权限和请求包装口径
- 必要的管理端运营页面、权限口径和运维告警口径
