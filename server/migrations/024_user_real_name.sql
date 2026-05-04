-- User real-name verification contracts and admin permissions.

SET NAMES utf8mb4;

USE `pvecloud`;

CREATE TABLE IF NOT EXISTS `user_real_name_applications` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '实名申请ID',
  `application_no` VARCHAR(64) NOT NULL COMMENT '实名申请编号',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `real_name` VARCHAR(64) NOT NULL COMMENT '真实姓名',
  `id_type` VARCHAR(32) NOT NULL DEFAULT 'id_card' COMMENT '证件类型：id_card',
  `id_number_digest` CHAR(64) NOT NULL COMMENT '证件号码查询摘要',
  `id_number_masked` VARCHAR(64) NOT NULL COMMENT '脱敏证件号码',
  `id_card_front_file_id` BIGINT UNSIGNED NULL COMMENT '身份证人像面附件ID',
  `id_card_back_file_id` BIGINT UNSIGNED NULL COMMENT '身份证国徽面附件ID',
  `hold_card_file_id` BIGINT UNSIGNED NULL COMMENT '手持证件附件ID',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态：pending/approved/rejected',
  `review_admin_id` BIGINT UNSIGNED NULL COMMENT '审核管理员ID',
  `reviewed_at` DATETIME(3) NULL COMMENT '审核时间',
  `reject_reason` VARCHAR(500) NULL COMMENT '拒绝原因',
  `submit_attempt` INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '用户提交次数快照',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_real_name_applications_no` (`application_no`),
  KEY `idx_user_real_name_applications_user_status` (`user_id`, `status`),
  KEY `idx_user_real_name_applications_status_created` (`status`, `created_at`),
  KEY `idx_user_real_name_applications_digest` (`id_number_digest`),
  CONSTRAINT `fk_real_name_applications_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_real_name_applications_front_file` FOREIGN KEY (`id_card_front_file_id`) REFERENCES `file_attachments` (`id`),
  CONSTRAINT `fk_real_name_applications_back_file` FOREIGN KEY (`id_card_back_file_id`) REFERENCES `file_attachments` (`id`),
  CONSTRAINT `fk_real_name_applications_hold_file` FOREIGN KEY (`hold_card_file_id`) REFERENCES `file_attachments` (`id`),
  CONSTRAINT `fk_real_name_applications_review_admin` FOREIGN KEY (`review_admin_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户实名申请';

SET @fk_file_uploader_exists := (
  SELECT COUNT(*)
  FROM information_schema.REFERENTIAL_CONSTRAINTS
  WHERE CONSTRAINT_SCHEMA = DATABASE()
    AND TABLE_NAME = 'file_attachments'
    AND CONSTRAINT_NAME = 'fk_file_attachments_uploader'
);
SET @drop_file_uploader_fk_sql := IF(
  @fk_file_uploader_exists > 0,
  'ALTER TABLE `file_attachments` DROP FOREIGN KEY `fk_file_attachments_uploader`',
  'SELECT 1'
);
PREPARE drop_file_uploader_fk_stmt FROM @drop_file_uploader_fk_sql;
EXECUTE drop_file_uploader_fk_stmt;
DEALLOCATE PREPARE drop_file_uploader_fk_stmt;

ALTER TABLE `file_attachments`
  MODIFY COLUMN `uploader_id` BIGINT UNSIGNED NULL COMMENT '上传者管理员ID';

SET @file_uploader_user_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'file_attachments'
    AND COLUMN_NAME = 'uploader_user_id'
);
SET @add_file_uploader_user_column_sql := IF(
  @file_uploader_user_column_exists = 0,
  'ALTER TABLE `file_attachments` ADD COLUMN `uploader_user_id` BIGINT UNSIGNED NULL COMMENT ''上传者用户ID'' AFTER `uploader_id`',
  'SELECT 1'
);
PREPARE add_file_uploader_user_column_stmt FROM @add_file_uploader_user_column_sql;
EXECUTE add_file_uploader_user_column_stmt;
DEALLOCATE PREPARE add_file_uploader_user_column_stmt;

SET @file_uploader_user_index_exists := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'file_attachments'
    AND INDEX_NAME = 'idx_file_attachments_uploader_user_id'
);
SET @add_file_uploader_user_index_sql := IF(
  @file_uploader_user_index_exists = 0,
  'ALTER TABLE `file_attachments` ADD KEY `idx_file_attachments_uploader_user_id` (`uploader_user_id`)',
  'SELECT 1'
);
PREPARE add_file_uploader_user_index_stmt FROM @add_file_uploader_user_index_sql;
EXECUTE add_file_uploader_user_index_stmt;
DEALLOCATE PREPARE add_file_uploader_user_index_stmt;

