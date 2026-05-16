-- Log management center Phase 2-5 tables, menus, permissions and retention configs.

SET NAMES utf8mb4;

USE `pvecloud`;

CREATE TABLE IF NOT EXISTS `user_security_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `user_id` BIGINT UNSIGNED NULL COMMENT '用户 ID，失败登录等未知账号场景可为空',
  `username` VARCHAR(64) NULL COMMENT '事件发生时用户名快照或登录标识',
  `email` VARCHAR(191) NULL COMMENT '事件发生时邮箱快照',
  `session_id` VARCHAR(128) NULL COMMENT '用户端会话 ID',
  `request_id` VARCHAR(64) NULL COMMENT '请求链路 ID',
  `request_method` VARCHAR(16) NULL COMMENT '请求方法',
  `request_path` VARCHAR(255) NULL COMMENT '请求路径',
  `action` VARCHAR(96) NOT NULL COMMENT '安全动作',
  `result` VARCHAR(32) NOT NULL DEFAULT 'success' COMMENT '结果：success/failed/limited',
  `ip` VARCHAR(64) NULL COMMENT '客户端 IP',
  `user_agent` VARCHAR(500) NULL COMMENT 'User-Agent 摘要',
  `remark` VARCHAR(500) NULL COMMENT '补充说明',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_security_logs_user_time` (`user_id`, `created_at`),
  KEY `idx_user_security_logs_action_time` (`action`, `created_at`),
  KEY `idx_user_security_logs_request_id` (`request_id`),
  KEY `idx_user_security_logs_ip_time` (`ip`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户安全日志';

CREATE TABLE IF NOT EXISTS `user_business_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
  `username` VARCHAR(64) NULL COMMENT '事件发生时用户名快照',
  `email` VARCHAR(191) NULL COMMENT '事件发生时邮箱快照',
  `request_id` VARCHAR(64) NULL COMMENT '请求链路 ID',
  `request_method` VARCHAR(16) NULL COMMENT '请求方法',
  `request_path` VARCHAR(255) NULL COMMENT '请求路径',
  `module` VARCHAR(64) NOT NULL COMMENT '业务模块',
  `action` VARCHAR(96) NOT NULL COMMENT '业务动作',
  `object_type` VARCHAR(64) NOT NULL COMMENT '业务对象类型',
  `object_id` VARCHAR(128) NULL COMMENT '业务对象标识',
  `summary` VARCHAR(500) NULL COMMENT '脱敏摘要',
  `ip` VARCHAR(64) NULL COMMENT '客户端 IP',
  `user_agent` VARCHAR(500) NULL COMMENT 'User-Agent 摘要',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_business_logs_user_time` (`user_id`, `created_at`),
  KEY `idx_user_business_logs_module_action_time` (`module`, `action`, `created_at`),
  KEY `idx_user_business_logs_object` (`object_type`, `object_id`),
  KEY `idx_user_business_logs_request_id` (`request_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户业务日志';

CREATE TABLE IF NOT EXISTS `frontend_error_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `source_app` VARCHAR(16) NOT NULL COMMENT '来源应用：admin/web',
  `user_id` BIGINT UNSIGNED NULL COMMENT 'Web 用户 ID',
  `admin_id` BIGINT UNSIGNED NULL COMMENT '管理员 ID',
  `request_id` VARCHAR(64) NULL COMMENT '请求链路 ID',
  `page_path` VARCHAR(255) NOT NULL COMMENT '页面路径',
  `error_type` VARCHAR(64) NOT NULL COMMENT '错误类型',
  `message` VARCHAR(500) NOT NULL COMMENT '错误消息摘要',
  `stack` TEXT NULL COMMENT '脱敏堆栈摘要',
  `api_path` VARCHAR(255) NULL COMMENT '关联 API 路径',
  `http_status` INT NULL COMMENT 'HTTP 状态码',
  `business_code` INT NULL COMMENT '业务错误码',
  `browser` VARCHAR(255) NULL COMMENT '浏览器摘要',
  `os` VARCHAR(255) NULL COMMENT '系统摘要',
  `app_version` VARCHAR(64) NULL COMMENT '前端应用版本',
  `ip` VARCHAR(64) NULL COMMENT '客户端 IP',
  `user_agent` VARCHAR(500) NULL COMMENT 'User-Agent 摘要',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_frontend_error_logs_source_time` (`source_app`, `created_at`),
  KEY `idx_frontend_error_logs_type_time` (`error_type`, `created_at`),
  KEY `idx_frontend_error_logs_request_id` (`request_id`),
  KEY `idx_frontend_error_logs_api_path` (`api_path`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='前端错误日志';

CREATE TABLE IF NOT EXISTS `backend_runtime_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `level` VARCHAR(16) NOT NULL COMMENT '日志级别',
  `category` VARCHAR(32) NOT NULL COMMENT '日志分类：access/panic/runtime',
  `request_id` VARCHAR(64) NULL COMMENT '请求链路 ID',
  `request_method` VARCHAR(16) NULL COMMENT '请求方法',
  `request_path` VARCHAR(255) NULL COMMENT '请求路径',
  `status` INT NULL COMMENT 'HTTP 状态码',
  `latency_ms` BIGINT NULL COMMENT '请求耗时毫秒',
  `client_ip` VARCHAR(64) NULL COMMENT '客户端 IP',
  `message` VARCHAR(500) NOT NULL COMMENT '日志消息摘要',
  `detail` TEXT NULL COMMENT '脱敏详情',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_backend_runtime_logs_category_time` (`category`, `created_at`),
  KEY `idx_backend_runtime_logs_level_time` (`level`, `created_at`),
  KEY `idx_backend_runtime_logs_request_id` (`request_id`),
  KEY `idx_backend_runtime_logs_status_time` (`status`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后端运行日志';

CREATE TABLE IF NOT EXISTS `log_export_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `admin_id` BIGINT UNSIGNED NOT NULL COMMENT '导出管理员 ID',
  `log_type` VARCHAR(64) NOT NULL COMMENT '导出日志类型',
  `filters` JSON NULL COMMENT '导出筛选条件',
  `row_count` INT NOT NULL DEFAULT 0 COMMENT '导出行数',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_log_export_records_admin_time` (`admin_id`, `created_at`),
  KEY `idx_log_export_records_type_time` (`log_type`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='日志导出记录';

INSERT INTO `admin_permissions` (`code`, `name`, `type`, `parent_code`, `path`, `icon`, `sort_order`, `visible_in_menu`, `group_name`, `description`)
VALUES
  ('page.logs.user-security', '用户安全日志', 'menu', 'page.logs', '/logs/user-security', NULL, 30, 1, '菜单', '显示用户安全日志页面'),
  ('page.logs.user-business', '用户业务日志', 'menu', 'page.logs', '/logs/user-business', NULL, 40, 1, '菜单', '显示用户业务日志页面'),
  ('page.logs.frontend-errors', '前端错误日志', 'menu', 'page.logs', '/logs/frontend-errors', NULL, 50, 1, '菜单', '显示前端错误日志页面'),
  ('page.logs.backend-runtime', '后端运行日志', 'menu', 'page.logs', '/logs/backend-runtime', NULL, 60, 1, '菜单', '显示后端运行日志页面'),
  ('user-security-log:*', '用户安全日志全权限', 'action', 'page.logs.user-security', NULL, NULL, 100, 0, '用户安全日志', '用户安全日志模块全部能力'),
  ('user-security-log:view', '查看用户安全日志', 'action', 'page.logs.user-security', NULL, NULL, 120, 0, '用户安全日志', '查看用户安全日志列表'),
  ('user-business-log:*', '用户业务日志全权限', 'action', 'page.logs.user-business', NULL, NULL, 100, 0, '用户业务日志', '用户业务日志模块全部能力'),
  ('user-business-log:view', '查看用户业务日志', 'action', 'page.logs.user-business', NULL, NULL, 120, 0, '用户业务日志', '查看用户业务日志列表'),
  ('frontend-error-log:*', '前端错误日志全权限', 'action', 'page.logs.frontend-errors', NULL, NULL, 100, 0, '前端错误日志', '前端错误日志模块全部能力'),
  ('frontend-error-log:view', '查看前端错误日志', 'action', 'page.logs.frontend-errors', NULL, NULL, 120, 0, '前端错误日志', '查看前端错误日志列表'),
  ('backend-runtime-log:*', '后端运行日志全权限', 'action', 'page.logs.backend-runtime', NULL, NULL, 100, 0, '后端运行日志', '后端运行日志模块全部能力'),
  ('backend-runtime-log:view', '查看后端运行日志', 'action', 'page.logs.backend-runtime', NULL, NULL, 120, 0, '后端运行日志', '查看后端运行日志列表'),
  ('log:export', '导出日志', 'action', 'page.logs', NULL, NULL, 200, 0, '日志管理', '导出日志列表'),
  ('log:cleanup', '清理日志', 'action', 'page.logs', NULL, NULL, 220, 0, '日志管理', '按留存策略清理日志')
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

INSERT INTO `system_configs` (`config_key`, `config_value`, `value_type`, `group_name`, `is_secret`, `description`)
VALUES
  ('logs.retention.user_security_days', '180', 'int', '日志管理', 0, '用户安全日志保留天数'),
  ('logs.retention.user_business_days', '365', 'int', '日志管理', 0, '用户业务日志保留天数'),
  ('logs.retention.frontend_error_days', '90', 'int', '日志管理', 0, '前端错误日志保留天数'),
  ('logs.retention.backend_runtime_days', '30', 'int', '日志管理', 0, '后端运行日志保留天数')
ON DUPLICATE KEY UPDATE
  `value_type` = VALUES(`value_type`),
  `group_name` = VALUES(`group_name`),
  `is_secret` = VALUES(`is_secret`),
  `description` = VALUES(`description`);

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
