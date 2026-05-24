# 后台审计动作目录

本文档记录当前管理端写入 `admin_audit_logs.action` 的动作值和覆盖现状。它描述当前实现，不新增动作、接口、表结构或权限。

## 使用边界

- `action` 是后台操作审计的动作标识，用于日志管理筛选、排查和后续展示映射。
- `object_type` 标识被操作对象类型，`object_id` 保存对象主标识。
- 登录安全页面不使用独立表；它通过 `object_type=admin_auth` 查询认证相关动作。
- 后台审计写入应携带请求上下文，包含操作者快照、会话、请求 ID、请求方法、请求路径、IP 和 User-Agent。

## 动作目录

### 管理端认证

| action | object_type | 说明 |
| --- | --- | --- |
| `admin.login.success` | `admin_auth` | 管理员登录成功 |
| `admin.login.failed` | `admin_auth` | 管理员登录失败 |
| `admin.login.limited` | `admin_auth` | 登录失败次数过多触发限流 |
| `admin.captcha.limited` | `admin_auth` | 管理端登录验证码获取过于频繁 |
| `admin.logout` | `admin_auth` | 管理员主动退出登录 |
| `admin.refresh` | `admin_auth` | 管理端登录会话刷新 |

### 管理员与角色

| action | object_type | 说明 |
| --- | --- | --- |
| `admin.user.create` | `admin_user` | 创建管理员账号 |
| `admin.user.update` | `admin_user` | 更新管理员账号基础资料 |
| `admin.user.disable` | `admin_user` | 禁用管理员账号 |
| `admin.user.role_update` | `admin_user` | 更新管理员账号角色 |
| `admin.user.password_reset` | `admin_user` | 重置管理员账号密码 |
| `admin.role.create` | `admin_role` | 创建管理端角色 |
| `admin.role.update` | `admin_role` | 更新管理端角色基础资料 |
| `admin.role.disable` | `admin_role` | 禁用管理端角色 |
| `admin.role.permission_update` | `admin_role` | 更新管理端角色权限 |
| `admin.session.revoke` | `admin_session` | 后台吊销管理员会话 |

### 系统配置与文件

| action | object_type | 说明 |
| --- | --- | --- |
| `system.config.update` | `system_config` | 更新非敏感系统配置 |
| `system.config.secret_update` | `system_config` | 更新敏感系统配置 |
| `file.upload` | `file_attachment` | 上传管理端文件 |
| `file.delete` | `file_attachment` | 软删除管理端文件 |

### Web 用户管理

| action | object_type | 说明 |
| --- | --- | --- |
| `web.user.create` | `web_user` | 创建 Web 用户账号 |
| `web.user.update` | `web_user` | 更新 Web 用户账号资料或状态 |
| `web.user.password_reset` | `web_user` | 重置 Web 用户密码 |
| `web.user_session.revoke` | `web_user_session` | 吊销 Web 用户会话 |

### 实名管理

| action | object_type | 说明 |
| --- | --- | --- |
| `real_name.sync` | `user_real_name` | 同步实名供应商结果；同步失败也使用该动作写入失败备注 |
| `real_name.review` | `user_real_name` | 人工审核实名申请 |

### 产品目录

产品目录当前统一使用 `object_type=product_catalog`，通过 `action` 区分资源类型和操作。

| action | object_type | 说明 |
| --- | --- | --- |
| `product.create` | `product_catalog` | 创建产品 |
| `product.update` | `product_catalog` | 更新产品 |
| `product.status.update` | `product_catalog` | 更新产品状态 |
| `product.delete` | `product_catalog` | 删除产品 |
| `product_plan.create` | `product_catalog` | 创建服务器套餐 |
| `product_plan.update` | `product_catalog` | 更新服务器套餐 |
| `product_plan.status.update` | `product_catalog` | 更新服务器套餐状态 |
| `product_plan.delete` | `product_catalog` | 删除服务器套餐 |
| `product_plan.prices.update` | `product_catalog` | 更新套餐价格 |
| `product_plan.regions.update` | `product_catalog` | 更新套餐销售地域 |
| `product_plan.os_templates.update` | `product_catalog` | 更新套餐系统模板 |
| `product_plan.network_types.update` | `product_catalog` | 更新套餐网络类型 |
| `sales_region.create` | `product_catalog` | 创建销售地域 |
| `sales_region.update` | `product_catalog` | 更新销售地域 |
| `sales_region.delete` | `product_catalog` | 删除销售地域 |
| `server_os_template.create` | `product_catalog` | 创建服务器系统模板 |
| `server_os_template.update` | `product_catalog` | 更新服务器系统模板 |
| `server_os_template.delete` | `product_catalog` | 删除服务器系统模板 |
| `network_type.create` | `product_catalog` | 创建网络类型 |
| `network_type.update` | `product_catalog` | 更新网络类型 |
| `network_type.delete` | `product_catalog` | 删除网络类型 |

