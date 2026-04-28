# Login 页面契约

## 页面定位

`Login` 是管理端登录入口。

它不挂载在后台壳内，不出现在侧栏菜单中。

## 行为范围

- 首次进入登录页时获取验证码。
- 登录成功后写入 token、本地登录态和最小管理员摘要。
- 登录错误留在登录页处理。
- 已登录用户访问登录页时，可按路由重定向规则进入默认受保护页面。

## 会话恢复

- 首次进入或刷新后，如果本地存在 token，前端优先调用 `GET /admin-api/auth/me` 恢复登录态。
- `GET /admin-api/dashboard` 只负责首页数据，不承担登录态恢复职责。
- `POST /admin-api/auth/logout` 无论成功失败都要清理本地状态。
- `POST /admin-api/auth/refresh` 失败按登录过期处理。

## 关联接口

以 `docs/server/api/` 中认证相关契约为准。

管理端只调用 `/admin-api/*`。

## 权限

登录页本身不要求页面权限。

登录成功后的可访问页面由路由 `meta.permission`、权限 store 和后端返回的权限码共同决定。
