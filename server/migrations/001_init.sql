-- pveCloud admin-only initial database schema.
-- Target: MariaDB 11.4.9 / InnoDB / utf8mb4.

SET NAMES utf8mb4;

CREATE DATABASE IF NOT EXISTS `pvecloud`
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE `pvecloud`;

CREATE TABLE IF NOT EXISTS `admin_users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '管理员ID',
  `username` VARCHAR(64) NOT NULL COMMENT '管理员用户名',
  `email` VARCHAR(191) NULL COMMENT '管理员邮箱',
  `password_hash` VARCHAR(255) NOT NULL COMMENT '密码哈希',
  `display_name` VARCHAR(64) NOT NULL COMMENT '显示名称',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '管理员状态',
  `last_login_at` DATETIME(3) NULL COMMENT '最后登录时间',
  `last_login_ip` VARCHAR(64) NULL COMMENT '最后登录 IP',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `deleted_at` DATETIME(3) NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_users_username` (`username`),
  UNIQUE KEY `uk_admin_users_email` (`email`),
  KEY `idx_admin_users_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员账号';

CREATE TABLE IF NOT EXISTS `admin_roles` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '角色ID',
  `code` VARCHAR(64) NOT NULL COMMENT '角色编码',
  `name` VARCHAR(64) NOT NULL COMMENT '角色名称',
  `description` VARCHAR(255) NULL COMMENT '角色说明',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '角色状态',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_roles_code` (`code`),
  KEY `idx_admin_roles_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理端角色';

CREATE TABLE IF NOT EXISTS `admin_permissions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '权限ID',
  `code` VARCHAR(96) NOT NULL COMMENT '权限码',
  `name` VARCHAR(96) NOT NULL COMMENT '权限名称',
  `type` VARCHAR(32) NOT NULL DEFAULT 'action' COMMENT '权限节点类型：menu/action',
  `parent_code` VARCHAR(96) NULL COMMENT '父级菜单权限码',
  `path` VARCHAR(255) NULL COMMENT '菜单路径',
  `icon` VARCHAR(64) NULL COMMENT '菜单图标',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '同级排序',
  `visible_in_menu` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否在侧栏菜单展示',
  `group_name` VARCHAR(64) NOT NULL COMMENT '权限分组',
  `description` VARCHAR(255) NULL COMMENT '权限说明',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_permissions_code` (`code`),
  KEY `idx_admin_permissions_parent` (`parent_code`),
  KEY `idx_admin_permissions_type_sort` (`type`, `sort_order`),
  KEY `idx_admin_permissions_group` (`group_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理端权限码';