SET @fk_file_admin_uploader_exists := (
  SELECT COUNT(*)
  FROM information_schema.REFERENTIAL_CONSTRAINTS
  WHERE CONSTRAINT_SCHEMA = DATABASE()
    AND TABLE_NAME = 'file_attachments'
    AND CONSTRAINT_NAME = 'fk_file_attachments_uploader'
);
SET @add_file_admin_uploader_fk_sql := IF(
  @fk_file_admin_uploader_exists = 0,
  'ALTER TABLE `file_attachments` ADD CONSTRAINT `fk_file_attachments_uploader` FOREIGN KEY (`uploader_id`) REFERENCES `admin_users` (`id`)',
  'SELECT 1'
);
PREPARE add_file_admin_uploader_fk_stmt FROM @add_file_admin_uploader_fk_sql;
EXECUTE add_file_admin_uploader_fk_stmt;
DEALLOCATE PREPARE add_file_admin_uploader_fk_stmt;

SET @fk_file_user_uploader_exists := (
  SELECT COUNT(*)
  FROM information_schema.REFERENTIAL_CONSTRAINTS
  WHERE CONSTRAINT_SCHEMA = DATABASE()
    AND TABLE_NAME = 'file_attachments'
    AND CONSTRAINT_NAME = 'fk_file_attachments_uploader_user'
);
SET @add_file_user_uploader_fk_sql := IF(
  @fk_file_user_uploader_exists = 0,
  'ALTER TABLE `file_attachments` ADD CONSTRAINT `fk_file_attachments_uploader_user` FOREIGN KEY (`uploader_user_id`) REFERENCES `users` (`id`)',
  'SELECT 1'
);
PREPARE add_file_user_uploader_fk_stmt FROM @add_file_user_uploader_fk_sql;
EXECUTE add_file_user_uploader_fk_stmt;
DEALLOCATE PREPARE add_file_user_uploader_fk_stmt;

INSERT INTO `system_configs` (`config_key`, `config_value`, `value_type`, `group_name`, `is_secret`, `description`) VALUES
  ('real_name.enabled', 'false', 'bool', '实名设置', 0, '是否开放用户端实名入口'),
  ('real_name.required_for_order', 'true', 'bool', '实名设置', 0, '购买机器前是否要求实名通过'),
  ('real_name.manual_review_enabled', 'true', 'bool', '实名设置', 0, '是否启用后台人工审核'),
  ('real_name.resubmit_enabled', 'true', 'bool', '实名设置', 0, '审核拒绝后是否允许用户重新提交'),
  ('real_name.max_submit_attempts', '3', 'int', '实名设置', 0, '同一用户最大实名提交次数'),
  ('real_name.id_card_front_required', 'true', 'bool', '实名设置', 0, '是否要求身份证人像面图片'),
  ('real_name.id_card_back_required', 'true', 'bool', '实名设置', 0, '是否要求身份证国徽面图片'),
  ('real_name.hold_card_required', 'false', 'bool', '实名设置', 0, '是否要求手持证件图片'),
  ('real_name.image_max_size_mb', '5', 'int', '实名设置', 0, '实名图片单张最大尺寸MB'),
  ('real_name.allowed_image_types', 'image/jpeg,image/png,image/webp', 'string', '实名设置', 0, '实名图片允许的MIME类型，英文逗号分隔'),
  ('real_name.review_notice', '', 'string', '实名设置', 0, '用户端实名说明文案')
ON DUPLICATE KEY UPDATE
  `value_type` = VALUES(`value_type`),
  `group_name` = VALUES(`group_name`),
  `is_secret` = VALUES(`is_secret`),
  `description` = VALUES(`description`);

INSERT INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`) VALUES
  ('page.real-name-management', '实名管理', 'menu', NULL, '/web/real-names', 'Checked', 50, 1, '菜单', '显示实名管理菜单和页面入口'),
  ('real-name:*', '实名管理全权限', 'action', 'page.real-name-management', NULL, NULL, 100, 0, '实名管理', '实名管理模块全部能力'),
  ('real-name:view', '查看实名申请', 'action', 'page.real-name-management', NULL, NULL, 110, 0, '实名管理', '查看实名申请列表和详情'),
  ('real-name:review', '审核实名申请', 'action', 'page.real-name-management', NULL, NULL, 120, 0, '实名管理', '审核通过或拒绝实名申请')
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
ON DUPLICATE KEY UPDATE
  `role_id` = VALUES(`role_id`);
