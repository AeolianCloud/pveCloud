-- Worker, instance lifecycle, renewal order and notification schema.
-- Target: MariaDB 11.4.x / InnoDB / utf8mb4.

SET NAMES utf8mb4;

USE `pvecloud`;

SET @orders_order_type_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'orders'
    AND COLUMN_NAME = 'order_type'
);
SET @add_orders_order_type_sql := IF(
  @orders_order_type_column_exists = 0,
  'ALTER TABLE `orders` ADD COLUMN `order_type` VARCHAR(32) NOT NULL DEFAULT ''purchase'' COMMENT ''订单类型：purchase/renewal'' AFTER `status`',
  'SELECT 1'
);
PREPARE add_orders_order_type_stmt FROM @add_orders_order_type_sql;
EXECUTE add_orders_order_type_stmt;
DEALLOCATE PREPARE add_orders_order_type_stmt;

SET @orders_related_instance_no_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'orders'
    AND COLUMN_NAME = 'related_instance_no'
);
SET @add_orders_related_instance_no_sql := IF(
  @orders_related_instance_no_column_exists = 0,
  'ALTER TABLE `orders` ADD COLUMN `related_instance_no` VARCHAR(64) NULL COMMENT ''续费关联实例编号'' AFTER `order_type`',
  'SELECT 1'
);
PREPARE add_orders_related_instance_no_stmt FROM @add_orders_related_instance_no_sql;
EXECUTE add_orders_related_instance_no_stmt;
DEALLOCATE PREPARE add_orders_related_instance_no_stmt;

SET @orders_payment_status_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'orders'
    AND COLUMN_NAME = 'payment_status'
);
SET @add_orders_payment_status_sql := IF(
  @orders_payment_status_column_exists = 0,
  'ALTER TABLE `orders` ADD COLUMN `payment_status` VARCHAR(32) NOT NULL DEFAULT ''unpaid'' COMMENT ''支付状态：unpaid/paid/manual_confirmed'' AFTER `total_amount_cents`',
  'SELECT 1'
);
PREPARE add_orders_payment_status_stmt FROM @add_orders_payment_status_sql;
EXECUTE add_orders_payment_status_stmt;
DEALLOCATE PREPARE add_orders_payment_status_stmt;

SET @orders_paid_at_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'orders'
    AND COLUMN_NAME = 'paid_at'
);
SET @add_orders_paid_at_sql := IF(
  @orders_paid_at_column_exists = 0,
  'ALTER TABLE `orders` ADD COLUMN `paid_at` DATETIME(3) NULL COMMENT ''支付或人工确认时间'' AFTER `payment_status`',
  'SELECT 1'
);
PREPARE add_orders_paid_at_stmt FROM @add_orders_paid_at_sql;
EXECUTE add_orders_paid_at_stmt;
DEALLOCATE PREPARE add_orders_paid_at_stmt;

SET @orders_payment_provider_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'orders'
    AND COLUMN_NAME = 'payment_provider'
);
SET @add_orders_payment_provider_sql := IF(
  @orders_payment_provider_column_exists = 0,
  'ALTER TABLE `orders` ADD COLUMN `payment_provider` VARCHAR(32) NULL COMMENT ''支付供应商预留'' AFTER `paid_at`',
  'SELECT 1'
);
PREPARE add_orders_payment_provider_stmt FROM @add_orders_payment_provider_sql;
EXECUTE add_orders_payment_provider_stmt;
DEALLOCATE PREPARE add_orders_payment_provider_stmt;

SET @orders_payment_trade_no_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'orders'
    AND COLUMN_NAME = 'payment_trade_no'
);
SET @add_orders_payment_trade_no_sql := IF(
  @orders_payment_trade_no_column_exists = 0,
  'ALTER TABLE `orders` ADD COLUMN `payment_trade_no` VARCHAR(128) NULL COMMENT ''支付交易号预留'' AFTER `payment_provider`',
  'SELECT 1'
);
PREPARE add_orders_payment_trade_no_stmt FROM @add_orders_payment_trade_no_sql;
EXECUTE add_orders_payment_trade_no_stmt;
DEALLOCATE PREPARE add_orders_payment_trade_no_stmt;

