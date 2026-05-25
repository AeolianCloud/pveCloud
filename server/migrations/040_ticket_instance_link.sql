-- Ticket instance association contract.
-- Target: MariaDB 11.4.x / InnoDB / utf8mb4.
--
-- This migration lets support tickets optionally reference a business instance
-- for troubleshooting. It only adds nullable association columns and indexes:
-- existing tickets remain valid and are not backfilled because historical
-- tickets did not capture an instance selection at creation time.

SET NAMES utf8mb4;

USE `pvecloud`;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'tickets' AND COLUMN_NAME = 'instance_id') = 0,
  'ALTER TABLE `tickets` ADD COLUMN `instance_id` BIGINT UNSIGNED NULL COMMENT ''关联实例ID'' AFTER `order_no`',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'tickets' AND COLUMN_NAME = 'instance_no') = 0,
  'ALTER TABLE `tickets` ADD COLUMN `instance_no` VARCHAR(64) NULL COMMENT ''关联实例编号快照'' AFTER `instance_id`',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'tickets' AND INDEX_NAME = 'idx_tickets_instance_id') = 0,
  'CREATE INDEX `idx_tickets_instance_id` ON `tickets` (`instance_id`)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'tickets' AND INDEX_NAME = 'idx_tickets_instance_no') = 0,
  'CREATE INDEX `idx_tickets_instance_no` ON `tickets` (`instance_no`)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @fk_exists := (
  SELECT COUNT(*)
  FROM information_schema.TABLE_CONSTRAINTS
  WHERE CONSTRAINT_SCHEMA = DATABASE()
    AND TABLE_NAME = 'tickets'
    AND CONSTRAINT_NAME = 'fk_tickets_instance'
);
SET @sql := IF(@fk_exists = 0,
  'ALTER TABLE `tickets` ADD CONSTRAINT `fk_tickets_instance` FOREIGN KEY (`instance_id`) REFERENCES `instances` (`id`)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
