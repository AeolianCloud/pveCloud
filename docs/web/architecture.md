# Web 前端架构

本文件描述用户端前端的目标边界、阶段范围和实现契约。

## 当前状态

- Web 基础前台阶段已经落地用户端前端壳、路由、静态页面、站点基础展示配置和请求边界准备。
- 当前在服务器产品目录能力之上开放用户注册、密码找回、自动刷新、用户资料编辑和支付宝/微信侧个人实名；订单、支付、实例或工单等业务 API 仍不开放。

## 文档入口

- Web 登录页契约：`docs/web/pages/login.md`
- Web 忘记密码页契约：`docs/web/pages/forgot-password.md`
- Web 重置密码页契约：`docs/web/pages/reset-password.md`
- Web 用户资料页契约：`docs/web/pages/account-profile.md`
- Web 实名页契约：`docs/web/pages/real-name.md`

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
- 用户端不得展示尚未由 owner docs 开放的订单、支付、实例、资源交付或 SLA 承诺。

## 当前阶段范围

当前阶段在基础前台和服务器产品目录之上开放用户账号自助和支付宝/微信侧个人实名能力：

- Home
- Products 服务器产品展示页
- Pricing 服务器套餐价格展示页
- Login / Register / Forgot Password / Reset Password 用户认证页
- User Center 控制台入口页
- Account Profile 用户资料页
- Real Name 个人实名页，支持支付宝/微信侧实名核验
- 404

这些页面承载信息架构、服务器产品目录展示、用户登录态、账号自助和未来接入点，不承载下单、支付、实例或工单流程。

## 本阶段不开放能力

- 用户端业务接口（公开站点配置、用户账号自助、用户实名和服务器产品目录接口除外）
- 订单、支付、实例或工单流程
- 订单创建和订单列表
- 支付发起、回调和支付结果确认
- 实例列表、实例开通和实例操作
- 工单创建、回复和列表
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

`web/` 已沿用上述技术栈。后续如需引入 UI 组件库、CSS 框架或 SSR/SSG 能力，必须先更新本文档并确认。

## 目标目录口径

```text
web/src/
  api/
  assets/
  components/
  layouts/
  router/
  store/
  styles/
  views/
```

说明：

- `api/`：只放 `/api/*` 请求封装，本阶段接公开站点配置、用户账号自助、用户实名和服务器产品目录接口
- `router/`：用户端路由定义
- `store/`：跨页面 concern，本阶段只保留必要基础状态
- `views/`：页面入口
- `layouts/`：用户端页面壳
- `components/`：仅限 `web/` 内部复用组件
- `styles/`：用户端 tokens、reset 和页面基础样式

## 路由口径

本阶段目标路由：

| 路由 | 页面 | 说明 |
| --- | --- | --- |
| `/` | Home | 用户端首页和推荐服务器套餐展示 |
| `/products` | Products | 服务器产品能力、销售地域和系统模板展示 |
| `/pricing` | Pricing | 服务器套餐规格和周期价格展示 |
| `/login` | Login | 用户登录入口，已登录时跳转 `/user` |
| `/register` | Register | 用户注册入口，已登录时跳转 `/user` |
| `/forgot-password` | Forgot Password | 申请密码重置邮件 |
| `/reset-password` | Reset Password | 通过一次性 token 重置密码 |
| `/user` | User Center | 控制台入口，要求用户登录态，未登录跳转 `/login` |
| `/user/profile` | Account Profile | 用户资料编辑页，要求用户登录态 |
| `/user/real-name` | Real Name | 个人实名页，要求用户登录态 |
| `/:pathMatch(.*)*` | 404 | 未匹配路由 |

本阶段 `/user`、`/user/profile` 和 `/user/real-name` 为用户登录保护路由，未登录跳转 `/login`；其它路由公开访问。本阶段不做用户权限判断。

未登录访问 `/user` 时可以携带站内 `redirect` 参数；登录成功后仅允许跳转合法站内路径，非法或缺失时回退 `/user`。

## 站点基础展示配置

Web 左上角品牌区域由后台系统配置驱动：

- `site.name`：站点显示名称，默认 `pveCloud`。
- `site.logo_url`：站点 Logo 图片 URL，默认空；为空时使用前端默认字母标识。
- `web.auth.login_captcha_enabled`：是否为登录页开启图形验证码，默认 `false`。
- `web.auth.register_captcha_enabled`：是否为注册页开启图形验证码，默认 `false`。
- `web.auth.password_reset_request_captcha_enabled`：是否为忘记密码申请页开启图形验证码，默认 `false`。
- `web.auth.password_reset_confirm_captcha_enabled`：是否为重置密码确认页开启图形验证码，默认 `false`。
- `real_name.*`：用户实名开关、供应商选择、提交要求和说明文案，业务开关和供应商接入参数由后台系统配置维护；支付宝、微信/腾讯云密钥属于后台敏感配置，不对前端暴露。

