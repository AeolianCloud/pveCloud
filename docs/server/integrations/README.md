# 外部系统集成

本目录原本用于描述外部系统接入的职责边界。

## 当前状态

当前用户账号自助、支付宝/微信侧个人实名和实例交付阶段开放以下外部集成契约：

- SMTP：用于用户端密码找回。
- SMTP：用于实例到期邮件提醒。
- 支付宝开放平台身份认证：仅用于个人实名核验。
- 微信侧实名核验：当前按腾讯云人脸核身/实名核验能力接入，仅用于个人实名核验。
- MCP PVE client API：仅作为后端内部上游，用于查询 PVE 节点/存储、创建 VM、查询 VM、启动 VM、停止 VM、删除 VM 和查询异步操作。

真实支付交易和其它外部业务集成仍不属于现阶段文档承诺范围。支付宝和微信在本文中的使用只表示实名核验供应商，不表示支付能力开放。MCP PVE client API 不作为 pveCloud 对外接口暴露，用户端和管理端只能通过订单、实例和实例操作业务接口间接使用。

相关目录可以继续保留为未来实现预留。

## 供应商参考入口

- 支付宝开放平台身份认证初始化：`https://opendocs.alipay.com/apis/api_2/alipay.user.certify.open.initialize`
- 支付宝开放平台身份认证开始认证：`https://opendocs.alipay.com/apis/api_2/alipay.user.certify.open.certify`
- 支付宝开放平台身份认证记录查询：`https://opendocs.alipay.com/apis/api_2/alipay.user.certify.open.query`
- 腾讯云人脸核身实名核身鉴权 DetectAuth：`https://cloud.tencent.com/document/product/1007/31816`
- 腾讯云人脸核身结果查询 GetDetectInfoEnhanced：`https://cloud.tencent.com/document/product/1007/41957`
- MCP PVE client API：以当前 MCP OpenAPI 文档为准，服务端只接入其已提供的 `/api/pve/*` 能力。

## 其它集成重新开放前的前置条件

- 先恢复对应数据库迁移和配置契约
- 先更新 `docs/server/architecture.md`
- 先更新 `docs/server/database/design.md`
- 先更新相关 API 文档和验收口径
