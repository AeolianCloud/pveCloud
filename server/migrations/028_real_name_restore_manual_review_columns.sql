-- Restore manual review columns for environments that applied an earlier cleanup.

SET NAMES utf8mb4;

USE `pvecloud`;

SET @real_name_review_admin_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_real_name_applications'
    AND COLUMN_NAME = 'review_admin_id'
);
SET @add_real_name_review_admin_column_sql := IF(
  @real_name_review_admin_column_exists = 0,
  'ALTER TABLE `user_real_name_applications` ADD COLUMN `review_admin_id` BIGINT UNSIGNED NULL COMMENT ''审核管理员ID'' AFTER `status`',
  'SELECT 1'
);
PREPARE add_real_name_review_admin_column_stmt FROM @add_real_name_review_admin_column_sql;
EXECUTE add_real_name_review_admin_column_stmt;
DEALLOCATE PREPARE add_real_name_review_admin_column_stmt;

SET @real_name_reviewed_at_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_real_name_applications'
    AND COLUMN_NAME = 'reviewed_at'
);
SET @add_real_name_reviewed_at_column_sql := IF(
  @real_name_reviewed_at_column_exists = 0,
  'ALTER TABLE `user_real_name_applications` ADD COLUMN `reviewed_at` DATETIME(3) NULL COMMENT ''审核时间'' AFTER `review_admin_id`',
  'SELECT 1'
);
PREPARE add_real_name_reviewed_at_column_stmt FROM @add_real_name_reviewed_at_column_sql;
EXECUTE add_real_name_reviewed_at_column_stmt;
DEALLOCATE PREPARE add_real_name_reviewed_at_column_stmt;

SET @real_name_review_admin_fk_exists := (
  SELECT COUNT(*)
  FROM information_schema.REFERENTIAL_CONSTRAINTS
  WHERE CONSTRAINT_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_real_name_applications'
    AND CONSTRAINT_NAME = 'fk_real_name_applications_review_admin'
);
SET @add_real_name_review_admin_fk_sql := IF(
  @real_name_review_admin_fk_exists = 0,
  'ALTER TABLE `user_real_name_applications` ADD CONSTRAINT `fk_real_name_applications_review_admin` FOREIGN KEY (`review_admin_id`) REFERENCES `admin_users` (`id`)',
  'SELECT 1'
);
PREPARE add_real_name_review_admin_fk_stmt FROM @add_real_name_review_admin_fk_sql;
EXECUTE add_real_name_review_admin_fk_stmt;
DEALLOCATE PREPARE add_real_name_review_admin_fk_stmt;