### 订单

| action | object_type | 说明 |
| --- | --- | --- |
| `order.admin_note.update` | `order` | 更新订单后台备注 |
| `order.cancel` | `order` | 管理端取消订单 |
| `order.close` | `order` | 管理端关闭订单 |
| `order.renewal.confirm` | `order` | 管理端人工确认续费订单 |

### 支付

| action | object_type | 说明 |
| --- | --- | --- |
| `payment.sync` | `payment` | 管理端主动同步支付渠道状态 |
| `payment.refund.create` | `refund` | 管理端发起全额退款 |
| `payment.refund.succeeded` | `refund` | 退款渠道成功并完成本地回滚 |
| `payment.refund.failed` | `refund` | 退款渠道失败或本地回滚失败 |
| `payment.provision.retry` | `payment` | 重试真实支付后自动交付失败的新购订单 |

### 钱包

钱包 v1 管理端只读，无管理端写接口，不新增后台调账审计动作。钱包充值回调、余额支付扣款和余额支付退款退回钱包必须写入钱包流水；供应商回调不写后台操作审计，但必须保存脱敏业务摘要和请求链路标识。

### 工单

| action | object_type | 说明 |
| --- | --- | --- |
| `ticket.reply` | `ticket` | 管理端回复工单 |
| `ticket.close` | `ticket` | 管理端关闭工单 |
| `ticket.assign` | `ticket` | 管理端指派或转派工单 |
| `ticket.collaborate` | `ticket` | 添加或移除工单协作者 |
| `ticket.note` | `ticket` | 追加工单内部备注 |
| `ticket.priority_upgrade` | `ticket` | 升级工单优先级 |
| `ticket.tags_replace` | `ticket` | 替换工单标签绑定 |
| `ticket.tag.create` | `ticket_tag` | 创建工单标签 |
| `ticket.tag.update` | `ticket_tag` | 更新工单标签 |

## 覆盖检查

当前管理端写接口的审计覆盖如下：

- 认证：登录成功、登录失败、登录限流、验证码限流、退出登录和会话刷新已写审计；普通获取验证码不写审计。
- 管理员、角色、会话：创建、更新、禁用、权限/角色变更、密码重置和会话吊销已写审计。
- 系统配置：配置更新已写审计，敏感配置使用独立动作。
- Web 用户：账号创建、更新、密码重置和会话吊销已写审计。
- 实名：供应商同步、同步失败备注和人工审核已写审计。
- 文件：上传和软删除已写审计；下载、详情、引用查询不写审计。
- 产品目录：产品、套餐、价格、关联、销售地域、系统模板和网络类型的写操作已写审计。
- 订单：后台备注、取消、关闭和人工续费确认已写审计。
- 支付：渠道状态同步、退款发起、退款成功/失败和自动交付失败重试必须写审计；支付供应商回调不写后台操作审计，但必须保存脱敏业务摘要和请求链路标识。
- 钱包：管理端 v1 只读；充值入账、余额支付扣款和余额支付退款退回钱包通过 `wallet_ledger_entries` 形成资金流水审计，不写普通后台操作审计。
- 工单：回复、关闭、指派/转派、协作者、内部备注、优先级升级、标签绑定和标签字典写操作已写审计。

当前未发现管理端受保护写接口缺少成功审计锚点。只读查询、文件下载、候选人查询、附件下载和普通验证码获取不写普通后台操作审计。

## 待确认项

以下不是当前代码缺口，但后续如要做展示、统计或更细排查，应先更新 owner docs 再实现：

- 是否为 `ticket.collaborate` 拆分 `ticket.collaborator.add` 与 `ticket.collaborator.remove`。
- 是否为 `ticket.assign` 拆分首次指派和转派动作。
- 是否为产品目录拆分更细的 `object_type`，例如 `product`、`product_plan`、`sales_region`、`server_os_template`、`network_type`。
- 是否为文件下载、审计详情查看、敏感配置查看等读操作增加审计；这会改变审计范围和数据量。
- 是否沉淀 action 的中文展示映射和前端模块筛选选项。
