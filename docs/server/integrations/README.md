# 外部系统集成

本目录原本用于描述外部系统接入的职责边界。

## 当前状态

当前用户账号自助、支付宝/微信侧个人实名、真实支付一期、发票 v1 和实例交付阶段开放以下外部集成契约：

- SMTP：用于用户端密码找回。
- SMTP：用于实例到期邮件提醒。
- 支付宝开放平台身份认证：仅用于个人实名核验。
- 微信侧实名核验：当前按腾讯云人脸核身/实名核验能力接入，仅用于个人实名核验。
- 支付宝支付：仅用于电脑网页支付、手机网页支付、支付回调、退款和主动查询。
- 微信支付：仅用于 Native 扫码、H5 支付、支付回调、退款和主动查询。
- MCP PVE client API：仅作为后端内部上游，用于查询 PVE 节点/存储、创建 VM、查询 VM、启动 VM、停止 VM、删除 VM 和查询异步操作。
- 发票 v1 不接第三方开票平台；运营在线下开票后在管理端登记发票号码并绑定 PDF。

JSAPI/openid、小程序支付、专票、红冲、第三方开票平台、部分退款、提现、人工调账、余额转账、自动对账批处理和其它外部业务集成仍不属于现阶段文档承诺范围。钱包 v1 只复用现有支付宝/微信支付适配做充值，不新增外部钱包供应商。MCP PVE client API 不作为 pveCloud 对外接口暴露，用户端和管理端只能通过订单、支付、钱包、发票、实例和实例操作业务接口间接使用。

相关目录可以继续保留为未来实现预留。

## 供应商参考入口

- 支付宝开放平台身份认证初始化：`https://opendocs.alipay.com/apis/api_2/alipay.user.certify.open.initialize`
- 支付宝开放平台身份认证开始认证：`https://opendocs.alipay.com/apis/api_2/alipay.user.certify.open.certify`
- 支付宝开放平台身份认证记录查询：`https://opendocs.alipay.com/apis/api_2/alipay.user.certify.open.query`
- 支付宝电脑网站支付：`alipay.trade.page.pay`
- 支付宝手机网站支付：`alipay.trade.wap.pay`
- 支付宝统一收单交易查询：`alipay.trade.query`
- 支付宝统一收单交易退款：`alipay.trade.refund`
- 腾讯云人脸核身实名核身鉴权 DetectAuth：`https://cloud.tencent.com/document/product/1007/31816`
- 腾讯云人脸核身结果查询 GetDetectInfoEnhanced：`https://cloud.tencent.com/document/product/1007/41957`
- 微信支付 Native 下单、H5 下单、支付通知、退款和查询：以微信支付 API v3 官方文档为准。
- MCP PVE client API：以当前 MCP OpenAPI 文档为准，服务端只接入其已提供的 `/api/pve/*` 能力。

## 支付 SDK 选择

- 微信支付适配优先使用微信支付 API v3 官方 Go SDK：`github.com/wechatpay-apiv3/wechatpay-go`。服务端不得自行拼接平台签名、通知解密、退款和查询协议。
- 支付宝国内开放平台支付适配优先使用成熟社区库 `github.com/smartwalle/alipay/v3` 对接 `alipay.trade.page.pay`、`alipay.trade.wap.pay`、`alipay.trade.query`、`alipay.trade.refund` 和通知验签。若后续支付宝提供更匹配国内开放平台交易 API 的官方 Go SDK，可在文档确认后替换。
- 测试环境可以使用仓库内 mock adapter 覆盖成功、失败、验签失败、查询不可用和退款不可确认分支；生产路径不得使用 mock adapter。

## 其它集成重新开放前的前置条件

- 先恢复对应数据库迁移和配置契约
- 先更新 `docs/server/architecture.md`
- 先更新 `docs/server/database/design.md`
- 先更新相关 API 文档和验收口径
