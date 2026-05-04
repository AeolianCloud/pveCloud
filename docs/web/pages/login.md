# Login 页面契约

## 页面定位

`Login` 是用户端登录入口，对应路由 `/login`。

本阶段开放用户注册、密码找回、登录、登录态恢复、自动刷新和退出，不开放订单、实例或工单流程。

## 行为范围

- 未登录用户访问 `/login` 时展示登录表单。
- 已登录用户访问 `/login` 时跳转 `/user`。
- 未登录用户访问 `/user` 时跳转 `/login`，并携带站内 `redirect` 参数用于登录后返回。
- 登录成功后写入用户端 access token、用户摘要和当前会话摘要。
- 登录成功后优先跳转合法 `redirect`，没有合法 `redirect` 时跳转 `/user`。
- 登录失败停留在登录页；账号不存在或密码错误使用统一错误提示，用户被禁用时展示明确禁用提示。
- 用户端退出后无论接口成功失败都清理本地登录态，并跳转 `/login`。

## 注册范围

- 未登录用户访问 `/register` 时展示注册表单。
- 已登录用户访问 `/register` 时跳转 `/user`。
- 注册字段包含用户名、邮箱、密码、确认密码和可选显示名称。
- 用户名和邮箱必须唯一；重复时展示明确重复提示。
- 注册成功后写入用户端 access token、用户摘要和当前会话摘要，并跳转 `/user`。
- 注册不创建订单、实例、钱包、余额或任何业务资源。

## 密码找回范围

- 未登录用户访问 `/forgot-password` 时展示密码找回申请表单。
- 申请字段只包含邮箱。
- 申请接口无论邮箱是否存在都返回统一成功提示，不暴露账号存在性。
- 如果邮箱对应 `active` 用户，服务端生成一次性重置 token 并发送重置链接。
- 未配置邮件发送能力时，申请接口应返回明确服务不可用错误；前端展示“密码找回服务暂不可用，请稍后再试”。
- 用户通过 `/reset-password?token=...` 打开重置页，提交新密码和确认密码。
- token 只能使用一次，过期、已使用、已吊销或不存在时返回明确失效提示。
- 如果 token 对应用户已被禁用，重置必须失败并吊销该 token。
- 密码重置成功后吊销该用户现有 active 会话，并跳转 `/login` 重新登录。

## Redirect 规则

- `redirect` 只允许站内路径。
- 合法 `redirect` 必须以单个 `/` 开头。
- `redirect` 不得以 `//` 开头，不得包含协议、域名或其它站外跳转语义。
- 非法、空值或缺失的 `redirect` 一律回退 `/user`。

## 登录态恢复

- Web 本地可保存用户端 access token，用于页面刷新后恢复登录态。
- Web 启动和进入 `/user` 前，如果本地存在 token，必须调用 `GET /api/auth/me` 恢复用户摘要和会话摘要。
- `GET /api/auth/me` 成功后视为登录态有效，成功响应中的用户摘要为服务端当前真实摘要。
- `GET /api/auth/me` 失败时必须清理本地 token、用户摘要和会话摘要。
- 用户被禁用后，`GET /api/auth/me` 不能恢复为有效登录态，并返回明确禁用错误。
- 当前阶段自动调用 `POST /api/auth/refresh`；当前 token 接近过期且会话仍有效时轮换新 token 和新会话。
- 自动刷新成功后更新本地 access token、用户摘要和当前会话摘要。
- 自动刷新失败时清理本地 token、用户摘要和会话摘要；如果当前访问受保护路由，跳转 `/login`。
- token 已过期、无效或对应会话已失效时按未登录处理，不循环重试 refresh。

## 错误处理

- `401xx` 表示未登录、token 无效、token 过期或会话失效，前端必须清理本地登录态；如果当前访问受保护路由，跳转 `/login`。
- `403xx` 表示无权限，本阶段用户端不引入权限码，前端不应把 `403xx` 当作未登录处理。
- 请求层遇到 HTTP 401 或响应包业务码 `401xx` 时必须清理本地 token；HTTP 403 或业务码 `403xx` 不触发本地 token 清理。
- 登录失败提示不得区分账号不存在和密码错误；用户被禁用时可以展示明确禁用提示。

## 关联接口

- `POST /api/auth/login`
- `POST /api/auth/register`
- `GET /api/auth/me`
- `POST /api/auth/logout`
- `POST /api/auth/refresh`
- `POST /api/auth/password-reset/request`
- `POST /api/auth/password-reset/confirm`

具体字段、响应和错误码以 `docs/server/api/` 为准。

用户端只调用 `/api/*`，不得调用 `/admin-api/*`。

## 验收重点

- `/login` 已登录时进入 `/user`。
- `/user` 未登录时进入 `/login`，登录成功后可回到合法站内 `redirect`。
- 非法 `redirect` 不会触发站外跳转。
- 登录失败不区分账号不存在和密码错误；禁用账号展示明确禁用提示。
- 退出后清理本地登录态并回到 `/login`。
- `GET /api/auth/me` 失败后清理本地登录态。
- HTTP 401 或业务码 `401xx` 会清理本地 token。
- `/register` 注册成功后进入 `/user`，重复用户名或邮箱有明确提示。
- `/forgot-password` 不暴露邮箱是否存在。
- `/reset-password` 成功后吊销旧会话并回到 `/login`。
- token 接近过期时会自动 refresh；refresh 失败后清理本地登录态。
