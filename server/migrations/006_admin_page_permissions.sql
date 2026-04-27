-- Add page entry permissions for current admin frontend navigation.

SET NAMES utf8mb4;

USE `pvecloud`;

INSERT INTO `admin_permissions` (`code`, `name`, `group_name`, `description`) VALUES
  ('page.dashboard', '进入控制台', '页面入口', '显示控制台菜单和页面入口'),
  ('page.system-settings.config', '进入系统配置', '页面入口', '显示系统设置下的系统配置入口'),
  ('page.system-settings.admin-users', '进入管理员账号', '页面入口', '显示系统设置下的管理员账号入口'),
  ('page.system-settings.admin-roles', '进入管理组权限', '页面入口', '显示系统设置下的管理组权限入口')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `group_name` = VALUES(`group_name`),
  `description` = VALUES(`description`);

INSERT INTO `admin_role_permissions` (`role_id`, `permission_id`)
SELECT `admin_roles`.`id`, `admin_permissions`.`id`
FROM `admin_roles`
JOIN `admin_permissions`
WHERE `admin_roles`.`code` = 'super_admin'
  AND `admin_permissions`.`code` IN (
    'page.dashboard',
    'page.system-settings.config',
    'page.system-settings.admin-users',
    'page.system-settings.admin-roles'
  )
ON DUPLICATE KEY UPDATE
  `role_id` = VALUES(`role_id`);
