# Real Name Management 页面契约

## 页面定位

`Real Name Management` 用于后台查看和审核用户端个人实名申请。

当前只承载个人实名审核，不承载企业实名、第三方核验配置、订单拦截记录或证件库管理。

## 路由结构

页面：

- 实名管理：`/web/real-names`

`实名管理` 作为管理端侧栏中的独立菜单项展示，不挂在 `System Settings` 下。

## 页面职责

- 分页展示用户实名申请。
- 支持按用户关键词、实名状态、证件类型和提交时间筛选。
- 查看实名申请详情，展示用户摘要、真实姓名、脱敏证件号码、证件类型、提交时间、审核状态、审核结果和拒绝原因。
- 对 `pending` 状态申请执行审核通过或拒绝。
- 拒绝申请时必须填写拒绝原因。
- 不支持后台代用户新增实名申请。
- 不支持后台直接修改用户实名姓名或证件号码。
- 不支持删除实名申请。

## 权限建议

- 页面入口：`page.real-name-management`
- 列表和详情资源：`real-name:view` 或 `real-name:*`
- 审核操作：`real-name:review` 或 `real-name:*`

## 关联接口

- `GET /admin-api/real-name-applications`
- `GET /admin-api/real-name-applications/{id}`
- `POST /admin-api/real-name-applications/{id}/review`

具体字段、响应和错误码以 `docs/server/api/` 为准。

## 验收重点

- 管理端页面只调用 `/admin-api/*`。
- 侧栏入口由服务端 `menus` 下发。
- 页面按钮显隐使用统一权限能力判断。
- 只有 `pending` 申请允许审核。
- 拒绝审核必须填写拒绝原因。
- 证件号码在列表和详情中默认脱敏展示。
