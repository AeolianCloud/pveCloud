# 发票 API

本文档维护用户端发票申请、查询、取消、下载和管理端发票运营接口。跨接口通用约定见 `docs/server/api/conventions.md`，发票业务流程见 `docs/server/invoices.md`。

## 发票

发票 v1 只开放已支付订单的人工电子普通发票申请。用户端可以把多笔自己的可开票订单合并提交一张发票申请；管理端线下开票后登记发票号码、开票时间并绑定 PDF 文件。v1 不开放增值税专用发票、红冲、作废、第三方开票平台、发票邮件发送、钱包充值开票、部分退款或开票后在线退款。

发票申请状态允许：`pending`、`processing`、`issued`、`rejected`、`cancelled`。`pending`、`processing` 和 `issued` 会占用订单开票资格；`rejected` 和 `cancelled` 不继续占用。订单存在有效发票申请时，支付退款接口必须返回 `409xx`。

### `GET /api/invoice-eligible-orders`

- 鉴权：用户端 Bearer Token
- 作用：分页查询当前用户可申请发票的订单
- 查询参数支持：`page`、`per_page`、`keyword`、`date_from`、`date_to`
- 列表项包含订单编号、订单类型、订单金额、币种、支付状态、支付完成时间、产品和套餐摘要、关联实例编号、是否已被有效发票申请占用
- 约束：
  - 只返回当前用户自己的订单
  - 仅返回 `payment_status=paid` 或 `manual_confirmed`、未退款且未被有效发票申请占用的订单
  - 钱包充值不出现在可开票订单列表

### `POST /api/invoices`

- 鉴权：用户端 Bearer Token
- 作用：创建发票申请
- 请求字段：`order_nos`、`title_type`、`title`、`tax_no`、`email`、`remark`、`client_token`
- `order_nos` 必填，至少 1 个订单编号，全部订单必须属于当前用户且满足可开票条件
- `title_type` 允许：`personal`、`company`
- `title` 必填，最多 100 字
- `tax_no` 在 `title_type=company` 时必填，最多 64 字；个人抬头可为空
- `email` 可选，最多 128 字；v1 仅保存展示，不自动发送邮件
- `remark` 可选，最多 500 字
- `client_token` 必填，用于当前用户创建申请幂等
- 成功数据包含发票申请详情
- 约束：
  - 同一用户同一 `client_token` 重复提交返回已有申请，不重复占用订单
  - 服务端必须同事务锁定订单、校验开票资格、写申请和写明细
  - 同一订单同时最多被一条 `pending`、`processing` 或 `issued` 发票申请占用；重复占用返回 `409xx`
  - 申请金额由服务端按订单事实汇总，不信任前端提交金额

### `GET /api/invoices`

- 鉴权：用户端 Bearer Token
- 作用：分页查询当前用户自己的发票申请
- 查询参数支持：`page`、`per_page`、`status`、`date_from`、`date_to`
- 列表项包含发票申请编号、状态、抬头类型、抬头、金额、币种、订单数量、发票号码、开票时间、创建时间和可用操作

### `GET /api/invoices/{invoice_no}`

- 鉴权：用户端 Bearer Token
- 作用：查看当前用户自己的发票申请详情
- 成功数据包含申请基础信息、抬头资料、订单明细、状态时间、驳回原因、发票号码、开票时间、PDF 下载可用状态
- 约束：只允许查看当前用户自己的发票申请；不得返回后台备注、管理员信息、文件物理路径或内部自增 ID

### `POST /api/invoices/{invoice_no}/cancel`

- 鉴权：用户端 Bearer Token
- 作用：取消当前用户自己的发票申请
- 请求字段：`reason` 可选，最多 500 字
- 约束：
  - 仅 `pending` 申请可由用户取消
  - 成功取消后释放订单开票占用
  - 非法状态返回 `409xx`

### `GET /api/invoices/{invoice_no}/download`

- 鉴权：用户端 Bearer Token
- 作用：下载当前用户自己的已开票 PDF
- 约束：
  - 仅 `issued` 申请可下载
  - 必须校验发票归属和文件引用关系
  - 下载响应不得暴露物理存储路径
  - 文件名进入响应头前必须安全编码
  - 受保护下载响应不得被共享缓存长期保存

## 管理端发票运营

发票运营用于查看发票申请、受理、驳回、登记开票和维护后台备注。管理端不创建用户订单，不代用户提交申请，不接第三方开票平台。

### `GET /admin-api/invoices`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.invoices`
- 作用：分页查询发票申请
- 查询参数支持：`page`、`per_page`、`status`、`invoice_no`、`order_no`、`user_keyword`、`title_keyword`、`date_from`、`date_to`
- 列表项包含发票申请编号、状态、用户摘要、抬头类型、抬头、金额、币种、订单数量、发票号码、申请时间、受理时间和开票时间
- 约束：不得返回 PDF 物理路径、完整文件存储路径或内部锁信息

### `GET /admin-api/invoices/{invoice_no}`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.invoices`
- 作用：查看发票申请详情
- 成功数据包含用户摘要、申请基础信息、抬头资料、订单明细、文件摘要、后台备注和状态时间
- 约束：敏感字段按安全文档脱敏展示；不得返回文件物理路径

### `POST /admin-api/invoices/{invoice_no}/accept`

- 鉴权：管理端 Bearer Token
- 操作权限：`invoice:update` 或 `invoice:*`
- 作用：受理发票申请
- 约束：仅允许 `pending -> processing`；非法状态返回 `409xx`
- 审计：`invoice.accept`

### `POST /admin-api/invoices/{invoice_no}/reject`

- 鉴权：管理端 Bearer Token
- 操作权限：`invoice:reject` 或 `invoice:*`
- 作用：驳回发票申请
- 请求字段：`reason` 必填，最多 500 字
- 约束：允许 `pending -> rejected` 或 `processing -> rejected`；成功驳回后释放订单开票占用；非法状态返回 `409xx`
- 审计：`invoice.reject`

### `POST /admin-api/invoices/{invoice_no}/issue`

- 鉴权：管理端 Bearer Token
- 操作权限：`invoice:issue` 或 `invoice:*`
- 作用：登记已开票结果
- 请求字段：`invoice_code` 可选、`invoice_number` 必填、`issued_at` 必填、`file_id` 必填
- 约束：
  - 仅允许 `processing -> issued`
  - `file_id` 必须指向已存在、未删除且 MIME 类型为 PDF 的文件附件
  - 成功开票必须同事务更新申请、绑定文件引用和写后台审计
  - 已开票申请继续占用订单开票资格
- 审计：`invoice.issue`

### `PATCH /admin-api/invoices/{invoice_no}/admin-note`

- 鉴权：管理端 Bearer Token
- 操作权限：`invoice:update` 或 `invoice:*`
- 作用：更新发票申请后台备注
- 请求字段：`admin_note` 可为空，最多 1000 字
- 约束：后台备注不返回用户端
- 审计：`invoice.admin_note.update`

### `GET /admin-api/invoices/{invoice_no}/download`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.invoices`
- 作用：下载或预览发票 PDF
- 约束：
  - 必须校验申请状态、文件引用关系和管理端权限
  - 下载响应不得暴露物理存储路径
  - 文件名进入响应头前必须安全编码
  - 受保护下载响应不得被共享缓存长期保存