SET @orders_payment_callback_payload_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'orders'
    AND COLUMN_NAME = 'payment_callback_payload'
);
SET @add_orders_payment_callback_payload_sql := IF(
  @orders_payment_callback_payload_column_exists = 0,
  'ALTER TABLE `orders` ADD COLUMN `payment_callback_payload` TEXT NULL COMMENT ''支付回调摘要预留，不保存敏感原文'' AFTER `payment_trade_no`',
  'SELECT 1'
);
PREPARE add_orders_payment_callback_payload_stmt FROM @add_orders_payment_callback_payload_sql;
EXECUTE add_orders_payment_callback_payload_stmt;
DEALLOCATE PREPARE add_orders_payment_callback_payload_stmt;

SET @idx_orders_type_status_created_exists := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'orders'
    AND INDEX_NAME = 'idx_orders_type_status_created'
);
SET @add_idx_orders_type_status_created_sql := IF(
  @idx_orders_type_status_created_exists = 0,
  'ALTER TABLE `orders` ADD KEY `idx_orders_type_status_created` (`order_type`, `status`, `created_at`)',
  'SELECT 1'
);
PREPARE add_idx_orders_type_status_created_stmt FROM @add_idx_orders_type_status_created_sql;
EXECUTE add_idx_orders_type_status_created_stmt;
DEALLOCATE PREPARE add_idx_orders_type_status_created_stmt;

SET @idx_orders_related_instance_exists := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'orders'
    AND INDEX_NAME = 'idx_orders_related_instance'
);
SET @add_idx_orders_related_instance_sql := IF(
  @idx_orders_related_instance_exists = 0,
  'ALTER TABLE `orders` ADD KEY `idx_orders_related_instance` (`related_instance_no`, `order_type`, `status`)',
  'SELECT 1'
);
PREPARE add_idx_orders_related_instance_stmt FROM @add_idx_orders_related_instance_sql;
EXECUTE add_idx_orders_related_instance_stmt;
DEALLOCATE PREPARE add_idx_orders_related_instance_stmt;

SET @idx_orders_payment_status_exists := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'orders'
    AND INDEX_NAME = 'idx_orders_payment_status'
);
SET @add_idx_orders_payment_status_sql := IF(
  @idx_orders_payment_status_exists = 0,
  'ALTER TABLE `orders` ADD KEY `idx_orders_payment_status` (`payment_status`, `created_at`)',
  'SELECT 1'
);
PREPARE add_idx_orders_payment_status_stmt FROM @add_idx_orders_payment_status_sql;
EXECUTE add_idx_orders_payment_status_stmt;
DEALLOCATE PREPARE add_idx_orders_payment_status_stmt;

SET @instances_service_started_at_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'instances'
    AND COLUMN_NAME = 'service_started_at'
);
SET @add_instances_service_started_at_sql := IF(
  @instances_service_started_at_column_exists = 0,
  'ALTER TABLE `instances` ADD COLUMN `service_started_at` DATETIME(3) NULL COMMENT ''服务开始时间'' AFTER `last_error_message`',
  'SELECT 1'
);
PREPARE add_instances_service_started_at_stmt FROM @add_instances_service_started_at_sql;
EXECUTE add_instances_service_started_at_stmt;
DEALLOCATE PREPARE add_instances_service_started_at_stmt;

SET @instances_expires_at_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'instances'
    AND COLUMN_NAME = 'expires_at'
);
SET @add_instances_expires_at_sql := IF(
  @instances_expires_at_column_exists = 0,
  'ALTER TABLE `instances` ADD COLUMN `expires_at` DATETIME(3) NULL COMMENT ''服务到期时间'' AFTER `service_started_at`',
  'SELECT 1'
);
PREPARE add_instances_expires_at_stmt FROM @add_instances_expires_at_sql;
EXECUTE add_instances_expires_at_stmt;
DEALLOCATE PREPARE add_instances_expires_at_stmt;

