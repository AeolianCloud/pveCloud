-- Unify admin menus and action permissions in admin_permissions.

SET NAMES utf8mb4;

USE `pvecloud`;

SET @sql = IF(
  (SELECT COUNT(*) FROM `information_schema`.`COLUMNS` WHERE `TABLE_SCHEMA` = DATABASE() AND `TABLE_NAME` = 'admin_permissions' AND `COLUMN_NAME` = 'type') = 0,
  'ALTER TABLE `admin_permissions` ADD COLUMN `type` VARCHAR(32) NOT NULL DEFAULT ''action'' COMMENT ''权限节点类型：menu/action'' AFTER `name`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = IF(
  (SELECT COUNT(*) FROM `information_schema`.`COLUMNS` WHERE `TABLE_SCHEMA` = DATABASE() AND `TABLE_NAME` = 'admin_permissions' AND `COLUMN_NAME` = 'parent_code') = 0,
  'ALTER TABLE `admin_permissions` ADD COLUMN `parent_code` VARCHAR(96) NULL COMMENT ''父级菜单权限码'' AFTER `type`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = IF(
  (SELECT COUNT(*) FROM `information_schema`.`COLUMNS` WHERE `TABLE_SCHEMA` = DATABASE() AND `TABLE_NAME` = 'admin_permissions' AND `COLUMN_NAME` = 'path') = 0,
  'ALTER TABLE `admin_permissions` ADD COLUMN `path` VARCHAR(255) NULL COMMENT ''菜单路径'' AFTER `parent_code`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = IF(
  (SELECT COUNT(*) FROM `information_schema`.`COLUMNS` WHERE `TABLE_SCHEMA` = DATABASE() AND `TABLE_NAME` = 'admin_permissions' AND `COLUMN_NAME` = 'icon') = 0,
  'ALTER TABLE `admin_permissions` ADD COLUMN `icon` VARCHAR(64) NULL COMMENT ''菜单图标'' AFTER `path`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = IF(
  (SELECT COUNT(*) FROM `information_schema`.`COLUMNS` WHERE `TABLE_SCHEMA` = DATABASE() AND `TABLE_NAME` = 'admin_permissions' AND `COLUMN_NAME` = 'sort_order') = 0,
  'ALTER TABLE `admin_permissions` ADD COLUMN `sort_order` INT NOT NULL DEFAULT 0 COMMENT ''同级排序'' AFTER `icon`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = IF(
  (SELECT COUNT(*) FROM `information_schema`.`COLUMNS` WHERE `TABLE_SCHEMA` = DATABASE() AND `TABLE_NAME` = 'admin_permissions' AND `COLUMN_NAME` = 'visible_in_menu') = 0,
  'ALTER TABLE `admin_permissions` ADD COLUMN `visible_in_menu` TINYINT(1) NOT NULL DEFAULT 0 COMMENT ''是否在侧栏菜单展示'' AFTER `sort_order`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = IF(
  (SELECT COUNT(*) FROM `information_schema`.`STATISTICS` WHERE `TABLE_SCHEMA` = DATABASE() AND `TABLE_NAME` = 'admin_permissions' AND `INDEX_NAME` = 'idx_admin_permissions_parent') = 0,
  'ALTER TABLE `admin_permissions` ADD KEY `idx_admin_permissions_parent` (`parent_code`)',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = IF(
  (SELECT COUNT(*) FROM `information_schema`.`STATISTICS` WHERE `TABLE_SCHEMA` = DATABASE() AND `TABLE_NAME` = 'admin_permissions' AND `INDEX_NAME` = 'idx_admin_permissions_type_sort') = 0,
  'ALTER TABLE `admin_permissions` ADD KEY `idx_admin_permissions_type_sort` (`type`, `sort_order`)',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `admin_permissions`
SET
  `type` = 'menu',
  `parent_code` = NULL,
  `path` = '/dashboard',
  `icon` = 'Odometer',
  `sort_order` = 10,
  `visible_in_menu` = 1,
  `group_name` = '菜单'
WHERE `code` = 'page.dashboard';

UPDATE `admin_permissions`
SET
  `type` = 'menu',
  `parent_code` = NULL,
  `path` = '/system',
  `icon` = 'Setting',
  `sort_order` = 20,
  `visible_in_menu` = 1,
  `group_name` = '菜单'
WHERE `code` = 'page.system-settings';

INSERT INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`)
VALUES
  ('page.system-settings', '系统设置', 'menu', NULL, '/system', 'Setting', 20, 1, '菜单', '显示系统设置父级菜单')
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
  `type` = 'menu',
  `parent_code` = 'page.system-settings',
  `path` = '/system/settings',
  `icon` = NULL,
  `sort_order` = 10,
  `visible_in_menu` = 1,
  `group_name` = '菜单'
WHERE `code` = 'page.system-settings.config';

UPDATE `admin_permissions`
SET
  `type` = 'menu',
  `parent_code` = 'page.system-settings',
  `path` = '/system/admin-users',
  `icon` = NULL,
  `sort_order` = 20,
  `visible_in_menu` = 1,
  `group_name` = '菜单'
WHERE `code` = 'page.system-settings.admin-users';

UPDATE `admin_permissions`
SET
  `type` = 'menu',
  `parent_code` = 'page.system-settings.admin-users',
  `path` = NULL,
  `icon` = NULL,
  `sort_order` = 21,
  `visible_in_menu` = 0,
  `group_name` = '菜单'
WHERE `code` = 'page.system-settings.admin-roles';

UPDATE `admin_permissions`
SET
  `type` = 'menu',
  `parent_code` = 'page.system-settings.admin-users',
  `path` = NULL,
  `icon` = NULL,
  `sort_order` = 22,
  `visible_in_menu` = 0,
  `group_name` = '菜单'
WHERE `code` = 'page.system-settings.admin-sessions';

UPDATE `admin_permissions`
SET
  `type` = 'menu',
  `parent_code` = 'page.system-settings',
  `path` = '/system/audit-logs',
  `icon` = NULL,
  `sort_order` = 30,
  `visible_in_menu` = 1,
  `group_name` = '菜单'
WHERE `code` = 'page.system-settings.audit-logs';

DELETE `arp`
FROM `admin_role_permissions` AS `arp`
JOIN `admin_permissions` AS `ap` ON `ap`.`id` = `arp`.`permission_id`
WHERE `ap`.`code` IN (
  'dashboard:view',
  'system-config:view',
  'admin-user:view',
  'admin-role:view',
  'admin-session:view',
  'audit-log:view',
  'audit:sensitive_view'
);

DELETE FROM `admin_permissions`
WHERE `code` IN (
  'dashboard:view',
  'system-config:view',
  'admin-user:view',
  'admin-role:view',
  'admin-session:view',
  'audit-log:view',
  'audit:sensitive_view'
);

UPDATE `admin_permissions`
SET
  `type` = 'action',
  `parent_code` = 'page.dashboard',
  `visible_in_menu` = 0,
  `sort_order` = 100
WHERE `code` IN ('dashboard:*');

UPDATE `admin_permissions`
SET
  `type` = 'action',
  `parent_code` = 'page.system-settings.config',
  `visible_in_menu` = 0,
  `sort_order` = CASE `code`
    WHEN 'system-config:*' THEN 100
    WHEN 'system-config:update' THEN 120
    ELSE `sort_order`
  END
WHERE `code` IN ('system-config:*', 'system-config:update');

UPDATE `admin_permissions`
SET
  `type` = 'action',
  `parent_code` = 'page.system-settings.admin-users',
  `visible_in_menu` = 0,
  `sort_order` = CASE `code`
    WHEN 'admin-user:*' THEN 100
    WHEN 'admin-user:create' THEN 120
    WHEN 'admin-user:update' THEN 130
    WHEN 'admin-user:password-reset' THEN 140
    ELSE `sort_order`
  END
WHERE `code` IN ('admin-user:*', 'admin-user:create', 'admin-user:update', 'admin-user:password-reset');

UPDATE `admin_permissions`
SET
  `type` = 'action',
  `parent_code` = 'page.system-settings.admin-roles',
  `visible_in_menu` = 0,
  `sort_order` = CASE `code`
    WHEN 'admin-role:*' THEN 100
    WHEN 'admin-role:create' THEN 120
    WHEN 'admin-role:update' THEN 130
    ELSE `sort_order`
  END
WHERE `code` IN ('admin-role:*', 'admin-role:create', 'admin-role:update');

UPDATE `admin_permissions`
SET
  `type` = 'action',
  `parent_code` = 'page.system-settings.admin-sessions',
  `visible_in_menu` = 0,
  `sort_order` = CASE `code`
    WHEN 'admin-session:*' THEN 100
    WHEN 'admin-session:revoke' THEN 120
    ELSE `sort_order`
  END
WHERE `code` IN ('admin-session:*', 'admin-session:revoke');

UPDATE `admin_permissions`
SET
  `type` = 'action',
  `parent_code` = 'page.system-settings.audit-logs',
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
SELECT DISTINCT `arp`.`role_id`, `parent`.`id`
FROM `admin_role_permissions` AS `arp`
JOIN `admin_permissions` AS `child` ON `child`.`id` = `arp`.`permission_id`
JOIN `admin_permissions` AS `parent` ON `parent`.`code` = `child`.`parent_code`
WHERE `child`.`parent_code` IS NOT NULL
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
