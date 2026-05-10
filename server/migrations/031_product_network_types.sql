-- Product network type schema and order snapshot fields.
-- Target: MariaDB 11.4.x / InnoDB / utf8mb4.

SET NAMES utf8mb4;

USE `pvecloud`;

CREATE TABLE IF NOT EXISTS `network_types` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '网络类型ID',
  `network_type_no` VARCHAR(64) NOT NULL COMMENT '对外网络类型编号',
  `code` VARCHAR(64) NOT NULL COMMENT '网络类型编码，后续可用于映射 PVE 网络',
  `name` VARCHAR(128) NOT NULL COMMENT '网络类型名称',
  `summary` VARCHAR(255) NULL COMMENT '网络类型简介',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '网络类型状态：active/inactive',
  `visible` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否 Web 展示',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_network_types_network_type_no` (`network_type_no`),
  UNIQUE KEY `uk_network_types_code` (`code`),
  KEY `idx_network_types_status_visible` (`status`, `visible`),
  KEY `idx_network_types_sort` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='网络类型';

CREATE TABLE IF NOT EXISTS `plan_network_types` (
  `plan_id` BIGINT UNSIGNED NOT NULL COMMENT '套餐ID',
  `network_type_id` BIGINT UNSIGNED NOT NULL COMMENT '网络类型ID',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '关联状态：active/inactive',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`plan_id`, `network_type_id`),
  KEY `idx_plan_network_types_network_type` (`network_type_id`),
  CONSTRAINT `fk_plan_network_types_plan` FOREIGN KEY (`plan_id`) REFERENCES `product_plans` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_plan_network_types_network_type` FOREIGN KEY (`network_type_id`) REFERENCES `network_types` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='套餐网络类型关联';

SET @orders_network_type_no_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'orders'
    AND COLUMN_NAME = 'network_type_no'
);
SET @add_orders_network_type_no_column_sql := IF(
  @orders_network_type_no_column_exists = 0,
  'ALTER TABLE `orders` ADD COLUMN `network_type_no` VARCHAR(64) NULL COMMENT ''网络类型编号快照'' AFTER `region_name`',
  'SELECT 1'
);
PREPARE add_orders_network_type_no_column_stmt FROM @add_orders_network_type_no_column_sql;
EXECUTE add_orders_network_type_no_column_stmt;
DEALLOCATE PREPARE add_orders_network_type_no_column_stmt;

SET @orders_network_type_code_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'orders'
    AND COLUMN_NAME = 'network_type_code'
);
SET @add_orders_network_type_code_column_sql := IF(
  @orders_network_type_code_column_exists = 0,
  'ALTER TABLE `orders` ADD COLUMN `network_type_code` VARCHAR(64) NULL COMMENT ''网络类型编码快照'' AFTER `network_type_no`',
  'SELECT 1'
);
PREPARE add_orders_network_type_code_column_stmt FROM @add_orders_network_type_code_column_sql;
EXECUTE add_orders_network_type_code_column_stmt;
DEALLOCATE PREPARE add_orders_network_type_code_column_stmt;

SET @orders_network_type_name_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'orders'
    AND COLUMN_NAME = 'network_type_name'
);
SET @add_orders_network_type_name_column_sql := IF(
  @orders_network_type_name_column_exists = 0,
  'ALTER TABLE `orders` ADD COLUMN `network_type_name` VARCHAR(128) NULL COMMENT ''网络类型名称快照'' AFTER `network_type_code`',
  'SELECT 1'
);
PREPARE add_orders_network_type_name_column_stmt FROM @add_orders_network_type_name_column_sql;
EXECUTE add_orders_network_type_name_column_stmt;
DEALLOCATE PREPARE add_orders_network_type_name_column_stmt;

INSERT INTO `network_types` (`network_type_no`, `code`, `name`, `summary`, `status`, `visible`, `sort_order`) VALUES
  ('NET-CLASSIC-001', 'classic', '经典网络', '默认经典网络类型，后续可映射到 PVE 网络', 'active', 1, 10)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `summary` = VALUES(`summary`),
  `status` = VALUES(`status`),
  `visible` = VALUES(`visible`),
  `sort_order` = VALUES(`sort_order`);

INSERT IGNORE INTO `plan_network_types` (`plan_id`, `network_type_id`, `status`, `sort_order`)
SELECT `product_plans`.`id`, `network_types`.`id`, 'active', 10
FROM `product_plans`
JOIN `network_types` ON `network_types`.`network_type_no` = 'NET-CLASSIC-001';
