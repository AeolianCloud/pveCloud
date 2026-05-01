-- Enhance normal admin operation logs with request and actor context.
-- Target: MySQL 8.x / MariaDB 11.x / InnoDB / utf8mb4.

SET NAMES utf8mb4;

USE `pvecloud`;

SET @schema_name = DATABASE();

SET @sql = (
  SELECT IF(
    COUNT(*) = 0,
    'ALTER TABLE `admin_audit_logs` ADD COLUMN `admin_username` VARCHAR(64) NULL COMMENT ''操作发生时的管理员用户名快照'' AFTER `admin_id`',
    'SELECT 1'
  )
  FROM `information_schema`.`COLUMNS`
  WHERE `TABLE_SCHEMA` = @schema_name
    AND `TABLE_NAME` = 'admin_audit_logs'
    AND `COLUMN_NAME` = 'admin_username'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (
  SELECT IF(
    COUNT(*) = 0,
    'ALTER TABLE `admin_audit_logs` ADD COLUMN `admin_display_name` VARCHAR(64) NULL COMMENT ''操作发生时的管理员显示名快照'' AFTER `admin_username`',
    'SELECT 1'
  )
  FROM `information_schema`.`COLUMNS`
  WHERE `TABLE_SCHEMA` = @schema_name
    AND `TABLE_NAME` = 'admin_audit_logs'
    AND `COLUMN_NAME` = 'admin_display_name'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (
  SELECT IF(
    COUNT(*) = 0,
    'ALTER TABLE `admin_audit_logs` ADD COLUMN `session_id` VARCHAR(64) NULL COMMENT ''触发操作的管理端会话标识'' AFTER `admin_display_name`',
    'SELECT 1'
  )
  FROM `information_schema`.`COLUMNS`
  WHERE `TABLE_SCHEMA` = @schema_name
    AND `TABLE_NAME` = 'admin_audit_logs'
    AND `COLUMN_NAME` = 'session_id'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (
  SELECT IF(
    COUNT(*) = 0,
    'ALTER TABLE `admin_audit_logs` ADD COLUMN `request_id` VARCHAR(64) NULL COMMENT ''请求链路ID'' AFTER `session_id`',
    'SELECT 1'
  )
  FROM `information_schema`.`COLUMNS`
  WHERE `TABLE_SCHEMA` = @schema_name
    AND `TABLE_NAME` = 'admin_audit_logs'
    AND `COLUMN_NAME` = 'request_id'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (
  SELECT IF(
    COUNT(*) = 0,
    'ALTER TABLE `admin_audit_logs` ADD COLUMN `request_method` VARCHAR(16) NULL COMMENT ''后台请求方法'' AFTER `request_id`',
    'SELECT 1'
  )
  FROM `information_schema`.`COLUMNS`
  WHERE `TABLE_SCHEMA` = @schema_name
    AND `TABLE_NAME` = 'admin_audit_logs'
    AND `COLUMN_NAME` = 'request_method'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (
  SELECT IF(
    COUNT(*) = 0,
    'ALTER TABLE `admin_audit_logs` ADD COLUMN `request_path` VARCHAR(255) NULL COMMENT ''后台请求路径'' AFTER `request_method`',
    'SELECT 1'
  )
  FROM `information_schema`.`COLUMNS`
  WHERE `TABLE_SCHEMA` = @schema_name
    AND `TABLE_NAME` = 'admin_audit_logs'
    AND `COLUMN_NAME` = 'request_path'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (
  SELECT IF(
    COUNT(*) = 0,
    'ALTER TABLE `admin_audit_logs` ADD KEY `idx_admin_audit_logs_session_created` (`session_id`, `created_at`)',
    'SELECT 1'
  )
  FROM `information_schema`.`STATISTICS`
  WHERE `TABLE_SCHEMA` = @schema_name
    AND `TABLE_NAME` = 'admin_audit_logs'
    AND `INDEX_NAME` = 'idx_admin_audit_logs_session_created'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (
  SELECT IF(
    COUNT(*) = 0,
    'ALTER TABLE `admin_audit_logs` ADD KEY `idx_admin_audit_logs_request_id` (`request_id`)',
    'SELECT 1'
  )
  FROM `information_schema`.`STATISTICS`
  WHERE `TABLE_SCHEMA` = @schema_name
    AND `TABLE_NAME` = 'admin_audit_logs'
    AND `INDEX_NAME` = 'idx_admin_audit_logs_request_id'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
