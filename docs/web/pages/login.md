# Login 页面契约

## 页面定位

`Login` 是用户端登录入口，对应路由 `/login`。

本阶段只支持已有用户登录，不开放用户注册、密码找回、资料编辑、订单、实例或工单流程。

## 行为范围

- 未登录用户访问 `/login` 时展示登录表单。
- 已登录用户访问 `/login` 时跳转 `/user`。
- 未登录用户访问 `/user` 时跳转 `/login`，并携带站内 `redirect` 参数用于登录后返回。
- 登录成功后写入用户端 access token、用户摘要和当前会话摘要。
- 登录成功后优先跳转合法 `redirect`，没有合法 `redirect` 时跳转 `/user`。
- 登录失败停留在登录页；账号不存在或密码错误使用统一错误提示，用户被禁用时展示明确禁用提示。
- 用户端退出后无论接口成功失败都清理本地登录态，并跳转 `/login`。

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
- 当前阶段不自动调用 `POST /api/auth/refresh`；token 过期、token 无效或会话失效时按未登录处理。

## 错误处理

- `401xx` 表示未登录、token 无效、token 过期或会话失效，前端必须清理本地登录态；如果当前访问受保护路由，跳转 `/login`。
- `403xx` 表示无权限，本阶段用户端不引入权限码，前端不应把 `403xx` 当作未登录处理。
- 请求层遇到 HTTP 401 或响应包业务码 `401xx` 时必须清理本地 token；HTTP 403 或业务码 `403xx` 不触发本地 token 清理。
- 登录失败提示不得区分账号不存在和密码错误；用户被禁用时可以展示明确禁用提示。

## 关联接口

- `POST /api/auth/login`
- `GET /api/auth/me`
- `POST /api/auth/logout`
- `POST /api/auth/refresh`

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
- 当前阶段 token 过期不会自动 refresh，会回到登录流程。
