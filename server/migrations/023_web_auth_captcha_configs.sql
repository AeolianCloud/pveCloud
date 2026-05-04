-- Web auth captcha switches for login, register and password reset flows.

SET NAMES utf8mb4;

USE `pvecloud`;

INSERT INTO `system_configs` (`config_key`, `config_value`, `value_type`, `group_name`, `is_secret`, `description`) VALUES
  ('web.auth.login_captcha_enabled', 'false', 'bool', '用户认证', 0, '是否开启 Web 登录验证码'),
  ('web.auth.register_captcha_enabled', 'false', 'bool', '用户认证', 0, '是否开启 Web 注册验证码'),
  ('web.auth.password_reset_request_captcha_enabled', 'false', 'bool', '用户认证', 0, '是否开启 Web 忘记密码申请验证码'),
  ('web.auth.password_reset_confirm_captcha_enabled', 'false', 'bool', '用户认证', 0, '是否开启 Web 重置密码确认验证码')
ON DUPLICATE KEY UPDATE
  `value_type` = VALUES(`value_type`),
  `group_name` = VALUES(`group_name`),
  `is_secret` = VALUES(`is_secret`),
  `description` = VALUES(`description`);
