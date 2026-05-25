# Invoice Management 页面契约

发票运营用于查看用户端发票申请，并执行人工电子普通发票的受理、驳回、开票登记和后台备注维护。页面只调用 `/admin-api/*`。

## 页面范围

- 发票申请列表
- 发票申请详情
- 用户摘要和订单明细
- 受理发票申请
- 驳回发票申请
- 登记发票号码、开票时间并绑定 PDF 文件
- 更新后台备注
- 下载或预览已开票 PDF

本页面不创建用户订单，不代用户提交发票申请，不接第三方开票平台，不支持专票、红冲、作废、发票邮件发送、钱包充值开票或开票后在线退款。

## 路由与权限

- 路由：`/invoices`
- 菜单权限：`page.invoices`
- 查看：`invoice:view` 或 `invoice:*`
- 受理和后台备注：`invoice:update` 或 `invoice:*`
- 登记开票：`invoice:issue` 或 `invoice:*`
- 驳回：`invoice:reject` 或 `invoice:*`

## 页面结构

发票运营包含列表、详情抽屉、受理/驳回/开票弹窗、PDF 绑定和权限按钮，按复杂管理页结构实现：

```text
admin/src/views/invoices/
  index.vue
  types.ts
  components/
```

## 行为约束

- 列表支持按状态、发票申请编号、订单编号、用户关键字、抬头关键字和创建时间范围筛选。
- 详情展示用户摘要、申请信息、抬头资料、订单明细、发票号码、文件摘要、状态时间和后台备注。
- `pending` 申请可受理或驳回；受理后进入 `processing`。
- `processing` 申请可登记开票或驳回；开票后进入 `issued`。
- `issued`、`rejected`、`cancelled` 不展示状态流转按钮。
- 开票登记必须选择 PDF 文件附件，前端只做 MIME 和文件摘要体验提示，后端仍最终校验文件类型、状态和引用关系。
- 低权限管理员不能看到或触发受理、驳回、开票和后台备注按钮；后端权限仍是最终裁决。
- 页面不得展示 PDF 物理路径、完整文件存储路径、完整税号进入控制台日志或错误提示。

## 关联接口

接口字段和响应结构以 `docs/server/api/endpoints.md` 为准。

- `GET /admin-api/invoices`
- `GET /admin-api/invoices/{invoice_no}`
- `POST /admin-api/invoices/{invoice_no}/accept`
- `POST /admin-api/invoices/{invoice_no}/reject`
- `POST /admin-api/invoices/{invoice_no}/issue`
- `PATCH /admin-api/invoices/{invoice_no}/admin-note`
- `GET /admin-api/invoices/{invoice_no}/download`
- `POST /admin-api/files/upload`

## 验收重点

- 侧栏菜单来自后端 `menus`，本地路由只作为页面组件白名单。
- 无权限访问 `/invoices` 时展示管理端 403 反馈。
- 状态流转按钮只在契约允许状态展示，非法状态以后端 `409xx` 为准。
- 发票 PDF 下载必须通过发票接口，不复用公开站点 Logo 或通用匿名文件入口。
- 发票运营不调用 `/api/*`，不复用用户端运行时代码。
