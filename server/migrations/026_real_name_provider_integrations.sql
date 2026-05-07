-- Provider-backed real-name verification contracts.

SET NAMES utf8mb4;

USE `pvecloud`;

SET @real_name_digest_version_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_real_name_applications'
    AND COLUMN_NAME = 'id_number_digest_version'
);
SET @add_real_name_digest_version_column_sql := IF(
  @real_name_digest_version_column_exists = 0,
  'ALTER TABLE `user_real_name_applications` ADD COLUMN `id_number_digest_version` VARCHAR(32) NOT NULL DEFAULT ''sha256-legacy'' COMMENT ''证件摘要版本'' AFTER `id_number_digest`',
  'SELECT 1'
);
PREPARE add_real_name_digest_version_column_stmt FROM @add_real_name_digest_version_column_sql;
EXECUTE add_real_name_digest_version_column_stmt;
DEALLOCATE PREPARE add_real_name_digest_version_column_stmt;

SET @real_name_provider_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_real_name_applications'
    AND COLUMN_NAME = 'verification_provider'
);
SET @add_real_name_provider_column_sql := IF(
  @real_name_provider_column_exists = 0,
  'ALTER TABLE `user_real_name_applications` ADD COLUMN `verification_provider` VARCHAR(32) NULL COMMENT ''实名核验供应商：alipay/wechat'' AFTER `id_type`',
  'SELECT 1'
);
PREPARE add_real_name_provider_column_stmt FROM @add_real_name_provider_column_sql;
EXECUTE add_real_name_provider_column_stmt;
DEALLOCATE PREPARE add_real_name_provider_column_stmt;

SET @real_name_provider_app_id_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_real_name_applications'
    AND COLUMN_NAME = 'provider_application_id'
);
SET @add_real_name_provider_app_id_column_sql := IF(
  @real_name_provider_app_id_column_exists = 0,
  'ALTER TABLE `user_real_name_applications` ADD COLUMN `provider_application_id` VARCHAR(128) NULL COMMENT ''供应商实名会话ID'' AFTER `verification_provider`',
  'SELECT 1'
);
PREPARE add_real_name_provider_app_id_column_stmt FROM @add_real_name_provider_app_id_column_sql;
EXECUTE add_real_name_provider_app_id_column_stmt;
DEALLOCATE PREPARE add_real_name_provider_app_id_column_stmt;

SET @real_name_provider_status_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_real_name_applications'
    AND COLUMN_NAME = 'provider_status'
);
SET @add_real_name_provider_status_column_sql := IF(
  @real_name_provider_status_column_exists = 0,
  'ALTER TABLE `user_real_name_applications` ADD COLUMN `provider_status` VARCHAR(64) NULL COMMENT ''供应商状态'' AFTER `provider_application_id`',
  'SELECT 1'
);
PREPARE add_real_name_provider_status_column_stmt FROM @add_real_name_provider_status_column_sql;
EXECUTE add_real_name_provider_status_column_stmt;
DEALLOCATE PREPARE add_real_name_provider_status_column_stmt;

SET @real_name_provider_result_code_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_real_name_applications'
    AND COLUMN_NAME = 'provider_result_code'
);
SET @add_real_name_provider_result_code_column_sql := IF(
  @real_name_provider_result_code_column_exists = 0,
  'ALTER TABLE `user_real_name_applications` ADD COLUMN `provider_result_code` VARCHAR(128) NULL COMMENT ''供应商结果码'' AFTER `provider_status`',
  'SELECT 1'
);
PREPARE add_real_name_provider_result_code_column_stmt FROM @add_real_name_provider_result_code_column_sql;
EXECUTE add_real_name_provider_result_code_column_stmt;
DEALLOCATE PREPARE add_real_name_provider_result_code_column_stmt;

SET @real_name_provider_result_message_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_real_name_applications'
    AND COLUMN_NAME = 'provider_result_message'
);
SET @add_real_name_provider_result_message_column_sql := IF(
  @real_name_provider_result_message_column_exists = 0,
  'ALTER TABLE `user_real_name_applications` ADD COLUMN `provider_result_message` VARCHAR(500) NULL COMMENT ''供应商结果说明'' AFTER `provider_result_code`',
  'SELECT 1'
);
PREPARE add_real_name_provider_result_message_column_stmt FROM @add_real_name_provider_result_message_column_sql;
EXECUTE add_real_name_provider_result_message_column_stmt;
DEALLOCATE PREPARE add_real_name_provider_result_message_column_stmt;

