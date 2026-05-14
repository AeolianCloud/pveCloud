-- Ticket enhancement contracts, collaboration tables, internal SLA fields and admin permissions.
-- Target: MariaDB 11.4.x / InnoDB / utf8mb4.

SET NAMES utf8mb4;

USE `pvecloud`;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'tickets' AND COLUMN_NAME = 'assignee_admin_id') = 0,
  'ALTER TABLE `tickets` ADD COLUMN `assignee_admin_id` BIGINT UNSIGNED NULL COMMENT ''当前处理管理员ID'' AFTER `status`',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'tickets' AND COLUMN_NAME = 'assigned_by_admin_id') = 0,
  'ALTER TABLE `tickets` ADD COLUMN `assigned_by_admin_id` BIGINT UNSIGNED NULL COMMENT ''最近指派管理员ID'' AFTER `assignee_admin_id`',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'tickets' AND COLUMN_NAME = 'assigned_at') = 0,
  'ALTER TABLE `tickets` ADD COLUMN `assigned_at` DATETIME(3) NULL COMMENT ''最近指派时间'' AFTER `assigned_by_admin_id`',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'tickets' AND COLUMN_NAME = 'first_response_due_at') = 0,
  'ALTER TABLE `tickets` ADD COLUMN `first_response_due_at` DATETIME(3) NULL COMMENT ''首次响应截止时间'' AFTER `last_admin_message_at`',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'tickets' AND COLUMN_NAME = 'first_responded_at') = 0,
  'ALTER TABLE `tickets` ADD COLUMN `first_responded_at` DATETIME(3) NULL COMMENT ''首次管理员响应时间'' AFTER `first_response_due_at`',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'tickets' AND COLUMN_NAME = 'resolution_due_at') = 0,
  'ALTER TABLE `tickets` ADD COLUMN `resolution_due_at` DATETIME(3) NULL COMMENT ''解决截止时间'' AFTER `first_responded_at`',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'tickets' AND COLUMN_NAME = 'resolved_at') = 0,
  'ALTER TABLE `tickets` ADD COLUMN `resolved_at` DATETIME(3) NULL COMMENT ''解决时间'' AFTER `resolution_due_at`',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @fk_exists := (
  SELECT COUNT(*)
  FROM information_schema.TABLE_CONSTRAINTS
  WHERE CONSTRAINT_SCHEMA = DATABASE()
    AND TABLE_NAME = 'tickets'
    AND CONSTRAINT_NAME = 'fk_tickets_assignee_admin'
);
SET @sql := IF(@fk_exists = 0,
  'ALTER TABLE `tickets` ADD CONSTRAINT `fk_tickets_assignee_admin` FOREIGN KEY (`assignee_admin_id`) REFERENCES `admin_users` (`id`)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @fk_exists := (
  SELECT COUNT(*)
  FROM information_schema.TABLE_CONSTRAINTS
  WHERE CONSTRAINT_SCHEMA = DATABASE()
    AND TABLE_NAME = 'tickets'
    AND CONSTRAINT_NAME = 'fk_tickets_assigned_by_admin'
);
SET @sql := IF(@fk_exists = 0,
  'ALTER TABLE `tickets` ADD CONSTRAINT `fk_tickets_assigned_by_admin` FOREIGN KEY (`assigned_by_admin_id`) REFERENCES `admin_users` (`id`)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'tickets' AND INDEX_NAME = 'idx_tickets_assignee_status_created') = 0,
  'CREATE INDEX `idx_tickets_assignee_status_created` ON `tickets` (`assignee_admin_id`, `status`, `created_at`)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'tickets' AND INDEX_NAME = 'idx_tickets_first_response_due') = 0,
  'CREATE INDEX `idx_tickets_first_response_due` ON `tickets` (`first_response_due_at`)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql := IF(
  (SELECT COUNT(*) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'tickets' AND INDEX_NAME = 'idx_tickets_resolution_due') = 0,
  'CREATE INDEX `idx_tickets_resolution_due` ON `tickets` (`resolution_due_at`)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `tickets`
SET
  `first_response_due_at` = CASE `priority`
    WHEN 'low' THEN DATE_ADD(`created_at`, INTERVAL 48 HOUR)
    WHEN 'high' THEN DATE_ADD(`created_at`, INTERVAL 8 HOUR)
    WHEN 'urgent' THEN DATE_ADD(`created_at`, INTERVAL 2 HOUR)
    ELSE DATE_ADD(`created_at`, INTERVAL 24 HOUR)
  END,
  `resolution_due_at` = CASE `priority`
    WHEN 'low' THEN DATE_ADD(`created_at`, INTERVAL 7 DAY)
    WHEN 'high' THEN DATE_ADD(`created_at`, INTERVAL 3 DAY)
    WHEN 'urgent' THEN DATE_ADD(`created_at`, INTERVAL 24 HOUR)
    ELSE DATE_ADD(`created_at`, INTERVAL 5 DAY)
  END
WHERE `first_response_due_at` IS NULL
   OR `resolution_due_at` IS NULL;

UPDATE `tickets`
SET `first_responded_at` = `last_admin_message_at`
WHERE `first_responded_at` IS NULL
  AND `last_admin_message_at` IS NOT NULL;

UPDATE `tickets`
SET `resolved_at` = `closed_at`
WHERE `resolved_at` IS NULL
  AND `closed_at` IS NOT NULL;

CREATE TABLE IF NOT EXISTS `ticket_tags` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '工单标签ID',
  `name` VARCHAR(40) NOT NULL COMMENT '标签名称',
  `color` VARCHAR(32) NULL COMMENT '标签颜色',
  `visibility` VARCHAR(32) NOT NULL DEFAULT 'internal' COMMENT '可见性：public/internal',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '状态：active/disabled',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  `created_by_admin_id` BIGINT UNSIGNED NULL COMMENT '创建管理员ID',
  `updated_by_admin_id` BIGINT UNSIGNED NULL COMMENT '更新管理员ID',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_ticket_tags_name` (`name`),
  KEY `idx_ticket_tags_visibility_status_sort` (`visibility`, `status`, `sort_order`),
  KEY `idx_ticket_tags_created_by` (`created_by_admin_id`),
  KEY `idx_ticket_tags_updated_by` (`updated_by_admin_id`),
  CONSTRAINT `fk_ticket_tags_created_by_admin` FOREIGN KEY (`created_by_admin_id`) REFERENCES `admin_users` (`id`),
  CONSTRAINT `fk_ticket_tags_updated_by_admin` FOREIGN KEY (`updated_by_admin_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工单标签字典';

CREATE TABLE IF NOT EXISTS `ticket_tag_bindings` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '工单标签绑定ID',
  `ticket_id` BIGINT UNSIGNED NOT NULL COMMENT '工单ID',
  `tag_id` BIGINT UNSIGNED NOT NULL COMMENT '标签ID',
  `created_by_admin_id` BIGINT UNSIGNED NULL COMMENT '绑定管理员ID',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_ticket_tag_bindings_ticket_tag` (`ticket_id`, `tag_id`),
  KEY `idx_ticket_tag_bindings_tag` (`tag_id`),
  KEY `idx_ticket_tag_bindings_created_by` (`created_by_admin_id`),
  CONSTRAINT `fk_ticket_tag_bindings_ticket` FOREIGN KEY (`ticket_id`) REFERENCES `tickets` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_ticket_tag_bindings_tag` FOREIGN KEY (`tag_id`) REFERENCES `ticket_tags` (`id`),
  CONSTRAINT `fk_ticket_tag_bindings_created_by_admin` FOREIGN KEY (`created_by_admin_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工单标签绑定';

CREATE TABLE IF NOT EXISTS `ticket_internal_notes` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '工单内部备注ID',
  `ticket_id` BIGINT UNSIGNED NOT NULL COMMENT '工单ID',
  `admin_id` BIGINT UNSIGNED NOT NULL COMMENT '备注管理员ID',
  `content` TEXT NOT NULL COMMENT '备注内容',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_ticket_internal_notes_ticket_created` (`ticket_id`, `created_at`),
  KEY `idx_ticket_internal_notes_admin` (`admin_id`),
  CONSTRAINT `fk_ticket_internal_notes_ticket` FOREIGN KEY (`ticket_id`) REFERENCES `tickets` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_ticket_internal_notes_admin` FOREIGN KEY (`admin_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工单内部备注';

CREATE TABLE IF NOT EXISTS `ticket_collaborators` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '工单协作者ID',
  `ticket_id` BIGINT UNSIGNED NOT NULL COMMENT '工单ID',
  `admin_id` BIGINT UNSIGNED NOT NULL COMMENT '协作管理员ID',
  `created_by_admin_id` BIGINT UNSIGNED NULL COMMENT '添加管理员ID',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_ticket_collaborators_ticket_admin` (`ticket_id`, `admin_id`),
  KEY `idx_ticket_collaborators_admin` (`admin_id`),
  KEY `idx_ticket_collaborators_created_by` (`created_by_admin_id`),
  CONSTRAINT `fk_ticket_collaborators_ticket` FOREIGN KEY (`ticket_id`) REFERENCES `tickets` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_ticket_collaborators_admin` FOREIGN KEY (`admin_id`) REFERENCES `admin_users` (`id`),
  CONSTRAINT `fk_ticket_collaborators_created_by_admin` FOREIGN KEY (`created_by_admin_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工单协作者';

CREATE TABLE IF NOT EXISTS `ticket_events` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '工单事件ID',
  `ticket_id` BIGINT UNSIGNED NOT NULL COMMENT '工单ID',
  `event_type` VARCHAR(64) NOT NULL COMMENT '事件类型',
  `actor_admin_id` BIGINT UNSIGNED NULL COMMENT '操作管理员ID',
  `actor_user_id` BIGINT UNSIGNED NULL COMMENT '操作用户ID',
  `before_data` JSON NULL COMMENT '变更前数据',
  `after_data` JSON NULL COMMENT '变更后数据',
  `remark` VARCHAR(500) NULL COMMENT '备注',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_ticket_events_ticket_created` (`ticket_id`, `created_at`),
  KEY `idx_ticket_events_type_created` (`event_type`, `created_at`),
  KEY `idx_ticket_events_actor_admin` (`actor_admin_id`),
  KEY `idx_ticket_events_actor_user` (`actor_user_id`),
  CONSTRAINT `fk_ticket_events_ticket` FOREIGN KEY (`ticket_id`) REFERENCES `tickets` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_ticket_events_actor_admin` FOREIGN KEY (`actor_admin_id`) REFERENCES `admin_users` (`id`),
  CONSTRAINT `fk_ticket_events_actor_user` FOREIGN KEY (`actor_user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工单操作历史';

INSERT INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`) VALUES
  ('ticket:assign', '指派工单', 'action', 'page.tickets', NULL, NULL, 130, 0, '工单管理', '指派和转派用户端工单'),
  ('ticket:collaborate', '工单协作', 'action', 'page.tickets', NULL, NULL, 140, 0, '工单管理', '维护工单协作者'),
  ('ticket:note', '内部备注', 'action', 'page.tickets', NULL, NULL, 150, 0, '工单管理', '追加工单内部备注'),
  ('ticket:priority', '升级优先级', 'action', 'page.tickets', NULL, NULL, 160, 0, '工单管理', '升级工单优先级'),
  ('ticket:tag', '绑定工单标签', 'action', 'page.tickets', NULL, NULL, 170, 0, '工单管理', '维护工单标签绑定'),
  ('ticket:tag-manage', '管理工单标签', 'action', 'page.tickets', NULL, NULL, 180, 0, '工单管理', '维护工单标签字典')
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
  AND `admin_permissions`.`code` IN (
    'ticket:assign',
    'ticket:collaborate',
    'ticket:note',
    'ticket:priority',
    'ticket:tag',
    'ticket:tag-manage'
  )
ON DUPLICATE KEY UPDATE
  `role_id` = VALUES(`role_id`);
