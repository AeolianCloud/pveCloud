# Web 前端架构

本文件描述用户端前端的目标边界、阶段范围和实现契约。

## 当前状态

- Web 用户端前端页面准备清空重做。
- 清空后 `web/` 仅保留独立请求边界和后续重做所需目录基础，不保留 Vue 应用入口。
- 当前不开放用户端页面、公开路由、用户账号自助页面、产品目录展示页、用户中心或实名页。
- 已存在的用户端接口、后端能力和历史页面契约不等于当前前端开放页面；重新开放任何页面前必须先更新本文档和对应页面契约。

## 文档入口

当前开放的页面契约：

- `docs/web/pages/home.md` - 首页
- `docs/web/pages/products.md` - 产品展示页
- `docs/web/pages/login.md` - 登录页
- `docs/web/pages/register.md` - 注册页
- `docs/web/pages/forgot-password.md` - 忘记密码页
- `docs/web/pages/reset-password.md` - 重置密码页
- `docs/web/pages/user-center.md` - 用户中心页
- `docs/web/pages/account-profile.md` - 账号资料页
- `docs/web/pages/real-name.md` - 实名认证页
- `docs/web/pages/not-found.md` - 404页面

后续重新设计页面时，先更新本文档和对应页面契约，再实现页面与路由。

## 定位

用户端前端面向官网、公区产品展示和用户中心。
接口边界固定为 `/api/*`。

## 独立边界

- `web/` 只调用 `/api/*`
- 不调用 `/admin-api/*`
- 不导入 `admin/` 的运行时代码
- 不创建跨前端共享运行时代码包
- 不依赖管理端权限、菜单、状态或请求封装

## 前后端边界配套

用户端前端的 API 消费边界，对应后端的用户端实现边界：

- 前端调用边界：`/api/*`
- 后端实现边界：`server/internal/delivery/http/web/*` 聚合路由，业务编排落在 `server/internal/usecase/web/*`，GORM model 和可复用查询对象落在 `server/internal/repository/mysql/*`

不要把用户端页面契约建立在管理端 `/admin-api/*`、`server/internal/delivery/http/admin/*` 或 `server/internal/usecase/admin/*` 之上。

## 安全边界

通用安全基线见 `docs/security.md`。

- 用户端前端只做登录态展示、路由保护和交互引导；用户状态、会话状态、实名状态、产品可售性和资源归属以后端为准。
- 用户端不得读取、调用或复用管理端权限模型、菜单、token、请求封装或运行时代码。
- 用户端本地保存的 access token 只用于请求认证，不代表当前会话一定有效；启动恢复和受保护路由进入前必须按本文状态边界校验。
- 用户端仅展示 owner docs 已开放的订单 MVP；不得展示尚未开放的支付、实例、资源交付或 SLA 承诺。

## 当前阶段范围

当前阶段开放官网展示、用户账号自助页面和订单 MVP，不开放支付、实例、工单等业务流程。

开放页面：
- 首页（Home）
- 产品展示页（Products）
- 登录页（Login）
- 注册页（Register）
- 忘记密码页（Forgot Password）
- 重置密码页（Reset Password）
- 用户中心页（User Center）
- 账号资料页（Account Profile）
- 实名认证页（Real Name）
- 订单列表页（Orders）
- 订单详情页（Order Detail）
- 404 页面

不开放页面：
- 价格页
- 支付页
- 实例管理页
- 工单页
- 钱包页

## 本阶段不开放能力

- 支付、实例或工单流程
- 支付发起、回调和支付结果确认
- 实例列表、实例开通和实例操作
- 工单创建、回复和列表
- 钱包、发票、SSH密钥管理
- 与 `admin/` 的任何运行时代码共享

## 技术倾向

当前规划沿用项目现有前端栈方向：

- Bun
- Vite
- Vue 3
- TypeScript
- Vue Router
- Pinia
- Axios
- Tailwind CSS

后续如需再引入其它 UI 组件库、CSS 框架或 SSR/SSG 能力，必须先更新本文档并确认。

## 目标目录口径

```text
web/src/
  api/
  assets/
  components/
  composables/
  layouts/
  router/
  stores/
  styles/
  views/
  App.vue
  main.ts
```

说明：

- `api/`：`/api/*` 请求封装，按业务域拆分。
- `assets/`：静态资源目录。
- `components/`：全局共享组件。
- `composables/`：组合式函数。
- `layouts/`：布局组件。
- `router/`：路由配置。
- `stores/`：Pinia 状态管理。
- `styles/`：全局样式。
- `views/`：页面组件，按路由结构组织。
- `App.vue`：应用根组件。
- `main.ts`：应用入口。

## 路由口径

当前开放以下路由：

- `/` - 首页
- `/products` - 产品展示页
- `/login` - 登录页
- `/register` - 注册页
- `/forgot-password` - 忘记密码页
- `/reset-password` - 重置密码页
- `/user` - 用户中心页（受保护）
- `/user/profile` - 账号资料页（受保护）
- `/user/real-name` - 实名认证页（受保护）
- `/user/orders` - 订单列表页（受保护）
- `/user/orders/:orderNo` - 订单详情页（受保护）
- `/:pathMatch(.*)*` - 404页面

