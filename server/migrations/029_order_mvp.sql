-- Order MVP schema.
-- Target: MariaDB 11.4.x / InnoDB / utf8mb4.

SET NAMES utf8mb4;

USE `pvecloud`;

CREATE TABLE IF NOT EXISTS `orders` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `order_no` VARCHAR(64) NOT NULL COMMENT '对外订单编号',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `client_token` VARCHAR(128) NOT NULL COMMENT '用户端幂等键',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '订单状态：pending/cancelled/closed',
  `product_no` VARCHAR(64) NOT NULL COMMENT '产品编号快照',
  `product_type` VARCHAR(32) NOT NULL COMMENT '产品类型快照',
  `product_name` VARCHAR(128) NOT NULL COMMENT '产品名称快照',
  `product_summary` VARCHAR(255) NULL COMMENT '产品简介快照',
  `plan_no` VARCHAR(64) NOT NULL COMMENT '套餐编号快照',
  `plan_code` VARCHAR(96) NOT NULL COMMENT '套餐编码快照',
  `plan_name` VARCHAR(128) NOT NULL COMMENT '套餐名称快照',
  `plan_summary` VARCHAR(255) NULL COMMENT '套餐简介快照',
  `cpu_cores` INT NOT NULL COMMENT 'CPU 核数快照',
  `memory_mb` INT NOT NULL COMMENT '内存 MB 快照',
  `system_disk_gb` INT NOT NULL COMMENT '系统盘 GB 快照',
  `data_disk_gb` INT NOT NULL DEFAULT 0 COMMENT '数据盘 GB 快照',
  `bandwidth_mbps` INT NOT NULL COMMENT '带宽 Mbps 快照',
  `traffic_gb` INT NULL COMMENT '月流量 GB 快照，NULL 表示暂不承诺',
  `public_ip_count` INT NOT NULL DEFAULT 1 COMMENT '公网 IP 数快照',
  `virtualization` VARCHAR(32) NOT NULL DEFAULT 'kvm' COMMENT '虚拟化方式快照',
  `architecture` VARCHAR(32) NOT NULL DEFAULT 'x86_64' COMMENT 'CPU 架构快照',
  `billing_cycle` VARCHAR(32) NOT NULL COMMENT '计费周期快照',
  `price_cents` BIGINT UNSIGNED NOT NULL COMMENT '单价，单位分',
  `original_price_cents` BIGINT UNSIGNED NULL COMMENT '划线价，单位分',
  `currency` VARCHAR(16) NOT NULL DEFAULT 'CNY' COMMENT '币种',
  `quantity` INT NOT NULL DEFAULT 1 COMMENT '数量，第一阶段固定为1',
  `total_amount_cents` BIGINT UNSIGNED NOT NULL COMMENT '订单总金额，单位分',
  `region_no` VARCHAR(64) NOT NULL COMMENT '销售地域编号快照',
  `region_code` VARCHAR(64) NOT NULL COMMENT '销售地域编码快照',
  `region_name` VARCHAR(128) NOT NULL COMMENT '销售地域名称快照',
  `template_no` VARCHAR(64) NOT NULL COMMENT '系统模板编号快照',
  `template_code` VARCHAR(96) NOT NULL COMMENT '系统模板编码快照',
  `template_name` VARCHAR(128) NOT NULL COMMENT '系统模板名称快照',
  `os_family` VARCHAR(32) NOT NULL COMMENT '系统族快照',
  `os_distribution` VARCHAR(64) NOT NULL COMMENT '系统发行版快照',
  `os_version` VARCHAR(64) NOT NULL COMMENT '系统版本快照',
  `os_architecture` VARCHAR(32) NOT NULL DEFAULT 'x86_64' COMMENT '系统架构快照',
  `user_note` VARCHAR(500) NULL COMMENT '用户备注',
  `admin_note` VARCHAR(1000) NULL COMMENT '后台备注',
  `cancel_reason` VARCHAR(500) NULL COMMENT '取消原因',
  `closed_reason` VARCHAR(500) NULL COMMENT '关闭原因',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `cancelled_at` DATETIME(3) NULL COMMENT '取消时间',
  `closed_at` DATETIME(3) NULL COMMENT '关闭时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_orders_order_no` (`order_no`),
  UNIQUE KEY `uk_orders_user_client_token` (`user_id`, `client_token`),
  KEY `idx_orders_user_status_created` (`user_id`, `status`, `created_at`),
  KEY `idx_orders_status_created` (`status`, `created_at`),
  KEY `idx_orders_product_plan` (`product_no`, `plan_no`),
  CONSTRAINT `fk_orders_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单';

INSERT INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`) VALUES
  ('page.orders', '订单管理', 'menu', NULL, '/orders', 'Tickets', 60, 1, '菜单', '显示订单管理菜单和页面入口'),
  ('order:*', '订单全权限', 'action', 'page.orders', NULL, NULL, 100, 0, '订单管理', '订单模块全部处理能力'),
  ('order:view', '查看订单', 'action', 'page.orders', NULL, NULL, 110, 0, '订单管理', '查看订单列表和详情'),
  ('order:update', '处理订单', 'action', 'page.orders', NULL, NULL, 120, 0, '订单管理', '更新后台备注和关闭订单'),
  ('order:cancel', '取消订单', 'action', 'page.orders', NULL, NULL, 130, 0, '订单管理', '取消用户端订单')
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
