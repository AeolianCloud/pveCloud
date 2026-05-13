-- Ticket MVP contracts and admin permissions.
-- Target: MariaDB 11.4.x / InnoDB / utf8mb4.

SET NAMES utf8mb4;

USE `pvecloud`;

CREATE TABLE IF NOT EXISTS `tickets` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '工单ID',
  `ticket_no` VARCHAR(64) NOT NULL COMMENT '工单编号',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `order_id` BIGINT UNSIGNED NULL COMMENT '关联订单ID',
  `order_no` VARCHAR(64) NULL COMMENT '关联订单号快照',
  `category` VARCHAR(32) NOT NULL COMMENT '分类：account/order/product/technical/billing/other',
  `priority` VARCHAR(32) NOT NULL DEFAULT 'normal' COMMENT '优先级：low/normal/high/urgent',
  `title` VARCHAR(160) NOT NULL COMMENT '标题',
  `status` VARCHAR(32) NOT NULL DEFAULT 'waiting_admin' COMMENT '状态：waiting_admin/waiting_user/closed',
  `last_message_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '最近消息时间',
  `last_user_message_at` DATETIME(3) NULL COMMENT '最近用户回复时间',
  `last_admin_message_at` DATETIME(3) NULL COMMENT '最近管理员回复时间',
  `closed_by_type` VARCHAR(32) NULL COMMENT '关闭方：user/admin',
  `closed_by_user_id` BIGINT UNSIGNED NULL COMMENT '关闭用户ID',
  `closed_by_admin_id` BIGINT UNSIGNED NULL COMMENT '关闭管理员ID',
  `closed_at` DATETIME(3) NULL COMMENT '关闭时间',
  `close_reason` VARCHAR(500) NULL COMMENT '关闭原因',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tickets_ticket_no` (`ticket_no`),
  KEY `idx_tickets_user_status_created` (`user_id`, `status`, `created_at`),
  KEY `idx_tickets_status_priority_created` (`status`, `priority`, `created_at`),
  KEY `idx_tickets_category_created` (`category`, `created_at`),
  KEY `idx_tickets_order_no` (`order_no`),
  KEY `idx_tickets_last_message_at` (`last_message_at`),
  CONSTRAINT `fk_tickets_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_tickets_order` FOREIGN KEY (`order_id`) REFERENCES `orders` (`id`),
  CONSTRAINT `fk_tickets_closed_user` FOREIGN KEY (`closed_by_user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_tickets_closed_admin` FOREIGN KEY (`closed_by_admin_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工单';

CREATE TABLE IF NOT EXISTS `ticket_messages` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '工单消息ID',
  `ticket_id` BIGINT UNSIGNED NOT NULL COMMENT '工单ID',
  `sender_type` VARCHAR(32) NOT NULL COMMENT '发送方：user/admin',
  `sender_user_id` BIGINT UNSIGNED NULL COMMENT '发送用户ID',
  `sender_admin_id` BIGINT UNSIGNED NULL COMMENT '发送管理员ID',
  `content` TEXT NOT NULL COMMENT '消息内容',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_ticket_messages_ticket_created` (`ticket_id`, `created_at`),
  KEY `idx_ticket_messages_sender_user` (`sender_user_id`),
  KEY `idx_ticket_messages_sender_admin` (`sender_admin_id`),
  CONSTRAINT `fk_ticket_messages_ticket` FOREIGN KEY (`ticket_id`) REFERENCES `tickets` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_ticket_messages_sender_user` FOREIGN KEY (`sender_user_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_ticket_messages_sender_admin` FOREIGN KEY (`sender_admin_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工单消息';

CREATE TABLE IF NOT EXISTS `ticket_message_attachments` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '工单消息附件ID',
  `ticket_id` BIGINT UNSIGNED NOT NULL COMMENT '工单ID',
  `message_id` BIGINT UNSIGNED NOT NULL COMMENT '工单消息ID',
  `file_id` BIGINT UNSIGNED NOT NULL COMMENT '附件ID',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_ticket_message_attachments_message_file` (`message_id`, `file_id`),
  KEY `idx_ticket_message_attachments_ticket` (`ticket_id`),
  KEY `idx_ticket_message_attachments_file` (`file_id`),
  CONSTRAINT `fk_ticket_message_attachments_ticket` FOREIGN KEY (`ticket_id`) REFERENCES `tickets` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_ticket_message_attachments_message` FOREIGN KEY (`message_id`) REFERENCES `ticket_messages` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_ticket_message_attachments_file` FOREIGN KEY (`file_id`) REFERENCES `file_attachments` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工单消息附件';

INSERT INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`) VALUES
  ('page.tickets', '工单管理', 'menu', NULL, '/tickets', 'Chatbubbles', 70, 1, '菜单', '显示工单管理菜单和页面入口'),
  ('ticket:*', '工单全权限', 'action', 'page.tickets', NULL, NULL, 100, 0, '工单管理', '工单模块全部处理能力'),
  ('ticket:reply', '回复工单', 'action', 'page.tickets', NULL, NULL, 110, 0, '工单管理', '回复用户端工单'),
  ('ticket:close', '关闭工单', 'action', 'page.tickets', NULL, NULL, 120, 0, '工单管理', '关闭用户端工单')
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
  AND `admin_permissions`.`code` IN ('page.tickets', 'ticket:*', 'ticket:reply', 'ticket:close')
ON DUPLICATE KEY UPDATE
  `role_id` = VALUES(`role_id`);