SET @real_name_provider_started_at_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_real_name_applications'
    AND COLUMN_NAME = 'provider_started_at'
);
SET @add_real_name_provider_started_at_column_sql := IF(
  @real_name_provider_started_at_column_exists = 0,
  'ALTER TABLE `user_real_name_applications` ADD COLUMN `provider_started_at` DATETIME(3) NULL COMMENT ''供应商会话开始时间'' AFTER `provider_result_message`',
  'SELECT 1'
);
PREPARE add_real_name_provider_started_at_column_stmt FROM @add_real_name_provider_started_at_column_sql;
EXECUTE add_real_name_provider_started_at_column_stmt;
DEALLOCATE PREPARE add_real_name_provider_started_at_column_stmt;

SET @real_name_provider_finished_at_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_real_name_applications'
    AND COLUMN_NAME = 'provider_finished_at'
);
SET @add_real_name_provider_finished_at_column_sql := IF(
  @real_name_provider_finished_at_column_exists = 0,
  'ALTER TABLE `user_real_name_applications` ADD COLUMN `provider_finished_at` DATETIME(3) NULL COMMENT ''供应商核验完成时间'' AFTER `provider_started_at`',
  'SELECT 1'
);
PREPARE add_real_name_provider_finished_at_column_stmt FROM @add_real_name_provider_finished_at_column_sql;
EXECUTE add_real_name_provider_finished_at_column_stmt;
DEALLOCATE PREPARE add_real_name_provider_finished_at_column_stmt;

SET @real_name_provider_response_digest_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_real_name_applications'
    AND COLUMN_NAME = 'provider_response_digest'
);
SET @add_real_name_provider_response_digest_column_sql := IF(
  @real_name_provider_response_digest_column_exists = 0,
  'ALTER TABLE `user_real_name_applications` ADD COLUMN `provider_response_digest` CHAR(64) NULL COMMENT ''供应商响应摘要'' AFTER `provider_finished_at`',
  'SELECT 1'
);
PREPARE add_real_name_provider_response_digest_column_stmt FROM @add_real_name_provider_response_digest_column_sql;
EXECUTE add_real_name_provider_response_digest_column_stmt;
DEALLOCATE PREPARE add_real_name_provider_response_digest_column_stmt;

SET @real_name_provider_trace_id_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_real_name_applications'
    AND COLUMN_NAME = 'provider_trace_id'
);
SET @add_real_name_provider_trace_id_column_sql := IF(
  @real_name_provider_trace_id_column_exists = 0,
  'ALTER TABLE `user_real_name_applications` ADD COLUMN `provider_trace_id` VARCHAR(128) NULL COMMENT ''供应商链路ID'' AFTER `provider_response_digest`',
  'SELECT 1'
);
PREPARE add_real_name_provider_trace_id_column_stmt FROM @add_real_name_provider_trace_id_column_sql;
EXECUTE add_real_name_provider_trace_id_column_stmt;
DEALLOCATE PREPARE add_real_name_provider_trace_id_column_stmt;

SET @real_name_provider_app_index_exists := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_real_name_applications'
    AND INDEX_NAME = 'uk_user_real_name_provider_application'
);
SET @add_real_name_provider_app_index_sql := IF(
  @real_name_provider_app_index_exists = 0,
  'ALTER TABLE `user_real_name_applications` ADD UNIQUE KEY `uk_user_real_name_provider_application` (`verification_provider`, `provider_application_id`)',
  'SELECT 1'
);
PREPARE add_real_name_provider_app_index_stmt FROM @add_real_name_provider_app_index_sql;
EXECUTE add_real_name_provider_app_index_stmt;
DEALLOCATE PREPARE add_real_name_provider_app_index_stmt;

SET @real_name_provider_status_index_exists := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_real_name_applications'
    AND INDEX_NAME = 'idx_user_real_name_provider_status'
);
SET @add_real_name_provider_status_index_sql := IF(
  @real_name_provider_status_index_exists = 0,
  'ALTER TABLE `user_real_name_applications` ADD KEY `idx_user_real_name_provider_status` (`verification_provider`, `provider_status`, `created_at`)',
  'SELECT 1'
);
PREPARE add_real_name_provider_status_index_stmt FROM @add_real_name_provider_status_index_sql;
EXECUTE add_real_name_provider_status_index_stmt;
DEALLOCATE PREPARE add_real_name_provider_status_index_stmt;

