-- Enforce one approved real-name record per identity digest.

SET NAMES utf8mb4;

USE `pvecloud`;

SET @approved_digest_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_real_name_applications'
    AND COLUMN_NAME = 'approved_id_number_digest'
);
SET @add_approved_digest_column_sql := IF(
  @approved_digest_column_exists = 0,
  'ALTER TABLE `user_real_name_applications` ADD COLUMN `approved_id_number_digest` CHAR(64) AS (CASE WHEN `status` = ''approved'' THEN `id_number_digest` ELSE NULL END) STORED COMMENT ''已通过实名证件摘要唯一键'' AFTER `id_number_digest`',
  'SELECT 1'
);
PREPARE add_approved_digest_column_stmt FROM @add_approved_digest_column_sql;
EXECUTE add_approved_digest_column_stmt;
DEALLOCATE PREPARE add_approved_digest_column_stmt;

SET @approved_digest_index_exists := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_real_name_applications'
    AND INDEX_NAME = 'uk_user_real_name_approved_digest'
);
SET @add_approved_digest_index_sql := IF(
  @approved_digest_index_exists = 0,
  'ALTER TABLE `user_real_name_applications` ADD UNIQUE KEY `uk_user_real_name_approved_digest` (`approved_id_number_digest`)',
  'SELECT 1'
);
PREPARE add_approved_digest_index_stmt FROM @add_approved_digest_index_sql;
EXECUTE add_approved_digest_index_stmt;
DEALLOCATE PREPARE add_approved_digest_index_stmt;
