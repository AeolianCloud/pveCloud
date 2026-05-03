# Product Management 页面契约

产品管理用于维护服务器产品目录，页面只调用 `/admin-api/*`。

## 页面范围

- 产品基础信息
- 固定服务器套餐
- 套餐周期价格
- 销售地域
- 服务器系统模板
- 套餐可用销售地域
- 套餐可用服务器系统模板

本页面不包含订单、支付、实例、库存扣减、PVE 节点或自动开通能力。

## 路由与权限

- 路由：`/products`
- 菜单权限：`page.products`
- 查看：`product:view` 或 `product:*`
- 创建：`product:create` 或 `product:*`
- 编辑：`product:update` 或 `product:*`
- 上架、下架、售罄和恢复：`product:publish` 或 `product:*`

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
    ProductEditorDialog.vue
    ProductPlanEditorDialog.vue
    SalesRegionEditorDialog.vue
    ServerOsTemplateEditorDialog.vue
    PlanPricesDialog.vue
    PlanRelationsDialog.vue
```

`index.vue` 只负责状态、请求、权限、事件编排和组件组合，不直接承载所有表格与弹窗表单。

页面采用一个页面多 tab，并通过套餐行内操作维护价格和关联配置：

- 产品
- 套餐
- 销售地域
- 系统模板

套餐列表提供：

- 价格弹窗：维护套餐周期价格
- 关联弹窗：维护套餐可用销售地域和服务器系统模板
- 公开检查列：提示产品、套餐、价格、销售地域或系统模板缺失导致 Web 不展示的原因

## 行为约束

- 第一版不做硬删除，只通过状态和可见性控制展示。
- 套餐为固定规格，不提供自定义配置器。
- 销售地域只表示销售展示和可售约束，不绑定 PVE 节点。
- 服务器系统模板不使用 `image` 命名，不绑定 PVE 模板 ID。
- 套餐公开展示需要产品、套餐、价格、销售地域和系统模板均满足公开目录条件。

## 关联接口

接口字段和响应结构以 `docs/server/api/endpoints.md` 为准。

- `GET /admin-api/products`
- `POST /admin-api/products`
- `PUT /admin-api/products/{id}`
- `PATCH /admin-api/products/{id}/status`
- `GET /admin-api/product-plans`
- `POST /admin-api/product-plans`
- `PUT /admin-api/product-plans/{id}`
- `PATCH /admin-api/product-plans/{id}/status`
- `GET /admin-api/product-plans/{id}/prices`
- `PUT /admin-api/product-plans/{id}/prices`
- `GET /admin-api/product-plans/{id}/regions`
- `PUT /admin-api/product-plans/{id}/regions`
- `GET /admin-api/product-plans/{id}/os-templates`
- `PUT /admin-api/product-plans/{id}/os-templates`
- `GET /admin-api/sales-regions`
- `POST /admin-api/sales-regions`
- `PUT /admin-api/sales-regions/{id}`
- `GET /admin-api/server-os-templates`
- `POST /admin-api/server-os-templates`
- `PUT /admin-api/server-os-templates/{id}`
