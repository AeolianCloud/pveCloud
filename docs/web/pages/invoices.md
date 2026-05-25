# Invoices 页面契约

`Invoices` 是用户端发票页面集合，对应 `/user/invoices`、`/user/invoices/new` 和 `/user/invoices/:invoiceNo`。

## 页面范围

- 发票申请列表
- 可开票订单选择
- 多订单合并申请电子普通发票
- 发票申请详情
- 取消待处理申请
- 下载已开票 PDF

本页面不支持增值税专用发票、红冲、作废、第三方开票平台、发票邮件发送、钱包充值开票、部分退款或开票后在线退款。

## 状态与展示

- 发票申请状态：`pending`、`processing`、`issued`、`rejected`、`cancelled`
- 发票类型：`electronic_normal`
- 抬头类型：`personal`、`company`
- 金额使用分为单位，前端只格式化展示，不作为最终金额事实。

列表展示：

- 发票申请编号
- 状态
- 抬头类型和抬头
- 申请金额和币种
- 订单数量
- 发票号码
- 开票时间
- 创建时间
- 可用操作

详情展示：

- 发票申请基础信息
- 抬头资料
- 订单明细
- 状态时间
- 驳回原因
- 发票号码和开票时间
- PDF 下载入口

## 行为约束

- 页面必须是受保护路由，未登录访问跳转 `/login`。
- 页面只展示当前登录用户自己的发票申请和可开票订单。
- 创建申请必须提交 `order_nos`、`title_type`、`title`、`tax_no`、`email`、`remark` 和 `client_token`。
- 企业抬头必须填写税号；个人抬头税号可为空。
- 可开票订单列表必须来自 `GET /api/invoice-eligible-orders`；前端不得自行裁决开票资格。
- 同一 `client_token` 重复提交时展示后端返回的已有申请。
- 用户只能取消自己的 `pending` 申请。
- `issued` 申请展示 PDF 下载入口；其它状态不展示下载主按钮。
- 页面不得展示后台备注、管理员信息、PDF 物理路径、完整文件存储路径或完整税号进入控制台日志和错误提示。

## 关联接口

接口字段和响应结构以 `docs/server/api/invoices.md` 为准。

- `GET /api/invoice-eligible-orders`
- `POST /api/invoices`
- `GET /api/invoices`
- `GET /api/invoices/{invoice_no}`
- `POST /api/invoices/{invoice_no}/cancel`
- `GET /api/invoices/{invoice_no}/download`

## 验收重点

- 未登录访问发票页面会进入 `/login`。
- 跨用户发票申请、订单和 PDF 不可访问。
- 多订单合并申请金额以后端返回为准。
- 待处理申请可取消，处理中、已开票、已驳回和已取消申请不可取消。
- PDF 下载只走发票下载接口。
- 页面响应式适配移动端和桌面端。