SET @instances_expire_notice_sent_at_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'instances'
    AND COLUMN_NAME = 'expire_notice_sent_at'
);
SET @add_instances_expire_notice_sent_at_sql := IF(
  @instances_expire_notice_sent_at_column_exists = 0,
  'ALTER TABLE `instances` ADD COLUMN `expire_notice_sent_at` DATETIME(3) NULL COMMENT ''最近到期提醒发送时间'' AFTER `expires_at`',
  'SELECT 1'
);
PREPARE add_instances_expire_notice_sent_at_stmt FROM @add_instances_expire_notice_sent_at_sql;
EXECUTE add_instances_expire_notice_sent_at_stmt;
DEALLOCATE PREPARE add_instances_expire_notice_sent_at_stmt;

SET @instances_expire_release_scheduled_at_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'instances'
    AND COLUMN_NAME = 'expire_release_scheduled_at'
);
SET @add_instances_expire_release_scheduled_at_sql := IF(
  @instances_expire_release_scheduled_at_column_exists = 0,
  'ALTER TABLE `instances` ADD COLUMN `expire_release_scheduled_at` DATETIME(3) NULL COMMENT ''到期自动释放计划时间'' AFTER `expire_notice_sent_at`',
  'SELECT 1'
);
PREPARE add_instances_expire_release_scheduled_at_stmt FROM @add_instances_expire_release_scheduled_at_sql;
EXECUTE add_instances_expire_release_scheduled_at_stmt;
DEALLOCATE PREPARE add_instances_expire_release_scheduled_at_stmt;

SET @instances_expire_released_at_column_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'instances'
    AND COLUMN_NAME = 'expire_released_at'
);
SET @add_instances_expire_released_at_sql := IF(
  @instances_expire_released_at_column_exists = 0,
  'ALTER TABLE `instances` ADD COLUMN `expire_released_at` DATETIME(3) NULL COMMENT ''因到期自动释放完成时间'' AFTER `expire_release_scheduled_at`',
  'SELECT 1'
);
PREPARE add_instances_expire_released_at_stmt FROM @add_instances_expire_released_at_sql;
EXECUTE add_instances_expire_released_at_stmt;
DEALLOCATE PREPARE add_instances_expire_released_at_stmt;

SET @idx_instances_expires_status_exists := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'instances'
    AND INDEX_NAME = 'idx_instances_expires_status'
);
SET @add_idx_instances_expires_status_sql := IF(
  @idx_instances_expires_status_exists = 0,
  'ALTER TABLE `instances` ADD KEY `idx_instances_expires_status` (`expires_at`, `status`)',
  'SELECT 1'
);
PREPARE add_idx_instances_expires_status_stmt FROM @add_idx_instances_expires_status_sql;
EXECUTE add_idx_instances_expires_status_stmt;
DEALLOCATE PREPARE add_idx_instances_expires_status_stmt;

SET @idx_instances_expire_release_exists := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'instances'
    AND INDEX_NAME = 'idx_instances_expire_release'
);
SET @add_idx_instances_expire_release_sql := IF(
  @idx_instances_expire_release_exists = 0,
  'ALTER TABLE `instances` ADD KEY `idx_instances_expire_release` (`expire_release_scheduled_at`, `status`)',
  'SELECT 1'
);
PREPARE add_idx_instances_expire_release_stmt FROM @add_idx_instances_expire_release_sql;
EXECUTE add_idx_instances_expire_release_stmt;
DEALLOCATE PREPARE add_idx_instances_expire_release_stmt;