路由元信息必须包含标题、图标、权限、菜单可见性等页面契约。

## 站点基础展示配置

官网展示站点名称、Logo、验证码区域和实名入口。

可继续复用现有公开配置接口 `GET /api/site-config`，但必须先恢复对应页面契约。

## 请求边界

- 请求基础路径为 `/api/*`。
- 当前可调用的接口：
  - 站点配置：`GET /api/site-config`
  - 站点 Logo：`GET /api/site-logo/{id}`，仅用于公开展示当前站点 Logo
  - 认证相关：登录、注册、密码找回、密码重置、登录态恢复、刷新、退出
  - 用户资料：`GET /api/auth/me`、`PATCH /api/user/profile`、`POST /api/user/password`
  - 实名认证：`GET /api/user/real-name`、`POST /api/user/real-name`、`POST /api/user/real-name/sync`
  - 产品目录：`GET /api/server-catalog`，返回套餐可选价格、销售地域、系统模板和网络类型
  - 订单：`POST /api/orders`、`GET /api/orders`、`GET /api/orders/{order_no}`、`POST /api/orders/{order_no}/cancel`
- 后续接入真实业务接口时，必须先更新 `docs/server/api/` 和必要的数据库契约。

## 状态边界

当前支持登录态恢复、token 刷新、用户资料展示和实名申请。

登录态恢复：
- Web 本地可保存用户端 access token，用于页面刷新后恢复登录态。
- Web 启动和进入受保护路由前，如果本地存在 token，必须调用 `GET /api/auth/me` 恢复用户摘要和会话摘要。
- `GET /api/auth/me` 成功后视为登录态有效，成功响应中的用户摘要为服务端当前真实摘要。
- `GET /api/auth/me` 失败时必须清理本地 token、用户摘要和会话摘要。

token 刷新：
- 当前阶段自动调用 `POST /api/auth/refresh`；当前 token 接近过期且会话仍有效时轮换新 token 和新会话。
- 自动刷新成功后更新本地 access token、用户摘要和当前会话摘要。
- 自动刷新失败时清理本地 token、用户摘要和会话摘要；如果当前访问受保护路由，跳转 `/login`。

实名状态：
- 实名状态以后端返回为准，前端展示当前状态：未实名、核验中、已通过、已拒绝。
- 实名功能开关和配置来自 `GET /api/site-config`。

## 服务器产品目录展示

官网展示服务器产品目录，调用 `GET /api/server-catalog`。

展示范围：

- 云服务器产品基础信息
- 固定服务器套餐
- 产品和套餐简介
- 月付、季付、半年付和年付价格
- 销售地域
- 服务器系统模板
- 网络类型
- 推荐套餐、上架状态和售罄状态

展示限制：

- 展示订单 MVP 下单入口，但不展示支付、实例、PVE 节点、资源池或库存扣减信息。
- CTA 可以进入订单确认或创建流程；不得承诺支付成功、自动开通、实例交付或 SLA。
- 产品价格、可售性、地域、系统模板和网络类型最终以后端公开目录返回为准。

## 订单 MVP

用户端订单只表示购买意向和后台人工处理入口。

页面范围：

- 产品页从可售套餐进入订单创建。
- 产品页创建订单前必须展示购买配置确认，用户选择计费周期、销售地域、系统模板和网络类型后再提交订单。
- 订单创建前必须登录；未登录时跳转 `/login` 并携带站内 `redirect`。
- 当后台配置 `real_name.required_for_order=true` 时，未通过实名的用户不能创建订单，应引导到 `/user/real-name`。
- `/user/orders` 展示当前登录用户自己的订单列表。
- `/user/orders/:orderNo` 展示当前登录用户自己的订单详情。
- 用户可以取消自己的 `pending` 订单。

展示限制：

- 只展示订单状态 `pending`、`cancelled`、`closed`。
- 不展示支付状态、实例状态、PVE 节点、资源池、库存扣减或自动开通进度。
- 订单金额、产品可售性、地域、系统模板和网络类型最终以后端订单接口返回为准。

## 契约原则

- 页面行为、状态语义、请求包装和登录流程必须以 `docs/server/api/` 和对应业务设计文档为准
- 订单金额、产品可售性、服务器系统模板、地域、网络类型等最终以后端为准
- 用户端前端负责展示与交互，不负责最终业务裁决
- 用户端展示页不能隐式承诺后台尚未开放的支付、实例或工单能力；订单只按 MVP 契约展示购买意向和后台人工处理入口

## 后续阶段要求

后续开放支付、实例和工单等真实用户端业务前，至少需要补齐：

- 支付、实例和工单相关用户端 `/api/*` 接口契约
- 支付、实例和工单所需数据库契约
- 支付流程说明
- 工单与实例相关页面范围
- 用户端路由、权限和请求包装口径的行为细节
- 必要的管理端运营页面和权限口径

这些内容确认后，再进入对应业务实现。

## 验收口径

- `web/` 与 `admin/` 独立，不共享运行时代码。
- `web/` 不调用 `/admin-api/*`。
- 官网页面实现完整，路由注册正确，Vue 应用入口正常。
- 页面行为符合对应页面契约。
- 登录态恢复、token 刷新、401/403 处理正确。
- 产品目录展示符合契约限制。
- `web` 构建通过。
