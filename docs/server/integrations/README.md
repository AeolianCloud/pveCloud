# 外部系统集成

外部客户端只负责协议适配，业务服务和 Worker 负责业务规则、幂等、状态机和补偿。

## PVE

PVE 客户端位于 `server/internal/integrations/pve/`，可处理登录、token/cookie、节点、存储、模板、VM 状态、创建、删除、重装、电源操作和 task/UPID 查询。

PVE 客户端不得决定订单资格、实例归属或交付完成状态。

## 支付

支付客户端位于 `server/internal/integrations/payment/`，可创建、查询、关闭支付订单，校验回调，请求和查询退款。

支付编排属于 `payment_service.go`，支付回调必须幂等。

## 通知

通知适配器位于 `server/internal/integrations/notify/`。第一阶段可保留邮件和短信能力入口，不要求真实供应商接入。

## 恢复原则

- MariaDB 是业务事实来源。
- 外部系统成功不等于本地业务完成。
- 本地失败、远端成功的情况必须能通过本地恢复锚点补偿。
