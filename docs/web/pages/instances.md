# Instances 页面契约

`Instances` 是用户端实例页面集合，用于当前登录用户查看自己订单交付后的云主机并执行基础电源操作。

对应路由：

- `/user/instances` - 实例列表
- `/user/instances/:instanceNo` - 实例详情

## 行为范围

- 所有实例页面都是受保护路由，未登录访问时跳转 `/login` 并携带站内 `redirect`。
- 页面进入前必须完成登录态恢复；恢复失败时清理本地登录态并回到 `/login`。
- 页面只展示当前登录用户自己的实例。
- 用户可以启动自己的 `stopped` 实例。
- 用户可以停止自己的 `running` 实例。
- 用户可以为自己的未释放实例创建续费订单。
- 用户可以从实例详情进入工单创建页并预填当前实例编号；工单仍只承载沟通和排障定位，不触发实例操作。

## 展示内容

实例列表和详情可展示：

- 实例编号
- 订单编号
- 实例状态：`creating`、`running`、`stopped`、`error`、`releasing`、`released`
- 产品名称
- 套餐名称和规格
- 销售地域
- 系统模板
- 创建时间
- 服务开始时间
- 到期时间
- 到期状态和释放倒计时
- 最近续费订单摘要
- 释放时间

## 关联接口

- `GET /api/instances` - 当前用户实例列表
- `GET /api/instances/{instance_no}` - 当前用户实例详情
- `POST /api/instances/{instance_no}/start` - 启动当前用户自己的实例
- `POST /api/instances/{instance_no}/stop` - 停止当前用户自己的实例
- `POST /api/instances/{instance_no}/renewal-orders` - 为当前用户自己的实例创建续费订单

具体字段、响应和错误码以 `docs/server/api/` 为准。

用户端只调用 `/api/*`，不得调用 `/admin-api/*`。

## 展示限制

- 不展示 MCP/PVE `node`、`storage`、`disk_source`、`snippets_storage`、`vmid`、operation ID 或管理端失败详情。
- 不展示重启、重装、重置密码、控制台、快照、备份、迁移、监控、网络防火墙或资源池管理入口。
- `error` 状态只提示联系后台处理，不展示上游原始错误。
- 续费订单只展示后端返回的最近续费订单摘要和支付状态，不在实例页内伪造支付结果。
- 实例关联工单只传递业务实例编号，不传递 PVE 节点、VMID、operation ID 或管理端失败详情。

## 验收重点

- 未登录访问会进入 `/login`。
- 只展示当前用户自己的实例。
- 非法状态不展示不可用操作。
- 启动和停止操作有明确加载、成功和失败反馈。
- 续费入口只对未释放实例展示，创建续费订单有明确反馈。
- 到期时间、提醒状态和释放倒计时展示正常。
- 页面不出现商户密钥、完整回调 payload、PVE 节点、资源池或自动交付承诺。
