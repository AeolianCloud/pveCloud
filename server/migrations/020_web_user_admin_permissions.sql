-- Admin permissions for Web user management.

SET NAMES utf8mb4;

USE `pvecloud`;

INSERT INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`) VALUES
  ('page.web-users', 'Web 用户管理', 'menu', NULL, '/web/users', 'User', 40, 1, '菜单', '显示 Web 用户管理菜单和页面入口'),
  ('page.web-user-sessions', '用户状态', 'menu', 'page.web-users', NULL, NULL, 45, 0, '菜单', '显示 Web 用户管理中的用户状态 tab'),
  ('web-user:*', 'Web 用户全权限', 'action', 'page.web-users', NULL, NULL, 100, 0, 'Web 用户', 'Web 用户账号模块全部能力'),
  ('web-user:create', '创建 Web 用户', 'action', 'page.web-users', NULL, NULL, 120, 0, 'Web 用户', '创建用户端账号'),
  ('web-user:update', '修改 Web 用户', 'action', 'page.web-users', NULL, NULL, 130, 0, 'Web 用户', '编辑用户端账号资料和状态'),
  ('web-user:password-reset', '重置 Web 用户密码', 'action', 'page.web-users', NULL, NULL, 140, 0, 'Web 用户', '重置用户端账号密码'),
  ('web-user-session:*', '用户状态全权限', 'action', 'page.web-user-sessions', NULL, NULL, 100, 0, '用户状态', '用户端登录状态模块全部能力'),
  ('web-user-session:revoke', '吊销用户会话', 'action', 'page.web-user-sessions', NULL, NULL, 120, 0, '用户状态', '吊销用户端登录会话')
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
