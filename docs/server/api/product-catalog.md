# 产品目录 API

本文档维护管理端产品、套餐、价格、销售地域、系统模板、网络类型和用户端公开产品目录接口。跨接口通用约定见 `docs/server/api/conventions.md`。

## 产品目录

产品目录维护服务器产品展示和可售约束。订单创建时读取产品目录并保存快照。产品目录本身不发起支付、不创建实例、不绑定 PVE 节点；实例交付通过独立交付映射把套餐、地域、系统模板和网络类型映射到 MCP PVE client API 参数。

### `GET /admin-api/products`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.products`
- 作用：分页查看产品主数据
- 支持按 `type`、`status`、`keyword` 查询

### `POST /admin-api/products`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:create` 或 `product:*`
- 作用：创建产品，当前 `type` 仅允许 `server`
- 审计：`product.create`

### `PUT /admin-api/products/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:update` 或 `product:*`
- 作用：编辑产品名称、slug、介绍、可见性和排序
- 审计：`product.update`

### `PATCH /admin-api/products/{id}/status`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:publish` 或 `product:*`
- 作用：切换产品 `draft`、`active`、`inactive` 状态
- 审计：`product.status.update`

### `DELETE /admin-api/products/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:delete` 或 `product:*`
- 作用：删除产品主数据
- 约束：产品存在套餐时不得删除，应先处理套餐；历史订单只依赖订单快照，不阻止删除
- 审计：`product.delete`

### `GET /admin-api/product-plans`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.products`
- 作用：分页查看服务器套餐和规格
- 支持按 `product_id`、`status`、`keyword` 查询

### `POST /admin-api/product-plans`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:create` 或 `product:*`
- 作用：创建固定服务器套餐
- 约束：套餐只保存固定规格，不提供自定义配置器
- 审计：`product_plan.create`

### `PUT /admin-api/product-plans/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:update` 或 `product:*`
- 作用：编辑套餐规格、介绍、推荐、可见性和排序
- 审计：`product_plan.update`

### `PATCH /admin-api/product-plans/{id}/status`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:publish` 或 `product:*`
- 作用：切换套餐 `draft`、`active`、`inactive`、`sold_out` 状态
- 审计：`product_plan.status.update`

### `DELETE /admin-api/product-plans/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:delete` 或 `product:*`
- 作用：删除固定服务器套餐
- 约束：同事务删除套餐周期价格、套餐销售地域关联和套餐系统模板关联；历史订单只依赖订单快照，不阻止删除
- 审计：`product_plan.delete`

### `PUT /admin-api/product-plans/{id}/prices`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:update` 或 `product:*`
- 作用：覆盖保存套餐周期价格
- 金额单位：分，不使用浮点数
- 支持周期：`monthly`、`quarterly`、`semi_yearly`、`yearly`
- 审计：`product_plan.prices.update`

### `GET /admin-api/product-plans/{id}/prices`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.products`
- 作用：读取套餐当前周期价格，用于产品管理页面回显

### `PUT /admin-api/product-plans/{id}/regions`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:update` 或 `product:*`
- 作用：覆盖保存套餐可用销售地域
- 审计：`product_plan.regions.update`

### `GET /admin-api/product-plans/{id}/regions`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.products`
- 作用：读取套餐当前可用销售地域，用于产品管理页面回显

### `PUT /admin-api/product-plans/{id}/os-templates`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:update` 或 `product:*`
- 作用：覆盖保存套餐可用服务器系统模板
- 审计：`product_plan.os_templates.update`

### `GET /admin-api/product-plans/{id}/os-templates`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.products`
- 作用：读取套餐当前可用服务器系统模板，用于产品管理页面回显

### `PUT /admin-api/product-plans/{id}/network-types`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:update` 或 `product:*`
- 作用：覆盖保存套餐可用网络类型
- 审计：`product_plan.network_types.update`

### `GET /admin-api/product-plans/{id}/network-types`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.products`
- 作用：读取套餐当前可用网络类型，用于产品管理页面回显

### `GET /admin-api/sales-regions`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.products`
- 作用：查看销售地域列表

### `POST /admin-api/sales-regions`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:create` 或 `product:*`
- 作用：创建销售地域。销售地域不直接绑定 PVE 节点；实例交付阶段通过交付映射选择上游节点。
- 审计：`sales_region.create`

### `PUT /admin-api/sales-regions/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:update` 或 `product:*`
- 作用：编辑销售地域
- 审计：`sales_region.update`

### `DELETE /admin-api/sales-regions/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:delete` 或 `product:*`
- 作用：删除销售地域
- 约束：销售地域仍被套餐关联时不得删除；历史订单只依赖订单快照，不阻止删除
- 审计：`sales_region.delete`

### `GET /admin-api/server-os-templates`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.products`
- 作用：查看服务器系统模板列表

### `POST /admin-api/server-os-templates`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:create` 或 `product:*`
- 作用：创建服务器系统模板。系统模板不直接绑定 PVE 模板；实例交付阶段通过交付映射选择上游磁盘来源。
- 审计：`server_os_template.create`

### `PUT /admin-api/server-os-templates/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:update` 或 `product:*`
- 作用：编辑服务器系统模板
- 审计：`server_os_template.update`

### `DELETE /admin-api/server-os-templates/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:delete` 或 `product:*`
- 作用：删除服务器系统模板
- 约束：系统模板仍被套餐关联时不得删除；历史订单只依赖订单快照，不阻止删除
- 审计：`server_os_template.delete`

### `GET /admin-api/network-types`

- 鉴权：管理端 Bearer Token
- 菜单权限：`page.products`
- 作用：查看网络类型列表

### `POST /admin-api/network-types`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:create` 或 `product:*`
- 作用：创建网络类型。网络类型不直接绑定 PVE 网络；实例交付阶段可参与交付映射匹配。
- 审计：`network_type.create`

### `PUT /admin-api/network-types/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:update` 或 `product:*`
- 作用：编辑网络类型
- 审计：`network_type.update`

### `DELETE /admin-api/network-types/{id}`

- 鉴权：管理端 Bearer Token
- 操作权限：`product:delete` 或 `product:*`
- 作用：删除网络类型
- 约束：网络类型仍被套餐关联时不得删除；历史订单只依赖订单快照，不阻止删除
- 审计：`network_type.delete`

### `GET /api/server-catalog`

- 鉴权：公开接口，不要求用户登录
- 作用：返回 Web 可展示服务器产品目录聚合数据
- 返回范围：已上架且可见的服务器产品、套餐、周期价格、销售地域、服务器系统模板和网络类型
- 展示约束：套餐需要至少有一个 active 周期价格、一个 active 且 visible 的销售地域、一个 active 且 visible 的服务器系统模板、一个 active 且 visible 的网络类型才进入公开目录
- 禁止返回：支付、库存扣减、PVE 节点、PVE 模板 ID、PVE 网络 ID、上游 VMID、存储、磁盘来源或资源池信息