CREATE TABLE IF NOT EXISTS `async_tasks` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '异步任务ID',
  `task_no` VARCHAR(64) NOT NULL COMMENT '对外任务编号',
  `task_type` VARCHAR(64) NOT NULL COMMENT '任务类型',
  `idempotency_key` VARCHAR(191) NULL COMMENT '幂等键',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态：pending/running/succeeded/failed/cancelled',
  `idempotency_active_key` VARCHAR(191) GENERATED ALWAYS AS (
    CASE
      WHEN `idempotency_key` IS NOT NULL AND `status` <> 'cancelled' THEN `idempotency_key`
      ELSE NULL
    END
  ) STORED COMMENT '未取消任务幂等投影',
  `object_type` VARCHAR(64) NULL COMMENT '业务对象类型',
  `object_no` VARCHAR(64) NULL COMMENT '业务对象编号',
  `payload` JSON NULL COMMENT '任务输入，不保存敏感原文',
  `result` JSON NULL COMMENT '任务结果摘要，不保存完整上游响应',
  `attempts` INT NOT NULL DEFAULT 0 COMMENT '已尝试次数',
  `max_attempts` INT NOT NULL DEFAULT 10 COMMENT '最大尝试次数',
  `scheduled_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '计划执行时间',
  `locked_by` VARCHAR(128) NULL COMMENT '领取 Worker ID',
  `locked_until` DATETIME(3) NULL COMMENT '领取锁过期时间',
  `last_error_code` VARCHAR(64) NULL COMMENT '最近错误码',
  `last_error_message` VARCHAR(500) NULL COMMENT '最近错误说明',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `completed_at` DATETIME(3) NULL COMMENT '完成时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_async_tasks_task_no` (`task_no`),
  UNIQUE KEY `uk_async_tasks_active_idempotency` (`task_type`, `idempotency_active_key`),
  KEY `idx_async_tasks_pickup` (`status`, `scheduled_at`, `locked_until`),
  KEY `idx_async_tasks_object` (`object_type`, `object_no`),
  KEY `idx_async_tasks_type_status` (`task_type`, `status`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='异步任务';

CREATE TABLE IF NOT EXISTS `notifications` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '通知ID',
  `notification_no` VARCHAR(64) NOT NULL COMMENT '对外通知编号',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `channel` VARCHAR(32) NOT NULL COMMENT '通道：email/sms',
  `scene` VARCHAR(64) NOT NULL COMMENT '场景',
  `target` VARCHAR(191) NOT NULL COMMENT '目标地址或手机号摘要',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态：pending/sent/failed/skipped',
  `subject` VARCHAR(191) NULL COMMENT '标题',
  `content_summary` VARCHAR(500) NULL COMMENT '内容摘要',
  `related_object_type` VARCHAR(64) NULL COMMENT '关联对象类型',
  `related_object_no` VARCHAR(64) NULL COMMENT '关联对象编号',
  `task_no` VARCHAR(64) NULL COMMENT '关联任务编号',
  `error_code` VARCHAR(64) NULL COMMENT '错误码',
  `error_message` VARCHAR(500) NULL COMMENT '错误说明',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `sent_at` DATETIME(3) NULL COMMENT '发送完成时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_notifications_notification_no` (`notification_no`),
  KEY `idx_notifications_user_created` (`user_id`, `created_at`),
  KEY `idx_notifications_scene_object` (`scene`, `related_object_type`, `related_object_no`),
  KEY `idx_notifications_task` (`task_no`),
  CONSTRAINT `fk_notifications_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知记录';

INSERT INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`) VALUES
  ('page.async-tasks', '异步任务', 'menu', NULL, '/async-tasks', 'Timer', 85, 1, '菜单', '显示异步任务管理菜单和页面入口'),
  ('async-task:*', '异步任务全权限', 'action', 'page.async-tasks', NULL, NULL, 100, 0, '异步任务', '异步任务查看和重试全部能力'),
  ('async-task:retry', '重试异步任务', 'action', 'page.async-tasks', NULL, NULL, 110, 0, '异步任务', '重试失败异步任务'),
  ('instance:renew', '实例续期', 'action', 'page.instances', NULL, NULL, 160, 0, '实例管理', '调整实例到期时间和处理续费')
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
