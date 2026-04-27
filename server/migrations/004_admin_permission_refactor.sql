-- Add fine-grained admin permission codes and keep legacy codes for compatibility.

SET NAMES utf8mb4;

USE `pvecloud`;

INSERT INTO `admin_permissions` (`code`, `name`, `group_name`, `description`) VALUES
  ('dashboard:*', '控制台全权限', 'dashboard', '管理控制台模块全部能力'),
  ('dashboard:view', '查看控制台', 'dashboard', '查看控制台页面和基础指标'),
  ('system-config:*', '系统配置全权限', 'system-config', '管理系统配置模块全部能力'),
  ('system-config:view', '查看系统配置', 'system-config', '查看系统配置页面和配置项'),
  ('system-config:update', '修改系统配置', 'system-config', '修改系统配置项'),
  ('admin-user:*', '管理员账号全权限', 'admin-user', '管理管理员账号模块全部能力'),
  ('admin-user:view', '查看管理员账号', 'admin-user', '查看管理员账号页面、列表和详情'),
  ('admin-user:create', '创建管理员账号', 'admin-user', '创建管理员账号'),
  ('admin-user:update', '修改管理员账号', 'admin-user', '编辑管理员资料、状态和角色'),
  ('admin-user:password-reset', '重置管理员密码', 'admin-user', '重置管理员登录密码'),
  ('admin-role:*', '管理组全权限', 'admin-role', '管理管理组和权限模块全部能力'),
  ('admin-role:view', '查看管理组', 'admin-role', '查看管理组列表、详情和权限树'),
  ('admin-role:create', '创建管理组', 'admin-role', '创建管理组'),
  ('admin-role:update', '修改管理组', 'admin-role', '编辑管理组、状态和权限分配'),
  ('admin-session:*', '管理端会话全权限', 'admin-session', '管理管理端会话模块全部能力'),
  ('admin-session:view', '查看管理端会话', 'admin-session', '查看管理端会话列表'),
  ('admin-session:revoke', '吊销管理端会话', 'admin-session', '吊销指定管理端会话'),
  ('audit-log:*', '审计日志全权限', 'audit-log', '管理审计日志模块全部能力'),
  ('audit-log:view', '查看审计日志', 'audit-log', '查看后台审计日志主信息'),
  ('audit-log:sensitive-view', '查看审计敏感详情', 'audit-log', '查看脱敏后的审计敏感字段'),
  ('risk-log:*', '高危操作日志全权限', 'risk-log', '管理高危操作日志模块全部能力'),
  ('risk-log:view', '查看高危操作日志', 'risk-log', '查看高危操作日志')
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
    'dashboard:*',
    'dashboard:view',
    'system-config:*',
    'system-config:view',
    'system-config:update',
    'admin-user:*',
    'admin-user:view',
    'admin-user:create',
    'admin-user:update',
    'admin-user:password-reset',
    'admin-role:*',
    'admin-role:view',
    'admin-role:create',
    'admin-role:update',
    'admin-session:*',
    'admin-session:view',
    'admin-session:revoke',
    'audit-log:*',
    'audit-log:view',
    'audit-log:sensitive-view',
    'risk-log:*',
    'risk-log:view'
  )
ON DUPLICATE KEY UPDATE
  `role_id` = VALUES(`role_id`);