展示规则：

- Web 顶部品牌区域同时支持文字和图片。
- 文字来自 `site.name`，为空时回退为 `pveCloud`。
- 图片来自 `site.logo_url`，为空时展示前端默认标识。
- 4 个用户认证验证码开关由 `GET /api/site-config` 暴露为布尔字段，前端按页面独立决定是否显示验证码区域。
- 用户实名公开配置由 `GET /api/site-config` 暴露为 `real_name` 对象，用于控制实名入口、可选供应商、默认供应商和说明文案；可选供应商必须是服务端过滤后的已启用且配置完整的供应商。
- `real_name.enabled` 跟随后台实名入口开关和可用实名方式；支付宝/微信侧供应商不可用、证件摘要密钥缺失或回调地址缺失时，不关闭已启用的人工审核入口，而是由服务端返回 `manual` 作为人工审核实名方式。
- 该配置为公开展示信息，不包含敏感配置。

## 请求边界

- 请求基础路径为 `/api/*`。
- 本阶段允许 Web 调用 `GET /api/site-config` 获取公开站点基础展示配置。
- 本阶段允许 Web 调用用户认证验证码接口：`GET /api/auth/login-captcha`、`GET /api/auth/register-captcha`、`GET /api/auth/password-reset-request-captcha`、`GET /api/auth/password-reset-confirm-captcha`。
- 本阶段允许 Web 调用用户账号自助接口：`POST /api/auth/login`、`POST /api/auth/register`、`GET /api/auth/me`、`POST /api/auth/logout`、`POST /api/auth/refresh`、`POST /api/auth/password-reset/request`、`POST /api/auth/password-reset/confirm`、`PATCH /api/user/profile`、`POST /api/user/password`。
- 本阶段允许 Web 调用用户实名接口：`GET /api/user/real-name`、`POST /api/user/real-name`、`POST /api/user/real-name/sync`。
- 本阶段允许 Web 调用 `GET /api/server-catalog` 获取公开服务器产品目录。
- 本阶段可以创建用户端请求封装骨架，但不得调用 `/admin-api/*`。
- 本阶段不新增站点配置、用户账号自助、用户实名和服务器产品目录以外的用户端业务接口契约。
- 后续接入真实业务接口时，必须先更新 `docs/server/api/` 和必要的数据库契约。

## 状态边界

- 用户端登录态由用户端 JWT access token 和服务端 `user_sessions` 共同决定。
- Web 可以将用户端 access token 保存到浏览器本地存储，用于页面刷新后恢复登录态。
- Web 启动和进入 `/user` 前通过 `GET /api/auth/me` 恢复登录态；失败时清理本地 token。
- 用户退出时调用 `POST /api/auth/logout`，无论接口成功失败都清理本地 token 并回到登录入口。
- 登录成功后保存用户摘要和当前会话摘要；登录失败停留在 `/login`，账号不存在或密码错误使用统一提示，禁用账号展示明确禁用提示。
- 当 `login_captcha_enabled=true` 时，登录页首屏拉取 `GET /api/auth/login-captcha` 并在提交 `POST /api/auth/login` 时额外提交 `captcha_id`、`captcha_code`；验证码错误、过期或缺失时刷新当前验证码。
- 注册成功后直接创建用户端会话并进入 `/user`；用户名和邮箱重复时返回明确重复提示。
- 当 `register_captcha_enabled=true` 时，注册页首屏拉取 `GET /api/auth/register-captcha` 并在提交 `POST /api/auth/register` 时额外提交 `captcha_id`、`captcha_code`；验证码错误、过期或缺失时刷新当前验证码。
- 忘记密码通过邮箱发送一次性重置链接；申请接口不得暴露账号是否存在。
- 当 `password_reset_request_captcha_enabled=true` 时，忘记密码页首屏拉取 `GET /api/auth/password-reset-request-captcha` 并在提交 `POST /api/auth/password-reset/request` 时额外提交 `captcha_id`、`captcha_code`；验证码错误、过期或缺失时刷新当前验证码。
- 当 `password_reset_confirm_captcha_enabled=true` 时，重置密码页首屏拉取 `GET /api/auth/password-reset-confirm-captcha` 并在提交 `POST /api/auth/password-reset/confirm` 时额外提交 `captcha_id`、`captcha_code`；验证码错误、过期或缺失时刷新当前验证码。
- 4 个认证流程的验证码互不通用，场景关闭时前端不得请求对应验证码接口。
- 用户资料编辑仅允许当前登录用户修改邮箱、显示名称和密码，不开放用户名修改。
- 用户实名仅允许当前登录用户提交个人实名资料、选择支付宝/微信侧供应商或人工审核方式、同步自己的外部供应商实名结果和查看自己的实名状态；实名状态不作为用户端权限码。
- `real_name.enabled=false` 时用户端不得提交实名申请。
- `pending` 和 `approved` 状态不得重复提交；`pending` 状态只允许同步供应商结果；`rejected` 状态是否允许重提由后台配置和提交次数决定。
- 前端不得根据支付宝、微信或腾讯云返回的 URL 参数直接判定实名通过；必须调用服务端同步接口，由服务端查询供应商结果。
- 用户端人工审核实名不展示证件图片上传，不能调用旧人工实名上传接口。
- `real_name.required_for_order=true` 时，用户购买机器前必须实名通过；未实名、核验中或被拒绝时引导到 `/user/real-name`。
- 已登录用户访问 `/login` 时跳转 `/user`。
- `401xx` 按未登录、token 无效、token 过期或会话失效处理，前端必须清理本地登录态；受保护路由访问失败时回到 `/login`。
- `403xx` 不作为未登录处理；当前阶段用户端不引入权限码。
- 请求层遇到 HTTP 401 或响应包业务码 `401xx` 时清理本地 token；HTTP 403 或业务码 `403xx` 不触发本地 token 清理。
- 当前阶段开放自动调用 `POST /api/auth/refresh`；当前 token 接近过期且当前会话仍有效时轮换新 token，刷新失败时回到登录流程。
- 本阶段不持久化订单、实例或工单状态。

