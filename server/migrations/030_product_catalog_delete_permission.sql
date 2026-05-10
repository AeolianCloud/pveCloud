-- Product catalog delete permission.
-- Target: MariaDB 11.4.x / InnoDB / utf8mb4.

SET NAMES utf8mb4;

USE `pvecloud`;

INSERT INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`) VALUES
  ('product:delete', '删除产品', 'action', 'page.products', NULL, NULL, 150, 0, '产品管理', '删除产品、套餐、销售地域和系统模板')
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
  AND `admin_permissions`.`code` = 'product:delete'
ON DUPLICATE KEY UPDATE
  `role_id` = VALUES(`role_id`);
