-- Re-open admin session permissions for the admin settings third tab.

SET NAMES utf8mb4;

USE `pvecloud`;

INSERT INTO `admin_permissions` (`code`, `name`, `group_name`, `description`) VALUES
  ('page.system-settings.admin-sessions', '进入管理员会话', '页面入口', '显示系统设置下的管理员会话入口'),
  ('admin-session:*', '管理员会话全权限', '管理员会话', '管理员会话模块全部能力'),
  ('admin-session:view', '查看管理员会话', '管理员会话', '查看管理员会话列表'),
  ('admin-session:revoke', '吊销管理员会话', '管理员会话', '吊销指定管理员会话')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `group_name` = VALUES(`group_name`),
  `description` = VALUES(`description`);

INSERT INTO `admin_role_permissions` (`role_id`, `permission_id`)
SELECT `admin_roles`.`id`, `admin_permissions`.`id`
FROM `admin_roles`
JOIN `admin_permissions`
WHERE `admin_roles`.`code` = 'super_admin'
ON DUPLICATE KEY UPDATE
  `role_id` = VALUES(`role_id`);
