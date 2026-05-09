-- Keep real-name verification open with manual review fallback.

SET NAMES utf8mb4;

USE `pvecloud`;

ALTER TABLE `user_real_name_applications`
  MODIFY COLUMN `id_number_digest` CHAR(64) NULL COMMENT '证件号码查询摘要；人工审核且未配置摘要密钥时为空';

ALTER TABLE `user_real_name_applications`
  MODIFY COLUMN `id_number_digest_version` VARCHAR(32) NULL COMMENT '证件摘要版本；人工审核且未配置摘要密钥时为空';

INSERT INTO `system_configs` (`config_key`, `config_value`, `value_type`, `group_name`, `is_secret`, `description`) VALUES
  ('real_name.manual_review_enabled', 'true', 'bool', '实名设置', 0, '是否启用后台人工审核实名兜底')
ON DUPLICATE KEY UPDATE
  `value_type` = VALUES(`value_type`),
  `group_name` = VALUES(`group_name`),
  `is_secret` = VALUES(`is_secret`),
  `description` = VALUES(`description`);

INSERT INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`) VALUES
  ('real-name:review', '审核人工实名申请', 'action', 'page.real-name-management', NULL, NULL, 120, 0, '实名管理', '通过或拒绝人工审核实名申请')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `type` = VALUES(`type`),
  `parent_code` = VALUES(`parent_code`),
  `path` = VALUES(`path`),
  `icon` = VALUES(`icon`),
  `sort_order` = VALUES(`sort_order`),
  `visible_in_menu` = VALUES(`visible_in_menu`),
  `group_name` = VALUES(`group_name`),
  `description` = VALUES(`description`);

INSERT INTO `admin_role_permissions` (`role_id`, `permission_id`)
SELECT `admin_roles`.`id`, `admin_permissions`.`id`
FROM `admin_roles`
JOIN `admin_permissions`
WHERE `admin_roles`.`code` = 'super_admin'
  AND `admin_permissions`.`code` IN ('real-name:review')
ON DUPLICATE KEY UPDATE
  `role_id` = VALUES(`role_id`);
