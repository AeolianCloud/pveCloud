# 外部系统集成

本目录对应未来代码目录 `server/internal/integrations/`。

## 集成边界

外部系统 client 只负责协议适配，不拥有业务规则。

```text
Service / Job
  ↓
integration client
  ↓
PVE / payment / notify provider
```

本目录只定义后端如何对接外部系统，包括 client 代码位置、协议封装、回调验签、错误映射、重试语义和 Service/Job 调用边界。部署拓扑、凭据保存、反向代理、备份恢复和运行手册放在 `docs/operations/`。

## PVE

PVE 相关代码放在：

```text
server/internal/integrations/pve/
```

PVE client 可以封装：

- 登录和 token/cookie 管理。
- 查询节点、存储、模板、VM 状态。
- 创建、删除、重装 VM。
- 开机、关机、重启、强制关机。
- 查询 PVE task/UPID 执行结果。

PVE client 不做：

- 判断订单是否可开通。
- 判断用户是否拥有实例。
- 修改订单、支付、钱包状态。
- 把 HTTP 成功当作实例交付成功。

## 支付

支付相关代码放在：

```text
server/internal/integrations/payment/
```

支付 client 可以封装：

- 创建支付单。
- 查询支付状态。
- 支付回调验签。
- 关闭支付单。
- 退款申请和退款查询。

支付业务编排放在 `payment_service.go`，回调必须幂等。

## 通知

通知相关代码放在：

```text
server/internal/integrations/notify/
```

第一阶段可以预留邮件和短信接口，不强制接真实供应商。
