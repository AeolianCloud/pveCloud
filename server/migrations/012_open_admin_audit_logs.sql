  -- Open normal admin operation logs under System Settings.
  -- Target: MariaDB 11.4.x / InnoDB / utf8mb4.

  SET NAMES utf8mb4;

  USE `pvecloud`;

  INSERT INTO `admin_permissions` (`code`, `name`, `group_name`, `description`) VALUES
    ('page.system-settings.audit-logs', '进入操作日志', '页面入口', '显示系统设置下的操作日志入口'),
    ('audit-log:*', '操作日志全权限', '操作日志', '操作日志模块全部能力'),
    ('audit-log:view', '查看操作日志', '操作日志', '查看普通后台操作日志列表'),
    ('audit-log:sensitive-view', '查看操作日志敏感详情', '操作日志', '查看操作日志中的前后快照和 User-Agent')
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
      'page.system-settings.audit-logs',
      'audit-log:*',
      'audit-log:view',
      'audit-log:sensitive-view'
    )
  ON DUPLICATE KEY UPDATE
    `role_id` = VALUES(`role_id`);