CREATE TABLE IF NOT EXISTS `admin_user_roles` (
  `admin_id` BIGINT UNSIGNED NOT NULL COMMENT '管理员ID',
  `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`admin_id`, `role_id`),
  CONSTRAINT `fk_admin_user_roles_admin_user` FOREIGN KEY (`admin_id`) REFERENCES `admin_users` (`id`),
  CONSTRAINT `fk_admin_user_roles_role` FOREIGN KEY (`role_id`) REFERENCES `admin_roles` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员角色关联';

CREATE TABLE IF NOT EXISTS `admin_role_permissions` (
  `role_id` BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
  `permission_id` BIGINT UNSIGNED NOT NULL COMMENT '权限ID',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`role_id`, `permission_id`),
  CONSTRAINT `fk_admin_role_permissions_role` FOREIGN KEY (`role_id`) REFERENCES `admin_roles` (`id`),
  CONSTRAINT `fk_admin_role_permissions_permission` FOREIGN KEY (`permission_id`) REFERENCES `admin_permissions` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限关联';

CREATE TABLE IF NOT EXISTS `admin_sessions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '管理端会话ID',
  `session_id` VARCHAR(64) NOT NULL COMMENT 'JWT jti，会话唯一标识',
  `admin_id` BIGINT UNSIGNED NOT NULL COMMENT '管理员ID',
  `status` VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '会话状态：active/revoked/expired',
  `issued_at` DATETIME(3) NOT NULL COMMENT '签发时间',
  `expires_at` DATETIME(3) NOT NULL COMMENT '过期时间',
  `last_seen_at` DATETIME(3) NULL COMMENT '最后访问时间',
  `last_seen_ip` VARCHAR(64) NULL COMMENT '最后访问 IP',
  `user_agent` VARCHAR(500) NULL COMMENT '登录或最近访问 User-Agent',
  `revoked_at` DATETIME(3) NULL COMMENT '吊销时间',
  `revoke_reason` VARCHAR(64) NULL COMMENT '吊销原因：logout/refresh/admin_disabled/expired',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_sessions_session_id` (`session_id`),
  KEY `idx_admin_sessions_admin_status` (`admin_id`, `status`),
  KEY `idx_admin_sessions_status_expires` (`status`, `expires_at`),
  KEY `idx_admin_sessions_expires_at` (`expires_at`),
  CONSTRAINT `fk_admin_sessions_admin_user` FOREIGN KEY (`admin_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理端登录会话';

CREATE TABLE IF NOT EXISTS `system_configs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '配置ID',
  `config_key` VARCHAR(128) NOT NULL COMMENT '配置键',
  `config_value` TEXT NULL COMMENT '配置值',
  `value_type` VARCHAR(32) NOT NULL DEFAULT 'string' COMMENT '值类型',
  `group_name` VARCHAR(64) NOT NULL COMMENT '配置分组',
  `is_secret` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否敏感配置',
  `description` VARCHAR(255) NULL COMMENT '配置说明',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_system_configs_key` (`config_key`),
  KEY `idx_system_configs_group` (`group_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置';

CREATE TABLE IF NOT EXISTS `admin_audit_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '审计日志ID',
  `admin_id` BIGINT UNSIGNED NULL COMMENT '管理员ID',
  `admin_username` VARCHAR(64) NULL COMMENT '操作发生时的管理员用户名快照',
  `admin_display_name` VARCHAR(64) NULL COMMENT '操作发生时的管理员显示名快照',
  `session_id` VARCHAR(64) NULL COMMENT '触发操作的管理端会话标识',
  `request_id` VARCHAR(64) NULL COMMENT '请求链路ID',
  `request_method` VARCHAR(16) NULL COMMENT '后台请求方法',
  `request_path` VARCHAR(255) NULL COMMENT '后台请求路径',
  `action` VARCHAR(96) NOT NULL COMMENT '操作动作',
  `object_type` VARCHAR(64) NOT NULL COMMENT '操作对象类型',
  `object_id` VARCHAR(64) NULL COMMENT '操作对象ID',
  `before_data` JSON NULL COMMENT '操作前数据，需脱敏',
  `after_data` JSON NULL COMMENT '操作后数据，需脱敏',
  `ip` VARCHAR(64) NULL COMMENT '操作IP',
  `user_agent` VARCHAR(500) NULL COMMENT '浏览器 User-Agent',
  `remark` VARCHAR(500) NULL COMMENT '操作备注',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_admin_audit_logs_admin_created` (`admin_id`, `created_at`),
  KEY `idx_admin_audit_logs_session_created` (`session_id`, `created_at`),
  KEY `idx_admin_audit_logs_request_id` (`request_id`),
  KEY `idx_admin_audit_logs_object` (`object_type`, `object_id`),
  KEY `idx_admin_audit_logs_action_created` (`action`, `created_at`),
  CONSTRAINT `fk_admin_audit_logs_admin_user` FOREIGN KEY (`admin_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台操作审计日志';

INSERT INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`) VALUES
  ('page.dashboard', '控制台', 'menu', NULL, '/dashboard', 'Odometer', 10, 1, '菜单', '显示控制台菜单和页面入口'),
  ('page.system-settings', '系统设置', 'menu', NULL, '/system', 'Setting', 20, 1, '菜单', '显示系统设置父级菜单'),
  ('page.system-settings.config', '系统配置', 'menu', 'page.system-settings', '/system/settings', NULL, 10, 1, '菜单', '显示系统设置下的系统配置入口'),
  ('page.system-settings.admin-users', '管理员设置', 'menu', 'page.system-settings', '/system/admin-users', NULL, 20, 1, '菜单', '显示系统设置下的管理员设置入口'),
  ('page.system-settings.admin-roles', '管理组权限', 'menu', 'page.system-settings.admin-users', NULL, NULL, 21, 0, '菜单', '显示管理员设置中的管理组权限入口'),
  ('page.system-settings.admin-sessions', '管理员会话', 'menu', 'page.system-settings.admin-users', NULL, NULL, 22, 0, '菜单', '显示管理员设置中的管理员会话入口'),
  ('page.system-settings.audit-logs', '操作日志', 'menu', 'page.system-settings', '/system/audit-logs', NULL, 30, 1, '菜单', '显示系统设置下的操作日志入口'),
  ('page.web-users', 'Web 用户管理', 'menu', NULL, '/web/users', 'User', 40, 1, '菜单', '显示 Web 用户管理菜单和页面入口'),
  ('page.web-user-sessions', '用户状态', 'menu', 'page.web-users', NULL, NULL, 45, 0, '菜单', '显示 Web 用户管理中的用户状态 tab'),
  ('dashboard:*', '控制台全权限', 'action', 'page.dashboard', NULL, NULL, 100, 0, '控制台', '控制台模块全部能力'),
  ('system-config:*', '系统配置全权限', 'action', 'page.system-settings.config', NULL, NULL, 100, 0, '系统配置', '系统配置模块全部能力'),
  ('system-config:update', '修改系统配置', 'action', 'page.system-settings.config', NULL, NULL, 120, 0, '系统配置', '修改系统配置项'),
  ('admin-user:*', '管理员账号全权限', 'action', 'page.system-settings.admin-users', NULL, NULL, 100, 0, '管理员账号', '管理员账号模块全部能力'),
  ('admin-user:create', '创建管理员账号', 'action', 'page.system-settings.admin-users', NULL, NULL, 120, 0, '管理员账号', '创建管理员账号'),
  ('admin-user:update', '修改管理员账号', 'action', 'page.system-settings.admin-users', NULL, NULL, 130, 0, '管理员账号', '编辑管理员资料、状态和角色'),
  ('admin-user:password-reset', '重置管理员密码', 'action', 'page.system-settings.admin-users', NULL, NULL, 140, 0, '管理员账号', '重置管理员登录密码'),
  ('admin-role:*', '管理组权限全权限', 'action', 'page.system-settings.admin-roles', NULL, NULL, 100, 0, '管理组权限', '管理组权限模块全部能力'),
  ('admin-role:create', '创建管理组', 'action', 'page.system-settings.admin-roles', NULL, NULL, 120, 0, '管理组权限', '创建管理组'),
  ('admin-role:update', '修改管理组', 'action', 'page.system-settings.admin-roles', NULL, NULL, 130, 0, '管理组权限', '编辑管理组、状态和权限分配'),
  ('admin-session:*', '管理员会话全权限', 'action', 'page.system-settings.admin-sessions', NULL, NULL, 100, 0, '管理员会话', '管理员会话模块全部能力'),
  ('admin-session:revoke', '吊销管理员会话', 'action', 'page.system-settings.admin-sessions', NULL, NULL, 120, 0, '管理员会话', '吊销指定管理员会话'),
  ('audit-log:*', '操作日志全权限', 'action', 'page.system-settings.audit-logs', NULL, NULL, 100, 0, '操作日志', '操作日志模块全部能力'),
  ('audit-log:sensitive-view', '查看操作日志敏感详情', 'action', 'page.system-settings.audit-logs', NULL, NULL, 120, 0, '操作日志', '查看操作日志中的前后快照和 User-Agent'),
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

INSERT INTO `admin_roles` (`code`, `name`, `description`, `status`) VALUES
  ('super_admin', '超级管理员', '拥有全部基础后台权限', 'active')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `description` = VALUES(`description`),
  `status` = VALUES(`status`);

INSERT INTO `admin_role_permissions` (`role_id`, `permission_id`)
SELECT `admin_roles`.`id`, `admin_permissions`.`id`
FROM `admin_roles`
JOIN `admin_permissions`
WHERE `admin_roles`.`code` = 'super_admin'
ON DUPLICATE KEY UPDATE
  `role_id` = VALUES(`role_id`);

INSERT INTO `system_configs` (`config_key`, `config_value`, `value_type`, `group_name`, `is_secret`, `description`) VALUES
  ('site.name', 'pveCloud', 'string', '站点设置', 0, '站点名称'),
  ('site.logo_url', '', 'string', '站点设置', 0, 'Web 左上角 Logo 图片 URL')
ON DUPLICATE KEY UPDATE
  `config_value` = VALUES(`config_value`),
  `value_type` = VALUES(`value_type`),
  `group_name` = VALUES(`group_name`),
  `is_secret` = VALUES(`is_secret`),
  `description` = VALUES(`description`);

-- Default admin accounts are intentionally not inserted here.
-- Create the first admin through a setup command so no default password is stored in git.
