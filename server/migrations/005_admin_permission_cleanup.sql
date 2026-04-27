-- Keep only admin permissions used by current frontend scope,
-- and rename permission groups / labels to Chinese display names.

SET NAMES utf8mb4;

USE `pvecloud`;

DELETE `arp`
FROM `admin_role_permissions` AS `arp`
JOIN `admin_permissions` AS `ap` ON `ap`.`id` = `arp`.`permission_id`
WHERE `ap`.`code` IN (
  'admin:manage',
  'audit:view',
  'audit:sensitive_view',
  'system:update',
  'admin-session:*',
  'admin-session:view',
  'admin-session:revoke',
  'audit-log:*',
  'audit-log:view',
  'audit-log:sensitive-view',
  'risk-log:*',
  'risk-log:view',
  'user:view',
  'user:update',
  'product:create',
  'product:update',
  'order:view',
  'order:cancel',
  'payment:view',
  'payment:manual_credit',
  'instance:view',
  'instance:operate',
  'ticket:reply'
);

DELETE FROM `admin_permissions`
WHERE `code` IN (
  'admin:manage',
  'audit:view',
  'audit:sensitive_view',
  'system:update',
  'admin-session:*',
  'admin-session:view',
  'admin-session:revoke',
  'audit-log:*',
  'audit-log:view',
  'audit-log:sensitive-view',
  'risk-log:*',
  'risk-log:view',
  'user:view',
  'user:update',
  'product:create',
  'product:update',
  'order:view',
  'order:cancel',
  'payment:view',
  'payment:manual_credit',
  'instance:view',
  'instance:operate',
  'ticket:reply'
);

UPDATE `admin_permissions`
SET
  `group_name` = CASE `code`
    WHEN 'dashboard:*' THEN '控制台'
    WHEN 'dashboard:view' THEN '控制台'
    WHEN 'system-config:*' THEN '系统配置'
    WHEN 'system-config:view' THEN '系统配置'
    WHEN 'system-config:update' THEN '系统配置'
    WHEN 'admin-user:*' THEN '管理员账号'
    WHEN 'admin-user:view' THEN '管理员账号'
    WHEN 'admin-user:create' THEN '管理员账号'
    WHEN 'admin-user:update' THEN '管理员账号'
    WHEN 'admin-user:password-reset' THEN '管理员账号'
    WHEN 'admin-role:*' THEN '管理组权限'
    WHEN 'admin-role:view' THEN '管理组权限'
    WHEN 'admin-role:create' THEN '管理组权限'
    WHEN 'admin-role:update' THEN '管理组权限'
    ELSE `group_name`
  END,
  `name` = CASE `code`
    WHEN 'dashboard:*' THEN '控制台全权限'
    WHEN 'dashboard:view' THEN '查看控制台'
    WHEN 'system-config:*' THEN '系统配置全权限'
    WHEN 'system-config:view' THEN '查看系统配置'
    WHEN 'system-config:update' THEN '修改系统配置'
    WHEN 'admin-user:*' THEN '管理员账号全权限'
    WHEN 'admin-user:view' THEN '查看管理员账号'
    WHEN 'admin-user:create' THEN '创建管理员账号'
    WHEN 'admin-user:update' THEN '修改管理员账号'
    WHEN 'admin-user:password-reset' THEN '重置管理员密码'
    WHEN 'admin-role:*' THEN '管理组权限全权限'
    WHEN 'admin-role:view' THEN '查看管理组权限'
    WHEN 'admin-role:create' THEN '创建管理组'
    WHEN 'admin-role:update' THEN '修改管理组权限'
    ELSE `name`
  END,
  `description` = CASE `code`
    WHEN 'dashboard:*' THEN '控制台模块全部能力'
    WHEN 'dashboard:view' THEN '查看控制台页面和基础指标'
    WHEN 'system-config:*' THEN '系统配置模块全部能力'
    WHEN 'system-config:view' THEN '查看系统配置页面和配置项'
    WHEN 'system-config:update' THEN '修改系统配置项'
    WHEN 'admin-user:*' THEN '管理员账号模块全部能力'
    WHEN 'admin-user:view' THEN '查看管理员账号页面、列表和详情'
    WHEN 'admin-user:create' THEN '创建管理员账号'
    WHEN 'admin-user:update' THEN '编辑管理员资料、状态和角色'
    WHEN 'admin-user:password-reset' THEN '重置管理员登录密码'
    WHEN 'admin-role:*' THEN '管理组权限模块全部能力'
    WHEN 'admin-role:view' THEN '查看管理组列表、详情和权限树'
    WHEN 'admin-role:create' THEN '创建管理组'
    WHEN 'admin-role:update' THEN '编辑管理组、状态和权限分配'
    ELSE `description`
  END
WHERE `code` IN (
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
  'admin-role:update'
);
