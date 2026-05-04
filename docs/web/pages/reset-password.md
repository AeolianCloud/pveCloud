# Reset Password 页面契约

## 页面定位

`Reset Password` 是用户端密码重置确认页，对应路由 `/reset-password`。

页面通过邮件中的一次性 `token` 完成密码重置确认，并在成功后引导用户回到登录页。

## 页面行为

- 用户通过 `/reset-password?token=...` 打开页面。
- 页面基础字段为 `token`、`password`、`confirm_password`，其中 `token` 来自 URL 查询参数。
- URL 缺少 `token` 时，页面展示明确错误提示，不允许提交。
- 密码重置成功后提示用户旧会话已失效，并跳转 `/login` 重新登录。
- `token` 过期、已使用、已吊销或无效时展示明确失效提示，并引导用户重新申请密码找回。

## 验证码行为

- 站点配置字段 `password_reset_confirm_captcha_enabled` 来自 `GET /api/site-config`。
- 当 `password_reset_confirm_captcha_enabled=false` 时：
  - 页面不显示验证码区域。
  - 页面不请求 `GET /api/auth/password-reset-confirm-captcha`。
  - 提交 `POST /api/auth/password-reset/confirm` 时不要求提交验证码字段。
- 当 `password_reset_confirm_captcha_enabled=true` 时：
  - 页面首屏加载后请求 `GET /api/auth/password-reset-confirm-captcha`。
  - 页面显示验证码图片、验证码输入框和刷新入口。
  - 提交 `POST /api/auth/password-reset/confirm` 时必须提交 `captcha_id`、`captcha_code`。
  - 验证码错误、缺失、过期或提交失败后，前端刷新当前场景验证码。
  - 当前页验证码不能用于登录、注册或忘记密码申请。
- 如果对应开关关闭，`GET /api/auth/password-reset-confirm-captcha` 应返回受控业务错误，前端不应继续重试。

## 关联接口

- `GET /api/site-config`
- `GET /api/auth/password-reset-confirm-captcha`
- `POST /api/auth/password-reset/confirm`

## 验收重点

- 页面只调用 `/api/*`。
- 开关开启时首屏能拉取验证码并在失败后刷新。
- 开关关闭时页面不请求验证码接口，表单流程保持原样。
- 重置成功后旧会话失效并返回登录页。
