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
  `group_name` VARCHAR(64) NOT NULL COMMENT '权限分组',
  `description` VARCHAR(255) NULL COMMENT '权限说明',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_permissions_code` (`code`),
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
  KEY `idx_admin_audit_logs_object` (`object_type`, `object_id`),
  KEY `idx_admin_audit_logs_action_created` (`action`, `created_at`),
  CONSTRAINT `fk_admin_audit_logs_admin_user` FOREIGN KEY (`admin_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台操作审计日志';

CREATE TABLE IF NOT EXISTS `admin_risk_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '高危操作日志ID',
  `audit_log_id` BIGINT UNSIGNED NULL COMMENT '关联普通审计日志ID',
  `admin_id` BIGINT UNSIGNED NULL COMMENT '管理员ID',
  `risk_level` VARCHAR(32) NOT NULL COMMENT '风险等级：medium/high/critical',
  `action` VARCHAR(96) NOT NULL COMMENT '高危操作动作',
  `object_type` VARCHAR(64) NOT NULL COMMENT '操作对象类型',
  `object_id` VARCHAR(64) NULL COMMENT '操作对象ID',
  `risk_reason` VARCHAR(255) NOT NULL COMMENT '判定为高危的原因',
  `before_data` JSON NULL COMMENT '操作前数据，需脱敏',
  `after_data` JSON NULL COMMENT '操作后数据，需脱敏',
  `ip` VARCHAR(64) NULL COMMENT '操作IP',
  `user_agent` VARCHAR(500) NULL COMMENT '浏览器 User-Agent',
  `remark` VARCHAR(500) NULL COMMENT '操作备注',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_admin_risk_logs_audit_log` (`audit_log_id`),
  KEY `idx_admin_risk_logs_admin_created` (`admin_id`, `created_at`),
  KEY `idx_admin_risk_logs_level_created` (`risk_level`, `created_at`),
  KEY `idx_admin_risk_logs_object` (`object_type`, `object_id`),
  KEY `idx_admin_risk_logs_action_created` (`action`, `created_at`),
  CONSTRAINT `fk_admin_risk_logs_audit_log` FOREIGN KEY (`audit_log_id`) REFERENCES `admin_audit_logs` (`id`),
  CONSTRAINT `fk_admin_risk_logs_admin_user` FOREIGN KEY (`admin_id`) REFERENCES `admin_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台高危操作日志';

INSERT INTO `admin_permissions` (`code`, `name`, `group_name`, `description`) VALUES
  ('dashboard:*', '控制台全权限', '控制台', '控制台模块全部能力'),
  ('dashboard:view', '查看控制台', '控制台', '查看控制台页面和基础指标'),
  ('system-config:*', '系统配置全权限', '系统配置', '系统配置模块全部能力'),
  ('system-config:view', '查看系统配置', '系统配置', '查看系统配置页面和配置项'),
  ('system-config:update', '修改系统配置', '系统配置', '修改系统配置项'),
  ('admin-user:*', '管理员账号全权限', '管理员账号', '管理员账号模块全部能力'),
  ('admin-user:view', '查看管理员账号', '管理员账号', '查看管理员账号页面、列表和详情'),
  ('admin-user:create', '创建管理员账号', '管理员账号', '创建管理员账号'),
  ('admin-user:update', '修改管理员账号', '管理员账号', '编辑管理员资料、状态和角色'),
  ('admin-user:password-reset', '重置管理员密码', '管理员账号', '重置管理员登录密码'),
  ('admin-role:*', '管理组权限全权限', '管理组权限', '管理组权限模块全部能力'),
  ('admin-role:view', '查看管理组', '管理组权限', '查看管理组列表、详情和权限树'),
  ('admin-role:create', '创建管理组', '管理组权限', '创建管理组'),
  ('admin-role:update', '修改管理组', '管理组权限', '编辑管理组、状态和权限分配'),
  ('page.dashboard', '进入控制台', '页面入口', '显示控制台菜单和页面入口'),
  ('page.system-settings.config', '进入系统配置', '页面入口', '显示系统设置下的系统配置入口'),
  ('page.system-settings.admin-users', '进入管理员账号', '页面入口', '显示系统设置下的管理员账号入口'),
  ('page.system-settings.admin-roles', '进入管理组权限', '页面入口', '显示系统设置下的管理组权限入口'),
  ('page.system-settings.admin-sessions', '进入管理员会话', '页面入口', '显示系统设置下的管理员会话入口'),
  ('admin-session:*', '管理员会话全权限', '管理员会话', '管理员会话模块全部能力'),
  ('admin-session:view', '查看管理员会话', '管理员会话', '查看管理员会话列表'),
  ('admin-session:revoke', '吊销管理员会话', '管理员会话', '吊销指定管理员会话')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
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
  ('site.name', 'pveCloud', 'string', 'site', 0, '站点名称')
ON DUPLICATE KEY UPDATE
  `config_value` = VALUES(`config_value`),
  `value_type` = VALUES(`value_type`),
  `group_name` = VALUES(`group_name`),
  `is_secret` = VALUES(`is_secret`),
  `description` = VALUES(`description`);

-- Default admin accounts are intentionally not inserted here.
-- Create the first admin through a setup command so no default password is stored in git.
