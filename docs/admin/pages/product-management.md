# Product Management 页面契约

产品管理用于维护服务器产品目录，页面只调用 `/admin-api/*`。

## 页面范围

- 产品基础信息
- 固定服务器套餐
- 套餐周期价格
- 销售地域
- 服务器系统模板
- 网络类型
- 套餐可用销售地域
- 套餐可用服务器系统模板
- 套餐可用网络类型

本页面不包含订单、支付、实例操作、库存扣减、PVE 节点或自动开通能力；实例交付参数由实例管理的交付映射维护。

## 路由与权限

- 路由：`/products`
- 菜单权限：`page.products`
- 查看：`product:view` 或 `product:*`
- 创建：`product:create` 或 `product:*`
- 编辑：`product:update` 或 `product:*`
- 上架、下架、售罄和恢复：`product:publish` 或 `product:*`
- 删除：`product:delete` 或 `product:*`

## 页面结构

产品管理是复杂管理页，必须采用页面容器结构：

```text
admin/src/views/products/
  index.vue
  types.ts
  components/
    ProductsTab.vue
    ProductPlansTab.vue
    SalesRegionsTab.vue
    ServerOsTemplatesTab.vue
    NetworkTypesTab.vue
    ProductEditorDialog.vue
    ProductPlanEditorDialog.vue
    SalesRegionEditorDialog.vue
    ServerOsTemplateEditorDialog.vue
    NetworkTypeEditorDialog.vue
    PlanPricesDialog.vue
    PlanRelationsDialog.vue
```

`index.vue` 只负责状态、请求、权限、事件编排和组件组合，不直接承载所有表格与弹窗表单。

页面采用一个页面多 tab，并通过套餐行内操作维护价格和关联配置：

- 产品
- 套餐
- 销售地域
- 系统模板
- 网络类型

套餐列表提供：

- 价格弹窗：维护套餐周期价格
- 关联弹窗：维护套餐可用销售地域、服务器系统模板和网络类型
- 公开检查列：提示产品、套餐、价格、销售地域、系统模板或网络类型缺失导致 Web 不展示的原因

## 行为约束

- 删除采用受限硬删除：产品存在套餐时不可删除；销售地域或系统模板仍被套餐关联时不可删除；套餐删除会同事务删除套餐价格和套餐关联。
- 历史订单只依赖创建时保存的产品、套餐、价格、销售地域和系统模板快照，不阻止产品目录项删除。
- 删除前必须二次确认，删除成功后刷新对应列表。
- 套餐为固定规格，不提供自定义配置器。
- 销售地域只表示销售展示和可售约束，不绑定 PVE 节点。
- 服务器系统模板不使用 `image` 命名，不绑定 PVE 模板 ID。
- 网络类型由后台自定义维护，当前只作为销售展示、下单选择和订单快照，不绑定 PVE 网络；后续对接 PVE 时可基于网络类型编码映射真实网络。
- 套餐公开展示需要产品、套餐、价格、销售地域、系统模板和网络类型均满足公开目录条件。

## 关联接口

接口字段和响应结构以 `docs/server/api/endpoints.md` 为准。

- `GET /admin-api/products`
- `POST /admin-api/products`
- `PUT /admin-api/products/{id}`
- `PATCH /admin-api/products/{id}/status`
- `DELETE /admin-api/products/{id}`
- `GET /admin-api/product-plans`
- `POST /admin-api/product-plans`
- `PUT /admin-api/product-plans/{id}`
- `PATCH /admin-api/product-plans/{id}/status`
- `DELETE /admin-api/product-plans/{id}`
- `GET /admin-api/product-plans/{id}/prices`
- `PUT /admin-api/product-plans/{id}/prices`
- `GET /admin-api/product-plans/{id}/regions`
- `PUT /admin-api/product-plans/{id}/regions`
- `GET /admin-api/product-plans/{id}/os-templates`
- `PUT /admin-api/product-plans/{id}/os-templates`
- `GET /admin-api/product-plans/{id}/network-types`
- `PUT /admin-api/product-plans/{id}/network-types`
- `GET /admin-api/sales-regions`
- `POST /admin-api/sales-regions`
- `PUT /admin-api/sales-regions/{id}`
- `DELETE /admin-api/sales-regions/{id}`
- `GET /admin-api/server-os-templates`
- `POST /admin-api/server-os-templates`
- `PUT /admin-api/server-os-templates/{id}`
- `DELETE /admin-api/server-os-templates/{id}`
- `GET /admin-api/network-types`
- `POST /admin-api/network-types`
- `PUT /admin-api/network-types/{id}`
- `DELETE /admin-api/network-types/{id}`