## 服务器产品目录展示

Web 通过 `GET /api/server-catalog` 展示服务器产品目录。

展示范围：

- 云服务器产品基础信息
- 固定服务器套餐
- 产品和套餐简介
- 月付、季付、半年付和年付价格
- 销售地域
- 服务器系统模板
- 推荐套餐、上架状态和售罄状态

展示限制：

- 不展示订单、支付、实例、PVE 节点、资源池或库存扣减信息。
- “立即购买”类交易文案不得出现；CTA 使用“登录后查看购买入口”“购买功能即将开放”或“暂时售罄”。
- 产品价格、可售性、地域和系统模板最终以后端公开目录返回为准。

## 契约原则

- 页面行为、状态语义、请求包装和登录流程必须以 `docs/server/api/` 和对应业务设计文档为准
- 订单金额、产品可售性、服务器系统模板、地域、支付状态等最终以后端为准
- 用户端前端负责展示与交互，不负责最终业务裁决
- 用户端展示页不能隐式承诺后台尚未开放的产品、订单、支付、实例或工单能力

## 后续阶段要求

后续开放订单、支付、实例和工单等真实用户端业务前，至少需要补齐：

- 订单、支付、实例和工单相关用户端 `/api/*` 接口契约
- 产品、订单、支付、实例和工单所需数据库契约
- 订单和支付流程说明
- 工单与实例相关页面范围
- 用户端路由、权限和请求包装口径的行为细节
- 必要的管理端运营页面和权限口径

这些内容确认后，再进入对应业务实现。

## 验收口径

- `web/` 与 `admin/` 独立，不共享运行时代码
- `web/` 不调用 `/admin-api/*`
- 本阶段只新增公开站点配置、用户账号自助、用户实名和服务器产品目录接口，不新增订单、支付、实例或工单接口
- Home、Products 和 Pricing 通过公开服务器产品目录展示产品、套餐、价格、销售地域、服务器系统模板、简介和状态
- Web 左上角品牌区域可展示后台配置的站点名称和 Logo 图片，配置为空时有默认回退
- 顶部导航“登录”和“控制台”指向同一用户登录体系；未登录访问控制台进入 `/login`，登录成功进入 `/user`
- 登录、注册、密码找回、重置密码、自动刷新和用户资料编辑流程可用，且不调用 `/admin-api/*`
- 用户实名页可按后台配置展示实名要求、供应商选择和供应商跳转/同步状态，且不调用 `/admin-api/*`
- 4 个认证页面根据 `GET /api/site-config` 返回的布尔开关独立显示或隐藏验证码区域；开关关闭时不请求验证码接口
- 基础路由和 404 可访问
- 用户端静态页面在桌面和移动端可用
- `web` 构建通过
