# Web 前端

Web 前端面向官网、公区产品展示和用户中心。接口边界是 `/api/*`，最终接口契约以 `docs/server/api/openapi.yaml` 为准。

## 独立边界

- `web/` 只调用 `/api/*`。
- `web/` 不调用 `/admin-api/*`。
- `web/` 不导入 `admin/` 的页面、组件、请求、状态、类型、常量或工具。
- 不创建公共前端 `shared/` 包。

## 样式组织

- `src/style.css` 只承载全局设计变量、基础 reset、应用外壳布局和跨页面复用工具类。
- 页面或组件私有样式写在对应 Vue SFC 的 `<style scoped>` 中，避免把页面级 class 长期堆进全局 CSS。
- 需要在用户端多个页面复用的样式，只能在 `web/` 内部抽取，不与 `admin/` 共用样式包。
- 主题相关颜色、边框、阴影和交互态使用语义化 CSS 变量；页面局部变量可定义在页面根 class 上。

## 技术栈

- Bun
- Vue 3
- Vite
- TypeScript
- Vue Router
- Pinia
- Axios

## 页面范围

- Home
- Products
- Pricing
- Login/register
- Order creation
- Payment result
- User center
- My orders
- My instances
- Tickets

## 状态设计

- `auth`：用户 token 和用户资料。
- `cart`：可选订单配置草稿。
- `order`：当前订单状态。
- `instance`：可选实例列表缓存。

## 接口边界

用户端页面只能依赖 OpenAPI 中 `/api/*` 契约。订单金额、可购买套餐、镜像、地域和支付状态必须以后端返回为准，前端不做最终业务裁决。
