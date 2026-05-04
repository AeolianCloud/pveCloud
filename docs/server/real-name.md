# 用户实名设计

本文档定义用户实名能力的当前业务口径。接口字段以 `docs/server/api/endpoints.md` 为准，表结构以 `server/migrations/` 为准。

## 功能范围

实名能力用于记录用户提交的个人实名资料，并由后台审核。

当前阶段只开放个人实名，不开放企业实名、活体检测、OCR 自动识别、银行卡四要素、公安联网核验或第三方实名服务。

## 后台配置

实名功能的业务开关和审核要求全部来自后台 `system_configs`，不新增 `server/config.yaml` 运行配置项。

后台系统配置按中文分组“实名设置”展示，至少包含：

- `real_name.enabled`：是否开放用户端实名入口，默认 `false`。
- `real_name.required_for_order`：购买机器前是否要求实名通过，默认 `true`。
- `real_name.manual_review_enabled`：是否启用后台人工审核，默认 `true`；当前阶段固定依赖人工审核。
- `real_name.resubmit_enabled`：审核拒绝后是否允许用户重新提交，默认 `true`。
- `real_name.max_submit_attempts`：同一用户最大提交次数，默认 `3`。
- `real_name.id_card_front_required`：是否要求身份证人像面图片，默认 `true`。
- `real_name.id_card_back_required`：是否要求身份证国徽面图片，默认 `true`。
- `real_name.hold_card_required`：是否要求手持证件图片，默认 `false`。
- `real_name.image_max_size_mb`：单张实名图片最大尺寸 MB，默认 `5`。
- `real_name.allowed_image_types`：允许的图片 MIME 类型，默认 `image/jpeg,image/png,image/webp`。
- `real_name.review_notice`：用户端实名说明文案，默认空。

配置约束：

- 布尔配置使用 `value_type=bool`，保存字符串 `true` 或 `false`。
- 数值配置使用 `value_type=int`，服务端按正整数校验。
- MIME 列表使用 `value_type=string`，以英文逗号分隔。
- 这些配置不属于敏感配置，`is_secret=0`。
- 图片上传仍复用全局文件安全校验；实名配置只能收紧实名场景的文件类型和大小，不能放宽全局上传安全边界。

## 实名状态

用户当前实名状态由最新有效实名申请决定：

- `unverified`：未提交或无有效实名记录。
- `pending`：已提交，等待后台审核。
- `approved`：审核通过。
- `rejected`：审核拒绝。

实名申请状态使用字符串字段，不使用数据库 enum。

状态流转：

- 用户提交后状态为 `pending`。
- 后台审核通过后状态为 `approved`。
- 后台拒绝后状态为 `rejected`，必须填写拒绝原因。
- `pending` 状态不允许用户重复提交。
- `approved` 状态不允许用户端自行覆盖；如需重置必须由后台后续另行开放契约。
- `rejected` 是否可重新提交由 `real_name.resubmit_enabled` 和 `real_name.max_submit_attempts` 决定。

## 资料与附件

实名资料包括：

- 真实姓名
- 证件类型，当前仅 `id_card`
- 证件号码
- 证件图片附件：人像面、国徽面、可选手持证件

存储规则：

- 证件号码不保存明文；数据库只保存查询摘要和脱敏展示值，接口只返回脱敏号码。
- 证件图片使用文件管理能力落盘，并通过 `file_attachment_references` 记录引用关系。
- 普通用户端接口不得返回实名图片下载地址；后台详情可在权限控制下查看必要附件。

## 审核与审计

- 后台实名审核必须要求管理端登录态。
- 审核通过或拒绝必须写入 `admin_audit_logs`。
- 审核操作需要记录审核管理员、审核时间、审核结果和拒绝原因。
- 用户提交实名不写入管理端审计日志，但应保存申请创建时间和更新时间。

## 购买拦截

- 用户浏览产品、套餐和价格不要求实名。
- 用户购买机器或创建机器订单前，服务端必须读取 `real_name.required_for_order`。
- 当 `real_name.required_for_order=true` 时，只有实名状态为 `approved` 的用户允许继续购买。
- `unverified`、`pending`、`rejected` 状态都必须阻止购买，并返回明确错误，引导用户进入 `/user/real-name`。
- 该校验必须在服务端购买/下单接口执行，前端提示不能替代服务端裁决。

## 当前不开放

- 企业实名
- 自动 OCR 或第三方核验
- 实名资料后台直接代填
- 用户端删除实名记录
- 已通过实名的用户端自行修改实名资料
