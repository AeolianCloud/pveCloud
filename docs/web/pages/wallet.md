# Wallet 页面契约

`Wallet` 是用户端钱包页，对应路由 `/user/wallet`。

## 页面范围

- 展示当前用户钱包余额和钱包状态
- 展示钱包流水
- 创建钱包充值
- 查询充值状态
- 引导用户在订单支付时选择钱包余额支付

本页面不支持提现、余额转账、人工调账、退款入钱包、JSAPI/openid 或小程序支付。真实支付订单退款仍原路退回支付宝/微信；钱包余额支付订单退款退回钱包余额。钱包充值不开具发票；钱包余额支付形成的已支付订单可按订单申请发票。

## 状态与展示

- 钱包状态：`active`、`disabled`
- 钱包流水方向：`credit`、`debit`
- 钱包流水类型：`recharge`、`payment`、`refund`
- 钱包充值状态：`pending`、`paid`、`closed`、`failed`
- 金额使用分为单位，前端只格式化展示，不作为最终金额事实。

## 行为约束

- 页面必须是受保护路由，未登录访问跳转 `/login`。
- 页面只展示当前登录用户自己的钱包、充值和流水。
- 创建充值必须提交 `provider`、`method`、`amount_cents` 和 `client_token`；重复提交同一 `client_token` 展示已有充值结果。
- `pending` 充值可以按固定间隔轮询；进入 `paid`、`closed` 或 `failed` 后停止轮询。
- 钱包余额支付入口出现在订单支付流程中；余额不足时展示后端返回的失败提示，不在前端自行扣减余额。
- 页面不得展示完整回调 payload、签名串、商户密钥、内部任务 ID 或完整上游响应。

## 关联接口

接口字段和响应结构以 `docs/server/api/orders-payments-wallet.md` 为准。

- `GET /api/wallet`
- `GET /api/wallet/ledger`
- `POST /api/wallet/recharges`
- `GET /api/wallet/recharges/{recharge_no}`
- `POST /api/orders/{order_no}/payments`

## 验收重点

- 未登录访问 `/user/wallet` 会进入 `/login`。
- 钱包余额、充值状态和流水只来自后端接口，不使用前端缓存作为最终事实。
- 充值成功后余额和流水刷新正确，重复回调或重复轮询不会造成前端重复入账假象。
- 订单余额支付成功后跳转订单或支付结果页，余额不足时展示明确失败提示。
- 页面响应式适配移动端和桌面端。
