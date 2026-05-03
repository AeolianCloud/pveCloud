-- Public site brand configuration.

SET NAMES utf8mb4;

USE `pvecloud`;

INSERT INTO `system_configs` (`config_key`, `config_value`, `value_type`, `group_name`, `is_secret`, `description`) VALUES
  ('site.name', 'pveCloud', 'string', '站点设置', 0, '站点名称'),
  ('site.logo_url', '', 'string', '站点设置', 0, 'Web 左上角 Logo 图片 URL')
ON DUPLICATE KEY UPDATE
  `value_type` = VALUES(`value_type`),
  `group_name` = VALUES(`group_name`),
  `is_secret` = VALUES(`is_secret`),
  `description` = VALUES(`description`);
