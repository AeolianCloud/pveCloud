# 外部系统集成

本目录原本用于描述外部系统接入的职责边界。

## 当前状态

当前用户账号自助和支付宝/微信侧个人实名阶段开放以下外部集成契约：

- SMTP：用于用户端密码找回。
- 支付宝开放平台身份认证：仅用于个人实名核验。
- 微信侧实名核验：当前按腾讯云人脸核身/实名核验能力接入，仅用于个人实名核验。

PVE、支付交易和其它外部业务集成仍不属于现阶段文档承诺范围。支付宝和微信在本文中的使用只表示实名核验供应商，不表示支付能力开放。

相关目录可以继续保留为未来实现预留。

## 供应商参考入口

- 支付宝开放平台身份认证初始化：`https://opendocs.alipay.com/apis/api_2/alipay.user.certify.open.initialize`
- 支付宝开放平台身份认证开始认证：`https://opendocs.alipay.com/apis/api_2/alipay.user.certify.open.certify`
- 支付宝开放平台身份认证记录查询：`https://opendocs.alipay.com/apis/api_2/alipay.user.certify.open.query`
- 腾讯云人脸核身实名核身鉴权 DetectAuth：`https://cloud.tencent.com/document/product/1007/31816`
- 腾讯云人脸核身结果查询 GetDetectInfoEnhanced：`https://cloud.tencent.com/document/product/1007/41957`

## 其它集成重新开放前的前置条件

- 先恢复对应数据库迁移和配置契约
- 先更新 `docs/server/architecture.md`
- 先更新 `docs/server/database/design.md`
- 先更新相关 API 文档和验收口径
