# Forgot Password 页面契约

## 页面定位

`Forgot Password` 是用户端密码找回申请页，对应路由 `/forgot-password`。

页面职责是收集用户邮箱并触发密码重置邮件申请，不负责展示账号是否存在，也不负责直接重置密码。

## 页面行为

- 未登录用户访问 `/forgot-password` 时展示找回密码表单。
- 已登录用户访问 `/forgot-password` 时仍可访问，但页面不承载控制台业务能力。
- 表单基础字段为 `email`。
- 提交成功后统一展示“如果邮箱对应有效账号，重置链接会发送到该邮箱”，不暴露账号存在性。
- 邮件服务未配置或暂不可用时，页面展示明确服务不可用提示。

## 验证码行为

- 站点配置字段 `password_reset_request_captcha_enabled` 来自 `GET /api/site-config`。
- 当 `password_reset_request_captcha_enabled=false` 时：
  - 页面不显示验证码区域。
  - 页面不请求 `GET /api/auth/password-reset-request-captcha`。
  - 提交 `POST /api/auth/password-reset/request` 时不要求提交验证码字段。
- 当 `password_reset_request_captcha_enabled=true` 时：
  - 页面首屏加载后请求 `GET /api/auth/password-reset-request-captcha`。
  - 页面显示验证码图片、验证码输入框和刷新入口。
  - 提交 `POST /api/auth/password-reset/request` 时必须提交 `captcha_id`、`captcha_code`。
  - 验证码错误、缺失、过期或提交失败后，前端刷新当前场景验证码。
  - 当前页验证码不能用于登录、注册或重置密码确认。
- 如果对应开关关闭，`GET /api/auth/password-reset-request-captcha` 应返回受控业务错误，前端不应继续重试。

## 关联接口

- `GET /api/site-config`
- `GET /api/auth/password-reset-request-captcha`
- `POST /api/auth/password-reset/request`

## 验收重点

- 页面只调用 `/api/*`。
- 开关开启时首屏能拉取验证码并在失败后刷新。
- 开关关闭时页面不请求验证码接口，表单流程保持原样。
- 页面不暴露邮箱是否存在。
