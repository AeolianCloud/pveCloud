-- Log management center Phase 1 menus and permissions.
-- Phase 1 reuses admin_audit_logs and GET /admin-api/audit-logs.

SET NAMES utf8mb4;

USE `pvecloud`;

INSERT INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`)
VALUES
  ('page.logs', '日志管理中心', 'menu', NULL, '/logs', 'DocumentText', 45, 1, '菜单', '显示日志管理中心父级菜单'),
  ('page.logs.admin-operations', '操作审计', 'menu', 'page.logs', '/logs/admin-operations', NULL, 10, 1, '菜单', '显示日志管理中心的操作审计页面'),
  ('page.logs.admin-security', '登录安全', 'menu', 'page.logs', '/logs/admin-security', NULL, 20, 1, '菜单', '显示日志管理中心的登录安全页面'),
  ('admin-security-log:*', '登录安全日志全权限', 'action', 'page.logs.admin-security', NULL, NULL, 100, 0, '登录安全日志', '登录安全日志模块全部能力'),
  ('admin-security-log:view', '查看登录安全日志', 'action', 'page.logs.admin-security', NULL, NULL, 120, 0, '登录安全日志', '查看管理端登录安全日志列表')
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

UPDATE `admin_permissions`
SET
  `name` = '日志管理兼容入口',
  `type` = 'menu',
  `parent_code` = 'page.logs.admin-operations',
  `path` = NULL,
  `icon` = NULL,
  `sort_order` = 900,
  `visible_in_menu` = 0,
  `group_name` = '菜单',
  `description` = '旧 /system/audit-logs 菜单权限兼容节点，不再作为侧栏菜单展示'
WHERE `code` = 'page.system-settings.audit-logs';

UPDATE `admin_permissions`
SET
  `type` = 'action',
  `parent_code` = 'page.logs.admin-operations',
  `visible_in_menu` = 0,
  `sort_order` = CASE `code`
    WHEN 'audit-log:*' THEN 100
    WHEN 'audit-log:sensitive-view' THEN 120
    ELSE `sort_order`
  END
WHERE `code` IN ('audit-log:*', 'audit-log:sensitive-view');

INSERT INTO `admin_role_permissions` (`role_id`, `permission_id`)
SELECT `admin_roles`.`id`, `admin_permissions`.`id`
FROM `admin_roles`
JOIN `admin_permissions`
WHERE `admin_roles`.`code` = 'super_admin'
ON DUPLICATE KEY UPDATE
  `role_id` = VALUES(`role_id`);

INSERT INTO `admin_role_permissions` (`role_id`, `permission_id`)
SELECT DISTINCT `arp`.`role_id`, `parent`.`id`
FROM `admin_role_permissions` AS `arp`
JOIN `admin_permissions` AS `child` ON `child`.`id` = `arp`.`permission_id`
JOIN `admin_permissions` AS `parent` ON `parent`.`code` = `child`.`parent_code`
WHERE `child`.`parent_code` IS NOT NULL
ON DUPLICATE KEY UPDATE
  `role_id` = VALUES(`role_id`);

INSERT INTO `admin_role_permissions` (`role_id`, `permission_id`)
SELECT DISTINCT `arp`.`role_id`, `target`.`id`
FROM `admin_role_permissions` AS `arp`
JOIN `admin_permissions` AS `legacy` ON `legacy`.`id` = `arp`.`permission_id`
JOIN `admin_permissions` AS `target` ON `target`.`code` IN ('page.logs.admin-security', 'admin-security-log:view')
WHERE `legacy`.`code` = 'page.system-settings.audit-logs'
ON DUPLICATE KEY UPDATE
  `role_id` = VALUES(`role_id`);

INSERT INTO `admin_role_permissions` (`role_id`, `permission_id`)
SELECT DISTINCT `arp`.`role_id`, `parent`.`id`
FROM `admin_role_permissions` AS `arp`
JOIN `admin_permissions` AS `child` ON `child`.`id` = `arp`.`permission_id`
JOIN `admin_permissions` AS `parent` ON `parent`.`code` = `child`.`parent_code`
WHERE `child`.`parent_code` IS NOT NULL
ON DUPLICATE KEY UPDATE
  `role_id` = VALUES(`role_id`);