INSERT INTO `system_configs` (`config_key`, `config_value`, `value_type`, `group_name`, `is_secret`, `description`) VALUES
  ('real_name.enabled', 'false', 'bool', '实名设置', 0, '是否开放用户端实名入口'),
  ('real_name.required_for_order', 'true', 'bool', '实名设置', 0, '购买机器前是否要求实名通过'),
  ('real_name.allowed_providers', 'alipay,wechat', 'string', '实名设置', 0, '允许用户选择的实名供应商，英文逗号分隔'),
  ('real_name.default_provider', 'alipay', 'string', '实名设置', 0, '默认实名供应商'),
  ('real_name.identity_digest_secret', '', 'string', '实名设置', 1, '证件号码HMAC摘要密钥，启用实名前必须在后台设置'),
  ('real_name.callback_base_url', '', 'string', '实名设置', 0, '实名供应商回调基础地址，例如 https://api.example.com/api/real-name/provider-callbacks'),
  ('real_name.resubmit_enabled', 'true', 'bool', '实名设置', 0, '核验失败后是否允许用户重新提交'),
  ('real_name.max_submit_attempts', '3', 'int', '实名设置', 0, '同一用户最大实名提交次数'),
  ('real_name.review_notice', '', 'string', '实名设置', 0, '用户端实名说明文案'),
  ('real_name.alipay.enabled', 'false', 'bool', '实名设置', 0, '是否启用支付宝实名供应商'),
  ('real_name.alipay.app_id', '', 'string', '实名设置', 0, '支付宝开放平台应用ID'),
  ('real_name.alipay.gateway_url', 'https://openapi.alipay.com/gateway.do', 'string', '实名设置', 0, '支付宝开放平台网关地址'),
  ('real_name.alipay.app_private_key', '', 'string', '实名设置', 1, '支付宝应用私钥'),
  ('real_name.alipay.alipay_public_key', '', 'string', '实名设置', 1, '支付宝公钥或证书公钥内容'),
  ('real_name.alipay.return_url', '', 'string', '实名设置', 0, '用户完成支付宝认证后的Web返回地址'),
  ('real_name.alipay.notify_url', '', 'string', '实名设置', 0, '支付宝异步通知地址；为空时由回调基础地址拼接'),
  ('real_name.wechat.enabled', 'false', 'bool', '实名设置', 0, '是否启用微信侧实名供应商'),
  ('real_name.wechat.secret_id', '', 'string', '实名设置', 1, '腾讯云SecretId'),
  ('real_name.wechat.secret_key', '', 'string', '实名设置', 1, '腾讯云SecretKey'),
  ('real_name.wechat.region', 'ap-guangzhou', 'string', '实名设置', 0, '腾讯云FaceID接口地域'),
  ('real_name.wechat.endpoint', 'faceid.tencentcloudapi.com', 'string', '实名设置', 0, '腾讯云FaceID接口端点'),
  ('real_name.wechat.rule_id', '', 'string', '实名设置', 0, '腾讯云实名核身规则ID或业务规则标识'),
  ('real_name.wechat.redirect_url', '', 'string', '实名设置', 0, '用户完成微信侧核验后的Web返回地址')
ON DUPLICATE KEY UPDATE
  `value_type` = VALUES(`value_type`),
  `group_name` = VALUES(`group_name`),
  `is_secret` = VALUES(`is_secret`),
  `description` = VALUES(`description`);

DELETE FROM `system_configs`
WHERE `config_key` IN (
  'real_name.manual_review_enabled',
  'real_name.id_card_front_required',
  'real_name.id_card_back_required',
  'real_name.hold_card_required',
  'real_name.image_max_size_mb',
  'real_name.allowed_image_types'
);

DELETE `admin_role_permissions`
FROM `admin_role_permissions`
JOIN `admin_permissions`
  ON `admin_permissions`.`id` = `admin_role_permissions`.`permission_id`
WHERE `admin_permissions`.`code` IN ('real-name:review', 'real-name:sensitive-view');

DELETE FROM `admin_permissions`
WHERE `code` IN ('real-name:review', 'real-name:sensitive-view');

INSERT INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`) VALUES
  ('real-name:sync', '同步实名供应商结果', 'action', 'page.real-name-management', NULL, NULL, 130, 0, '实名管理', '从支付宝或微信实名供应商同步核验结果')
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
  AND `admin_permissions`.`code` IN ('real-name:sync')
ON DUPLICATE KEY UPDATE
  `role_id` = VALUES(`role_id`);
